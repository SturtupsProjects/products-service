package repo

import (
	"crm-admin/internal/entity"
	"crm-admin/internal/usecase"
	"errors"
	"fmt"
	"github.com/jmoiron/sqlx"
	"strings"
)

type salesRepoImpl struct {
	db *sqlx.DB
}

func NewSalesRepo(db *sqlx.DB) usecase.SalesRepo {
	return &salesRepoImpl{db: db}
}

func (r *salesRepoImpl) CreateSale(in *entity.SalesTotal) (*entity.SaleResponse, error) {
	sale := &entity.SaleResponse{}

	query := `INSERT INTO sales (client_id, sold_by, total_sale_price, payment_method)
	          VALUES (:client_id, :sold_by, :total_sale_price, :payment_method) RETURNING id, created_at`
	err := r.db.QueryRowx(query, in).StructScan(sale)
	if err != nil {
		return nil, err
	}

	for _, item := range in.SoldProducts {
		item.SaleID = sale.ID
		itemQuery := `INSERT INTO sales_items (sale_id, product_id, quantity, sale_price, total_price)
		              VALUES (:sale_id, :product_id, :quantity, :sale_price, :total_price)`
		_, err := r.db.NamedExec(itemQuery, item)
		if err != nil {
			return nil, err
		}
	}

	return sale, nil
}

func (r *salesRepoImpl) UpdateSale(in *entity.SaleUpdate) (*entity.SaleResponse, error) {
	query := `UPDATE sales SET `
	updates := []string{}
	params := map[string]interface{}{"id": in.ID}

	if in.ClientID != "" {
		updates = append(updates, "client_id = :client_id")
		params["client_id"] = in.ClientID
	}
	if in.PaymentMethod != "" {
		updates = append(updates, "payment_method = :payment_method")
		params["payment_method"] = in.PaymentMethod
	}

	if len(updates) == 0 {
		return nil, errors.New("no fields to update")
	}

	query += strings.Join(updates, ", ")
	query += " WHERE id = :id RETURNING id, client_id, sold_by, total_sale_price, payment_method, created_at"

	sale := &entity.SaleResponse{}
	err := r.db.QueryRowx(query, params).StructScan(sale)
	if err != nil {
		return nil, err
	}

	return sale, nil
}

func (r *salesRepoImpl) GetSale(in *entity.SaleID) (*entity.SaleResponse, error) {
	query := `SELECT id, client_id, sold_by, total_sale_price, payment_method, created_at
	          FROM sales WHERE id = $1`
	sale := &entity.SaleResponse{}
	err := r.db.Get(sale, query, in.ID)
	if err != nil {
		return nil, err
	}

	itemsQuery := `SELECT id, sale_id, product_id, quantity, sale_price, total_price
	               FROM sales_items WHERE sale_id = $1`
	err = r.db.Select(&sale.SoldProducts, itemsQuery, in.ID)
	if err != nil {
		return nil, err
	}

	return sale, nil
}

func (r *salesRepoImpl) GetSaleList(in *entity.SaleFilter) (*entity.SaleList, error) {
	var sales []entity.SaleResponse
	var queryBuilder strings.Builder
	var args []interface{}
	argIndex := 1

	queryBuilder.WriteString(`
		SELECT s.id, s.client_id, s.sold_by, s.total_sale_price, 
		       s.payment_method, s.created_at 
		FROM sales s JOIN sales_items i ON s.id = i.sale_id
		WHERE 1=1
	`)

	if in.ClientID != "" {
		queryBuilder.WriteString(" AND s.client_id ILIKE '%' || $" + fmt.Sprint(argIndex) + " || '%'")
		args = append(args, in.ClientID)
		argIndex++
	}

	if in.SoldBy != "" {
		queryBuilder.WriteString(" AND s.sold_by ILIKE '%' || $" + fmt.Sprint(argIndex) + " || '%'")
		args = append(args, in.SoldBy)
		argIndex++
	}

	if in.StartDate != "" {
		queryBuilder.WriteString(" AND DATE(s.created_at) >= DATE($" + fmt.Sprint(argIndex) + ")")
		args = append(args, in.StartDate)
		argIndex++
	}

	if in.EndDate != "" {
		queryBuilder.WriteString(" AND DATE(s.created_at) <= DATE($" + fmt.Sprint(argIndex) + ")")
		args = append(args, in.EndDate)
		argIndex++
	}

	queryBuilder.WriteString(" ORDER BY s.created_at DESC")

	query := queryBuilder.String()
	err := r.db.Select(&sales, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list sales: %w", err)
	}

	return &entity.SaleList{Sales: sales}, nil
}

func (r *salesRepoImpl) DeleteSale(in *entity.SaleID) (*entity.Message, error) {
	_, err := r.db.Exec(`DELETE FROM sales_items WHERE sale_id = $1`, in.ID)
	if err != nil {
		return nil, err
	}
	result, err := r.db.Exec(`DELETE FROM sales WHERE id = $1`, in.ID)
	if err != nil {
		return nil, err
	}
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return nil, errors.New("sale not found")
	}

	return &entity.Message{Message: "Sale deleted successfully"}, nil
}
