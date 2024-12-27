package repo

import (
	"crm-admin/internal/entity"
	pb "crm-admin/internal/generated/products"
	"crm-admin/internal/usecase"
	"database/sql"
	"errors"
	"fmt"
	"github.com/jmoiron/sqlx"
	"strings"
)

type purchasesRepoImpl struct {
	db *sqlx.DB
}

func NewPurchasesRepo(db *sqlx.DB) usecase.PurchasesRepo {
	return &purchasesRepoImpl{db: db}
}

// CreatePurchase создает новую закупку с товарами
func (r *purchasesRepoImpl) CreatePurchase(in *entity.PurchaseRequest) (*pb.PurchaseResponse, error) {
	tx, err := r.db.Beginx()
	if err != nil {
		return nil, fmt.Errorf("failed to start transaction: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	purchase := &pb.PurchaseResponse{}
	query := `
		INSERT INTO purchases (supplier_id, purchased_by, total_cost, payment_method, description, company_id)
		VALUES ($1, $2, $3, $4, $5, $6) RETURNING id, created_at
	`
	err = tx.QueryRowx(query, in.SupplierID, in.PurchasedBy, in.TotalCost, in.PaymentMethod, in.Description, in.CompanyID).
		Scan(&purchase.Id, &purchase.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to create purchase: %w", err)
	}

	itemQuery := `
		INSERT INTO purchase_items (purchase_id, product_id, quantity, purchase_price, total_price, company_id)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	for _, item := range in.PurchaseItems {
		if _, err := tx.Exec(itemQuery, purchase.Id, item.ProductID, item.Quantity, item.PurchasePrice, item.TotalPrice, in.CompanyID); err != nil {
			return nil, fmt.Errorf("failed to add purchase item: %w", err)
		}
	}

	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	// Заполнение дополнительных данных
	purchase.SupplierId = in.SupplierID
	purchase.PurchasedBy = in.PurchasedBy
	purchase.TotalCost = in.TotalCost
	purchase.PaymentMethod = in.PaymentMethod
	purchase.Description = in.Description

	return purchase, nil
}

// UpdatePurchase обновляет информацию о закупке
func (r *purchasesRepoImpl) UpdatePurchase(in *pb.PurchaseUpdate) (*pb.PurchaseResponse, error) {
	if in.Id == "" || in.CompanyId == "" {
		return nil, errors.New("missing required fields")
	}

	params := make(map[string]interface{})
	params["id"] = in.Id
	params["company_id"] = in.CompanyId

	var updates []string
	if in.SupplierId != "" {
		updates = append(updates, "supplier_id = :supplier_id")
		params["supplier_id"] = in.SupplierId
	}
	if in.Description != "" {
		updates = append(updates, "description = :description")
		params["description"] = in.Description
	}
	if in.PaymentMethod != "" {
		updates = append(updates, "payment_method = :payment_method")
		params["payment_method"] = in.PaymentMethod
	}

	if len(updates) == 0 {
		return nil, errors.New("no fields to update")
	}

	query := `
		UPDATE purchases SET ` + strings.Join(updates, ", ") + `
		WHERE id = :id AND company_id = :company_id
		RETURNING id, supplier_id, purchased_by, total_cost, description, payment_method, created_at
	`
	stmt, err := r.db.PrepareNamed(query)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare query: %w", err)
	}
	defer stmt.Close()

	purchase := &pb.PurchaseResponse{}
	err = stmt.QueryRowx(params).Scan(&purchase.Id, &purchase.SupplierId, &purchase.PurchasedBy, &purchase.TotalCost, &purchase.Description, &purchase.PaymentMethod, &purchase.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to update purchase: %w", err)
	}

	return purchase, nil
}

// GetPurchase возвращает информацию о закупке с деталями
func (r *purchasesRepoImpl) GetPurchase(in *pb.PurchaseID) (*pb.PurchaseResponse, error) {
	if in.Id == "" || in.CompanyId == "" {
		return nil, errors.New("missing required fields")
	}

	query := `
        SELECT p.id, p.supplier_id, p.purchased_by, p.total_cost, p.payment_method, p.description, p.created_at,
               i.id AS item_id, i.product_id, i.quantity, i.purchase_price, i.total_price
        FROM purchases p
        LEFT JOIN purchase_items i ON p.id = i.purchase_id
        WHERE p.id = $1 AND p.company_id = $2
    `

	purchase := &pb.PurchaseResponse{}
	var items []*pb.PurchaseItemResponse

	rows, err := r.db.Queryx(query, in.Id, in.CompanyId)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve purchase: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var item pb.PurchaseItemResponse
		var itemID sql.NullString
		var productID sql.NullString
		var quantity sql.NullInt32
		var purchasePrice sql.NullFloat64
		var totalPrice sql.NullFloat64

		err = rows.Scan(
			&purchase.Id,
			&purchase.SupplierId,
			&purchase.PurchasedBy,
			&purchase.TotalCost,
			&purchase.PaymentMethod,
			&purchase.Description,
			&purchase.CreatedAt,
			&itemID,
			&productID,
			&quantity,
			&purchasePrice,
			&totalPrice,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan purchase data: %w", err)
		}
		if itemID.Valid {
			item.Id = itemID.String
		}
		if productID.Valid {
			item.ProductId = productID.String
		}
		if quantity.Valid {
			item.Quantity = quantity.Int32
		}
		if purchasePrice.Valid {
			item.PurchasePrice = purchasePrice.Float64
		}
		if totalPrice.Valid {
			item.TotalPrice = totalPrice.Float64
		}

		if itemID.Valid {
			items = append(items, &item)
		}
	}

	purchase.Items = items

	return purchase, nil
}

// GetPurchaseList возвращает список закупок
func (r *purchasesRepoImpl) GetPurchaseList(in *pb.FilterPurchase) (*pb.PurchaseList, error) {
	var purchases []*pb.PurchaseResponse
	var queryBuilder strings.Builder
	var args []interface{}
	argIndex := 2

	queryBuilder.WriteString(`
        SELECT p.id, p.supplier_id, p.purchased_by, p.total_cost, p.payment_method, p.description, p.created_at,
               i.id AS item_id, i.product_id, i.quantity, i.purchase_price, i.total_price
        FROM purchases p
        LEFT JOIN purchase_items i ON p.id = i.purchase_id
        WHERE p.company_id = $1
    `)
	args = append(args, in.CompanyId)

	if in.SupplierId != "" {
		queryBuilder.WriteString(" AND p.supplier_id ILIKE '%' || $" + fmt.Sprint(argIndex) + " || '%'")
		args = append(args, in.SupplierId)
		argIndex++
	}

	queryBuilder.WriteString(" ORDER BY p.created_at DESC")
	query := queryBuilder.String()

	rows, err := r.db.Queryx(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list purchases: %w", err)
	}
	defer rows.Close()

	purchaseMap := make(map[string]*pb.PurchaseResponse)

	for rows.Next() {
		var purchase pb.PurchaseResponse
		var item pb.PurchaseItemResponse
		var itemID sql.NullString
		var productID sql.NullString
		var quantity sql.NullInt32
		var purchasePrice sql.NullFloat64
		var totalPrice sql.NullFloat64

		err = rows.Scan(
			&purchase.Id,
			&purchase.SupplierId,
			&purchase.PurchasedBy,
			&purchase.TotalCost,
			&purchase.PaymentMethod,
			&purchase.Description,
			&purchase.CreatedAt,
			&itemID,
			&productID,
			&quantity,
			&purchasePrice,
			&totalPrice,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan purchase list: %w", err)
		}

		if itemID.Valid {
			item.Id = itemID.String
		}
		if productID.Valid {
			item.ProductId = productID.String
		}
		if quantity.Valid {
			item.Quantity = quantity.Int32
		}
		if purchasePrice.Valid {
			item.PurchasePrice = purchasePrice.Float64
		}
		if totalPrice.Valid {
			item.TotalPrice = totalPrice.Float64
		}

		if _, exists := purchaseMap[purchase.Id]; !exists {
			purchaseMap[purchase.Id] = &purchase
		}

		if itemID.Valid {
			purchaseMap[purchase.Id].Items = append(purchaseMap[purchase.Id].Items, &item)
		}
	}

	for _, purchase := range purchaseMap {
		purchases = append(purchases, purchase)
	}

	return &pb.PurchaseList{Purchases: purchases}, nil
}

// DeletePurchase удаляет закупку и связанные товары
func (r *purchasesRepoImpl) DeletePurchase(in *pb.PurchaseID) (*pb.Message, error) {
	if in.Id == "" || in.CompanyId == "" {
		return nil, errors.New("missing required fields")
	}

	tx, err := r.db.Beginx()
	if err != nil {
		return nil, fmt.Errorf("failed to start transaction: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	_, err = tx.Exec(`DELETE FROM purchase_items WHERE purchase_id = $1 AND company_id = $2`, in.Id, in.CompanyId)
	if err != nil {
		return nil, fmt.Errorf("failed to delete purchase items: %w", err)
	}

	result, err := tx.Exec(`DELETE FROM purchases WHERE id = $1 AND company_id = $2`, in.Id, in.CompanyId)
	if err != nil {
		return nil, fmt.Errorf("failed to delete purchase: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return nil, errors.New("purchase not found")
	}

	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return &pb.Message{Message: "Purchase deleted successfully"}, nil
}
