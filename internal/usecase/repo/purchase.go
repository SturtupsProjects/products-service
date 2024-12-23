package repo

import (
	"crm-admin/internal/entity"
	pb "crm-admin/internal/generated/products"
	"crm-admin/internal/usecase"
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

func (r *purchasesRepoImpl) CreatePurchase(in *entity.PurchaseRequest) (*pb.PurchaseResponse, error) {
	purchase := &pb.PurchaseResponse{}

	query := `INSERT INTO purchases (supplier_id, purchased_by, total_cost, payment_method, description, company_id)
	          VALUES ($1, $2, $3, $4, $5, $6) RETURNING id, created_at`
	err := r.db.QueryRowx(query, in.SupplierID, in.PurchasedBy, in.TotalCost, in.PaymentMethod, in.Description, in.CompanyID).
		Scan(&purchase.Id, &purchase.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to create purchase: %w", err)
	}

	for _, item := range in.PurchaseItems {
		itemQuery := `INSERT INTO purchase_items (purchase_id, product_id, quantity, purchase_price, total_price, company_id)
		              VALUES ($1, $2, $3, $4, $5, $6)`
		_, err := r.db.Exec(itemQuery, purchase.Id, item.ProductID, item.Quantity, item.PurchasePrice, item.TotalPrice, in.CompanyID)
		if err != nil {
			return nil, fmt.Errorf("failed to create purchase item: %w", err)
		}
	}

	purchase.SupplierId = in.SupplierID
	purchase.PurchasedBy = in.PurchasedBy
	purchase.TotalCost = in.TotalCost
	purchase.PaymentMethod = in.PaymentMethod
	purchase.Description = in.Description

	return purchase, nil
}

func (r *purchasesRepoImpl) UpdatePurchase(in *pb.PurchaseUpdate) (*pb.PurchaseResponse, error) {
	query := `UPDATE purchases SET `
	updates := []string{}
	params := map[string]interface{}{"id": in.Id, "company_id": in.CompanyId}

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

	query += strings.Join(updates, ", ")
	query += " WHERE id = :id AND company_id = :company_id RETURNING id, supplier_id, purchased_by, total_cost, description, payment_method, created_at"

	purchase := &pb.PurchaseResponse{}
	err := r.db.QueryRowx(query, params).Scan(&purchase.Id, &purchase.SupplierId, &purchase.PurchasedBy, &purchase.TotalCost, &purchase.Description, &purchase.PaymentMethod, &purchase.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to update purchase: %w", err)
	}

	return purchase, nil
}

func (r *purchasesRepoImpl) GetPurchase(in *pb.PurchaseID) (*pb.PurchaseResponse, error) {
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
		err = rows.Scan(
			&purchase.Id,
			&purchase.SupplierId,
			&purchase.PurchasedBy,
			&purchase.TotalCost,
			&purchase.PaymentMethod,
			&purchase.Description,
			&purchase.CreatedAt,
			&item.Id,
			&item.ProductId,
			&item.Quantity,
			&item.PurchasePrice,
			&item.TotalPrice,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan purchase data: %w", err)
		}
		if item.Id != "" {
			items = append(items, &item)
		}
	}

	purchase.Items = items

	return purchase, nil
}

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
		err = rows.Scan(
			&purchase.Id,
			&purchase.SupplierId,
			&purchase.PurchasedBy,
			&purchase.TotalCost,
			&purchase.PaymentMethod,
			&purchase.Description,
			&purchase.CreatedAt,
			&item.Id,
			&item.ProductId,
			&item.Quantity,
			&item.PurchasePrice,
			&item.TotalPrice,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan purchase list: %w", err)
		}

		if _, exists := purchaseMap[purchase.Id]; !exists {
			purchaseMap[purchase.Id] = &purchase
		}

		if item.Id != "" {
			purchaseMap[purchase.Id].Items = append(purchaseMap[purchase.Id].Items, &item)
		}
	}

	for _, purchase := range purchaseMap {
		purchases = append(purchases, purchase)
	}

	return &pb.PurchaseList{Purchases: purchases}, nil
}

func (r *purchasesRepoImpl) DeletePurchase(in *pb.PurchaseID) (*pb.Message, error) {
	_, err := r.db.Exec(`DELETE FROM purchase_items WHERE purchase_id = $1 AND company_id = $2`, in.Id, in.CompanyId)
	if err != nil {
		return nil, fmt.Errorf("failed to delete purchase items: %w", err)
	}

	result, err := r.db.Exec(`DELETE FROM purchases WHERE id = $1 AND company_id = $2`, in.Id, in.CompanyId)
	if err != nil {
		return nil, fmt.Errorf("failed to delete purchase: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return nil, errors.New("purchase not found")
	}

	return &pb.Message{Message: "Purchase deleted successfully"}, nil
}
