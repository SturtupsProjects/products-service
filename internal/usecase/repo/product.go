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

// ---------------- Product Category CRUD -----------------------------------------------------------------------------

func (p *productRepo) CreateProductCategory(in *pb.CreateCategoryRequest) (*pb.Category, error) {
	var category pb.Category

	query := `INSERT INTO product_categories (name, image_url, created_by, company_id, branch_id) 
              VALUES ($1, $2, $3, $4, $5) 
              RETURNING id, name, image_url, created_by, company_id, branch_id, created_at`
	err := p.db.QueryRowx(query, in.Name, in.ImageUrl, in.CreatedBy, in.CompanyId, in.BranchId).
		Scan(&category.Id, &category.Name, &category.ImageUrl, &category.CreatedBy, &category.CompanyId, &category.BranchId, &category.CreatedAt)

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

	query = query[:len(query)-2] + fmt.Sprintf(" WHERE id = $%d AND company_id = $%d AND branch_id = $%d "+
		"RETURNING id, name, image_url, created_by, company_id, branch_id, created_at", argCounter, argCounter+1, argCounter+2)
	args = append(args, in.Id, in.CompanyId, in.BranchId)

	err := p.db.QueryRowx(query, args...).Scan(
		&category.Id,
		&category.Name,
		&category.ImageUrl,
		&category.CreatedBy,
		&category.CompanyId,
		&category.BranchId,
		&category.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to update product category: %w", err)
	}

	return category, nil
}

func (p *productRepo) DeleteProductCategory(in *pb.GetCategoryRequest) (*pb.Message, error) {
	query := `DELETE FROM product_categories WHERE id = $1 AND company_id = $2 AND branch_id = $3`

	res, err := p.db.Exec(query, in.Id, in.CompanyId, in.BranchId)
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

	query := `SELECT id, name, image_url, created_by, company_id, branch_id, created_at 
			  FROM product_categories WHERE id = $1 AND company_id = $2 AND branch_id = $3`

	var res pb.Category

	err := p.db.QueryRowx(query, in.Id, in.CompanyId, in.BranchId).Scan(
		&res.Id,
		&res.Name,
		&res.ImageUrl,
		&res.CreatedBy,
		&res.CompanyId,
		&res.BranchId,
		&res.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get product category: %w", err)
	}

	return &res, nil
}

func (p *productRepo) GetListProductCategory(in *pb.CategoryName) (*pb.CategoryList, error) {
	var categories []*pb.Category
	var args []interface{}

	var queryBuilder strings.Builder
	queryBuilder.WriteString("SELECT id, name, image_url, created_by, company_id, branch_id, created_at, COUNT(*) OVER() AS total_count FROM product_categories WHERE company_id = $1 AND branch_id = $2")

	args = append(args, in.CompanyId, in.BranchId)

	if in.Name != "" {
		queryBuilder.WriteString(" AND name ILIKE $3") // Use ILIKE for case-insensitive matching
		args = append(args, "%"+in.Name+"%")
	}

	queryBuilder.WriteString(" ORDER BY created_at DESC")

	if in.Limit > 0 && in.Page > 0 {
		queryBuilder.WriteString(fmt.Sprintf(" LIMIT $%d OFFSET $%d", len(args)+1, len(args)+2))
		args = append(args, in.Limit, (in.Page-1)*in.Limit)
	}

	query := queryBuilder.String()

	rows, err := p.db.Queryx(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}
	defer rows.Close()

	var totalCount int64
	for rows.Next() {
		var category pb.Category
		var total int64
		if err := rows.Scan(
			&category.Id,
			&category.Name,
			&category.ImageUrl,
			&category.CreatedBy,
			&category.CompanyId,
			&category.BranchId,
			&category.CreatedAt,
			&total, // Scanning total_count from the query
		); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		totalCount = total // Set the total count (will be the same for all rows)
		categories = append(categories, &category)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error while iterating rows: %w", err)
	}

	return &pb.CategoryList{
		Categories: categories,
		TotalCount: totalCount,
	}, nil
}

// ---------------- End Product Category CRUD ------------------------------------------------------------------------

// ------------------- Product CRUD ----------------------------------------------------------------------------------

func (p *productRepo) CreateProduct(in *pb.CreateProductRequest) (*pb.Product, error) {
	var product pb.Product

	query := `
		INSERT INTO products (category_id, name, image_url, bill_format, incoming_price, standard_price, company_id, branch_id, created_by, total_count)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, COALESCE($10, 0))
		RETURNING id, category_id, name, image_url, bill_format, incoming_price, standard_price, total_count, company_id, branch_id, created_by, created_at
	`
	err := p.db.QueryRowx(query, in.CategoryId, in.Name, in.ImageUrl, in.BillFormat, in.IncomingPrice, in.StandardPrice, in.CompanyId, in.BranchId, in.CreatedBy, 0).
		Scan(&product.Id, &product.CategoryId, &product.Name, &product.ImageUrl, &product.BillFormat, &product.IncomingPrice,
			&product.StandardPrice, &product.TotalCount, &product.CompanyId, &product.BranchId, &product.CreatedBy, &product.CreatedAt)

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
        INSERT INTO products (category_id, name, image_url, bill_format, incoming_price, standard_price, company_id, branch_id, created_by, total_count)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, COALESCE($10, 0))
        RETURNING id, category_id, name, image_url, bill_format, incoming_price, standard_price, total_count, created_by, created_at
    `

	var createdProducts []*pb.Product

	// Iterate over the list of products and insert each one
	for _, productReq := range in.Products {
		var product pb.Product
		err := tx.QueryRowx(query, in.CategoryId, productReq.Name, "https://smartadmin.uz/static/media/gif2.aff05f0cb04b5d100ae4.png", productReq.BillFormat, productReq.IncomingPrice,
			productReq.StandardPrice, in.CompanyId, in.BranchId, in.CreatedBy, productReq.TotalCount).
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

	// Updated query with branch_id check
	query = query[:len(query)-2] + fmt.Sprintf(" WHERE id = $%d AND company_id = $%d AND branch_id = $%d "+
		"RETURNING id, category_id, name, bill_format, incoming_price, standard_price, total_count, created_by, created_at, image_url",
		argCounter, argCounter+1, argCounter+2)
	args = append(args, in.Id, in.CompanyId, in.BranchId)

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
          total_count, created_by, created_at, branch_id FROM products WHERE id = $1 AND company_id = $2 AND branch_id = $3`

	err := p.db.QueryRowx(query, in.Id, in.CompanyId, in.BranchId).Scan(
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
		&res.BranchId, // Include branch_id in the result scan
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get product: %w", err)
	}

	return &res, nil
}

func (p *productRepo) GetProductList(in *pb.ProductFilter) (*pb.ProductList, error) {
	var products []*pb.Product
	var args []interface{}
	argIndex := 3

	// Условия фильтрации
	conditions := []string{"company_id = $1", "branch_id = $2"}
	args = append(args, in.CompanyId, in.BranchId)

	// Фильтр по категории
	if in.CategoryId != "" {
		conditions = append(conditions, fmt.Sprintf("category_id = $%d", argIndex))
		args = append(args, in.CategoryId)
		argIndex++
	}
	// Фильтр по имени
	if in.Name != "" {
		conditions = append(conditions, fmt.Sprintf("name ILIKE $%d", argIndex))
		args = append(args, "%"+in.Name+"%")
		argIndex++
	}
	// Фильтр по создателю
	if in.CreatedBy != "" {
		conditions = append(conditions, fmt.Sprintf("created_by = $%d", argIndex))
		args = append(args, in.CreatedBy)
		argIndex++
	}
	// Фильтр по дате создания
	if in.CreatedAt != "" {
		conditions = append(conditions, fmt.Sprintf("created_at = $%d", argIndex))
		args = append(args, in.CreatedAt)
		argIndex++
	}
	// Фильтр по минимальному количеству
	if in.TotalCount > 0 {
		conditions = append(conditions, fmt.Sprintf("total_count <= $%d", argIndex))
		args = append(args, in.TotalCount)
		argIndex++
	}

	// Формирование подзапроса для total_count
	countQuery := fmt.Sprintf(`
		(SELECT COUNT(*)
		 FROM products
		 WHERE %s
		) AS total_count`, strings.Join(conditions, " AND "))

	// Формирование основного запроса
	baseQuery := fmt.Sprintf(`
		SELECT 
			id, category_id, name, image_url, bill_format, incoming_price, 
			standard_price, total_count, created_by, created_at, branch_id,
			%s
		FROM products
		WHERE %s
		ORDER BY created_at DESC`, countQuery, strings.Join(conditions, " AND "))

	// Добавление пагинации
	if in.Limit > 0 && in.Page > 0 {
		baseQuery += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argIndex, argIndex+1)
		args = append(args, in.Limit, (in.Page-1)*in.Limit)
	}

	// Выполнение запроса
	rows, err := p.db.Queryx(baseQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}
	defer rows.Close()

	var totalCount int64
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
			&product.BranchId,
			&totalCount, // Сканируем результат подзапроса
		); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		products = append(products, &product)
	}

	// Проверка на ошибки при чтении строк
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error while iterating rows: %w", err)
	}

	// Возврат результата
	return &pb.ProductList{
		Products:   products,
		TotalCount: totalCount,
	}, nil
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

func (p *productQuantity) TransferProducts(in *pb.TransferReq) error {
	tx, err := p.db.Beginx()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
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

	// Проверить или создать категорию "transferred_from_branch" для целевого филиала
	var categoryID string
	err = tx.Get(&categoryID, `
		SELECT id
		FROM product_categories
		WHERE name = $1 AND branch_id = $2`,
		"transferred_from_branch", in.ToBranchId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			categoryID = uuid.New().String()
			_, err = tx.Exec(`
				INSERT INTO product_categories (id, name, branch_id, company_id, created_by, created_at)
				VALUES ($1, $2, $3, $4, $5, NOW())`,
				categoryID, "transferred_from_branch", in.ToBranchId, in.CompanyId, in.TransferredBy)
			if err != nil {
				return fmt.Errorf("failed to create category: %w", err)
			}
		} else {
			return fmt.Errorf("failed to fetch category: %w", err)
		}
	}

	// Создать запись о трансфере
	transferID := uuid.New().String()
	_, err = tx.Exec(`
		INSERT INTO transfers (id, transferred_by, from_branch_id, to_branch_id, description, company_id, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, NOW())`,
		transferID, in.TransferredBy, in.FromBranchId, in.ToBranchId, in.Description, in.CompanyId)
	if err != nil {
		return fmt.Errorf("failed to create transfer: %w", err)
	}

	// Обработать товары в трансфере
	for _, productReq := range in.Products {
		// Уменьшить количество на исходном филиале
		res, err := tx.Exec(`
			UPDATE products
			SET total_count = total_count - $1
			WHERE id = $2 AND branch_id = $3 AND total_count >= $1`,
			productReq.ProductQuantity, productReq.ProductId, in.FromBranchId)
		if err != nil {
			return fmt.Errorf("failed to reduce product quantity: %w", err)
		}

		rowsAffected, err := res.RowsAffected()
		if err != nil {
			return fmt.Errorf("failed to get affected rows: %w", err)
		}
		if rowsAffected == 0 {
			return fmt.Errorf("insufficient quantity or product not found in source branch")
		}

		// Добавить или обновить продукт в целевом филиале
		_, err = tx.Exec(`
			INSERT INTO products (id, category_id, name, image_url, bill_format, incoming_price, standard_price, total_count, branch_id, company_id, created_by, created_at)
			SELECT id, $5, name, image_url, bill_format, incoming_price, standard_price, $1, $2, company_id, $6, NOW()
			FROM products
			WHERE id = $3 AND branch_id = $4
			ON CONFLICT (id, branch_id) DO UPDATE
			SET total_count = products.total_count + EXCLUDED.total_count`,
			productReq.ProductQuantity, in.ToBranchId, productReq.ProductId, in.FromBranchId, categoryID, in.TransferredBy)
		if err != nil {
			return fmt.Errorf("failed to increase product quantity: %w", err)
		}

		// Создать запись в таблице transfer_products
		_, err = tx.Exec(`
			INSERT INTO transfer_products (id, product_transfers_id, product_id, quantity)
			VALUES ($1, $2, $3, $4)`,
			uuid.New().String(), transferID, productReq.ProductId, productReq.ProductQuantity)
		if err != nil {
			return fmt.Errorf("failed to create transfer product record: %w", err)
		}
	}

	return nil
}

//------------------- End Product Quantity CRUD ------------------------------------------------------------------
