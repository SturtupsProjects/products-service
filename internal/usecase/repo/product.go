package repo

import (
	"crm-admin/internal/entity"
	"crm-admin/internal/usecase"
	pb "crm-admin/pkg/gednerated/products"
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

func (p *productRepo) CreateProductCategory(in *entity.CategoryName) (*pb.Category, error) {
	var category pb.Category

	query := `INSERT INTO product_categories (name, created_by) VALUES ($1, $2) RETURNING id, name,created_by, created_at`
	err := p.db.QueryRowx(query, in.Name, in.CreatedBy).Scan(&category.Id, &category.Name, &category.CreatedBy, &category.CreatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to create product category: %w", err)
	}
	return &category, nil
}

func (p *productRepo) DeleteProductCategory(in *entity.CategoryID) (*pb.Message, error) {
	query := `DELETE FROM product_categories WHERE id = $1`

	res, err := p.db.Exec(query, in.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to delete product category: %w", err)
	}
	rows, _ := res.RowsAffected()

	return &pb.Message{Message: fmt.Sprintf("Deleted %d category(ies)", rows)}, nil
}

func (p *productRepo) GetProductCategory(in *entity.CategoryID) (*pb.Category, error) {
	var category *pb.Category
	query := `SELECT id, name, created_by, created_at FROM product_categories WHERE id = $1`

	err := p.db.Get(&category, query, in.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get product category: %w", err)
	}

	return category, nil
}

func (p *productRepo) GetListProductCategory(in *entity.CategoryName) (*pb.CategoryList, error) {
	var categories []*pb.Category
	query := `SELECT id, name, created_by,created_at FROM product_categories`
	args := make([]interface{}, 1)

	if in.Name != "" {
		query += "WHERE name LIKE $1"
		args = append(args, "%"+in.Name+"%")
	}

	err := p.db.Select(&categories, query, args)
	if err != nil {
		return nil, fmt.Errorf("failed to list product categories: %w", err)
	}
	return &pb.CategoryList{Categories: categories}, nil
}

// ---------------- End Product Category CRUD ------------------------------------------------------------------------

// ------------------- Product CRUD ------------------------------------------------------------------------

func (p *productRepo) CreateProduct(in *entity.ProductRequest) (*pb.Product, error) {
	var product pb.Product

	query := `
		INSERT INTO products (category_id, name, bill_format, incoming_price, standard_price, created_by)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, category_id, name, bill_format, incoming_price, standard_price, total_count, created_by, created_at
	`
	err := p.db.QueryRowx(query, in.CategoryID, in.Name, in.BillFormat, in.IncomingPrice, in.StandardPrice, in.CreatedBy).
		Scan(&product.Id, &product.CategoryId, &product.Name, &product.BillFormat, &product.IncomingPrice,
			&product.StandardPrice, &product.TotalCount, &product.CreatedBy, &product.CreatedAt)

	if err != nil {
		return nil, err
	}

	return &product, nil
}

func (p *productRepo) UpdateProduct(in *entity.ProductUpdate) (*pb.Product, error) {
	var product *pb.Product
	query := `UPDATE products SET `
	var args []interface{}
	argCounter := 1

	// Dynamically build the query based on non-empty fields
	if in.CategoryID != "" {
		query += fmt.Sprintf("category_id = $%d, ", argCounter)
		args = append(args, in.CategoryID)
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

	// Remove trailing comma and space, add WHERE clause
	query = query[:len(query)-2] + fmt.Sprintf(" WHERE id = $%d "+
		"RETURNING id, category_id, name, bill_format, incoming_price, standard_price, total_count, created_by, created_at", argCounter)
	args = append(args, in.ID)

	// Execute the query
	err := p.db.QueryRowx(query, args...).Scan(&product.Id, &product.CategoryId, &product.Name, &product.BillFormat,
		&product.IncomingPrice, &product.StandardPrice, &product.TotalCount, &product.CreatedBy, &product.CreatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to update product: %w", err)
	}

	return product, nil
}

func (p *productRepo) DeleteProduct(in *entity.ProductID) (*pb.Message, error) {
	query := `DELETE FROM products WHERE id = $1`

	res, err := p.db.Exec(query, in.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to delete product: %w", err)
	}
	rows, _ := res.RowsAffected()

	return &pb.Message{Message: fmt.Sprintf("Deleted %d product(s)", rows)}, nil
}

func (p *productRepo) GetProduct(in *entity.ProductID) (*pb.Product, error) {
	var product *pb.Product

	query := `SELECT id, category_id, name, bill_format, incoming_price, standard_price,
       total_count, created_by, created_at FROM products WHERE id = $1`

	err := p.db.Get(&product, query, in.ID)

	if err != nil {
		return nil, fmt.Errorf("failed to get product: %w", err)
	}

	return product, nil
}

func (p *productRepo) GetProductList(in *entity.FilterProduct) (*pb.ProductList, error) {
	var products []*pb.Product
	var args []interface{}
	var filters []string

	query := `
		SELECT id, category_id, name, bill_format, incoming_price, standard_price, total_count, created_by, created_at
		FROM products 
	`

	// Dynamically build the WHERE clause based on filters
	if in.CategoryId != "" {
		filters = append(filters, `category_id = ?`)
		args = append(args, in.CategoryId)
	}
	if in.Name != "" {
		filters = append(filters, `name ILIKE ?`)
		args = append(args, "%"+in.Name+"%")
	}
	if in.TotalCount != "" {
		filters = append(filters, `total_count = ?`)
		args = append(args, in.TotalCount)
	}
	if in.CreatedBy != "" {
		filters = append(filters, `created_by = ?`)
		args = append(args, in.CreatedBy)
	}

	// Add the filters to the query if there are any
	if len(filters) > 0 {
		query += " WHERE " + strings.Join(filters, " AND ")
	}

	// Add ordering to the query
	query += " ORDER BY created_at DESC"

	// Execute the query
	err := p.db.Select(&products, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list products: %w", err)
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
