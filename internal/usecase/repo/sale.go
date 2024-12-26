package repo

import (
	"crm-admin/internal/entity"
	pb "crm-admin/internal/generated/products"
	"crm-admin/internal/usecase"
	"errors"
	"fmt"
	"github.com/jmoiron/sqlx"
	"log"
	"strings"
	"time"
)

type salesRepoImpl struct {
	db *sqlx.DB
}

func NewSalesRepo(db *sqlx.DB) usecase.SalesRepo {
	return &salesRepoImpl{db: db}
}

// SaleUpdateParams структура для обновления продажи
type SaleUpdateParams struct {
	ID            string  `db:"id"`
	CompanyID     string  `db:"company_id"`
	ClientID      *string `db:"client_id,omitempty"`
	PaymentMethod *string `db:"payment_method,omitempty"`
}

// CreateSale создает новую продажу и соответствующие элементы продажи
func (r *salesRepoImpl) CreateSale(in *entity.SalesTotal) (*pb.SaleResponse, error) {
	sale := &pb.SaleResponse{}

	query := `
		INSERT INTO sales (company_id, client_id, sold_by, total_sale_price, payment_method)
		VALUES ($1, $2, $3, $4, $5) 
		RETURNING id, created_at
	`

	err := r.db.QueryRowx(query, in.CompanyID, in.ClientID, in.SoldBy, in.TotalSalePrice, in.PaymentMethod).
		Scan(&sale.Id, &sale.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to create sale: %w", err)
	}

	tx, err := r.db.Beginx()
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	for _, item := range in.SoldProducts {
		item.SaleID = sale.Id
		itemQuery := `
			INSERT INTO sales_items (company_id, sale_id, product_id, quantity, sale_price, total_price)
			VALUES ($1, $2, $3, $4, $5, $6)
		`
		_, err := tx.Exec(itemQuery, in.CompanyID, item.SaleID, item.ProductID, item.Quantity, item.SalePrice, item.TotalPrice)
		if err != nil {
			return nil, fmt.Errorf("failed to insert sales item: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	sale.TotalSalePrice = in.TotalSalePrice
	sale.ClientId = in.ClientID
	sale.SoldBy = in.SoldBy
	sale.PaymentMethod = in.PaymentMethod

	return sale, nil
}

// UpdateSale обновляет детали продажи
func (r *salesRepoImpl) UpdateSale(in *pb.SaleUpdate) (*pb.SaleResponse, error) {
	updates := []string{}
	params := SaleUpdateParams{
		ID:        in.Id,
		CompanyID: in.CompanyId,
	}

	if in.ClientId != "" {
		updates = append(updates, "client_id = :client_id")
		params.ClientID = &in.ClientId
	}
	if in.PaymentMethod != "" {
		updates = append(updates, "payment_method = :payment_method")
		params.PaymentMethod = &in.PaymentMethod
	}

	if len(updates) == 0 {
		return nil, errors.New("no fields to update")
	}

	query := fmt.Sprintf(`
		UPDATE sales SET %s
		WHERE id = :id AND company_id = :company_id 
		RETURNING id, client_id, sold_by, total_sale_price, payment_method, created_at
	`, strings.Join(updates, ", "))

	stmt, err := r.db.PrepareNamed(query)
	if err != nil {
		return nil, fmt.Errorf("error preparing query: %w", err)
	}
	defer stmt.Close()

	sale := &pb.SaleResponse{}
	err = stmt.QueryRowx(params).StructScan(sale)
	if err != nil {
		return nil, fmt.Errorf("error executing update: %w", err)
	}

	return sale, nil
}

// GetSale получает детали продажи по ID
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
		return nil, fmt.Errorf("failed to query sale: %w", err)
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
			return nil, fmt.Errorf("failed to scan sale row: %w", err)
		}

		if item.Id != "" {
			soldProducts = append(soldProducts, &item)
		}
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over sale rows: %w", err)
	}

	sale.SoldProducts = soldProducts

	return sale, nil
}

// GetSaleList получает список продаж с возможными фильтрами
func (r *salesRepoImpl) GetSaleList(in *pb.SaleFilter) (*pb.SaleList, error) {
	var sales []*pb.SaleResponse
	var queryBuilder strings.Builder
	var args []interface{}
	argIndex := 2 // Первый аргумент уже используется для CompanyId

	queryBuilder.WriteString(`
		SELECT s.id, s.client_id, s.sold_by, s.total_sale_price, s.payment_method, s.created_at,
			   i.id AS item_id, i.product_id, i.quantity, i.sale_price, i.total_price 
		FROM sales s 
		LEFT JOIN sales_items i ON s.id = i.sale_id
		WHERE s.company_id = $1
	`)
	args = append(args, in.CompanyId)

	if in.ClientId != "" {
		queryBuilder.WriteString(fmt.Sprintf(" AND s.client_id ILIKE '%%' || $%d || '%%'", argIndex))
		args = append(args, in.ClientId)
		argIndex++
	}

	if in.SoldBy != "" {
		queryBuilder.WriteString(fmt.Sprintf(" AND s.sold_by ILIKE '%%' || $%d || '%%'", argIndex))
		args = append(args, in.SoldBy)
		argIndex++
	}

	if in.StartDate != "" {
		queryBuilder.WriteString(fmt.Sprintf(" AND DATE(s.created_at) >= DATE($%d)", argIndex))
		args = append(args, in.StartDate)
		argIndex++
	}

	if in.EndDate != "" {
		queryBuilder.WriteString(fmt.Sprintf(" AND DATE(s.created_at) <= DATE($%d)", argIndex))
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
			return nil, fmt.Errorf("failed to scan sale row: %w", err)
		}

		if _, exists := salesMap[sale.Id]; !exists {
			salesMap[sale.Id] = &sale
		}

		if item.Id != "" {
			salesMap[sale.Id].SoldProducts = append(salesMap[sale.Id].SoldProducts, &item)
		}
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over sale rows: %w", err)
	}

	for _, sale := range salesMap {
		sales = append(sales, sale)
	}

	return &pb.SaleList{Sales: sales}, nil
}

// DeleteSale удаляет продажу и связанные с ней элементы
func (r *salesRepoImpl) DeleteSale(in *pb.SaleID) (*pb.Message, error) {
	tx, err := r.db.Beginx()
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	_, err = tx.Exec(`DELETE FROM sales_items WHERE sale_id = $1 AND company_id = $2`, in.Id, in.CompanyId)
	if err != nil {
		return nil, fmt.Errorf("failed to delete sales items: %w", err)
	}

	result, err := tx.Exec(`DELETE FROM sales WHERE id = $1 AND company_id = $2`, in.Id, in.CompanyId)
	if err != nil {
		return nil, fmt.Errorf("failed to delete sale: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return nil, fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return nil, errors.New("sale not found")
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return &pb.Message{Message: "Sale deleted successfully"}, nil
}

// GetSalesByDay получает данные о продажах, сгруппированные по дню и продукту
func (r *salesRepoImpl) GetSalesByDay(request *pb.MostSoldProductsRequest) ([]*pb.DailySales, error) {
	log.Println("CompanyID:", request.CompanyId)

	query := `
		SELECT 
			TO_CHAR(s.created_at, 'Day') AS day,
			p.name,
			SUM(si.quantity) AS total_quantity
		FROM sales s
		INNER JOIN sales_items si ON s.id = si.sale_id
		INNER JOIN products p ON si.product_id = p.id
		WHERE s.company_id = $1 AND s.created_at BETWEEN $2 AND $3
		GROUP BY day, p.name
		ORDER BY day, total_quantity DESC
	`

	startDate, err := time.Parse(time.RFC3339, request.GetStartDate())
	if err != nil {
		return nil, fmt.Errorf("invalid start date format: %w", err)
	}
	endDate, err := time.Parse(time.RFC3339, request.GetEndDate())
	if err != nil {
		return nil, fmt.Errorf("invalid end date format: %w", err)
	}

	rows, err := r.db.Query(query, request.GetCompanyId(), startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to query sales by day: %w", err)
	}
	defer rows.Close()

	var results []*pb.DailySales
	for rows.Next() {
		var day, productName string
		var totalQuantity int64

		if err := rows.Scan(&day, &productName, &totalQuantity); err != nil {
			return nil, fmt.Errorf("failed to scan sales by day row: %w", err)
		}

		results = append(results, &pb.DailySales{
			Day:           strings.TrimSpace(day),
			ProductName:   productName,
			TotalQuantity: totalQuantity,
		})
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over sales by day rows: %w", err)
	}

	return results, nil
}

// GetTopClients получает топ клиентов по общей сумме продаж
func (r *salesRepoImpl) GetTopClients(in *pb.GetTopEntitiesRequest) ([]*pb.TopEntity, error) {
	if in.Limit == 0 {
		in.Limit = 10
	}

	query := `
		SELECT client_id, SUM(total_sale_price) AS total_sum 
		FROM sales
		WHERE company_id = $1 
		GROUP BY client_id  
		ORDER BY total_sum DESC 
		LIMIT $2
	`

	rows, err := r.db.Query(query, in.CompanyId, in.Limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query top clients: %w", err)
	}
	defer rows.Close()

	var entities []*pb.TopEntity

	for rows.Next() {
		var entity pb.TopEntity
		if err := rows.Scan(&entity.SupplierId, &entity.TotalValue); err != nil {
			return nil, fmt.Errorf("failed to scan top client row: %w", err)
		}
		entities = append(entities, &entity)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over top clients rows: %w", err)
	}

	return entities, nil
}

// GetTopSuppliers получает топ поставщиков по общей сумме затрат
func (r *salesRepoImpl) GetTopSuppliers(in *pb.GetTopEntitiesRequest) ([]*pb.TopEntity, error) {
	if in.Limit == 0 {
		in.Limit = 10
	}

	query := `
		SELECT supplier_id, SUM(total_cost) AS total_sum 
		FROM purchases
		WHERE company_id = $1 
		GROUP BY supplier_id  
		ORDER BY total_sum DESC 
		LIMIT $2
	`

	rows, err := r.db.Query(query, in.CompanyId, in.Limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query top suppliers: %w", err)
	}
	defer rows.Close()

	var entities []*pb.TopEntity

	for rows.Next() {
		var entity pb.TopEntity
		if err := rows.Scan(&entity.SupplierId, &entity.TotalValue); err != nil {
			return nil, fmt.Errorf("failed to scan top supplier row: %w", err)
		}
		entities = append(entities, &entity)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over top suppliers rows: %w", err)
	}

	return entities, nil
}
