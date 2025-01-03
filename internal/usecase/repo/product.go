package repo

import (
	"crm-admin/internal/entity"
	pb "crm-admin/internal/generated/products"
	"crm-admin/internal/usecase"
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

	query := `INSERT INTO product_categories (name, image_url, created_by, company_id) 
              VALUES ($1, $2, $3, $4) 
              RETURNING id, name, image_url, created_by, company_id, created_at`
	err := p.db.QueryRowx(query, in.Name, in.ImageUrl, in.CreatedBy, in.CompanyId).
		Scan(&category.Id, &category.Name, &category.ImageUrl, &category.CreatedBy, &category.CompanyId, &category.CreatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to create product category: %w", err)
	}
	return &category, nil
}

func (p *productRepo) UpdateProductCategory(in *pb.UpdateCategoryRequest) (*pb.Category, error) {
	category := &pb.Category{}
	query := `UPDATE product_categories SET `
	var args []interface{}
	argCounter := 1

	if in.Name != "" {
		query += fmt.Sprintf("name = $%d, ", argCounter)
		args = append(args, in.Name)
		argCounter++
	}
	if in.ImageUrl != "" {
		query += fmt.Sprintf("image_url = $%d, ", argCounter)
		args = append(args, in.ImageUrl)
		argCounter++
	}

	if len(args) == 0 {
		return nil, fmt.Errorf("no fields to update")
	}

	query = query[:len(query)-2] + fmt.Sprintf(" WHERE id = $%d AND company_id = $%d "+
		"RETURNING id, name, image_url, created_by, company_id, created_at", argCounter, argCounter+1)
	args = append(args, in.Id, in.CompanyId)

	err := p.db.QueryRowx(query, args...).Scan(
		&category.Id,
		&category.Name,
		&category.ImageUrl,
		&category.CreatedBy,
		&category.CompanyId,
		&category.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to update product category: %w", err)
	}

	return category, nil
}

func (p *productRepo) DeleteProductCategory(in *pb.GetCategoryRequest) (*pb.Message, error) {
	query := `DELETE FROM product_categories WHERE id = $1 AND company_id = $2`

	res, err := p.db.Exec(query, in.Id, in.CompanyId)
	if err != nil {
		return nil, fmt.Errorf("failed to delete product category: %w", err)
	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		return nil, fmt.Errorf("no records deleted")
	}

	return &pb.Message{Message: fmt.Sprintf("Deleted %d category(ies)", rows)}, nil
}

func (p *productRepo) GetProductCategory(in *pb.GetCategoryRequest) (*pb.Category, error) {
	if in == nil {
		return nil, fmt.Errorf("input parameter is nil")
	}

	query := `SELECT id, name, image_url, created_by, created_at FROM product_categories WHERE id = $1 AND company_id = $2`

	var res pb.Category

	err := p.db.QueryRowx(query, in.Id, in.CompanyId).Scan(
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
	query := `SELECT id, name, image_url, created_by, created_at FROM product_categories WHERE company_id = $1`
	var args []interface{}

	args = append(args, in.CompanyId)

	if in.Name != "" {
		query += " AND name LIKE $2"
		args = append(args, "%"+in.Name+"%")
	}

	rows, err := p.db.Queryx(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var category pb.Category
		if err := rows.Scan(&category.Id, &category.Name, &category.ImageUrl, &category.CreatedBy, &category.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		categories = append(categories, &category)
	}

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
		INSERT INTO products (category_id, name, image_url, bill_format, incoming_price, standard_price, company_id, created_by, total_count)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, COALESCE($9, 0))
		RETURNING id, category_id, name, image_url, bill_format, incoming_price, standard_price, total_count, created_by, created_at
	`
	err := p.db.QueryRowx(query, in.CategoryId, in.Name, in.ImageUrl, in.BillFormat, in.IncomingPrice, in.StandardPrice, in.CompanyId, in.CreatedBy, in.TotalCount).
		Scan(&product.Id, &product.CategoryId, &product.Name, &product.ImageUrl, &product.BillFormat, &product.IncomingPrice,
			&product.StandardPrice, &product.TotalCount, &product.CreatedBy, &product.CreatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to create product: %w", err)
	}

	return &product, nil
}

func (p *productRepo) CreateBulkProducts(in *pb.CreateBulkProductsRequest) (*pb.BulkCreateResponse, error) {
	// Start a transaction
	tx, err := p.db.Beginx()
	if err != nil {
		return nil, fmt.Errorf("failed to start transaction: %w", err)
	}

	// Prepare the query for inserting products
	query := `
        INSERT INTO products (category_id, name, image_url, bill_format, incoming_price, standard_price, company_id, created_by, total_count)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, COALESCE($9, 0))
        RETURNING id, category_id, name, image_url, bill_format, incoming_price, standard_price, total_count, created_by, created_at
    `

	var createdProducts []*pb.Product

	// Iterate over the list of products and insert each one
	for _, productReq := range in.Products {
		var product pb.Product
		err := tx.QueryRowx(query, in.CategoryId, productReq.Name, productReq.ImageUrl, productReq.BillFormat, productReq.IncomingPrice,
			productReq.StandardPrice, in.CompanyId, in.CreatedBy, productReq.TotalCount).
			Scan(&product.Id, &product.CategoryId, &product.Name, &product.ImageUrl, &product.BillFormat, &product.IncomingPrice,
				&product.StandardPrice, &product.TotalCount, &product.CreatedBy, &product.CreatedAt)

		if err != nil {
			// Roll back the transaction on error
			if rollbackErr := tx.Rollback(); rollbackErr != nil {
				return nil, fmt.Errorf("failed to rollback transaction after error: %w", rollbackErr)
			}
			return nil, fmt.Errorf("failed to create product: %w", err)
		}

		// Add the successfully created product to the response list
		createdProducts = append(createdProducts, &product)
	}

	// Commit the transaction
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	// Build the bulk creation response
	response := &pb.BulkCreateResponse{
		Success:  true,
		Products: createdProducts,
		Message:  "Bulk products created successfully",
	}

	return response, nil
}

func (p *productRepo) UpdateProduct(in *pb.UpdateProductRequest) (*pb.Product, error) {
	product := &pb.Product{}
	query := `UPDATE products SET `
	var args []interface{}
	argCounter := 1

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

	if len(args) == 0 {
		return nil, fmt.Errorf("no fields to update")
	}

	query = query[:len(query)-2] + fmt.Sprintf(" WHERE id = $%d AND company_id = $%d "+
		"RETURNING id, category_id, name, bill_format, incoming_price, standard_price, total_count, created_by, created_at, image_url",
		argCounter, argCounter+1)
	args = append(args, in.Id, in.CompanyId)

	err := p.db.QueryRowx(query, args...).Scan(
		&product.Id,
		&product.CategoryId,
		&product.Name,
		&product.BillFormat,
		&product.IncomingPrice,
		&product.StandardPrice,
		&product.TotalCount,
		&product.CreatedBy,
		&product.CreatedAt,
		&product.ImageUrl,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to update product: %w", err)
	}

	return product, nil
}

func (p *productRepo) DeleteProduct(in *pb.GetProductRequest) (*pb.Message, error) {
	query := `DELETE FROM products WHERE id = $1 AND company_id = $2`

	res, err := p.db.Exec(query, in.Id, in.CompanyId)
	if err != nil {
		return nil, fmt.Errorf("failed to delete product: %w", err)
	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		return nil, fmt.Errorf("no records deleted")
	}

	return &pb.Message{Message: fmt.Sprintf("Deleted %d product(s)", rows)}, nil
}

func (p *productRepo) GetProduct(in *pb.GetProductRequest) (*pb.Product, error) {
	if in == nil {
		return nil, fmt.Errorf("input parameter is nil")
	}

	var res pb.Product
	query := `SELECT id, category_id, name, image_url, bill_format, incoming_price, standard_price,
          total_count, created_by, created_at FROM products WHERE id = $1 AND company_id = $2`

	err := p.db.QueryRowx(query, in.Id, in.CompanyId).Scan(
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
		return nil, fmt.Errorf("failed to get product: %w", err)
	}

	return &res, nil
}

func (p *productRepo) GetProductList(in *pb.ProductFilter) (*pb.ProductList, error) {
	var products []*pb.Product
	baseQuery := `
		SELECT id, category_id, name, image_url, bill_format, incoming_price, standard_price, total_count, created_by, created_at
		FROM products
		WHERE company_id = $1` // Жёсткое условие для company_id
	var args []interface{}
	args = append(args, in.CompanyId) // Первым параметром всегда будет company_id

	var conditions []string

	if in.CategoryId != "" {
		conditions = append(conditions, fmt.Sprintf("category_id = $%d", len(args)+1))
		args = append(args, in.CategoryId)
	}

	if in.Name != "" {
		conditions = append(conditions, fmt.Sprintf("name LIKE $%d", len(args)+1))
		args = append(args, "%"+in.Name+"%")
	}

	if in.CreatedBy != "" {
		conditions = append(conditions, fmt.Sprintf("created_by = $%d", len(args)+1))
		args = append(args, in.CreatedBy)
	}

	if in.CreatedAt != "" {
		conditions = append(conditions, fmt.Sprintf("created_at = $%d", len(args)+1))
		args = append(args, in.CreatedAt)
	}

	if len(conditions) > 0 {
		baseQuery += " AND " + strings.Join(conditions, " AND ")
	}

	// Выполнение запроса
	rows, err := p.db.Queryx(baseQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}
	defer rows.Close()

	// Чтение строк из результата
	for rows.Next() {
		var product pb.Product
		if err := rows.Scan(
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
		); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		products = append(products, &product)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error while iterating rows: %w", err)
	}

	return &pb.ProductList{Products: products}, nil
}

// ------------------- End Product CRUD ------------------------------------------------------------------------

//------------------- Product Quantity CRUD ------------------------------------------------------------------------

func (p *productQuantity) AddProduct(in *entity.CountProductReq) (*entity.ProductNumber, error) {
	product := &entity.ProductNumber{}

	query := `
		UPDATE products
	SET total_count = total_count + $1
	WHERE id = $2
	RETURNING id, total_count
	`

	err := p.db.QueryRowx(query, in.Count, in.ID).
		Scan(&product.ID, &product.TotalCount)
	if err != nil {
		return nil, fmt.Errorf("failed to add product stock: %w", err)
	}

	return product, nil
}

func (p *productQuantity) RemoveProduct(in *entity.CountProductReq) (*entity.ProductNumber, error) {
	res := &entity.ProductNumber{}

	query := `UPDATE products SET total_count = total_count - $1
	RETURNING id, total_count`

	err := p.db.Get(res, query, in)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (p *productQuantity) GetProductCount(in *entity.ProductID) (*entity.ProductNumber, error) {
	res := &entity.ProductNumber{}

	query := ` SELECT id, total_count from products WHERE id = $1`

	err := p.db.Get(res, query, in)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (p *productQuantity) ProductCountChecker(in *entity.CountProductReq) (bool, error) {
	var res bool

	query := ` select 'true' from products where id = $1 and total_count >= $2 `

	err := p.db.Get(&res, query, in.ID, in.Count)
	if err != nil {
		return false, err
	}

	return res, nil
}

//------------------- End Product Quantity CRUD ------------------------------------------------------------------
