package repo

import (
	"crm-admin/internal/entity"
	"crm-admin/internal/usecase"
	pb "crm-admin/pkg/generated/products"
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

	query := `INSERT INTO purchases (supplier_id, purchased_by, total_cost, payment_method, description)
	          VALUES (:supplier_id, :purchased_by, :total_cost, :payment_method, :description) RETURNING id, created_at`
	err := r.db.QueryRowx(query, in).StructScan(purchase)
	if err != nil {
		return nil, err
	}

	for _, item := range *in.PurchaseItem {
		itemQuery := `INSERT INTO purchase_items (purchase_id, product_id, quantity, purchase_price, total_price)
		              VALUES ($1, $2, $3, $4, $5)`
		_, err := r.db.Exec(itemQuery, purchase.Id, item.ProductID, item.Quantity, item.PurchasePrice, item.TotalPrice)
		if err != nil {
			return nil, err
		}
	}

	return purchase, nil
}

func (r *purchasesRepoImpl) UpdatePurchase(in *entity.PurchaseUpdate) (*pb.PurchaseResponse, error) {
	// Изначальный запрос с условиями
	query := `UPDATE purchases SET `
	updates := []string{}
	params := map[string]interface{}{"id": in.ID}

	// Добавляем поля в зависимости от их наличия
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

	// Если нет полей для обновления, возвращаем ошибку
	if len(updates) == 0 {
		return nil, errors.New("no fields to update")
	}

	// Объединяем части обновляемых полей
	query += strings.Join(updates, ", ")
	query += " WHERE id = :id RETURNING id, supplier_id, purchased_by, total_cost, description, payment_method, created_at"

	// Выполняем запрос
	purchase := &pb.PurchaseResponse{}
	err := r.db.QueryRowx(query, params).StructScan(purchase)
	if err != nil {
		return nil, err
	}

	return purchase, nil
}

// GetPurchase возвращает закупку по ID
func (r *purchasesRepoImpl) GetPurchase(in *entity.PurchaseID) (*pb.PurchaseResponse, error) {
	query := `SELECT id, supplier_id, purchased_by, total_cost, payment_method, description, created_at
	          FROM purchases WHERE id = $1`
	purchase := &pb.PurchaseResponse{}
	err := r.db.Get(purchase, query, in.ID)
	if err != nil {
		return nil, err
	}

	itemsQuery := `SELECT id, purchase_id, product_id, quantity, purchase_price, total_price
	               FROM purchase_items WHERE purchase_id = $1`
	err = r.db.Select(&purchase.Items, itemsQuery, in.ID)
	if err != nil {
		return nil, err
	}

	return purchase, nil
}

func (r *purchasesRepoImpl) GetPurchaseList(in *entity.FilterPurchase) (*pb.PurchaseList, error) {
	var purchases []*pb.PurchaseResponse
	var queryBuilder strings.Builder
	var args []interface{}
	argIndex := 1

	// Базовый запрос
	queryBuilder.WriteString(`
		SELECT p.id, p.supplier_id, p.purchased_by, p.total_cost, p.description, 
		       p.payment_method, p.created_at 
		FROM purchases p JOIN purchase_items i ON p.id = i.purchase_id
		WHERE 1=1
	`)

	// Фильтр по SupplierID с использованием ILIKE
	if in.SupplierID != "" {
		queryBuilder.WriteString(" AND p.supplier_id ILIKE '%' || $" + fmt.Sprint(argIndex) + " || '%'")
		args = append(args, in.SupplierID)
		argIndex++
	}

	// Фильтр по PurchasedBy с использованием ILIKE
	if in.PurchasedBy != "" {
		queryBuilder.WriteString(" AND p.purchased_by ILIKE '%' || $" + fmt.Sprint(argIndex) + " || '%'")
		args = append(args, in.PurchasedBy)
		argIndex++
	}

	// Фильтр по CreatedAt
	if in.CreatedAt != "" {
		queryBuilder.WriteString(" AND DATE(p.created_at) = DATE($" + fmt.Sprint(argIndex) + ")")
		args = append(args, in.CreatedAt)
		argIndex++
	}

	// Сортировка по дате создания
	queryBuilder.WriteString(" ORDER BY p.created_at DESC")

	// Выполнение запроса
	query := queryBuilder.String()
	err := r.db.Select(&purchases, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list purchases: %w", err)
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
