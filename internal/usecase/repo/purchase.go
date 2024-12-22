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

	// Insert into purchases table and get the ID and created_at
	query := `INSERT INTO purchases (supplier_id, purchased_by, total_cost, payment_method, description)
	          VALUES ($1, $2, $3, $4, $5) RETURNING id, created_at`

	err := r.db.QueryRowx(query, in.SupplierID, in.PurchasedBy, in.TotalCost, in.PaymentMethod, in.Description).Scan(&purchase.Id, &purchase.CreatedAt)
	if err != nil {
		return nil, err
	}

	// Insert purchase items
	for _, item := range *in.PurchaseItem {
		itemQuery := `INSERT INTO purchase_items (purchase_id, product_id, quantity, purchase_price, total_price)
		              VALUES ($1, $2, $3, $4, $5)`
		_, err := r.db.Exec(itemQuery, purchase.Id, item.ProductID, item.Quantity, item.PurchasePrice, item.TotalPrice)
		if err != nil {
			return nil, err
		}
	}

	// Fill the response object
	purchase.SupplierId = in.SupplierID
	purchase.PurchasedBy = in.PurchasedBy
	purchase.TotalCost = in.TotalCost
	purchase.PaymentMethod = in.PaymentMethod
	purchase.Description = in.Description

	return purchase, nil
}

func (r *purchasesRepoImpl) UpdatePurchase(in *entity.PurchaseUpdate) (*pb.PurchaseResponse, error) {

	query := `UPDATE purchases SET `
	updates := []string{}
	params := map[string]interface{}{"id": in.ID}

	// Add fields if they are provided
	if in.SupplierID != "" {
		updates = append(updates, "supplier_id = :supplier_id")
		params["supplier_id"] = in.SupplierID
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
	query += " WHERE id = :id RETURNING id, supplier_id, purchased_by, total_cost, description, payment_method, created_at"

	purchase := &pb.PurchaseResponse{}
	err := r.db.QueryRowx(query, params).Scan(&purchase.Id, &purchase.SupplierId, &purchase.PurchasedBy, &purchase.TotalCost, &purchase.Description, &purchase.PaymentMethod, &purchase.CreatedAt)
	if err != nil {
		return nil, err
	}

	return purchase, nil
}

// GetPurchase returns the purchase by ID
func (r *purchasesRepoImpl) GetPurchase(in *entity.PurchaseID) (*pb.PurchaseResponse, error) {
	query := `
        SELECT 
            p.id, 
            p.supplier_id, 
            p.purchased_by, 
            p.total_cost, 
            p.payment_method, 
            p.description, 
            p.created_at, 
            i.id AS item_id, 
            i.product_id, 
            i.quantity, 
            i.purchase_price, 
            i.total_price
        FROM purchases p
        LEFT JOIN purchase_items i ON p.id = i.purchase_id
        WHERE p.id = $1
    `

	purchase := &pb.PurchaseResponse{}
	var items []*pb.PurchaseItemResponse

	rows, err := r.db.Queryx(query, in.ID)
	if err != nil {
		return nil, err
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
			return nil, err
		}

		// If item exists, append it to items list
		if item.Id != "" {
			items = append(items, &item)
		}
	}

	purchase.Items = items

	// Handle any errors that occurred while iterating over rows
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return purchase, nil
}

func (r *purchasesRepoImpl) GetPurchaseList(in *entity.FilterPurchase) (*pb.PurchaseList, error) {
	var purchases []*pb.PurchaseResponse
	var queryBuilder strings.Builder
	var args []interface{}
	argIndex := 1

	queryBuilder.WriteString(`
        SELECT p.id, p.supplier_id, p.purchased_by, p.total_cost, p.description, 
               p.payment_method, p.created_at, 
               i.id AS item_id, i.product_id, i.quantity, i.purchase_price, i.total_price
        FROM purchases p 
        LEFT JOIN purchase_items i ON p.id = i.purchase_id
        WHERE 1=1
    `)

	if in.SupplierID != "" {
		queryBuilder.WriteString(" AND p.supplier_id ILIKE '%' || $" + fmt.Sprint(argIndex) + " || '%'")
		args = append(args, in.SupplierID)
		argIndex++
	}

	if in.PurchasedBy != "" {
		queryBuilder.WriteString(" AND p.purchased_by ILIKE '%' || $" + fmt.Sprint(argIndex) + " || '%'")
		args = append(args, in.PurchasedBy)
		argIndex++
	}

	if in.CreatedAt != "" {
		queryBuilder.WriteString(" AND DATE(p.created_at) = DATE($" + fmt.Sprint(argIndex) + ")")
		args = append(args, in.CreatedAt)
		argIndex++
	}

	queryBuilder.WriteString(" ORDER BY p.created_at DESC")

	query := queryBuilder.String()

	rows, err := r.db.Queryx(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list purchases: %w", err)
	}
	defer rows.Close()

	purchasesMap := make(map[string]*pb.PurchaseResponse)

	// Iterate over the result set and populate the purchases
	for rows.Next() {
		var purchase pb.PurchaseResponse
		var item pb.PurchaseItemResponse
		err := rows.Scan(
			&purchase.Id,
			&purchase.SupplierId,
			&purchase.PurchasedBy,
			&purchase.TotalCost,
			&purchase.Description,
			&purchase.PaymentMethod,
			&purchase.CreatedAt,
			&item.Id,
			&item.ProductId,
			&item.Quantity,
			&item.PurchasePrice,
			&item.TotalPrice,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan purchase: %w", err)
		}

		// If purchase is not in the map, add it
		if _, exists := purchasesMap[purchase.Id]; !exists {
			purchasesMap[purchase.Id] = &purchase
		}

		// Append items to the purchase
		purchasesMap[purchase.Id].Items = append(purchasesMap[purchase.Id].Items, &item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over purchases: %w", err)
	}

	// Convert map to slice
	for _, purchase := range purchasesMap {
		purchases = append(purchases, purchase)
	}

	return &pb.PurchaseList{Purchases: purchases}, nil
}

func (r *purchasesRepoImpl) DeletePurchase(in *entity.PurchaseID) (*pb.Message, error) {
	_, err := r.db.Exec(`DELETE FROM purchase_items WHERE purchase_id = $1`, in.ID)
	if err != nil {
		return nil, err
	}

	result, err := r.db.Exec(`DELETE FROM purchases WHERE id = $1`, in.ID)
	if err != nil {
		return nil, err
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return nil, errors.New("purchase not found")
	}

	return &pb.Message{Message: "Purchase deleted successfully"}, nil
}
