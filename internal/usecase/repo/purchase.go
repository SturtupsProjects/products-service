package repo

import (
	"crm-admin/internal/entity"
	pb "crm-admin/internal/generated/products"
	"crm-admin/internal/usecase"
	"database/sql"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"strings"
	"time"
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
		INSERT INTO purchases (supplier_id, purchased_by, total_cost, payment_method, description, company_id, branch_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id, created_at
	`
	err = tx.QueryRowx(query, in.SupplierID, in.PurchasedBy, in.TotalCost, in.PaymentMethod, in.Description, in.CompanyID, in.BranchID).
		Scan(&purchase.Id, &purchase.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to create purchase: %w", err)
	}

	itemQuery := `
		INSERT INTO purchase_items (purchase_id, product_id, quantity, purchase_price, total_price, company_id, branch_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	for _, item := range in.PurchaseItems {
		if _, err := tx.Exec(itemQuery, purchase.Id, item.ProductID, item.Quantity, item.PurchasePrice, item.TotalPrice, in.CompanyID, in.BranchID); err != nil {
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
	if in.Id == "" || in.CompanyId == "" || in.BranchId == "" {
		return nil, errors.New("missing required fields")
	}

	params := make(map[string]interface{})
	params["id"] = in.Id
	params["company_id"] = in.CompanyId
	params["branch_id"] = in.BranchId

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
		WHERE id = :id AND company_id = :company_id AND branch_id = :branch_id
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
	if in.Id == "" || in.CompanyId == "" || in.BranchId == "" {
		return nil, errors.New("missing required fields")
	}

	query := `
        SELECT p.id, p.supplier_id, p.purchased_by, p.total_cost, p.payment_method, p.description, p.created_at,
               i.id AS item_id, i.product_id, i.quantity, i.purchase_price, i.total_price, pd.name, pd.image_url
        FROM purchases p
        LEFT JOIN purchase_items i ON p.id = i.purchase_id
        LEFT JOIN products pd ON i.product_id = pd.id
        WHERE p.id = $1 AND p.company_id = $2 AND p.branch_id = $3
    `

	purchase := &pb.PurchaseResponse{}
	var items []*pb.PurchaseItemResponse

	rows, err := r.db.Queryx(query, in.Id, in.CompanyId, in.BranchId)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve purchase: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var item pb.PurchaseItemResponse
		var itemID sql.NullString
		var productID sql.NullString
		var productName sql.NullString
		var productImage sql.NullString
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
			&productName,
			&productImage,
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

		if productName.Valid {
			item.ProductName = productName.String
		}
		if productImage.Valid {
			item.ProductImage = productImage.String
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

func (r *purchasesRepoImpl) GetPurchaseList(in *pb.FilterPurchase) (*pb.PurchaseList, error) {
	var purchases []*pb.PurchaseResponse
	var args []interface{}
	argIndex := 3

	// Основные фильтры для закупок
	filters := []string{"p.company_id = $1", "p.branch_id = $2"}
	args = append(args, in.CompanyId, in.BranchId)

	// Фильтры
	if in.SupplierId != "" {
		filters = append(filters, fmt.Sprintf("p.supplier_id ILIKE '%%' || $%d || '%%'", argIndex))
		args = append(args, in.SupplierId)
		argIndex++
	}
	if in.Description != "" {
		filters = append(filters, fmt.Sprintf("p.description ILIKE '%%' || $%d || '%%'", argIndex))
		args = append(args, in.Description)
		argIndex++
	}

	// Подсчёт общего количества записей закупок
	countQuery := fmt.Sprintf(`
		SELECT COUNT(*)
		FROM purchases p
		WHERE %s`, strings.Join(filters, " AND "))

	var totalCount int64
	if err := r.db.Get(&totalCount, countQuery, args...); err != nil {
		return nil, fmt.Errorf("failed to get total count: %w", err)
	}

	// Запрос для получения списка закупок
	mainQuery := fmt.Sprintf(`
		SELECT 
			p.id, p.supplier_id, p.purchased_by, p.total_cost, p.payment_method, p.description, p.branch_id, p.company_id, p.created_at
		FROM purchases p
		WHERE %s
		ORDER BY p.created_at DESC`, strings.Join(filters, " AND "))

	// Пагинация
	if in.Limit > 0 && in.Page > 0 {
		mainQuery += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argIndex, argIndex+1)
		args = append(args, in.Limit, (in.Page-1)*in.Limit)
	}

	// Выполнение основного запроса
	rows, err := r.db.Queryx(mainQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch purchases: %w", err)
	}
	defer rows.Close()

	// Карта для хранения закупок
	purchaseMap := make(map[string]*pb.PurchaseResponse)

	for rows.Next() {
		var purchase pb.PurchaseResponse
		err = rows.Scan(
			&purchase.Id,
			&purchase.SupplierId,
			&purchase.PurchasedBy,
			&purchase.TotalCost,
			&purchase.PaymentMethod,
			&purchase.Description,
			&purchase.BranchId,
			&purchase.CompanyId,
			&purchase.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan purchase row: %w", err)
		}

		purchaseMap[purchase.Id] = &purchase
	}

	// Получение связанных товаров для каждой закупки
	for purchaseID, purchase := range purchaseMap {
		itemsQuery := `
			SELECT 
				i.id, i.product_id, i.quantity, i.purchase_price, i.total_price,
				pr.name AS product_name, pr.image_url
			FROM purchase_items i
			LEFT JOIN products pr ON i.product_id = pr.id
			WHERE i.purchase_id = $1`

		itemRows, err := r.db.Queryx(itemsQuery, purchaseID)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch purchase items for purchase_id %s: %w", purchaseID, err)
		}

		var items []*pb.PurchaseItemResponse
		for itemRows.Next() {
			var item pb.PurchaseItemResponse
			var productName sql.NullString
			var productImage sql.NullString

			err = itemRows.Scan(
				&item.Id,
				&item.ProductId,
				&item.Quantity,
				&item.PurchasePrice,
				&item.TotalPrice,
				&productName,
				&productImage,
			)
			if err != nil {
				return nil, fmt.Errorf("failed to scan purchase item: %w", err)
			}

			if productName.Valid {
				item.ProductName = productName.String
			}
			if productImage.Valid {
				item.ProductImage = productImage.String
			}

			items = append(items, &item)
		}
		itemRows.Close()

		// Добавляем товары к закупке
		purchase.Items = items
	}

	// Формируем список закупок
	for _, purchase := range purchaseMap {
		purchases = append(purchases, purchase)
	}

	return &pb.PurchaseList{
		Purchases:  purchases,
		TotalCount: totalCount,
	}, nil
}

// DeletePurchase удаляет закупку и связанные товары
func (r *purchasesRepoImpl) DeletePurchase(in *pb.PurchaseID) (*pb.Message, error) {
	if in.Id == "" || in.CompanyId == "" || in.BranchId == "" {
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

	_, err = tx.Exec(`DELETE FROM purchase_items WHERE purchase_id = $1 AND company_id = $2 AND branch_id = $3`, in.Id, in.CompanyId, in.BranchId)
	if err != nil {
		return nil, fmt.Errorf("failed to delete purchase items: %w", err)
	}

	result, err := tx.Exec(`DELETE FROM purchases WHERE id = $1 AND company_id = $2 AND branch_id = $3`, in.Id, in.CompanyId, in.BranchId)
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

// --------------------------------------------- Transfer Structs ----------------------------------------------------

func (r purchasesRepoImpl) CreateTransfers(in *pb.TransferReq) (*pb.Transfer, error) {
	transferID := uuid.New().String()

	tx, err := r.db.Beginx()
	if err != nil {
		return nil, err
	}

	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p)
		} else if err != nil {
			tx.Rollback()
		} else {
			tx.Commit()
		}
	}()

	_, err = tx.Exec(`
		INSERT INTO transfers (id, transferred_by, from_branch_id, to_branch_id, description, company_id, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, NOW())`,
		transferID, in.TransferredBy, in.FromBranchId, in.ToBranchId, in.Description, in.CompanyId,
	)
	if err != nil {
		return nil, err
	}

	for _, product := range in.Products {
		_, err = tx.Exec(`
			INSERT INTO transfer_products (id, product_transfers_id, product_id, quantity)
			VALUES ($1, $2, $3, $4)`,
			uuid.New().String(), transferID, product.ProductId, product.ProductQuantity,
		)
		if err != nil {
			return nil, err
		}
	}

	return &pb.Transfer{
		Id:            transferID,
		TransferredBy: in.TransferredBy,
		FromBranchId:  in.FromBranchId,
		ToBranchId:    in.ToBranchId,
		Description:   in.Description,
		CompanyId:     in.CompanyId,
		CreatedAt:     time.Now().String(),
	}, nil
}

func (r purchasesRepoImpl) GetTransfers(in *pb.TransferID) (*pb.Transfer, error) {
	var transfer pb.Transfer
	var products []struct {
		ID              string `db:"id"`
		ProductID       string `db:"product_id"`
		ProductQuantity int64  `db:"product_quantity"`
		ProductName     string `db:"product_name"`
		ProductImage    string `db:"product_image"`
	}

	err := r.db.Get(&transfer, `
		SELECT id, transferred_by, from_branch_id, to_branch_id, description, created_at, company_id
		FROM transfers
		WHERE id = $1`, in.Id,
	)
	if err != nil {
		return nil, errors.New("transfer not found")
	}

	err = r.db.Select(&products, `
		SELECT tp.id, tp.product_id, tp.quantity AS product_quantity, p.name AS product_name, p.image_url AS product_image
		FROM transfer_products tp
		JOIN products p ON tp.product_id = p.id
		WHERE tp.product_transfers_id = $1`,
		in.Id,
	)
	if err != nil {
		return nil, err
	}

	transferProducts := make([]*pb.TransfersProducts, len(products))
	for i, product := range products {
		transferProducts[i] = &pb.TransfersProducts{
			Id:              product.ID,
			ProductId:       product.ProductID,
			ProductQuantity: product.ProductQuantity,
			ProductName:     product.ProductName,
			ProductImage:    product.ProductImage,
		}
	}

	transfer.Products = transferProducts
	return &transfer, nil
}

func (r purchasesRepoImpl) GetTransferList(in *pb.TransferFilter) (*pb.TransferList, error) {
	filters := []string{}
	args := []interface{}{}

	if in.BranchId != "" {
		filters = append(filters, `(from_branch_id = $1 OR to_branch_id = $1)`)
		args = append(args, in.BranchId)
	}

	productFilter := strings.TrimSpace(in.ProductName)
	if productFilter != "" {
		filters = append(filters, `p.name ILIKE '%' || $2 || '%'`)
		args = append(args, productFilter)
	}

	whereClause := ""
	if len(filters) > 0 {
		whereClause = "WHERE " + strings.Join(filters, " AND ")
	}

	var totalCount int
	err := r.db.Get(&totalCount, fmt.Sprintf(`
		SELECT COUNT(*)
		FROM transfers t
		JOIN transfer_products tp ON t.id = tp.product_transfers_id
		JOIN products p ON tp.product_id = p.id
		%s`, whereClause), args...)
	if err != nil {
		return nil, err
	}

	if totalCount == 0 {
		return &pb.TransferList{TotalCount: 0, Transfers: []*pb.Transfer{}}, nil
	}

	query := fmt.Sprintf(`
		SELECT t.id, t.transferred_by, t.from_branch_id, t.to_branch_id, t.description, t.created_at, t.company_id
		FROM transfers t
		LEFT JOIN transfer_products tp ON t.id = tp.product_transfers_id
		LEFT JOIN products p ON tp.product_id = p.id
		%s
		ORDER BY t.created_at DESC`, whereClause)

	if in.Limit > 0 && in.Page > 0 {
		query += " LIMIT $3 OFFSET $4"
		args = append(args, in.Limit, (in.Page-1)*in.Limit)
	}

	rows, err := r.db.Queryx(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	transfers := []*pb.Transfer{}
	for rows.Next() {
		var transfer pb.Transfer
		if err := rows.StructScan(&transfer); err != nil {
			return nil, err
		}
		transfers = append(transfers, &transfer)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return &pb.TransferList{
		Transfers:  transfers,
		TotalCount: int64(totalCount),
	}, nil
}
