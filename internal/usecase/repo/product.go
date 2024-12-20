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

type productRepo struct {
	db *sqlx.DB
}

type productQuantity struct {
	db *sqlx.DB
}

func NewProductRepo(db *sqlx.DB) usecase.ProductsRepo {
	return &productRepo{db: db}
}

func NewProductQuantity(db *sqlx.DB) usecase.ProductQuantity {
	return &productQuantity{db: db}
}

//---------------- Product Category CRUD -----------------------------------------------------------------------------

func (p *productRepo) CreateProductCategory(in *pb.CreateCategoryRequest) (*pb.Category, error) {
	var category pb.Category

	query := `INSERT INTO product_categories (name, image_url, created_by) VALUES ($1, $2, $3) RETURNING id, name, image_url, created_by, created_at`
	err := p.db.QueryRowx(query, in.Name, in.ImageUrl, in.CreatedBy).Scan(&category.Id, &category.Name, &category.ImageUrl, &category.CreatedBy, &category.CreatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to create product category: %w", err)
	}
	return &category, nil
}

func (p *productRepo) DeleteProductCategory(in *pb.GetCategoryRequest) (*pb.Message, error) {
	query := `DELETE FROM product_categories WHERE id = $1`

	res, err := p.db.Exec(query, in.Id)
	if err != nil {
		return nil, fmt.Errorf("failed to delete product category: %w", err)
	}
	rows, _ := res.RowsAffected()

	return &pb.Message{Message: fmt.Sprintf("Deleted %d category(ies)", rows)}, nil
}

func (p *productRepo) GetProductCategory(in *pb.GetCategoryRequest) (*pb.Category, error) {

	if in == nil {
		return nil, fmt.Errorf("input parameter is nil")
	}

	query := `SELECT id, name, image_url,created_by, created_at FROM product_categories WHERE id = $1`

	var res pb.Category

	err := p.db.QueryRowx(query, in.Id).Scan(
		&res.Id,
		&res.Name,
		&res.ImageUrl,
		&res.CreatedBy,
		&res.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get product category: %w", err)
	}

	return &res, nil
}

func (p *productRepo) GetListProductCategory(in *pb.CategoryName) (*pb.CategoryList, error) {
	var categories []*pb.Category
	query := `SELECT id, name, image_url, created_by, created_at FROM product_categories`
	var args []interface{}

	if in.Name != "" {
		query += " WHERE name LIKE $1"
		args = append(args, "%"+in.Name+"%")
	}

	// Execute the query
	rows, err := p.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}
	defer rows.Close()

	// Iterate through the rows and build the categories list
	for rows.Next() {
		var category pb.Category
		if err := rows.Scan(&category.Id, &category.Name, &category.ImageUrl, &category.CreatedBy, &category.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		categories = append(categories, &category)
	}

	// Check for errors encountered during iteration
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error while iterating rows: %w", err)
	}

	return &pb.CategoryList{Categories: categories}, nil
}

// ---------------- End Product Category CRUD ------------------------------------------------------------------------

// ------------------- Product CRUD ------------------------------------------------------------------------

func (p *productRepo) CreateProduct(in *pb.CreateProductRequest) (*pb.Product, error) {
	var product pb.Product

	query := `
		INSERT INTO products (category_id, name, image_url, bill_format, incoming_price, standard_price, created_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, category_id, name, image_url, bill_format, incoming_price, standard_price, total_count, created_by, created_at
	`
	err := p.db.QueryRowx(query, in.CategoryId, in.Name, in.ImageUrl, in.BillFormat, in.IncomingPrice, in.StandardPrice, in.CreatedBy).
		Scan(&product.Id, &product.CategoryId, &product.Name, &product.ImageUrl, &product.BillFormat, &product.IncomingPrice,
			&product.StandardPrice, &product.TotalCount, &product.CreatedBy, &product.CreatedAt)

	if err != nil {
		return nil, err
	}

	return &product, nil
}

func (p *productRepo) UpdateProduct(in *pb.UpdateProductRequest) (*pb.Product, error) {
	// Initialize product struct
	product := &pb.Product{}
	query := `UPDATE products SET `
	var args []interface{}
	argCounter := 1

	// Dynamically build the query based on non-empty fields
	if in.CategoryId != "" {
		query += fmt.Sprintf("category_id = $%d, ", argCounter)
		args = append(args, in.CategoryId)
		argCounter++
	}
	if in.Name != "" {
		query += fmt.Sprintf("name = $%d, ", argCounter)
		args = append(args, in.Name)
		argCounter++
	}
	if in.BillFormat != "" {
		query += fmt.Sprintf("bill_format = $%d, ", argCounter)
		args = append(args, in.BillFormat)
		argCounter++
	}
	if in.IncomingPrice != 0 {
		query += fmt.Sprintf("incoming_price = $%d, ", argCounter)
		args = append(args, in.IncomingPrice)
		argCounter++
	}
	if in.StandardPrice != 0 {
		query += fmt.Sprintf("standard_price = $%d, ", argCounter)
		args = append(args, in.StandardPrice)
		argCounter++
	}
	if in.ImageUrl != "" {
		query += fmt.Sprintf("image_url = $%d, ", argCounter)
		args = append(args, in.ImageUrl)
		argCounter++
	}

	// Remove trailing comma and space, add WHERE clause
	query = query[:len(query)-2] + fmt.Sprintf(" WHERE id = $%d "+
		"RETURNING id, category_id, name, bill_format, incoming_price, standard_price, total_count, created_by, created_at", argCounter)
	args = append(args, in.Id)

	// Execute the query
	err := p.db.QueryRowx(query, args...).Scan(
		&product.Id,
		&product.CategoryId,
		&product.Name,
		&product.ImageUrl,
		&product.BillFormat,
		&product.IncomingPrice,
		&product.StandardPrice,
		&product.TotalCount,
		&product.CreatedBy,
		&product.CreatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to update product: %w", err)
	}

	return product, nil
}

func (p *productRepo) DeleteProduct(in *pb.GetProductRequest) (*pb.Message, error) {
	query := `DELETE FROM products WHERE id = $1`

	res, err := p.db.Exec(query, in.Id)
	if err != nil {
		return nil, fmt.Errorf("failed to delete product: %w", err)
	}
	rows, _ := res.RowsAffected()

	return &pb.Message{Message: fmt.Sprintf("Deleted %d product(s)", rows)}, nil
}

func (p *productRepo) GetProduct(in *pb.GetProductRequest) (*pb.Product, error) {
	if in == nil {
		return nil, fmt.Errorf("input parameter is nil")
	}

	var res pb.Product
	query := `SELECT id, category_id, name, image_url, bill_format, incoming_price, standard_price,
          total_count, created_by, created_at FROM products WHERE id = $1`

	// Используем QueryRowx и вручную маппим результат через Scan
	err := p.db.QueryRowx(query, in.Id).Scan(
		&res.Id,
		&res.CategoryId,
		&res.Name,
		&res.ImageUrl,
		&res.BillFormat,
		&res.IncomingPrice,
		&res.StandardPrice,
		&res.TotalCount,
		&res.CreatedBy,
		&res.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("product not found")
		}
		return nil, fmt.Errorf("failed to get product: %w", err)
	}

	return &res, nil

}

func (p *productRepo) GetProductList(in *pb.ProductFilter) (*pb.ProductList, error) {
	var products []*pb.Product
	var args []interface{}
	var filters []string
	argIndex := 1

	query := `
    SELECT id, category_id, name, image_url, bill_format, incoming_price, standard_price, total_count, created_by, created_at
    FROM products
`

	if in.CategoryId != "" {
		filters = append(filters, fmt.Sprintf(`category_id = $%d`, argIndex))
		args = append(args, in.CategoryId)
		argIndex++
	}
	if in.Name != "" {
		filters = append(filters, fmt.Sprintf(`name ILIKE $%d`, argIndex))
		args = append(args, "%"+in.Name+"%")
		argIndex++
	}
	if in.CreatedBy != "" {
		filters = append(filters, fmt.Sprintf(`created_by = $%d`, argIndex))
		args = append(args, in.CreatedBy)
		argIndex++
	}

	if len(filters) > 0 {
		query += " WHERE " + strings.Join(filters, " AND ")
	}

	query += " ORDER BY created_at DESC"

	rows, err := p.db.Queryx(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query products: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var product pb.Product
		err := rows.Scan(
			&product.Id,
			&product.CategoryId,
			&product.Name,
			&product.ImageUrl,
			&product.BillFormat,
			&product.IncomingPrice,
			&product.StandardPrice,
			&product.TotalCount,
			&product.CreatedBy,
			&product.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan product: %w", err)
		}
		products = append(products, &product)
	}

	// Проверяем ошибки при итерации
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over products: %w", err)
	}

	return &pb.ProductList{Products: products}, nil

}

// ------------------- End Product CRUD ------------------------------------------------------------------------

// -------------------------------------------- Must fix end Do Reflect -------------------------------------

func (p *productQuantity) AddProduct(in *entity.CountProductReq) (*entity.ProductNumber, error) {
	var product *entity.ProductNumber

	query := `
		UPDATE products
		SET total_count = total_count + $1
		WHERE id = $2
		RETURNING id, total_count
	`
	err := p.db.QueryRowx(query, in.Count, in.Id).
		Scan(product.ID, &product.TotalCount)
	if err != nil {
		return nil, fmt.Errorf("failed to add product stock: %w", err)
	}

	return product, nil
}

func (p *productQuantity) RemoveProduct(in *entity.CountProductReq) (*entity.ProductNumber, error) {
	var res *entity.ProductNumber

	query := `UPDATE products SET total_count = total_count - $1
		RETURNING id, total_count`

	err := p.db.Get(res, query, in)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (p *productQuantity) GetProductCount(in *entity.ProductID) (*entity.ProductNumber, error) {
	var res *entity.ProductNumber

	query := `SELECT id, total_count from products WHERE id = $1`

	err := p.db.Get(res, query, in)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (p *productQuantity) ProductCountChecker(in *entity.CountProductReq) (bool, error) {
	var res bool

	query := `select 'true' from products where id = $1 and total_count >= $2`

	err := p.db.Get(&res, query, in.Id, in.Count)
	if err != nil {
		return false, err
	}

	return res, nil
}
