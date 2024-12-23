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

type salesRepoImpl struct {
	db *sqlx.DB
}

func NewSalesRepo(db *sqlx.DB) usecase.SalesRepo {
	return &salesRepoImpl{db: db}
}

func (r *salesRepoImpl) CreateSale(in *entity.SalesTotal) (*pb.SaleResponse, error) {
	sale := &pb.SaleResponse{}

	query := `INSERT INTO sales (company_id, client_id, sold_by, total_sale_price, payment_method)
	          VALUES ($1, $2, $3, $4, $5) RETURNING id, created_at`

	err := r.db.QueryRowx(query, in.CompanyID, in.ClientID, in.SoldBy, in.TotalSalePrice, in.PaymentMethod).Scan(&sale.Id, &sale.CreatedAt)
	if err != nil {
		return nil, err
	}

	for _, item := range in.SoldProducts {
		item.SaleID = sale.Id
		itemQuery := `INSERT INTO sales_items (company_id, sale_id, product_id, quantity, sale_price, total_price)
		              VALUES ($1, $2, $3, $4, $5, $6)`
		_, err := r.db.Exec(itemQuery, in.CompanyID, item.SaleID, item.ProductID, item.Quantity, item.SalePrice, item.TotalPrice)
		if err != nil {
			return nil, err
		}
	}

	sale.TotalSalePrice = in.TotalSalePrice
	sale.ClientId = in.ClientID
	sale.SoldBy = in.SoldBy
	sale.PaymentMethod = in.PaymentMethod

	return sale, nil
}

func (r *salesRepoImpl) UpdateSale(in *pb.SaleUpdate) (*pb.SaleResponse, error) {
	query := `UPDATE sales SET `
	updates := []string{}
	params := map[string]interface{}{"id": in.Id, "company_id": in.CompanyId}

	if in.ClientId != "" {
		updates = append(updates, "client_id = :client_id")
		params["client_id"] = in.ClientId
	}
	if in.PaymentMethod != "" {
		updates = append(updates, "payment_method = :payment_method")
		params["payment_method"] = in.PaymentMethod
	}

	if len(updates) == 0 {
		return nil, errors.New("no fields to update")
	}

	query += strings.Join(updates, ", ")
	query += " WHERE id = :id AND company_id = :company_id RETURNING id, client_id, sold_by, total_sale_price, payment_method, created_at"

	sale := &pb.SaleResponse{}
	err := r.db.QueryRowx(query, params).Scan(&sale.Id, &sale.ClientId, &sale.SoldBy, &sale.TotalSalePrice, &sale.PaymentMethod, &sale.CreatedAt)
	if err != nil {
		return nil, err
	}

	return sale, nil
}

func (r *salesRepoImpl) GetSale(in *pb.SaleID) (*pb.SaleResponse, error) {
	query := `
        SELECT 
            s.id, 
            s.client_id, 
            s.sold_by, 
            s.total_sale_price, 
            s.payment_method, 
            s.created_at, 
            i.id AS item_id, 
            i.product_id, 
            i.quantity, 
            i.sale_price, 
            i.total_price
        FROM sales s
        LEFT JOIN sales_items i ON s.id = i.sale_id
        WHERE s.id = $1 AND s.company_id = $2
    `

	sale := &pb.SaleResponse{}
	var soldProducts []*pb.SalesItem

	rows, err := r.db.Queryx(query, in.Id, in.CompanyId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var item pb.SalesItem
		err = rows.Scan(
			&sale.Id,
			&sale.ClientId,
			&sale.SoldBy,
			&sale.TotalSalePrice,
			&sale.PaymentMethod,
			&sale.CreatedAt,
			&item.Id,
			&item.ProductId,
			&item.Quantity,
			&item.SalePrice,
			&item.TotalPrice,
		)
		if err != nil {
			return nil, err
		}

		if item.Id != "" {
			soldProducts = append(soldProducts, &item)
		}
	}

	sale.SoldProducts = soldProducts

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return sale, nil
}

func (r *salesRepoImpl) GetSaleList(in *pb.SaleFilter) (*pb.SaleList, error) {
	var sales []*pb.SaleResponse
	var queryBuilder strings.Builder
	var args []interface{}
	argIndex := 2

	queryBuilder.WriteString(`
        SELECT s.id, s.client_id, s.sold_by, s.total_sale_price, s.payment_method, s.created_at,
               i.id AS item_id, i.product_id, i.quantity, i.sale_price, i.total_price 
        FROM sales s 
        JOIN sales_items i ON s.id = i.sale_id
        WHERE s.company_id = $1
    `)
	args = append(args, in.CompanyId)

	if in.ClientId != "" {
		queryBuilder.WriteString(" AND s.client_id ILIKE '%' || $" + fmt.Sprint(argIndex) + " || '%'")
		args = append(args, in.ClientId)
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

	rows, err := r.db.Queryx(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list sales: %w", err)
	}
	defer rows.Close()

	salesMap := make(map[string]*pb.SaleResponse)

	for rows.Next() {
		var sale pb.SaleResponse
		var item pb.SalesItem
		err = rows.Scan(
			&sale.Id,
			&sale.ClientId,
			&sale.SoldBy,
			&sale.TotalSalePrice,
			&sale.PaymentMethod,
			&sale.CreatedAt,
			&item.Id,
			&item.ProductId,
			&item.Quantity,
			&item.SalePrice,
			&item.TotalPrice,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan sale: %w", err)
		}

		if _, exists := salesMap[sale.Id]; !exists {
			salesMap[sale.Id] = &sale
		}

		if item.Id != "" {
			salesMap[sale.Id].SoldProducts = append(salesMap[sale.Id].SoldProducts, &item)
		}
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over sales: %w", err)
	}

	for _, sale := range salesMap {
		sales = append(sales, sale)
	}

	return &pb.SaleList{Sales: sales}, nil
}

func (r *salesRepoImpl) DeleteSale(in *pb.SaleID) (*pb.Message, error) {
	_, err := r.db.Exec(`DELETE FROM sales_items WHERE sale_id = $1 AND company_id = $2`, in.Id, in.CompanyId)
	if err != nil {
		return nil, err
	}

	result, err := r.db.Exec(`DELETE FROM sales WHERE id = $1 AND company_id = $2`, in.Id, in.CompanyId)
	if err != nil {
		return nil, err
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return nil, errors.New("sale not found")
	}

	return &pb.Message{Message: "Sale deleted successfully"}, nil
}
