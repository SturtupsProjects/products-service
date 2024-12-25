package repo

import (
	"crm-admin/internal/entity"
	pb "crm-admin/internal/generated/products"
	"crm-admin/internal/usecase"
	"errors"
	"fmt"
	"github.com/jmoiron/sqlx"
	"strings"
	"time"
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

// GetSalesByDay retrieves sales data grouped by day and product.
func (r *salesRepoImpl) GetSalesByDay(request *pb.MostSoldProductsRequest) ([]*pb.DailySales, error) {
	query := `
        SELECT 
            TO_CHAR(sale_date, 'Day') AS day,
            p.product_name,
            SUM(si.quantity) AS total_quantity
        FROM sales s
        INNER JOIN sales_items si ON s.sale_id = si.sale_id
        INNER JOIN products p ON si.product_id = p.product_id
        WHERE s.company_id = $1 AND s.sale_date BETWEEN $2 AND $3
        GROUP BY day, p.product_name
        ORDER BY day, total_quantity DESC
    `

	startDate, err := time.Parse(time.RFC3339, request.GetStartDate())
	if err != nil {
		return nil, err
	}
	endDate, err := time.Parse(time.RFC3339, request.GetEndDate())
	if err != nil {
		return nil, err
	}

	rows, err := r.db.Query(query, request.GetCompanyId(), startDate, endDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []*pb.DailySales
	for rows.Next() {
		var day, productName string
		var totalQuantity int64

		if err := rows.Scan(&day, &productName, &totalQuantity); err != nil {
			return nil, err
		}

		results = append(results, &pb.DailySales{
			Day:           day,
			ProductName:   productName,
			TotalQuantity: totalQuantity,
		})
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return results, nil
}

func (r *salesRepoImpl) GetTopClients(ctx context.Context, companyID string, limit int32) ([]*pb.TopEntity, error) {
	query := `
        SELECT c.id, SUM(s.total_sale_price) AS total_value
        FROM sales s
        JOIN clients c ON s.client_id = c.id
        WHERE s.company_id = $1
        GROUP BY c.id
        ORDER BY total_value DESC
        LIMIT $2
    `
	rows, err := r.db.QueryContext(ctx, query, companyID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entities []*pb.TopEntity
	for rows.Next() {
		var entity pb.TopEntity
		if err := rows.Scan(&entity.Name, &entity.TotalValue); err != nil {
			return nil, err
		}
		entities = append(entities, &entity)
	}

	return entities, nil
}

func (r *salesRepoImpl) GetTopSuppliers(ctx context.Context, companyID string, limit int32) ([]*pb.TopEntity, error) {
	query := `
        SELECT s.id, SUM(p.total_cost) AS total_value
        FROM purchases p
        JOIN suppliers s ON p.supplier_id = s.id
        WHERE p.company_id = $1
        GROUP BY s.id
        ORDER BY total_value DESC
        LIMIT $2
    `
	rows, err := r.db.QueryContext(ctx, query, companyID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entities []*pb.TopEntity
	for rows.Next() {
		var entity pb.TopEntity
		if err := rows.Scan(&entity.Name, &entity.TotalValue); err != nil {
			return nil, err
		}
		entities = append(entities, &entity)
	}

	return entities, nil
}
