package entity

// Category represents a product category.
type Category struct {
	ID        string `json:"id" db:"id"`
	Name      string `json:"name" db:"name"`
	CreatedBy string `json:"created_by" db:"created_by"`
	CreatedAt string `json:"created_at" db:"created_at"`
}

type CategoryList struct {
	Categories []Category `json:"categories"`
}

// Product related structs

// ProductID represents a product's ID.
type ProductID struct {
	ID string `json:"id" db:"id"`
}

// FilterProduct defines filter criteria for product queries.
type FilterProduct struct {
	CategoryID string `json:"category_id" db:"category_id"`
	Name       string `json:"name" db:"name"`
	TotalCount string `json:"total_count" db:"total_count"`
	CreatedBy  string `json:"created_by" db:"created_by"`
}

// ProductRequest represents data for creating a new product.
type ProductRequest struct {
	CategoryID    string `json:"category_id" db:"category_id"`
	Name          string `json:"name" db:"name"`
	BillFormat    string `json:"bill_format" db:"bill_format"`
	IncomingPrice int64  `json:"incoming_price" db:"incoming_price"`
	StandardPrice int64  `json:"standard_price" db:"standard_price"`
	CreatedBy     string `json:"created_by" db:"created_by"`
}

// ProductUpdate represents data for updating an existing product.
type ProductUpdate struct {
	ID            string `json:"id" db:"id"`
	CategoryID    string `json:"category_id" db:"category_id"`
	Name          string `json:"name" db:"name"`
	BillFormat    string `json:"bill_format" db:"bill_format"`
	IncomingPrice int64  `json:"incoming_price" db:"incoming_price"`
	StandardPrice int64  `json:"standard_price" db:"standard_price"`
}

// Product represents a product with complete data.
type Product struct {
	ID            string `json:"id" db:"id"`
	CategoryID    string `json:"category_id" db:"category_id"`
	Name          string `json:"name" db:"name"`
	BillFormat    string `json:"bill_format" db:"bill_format"`
	IncomingPrice int64  `json:"incoming_price" db:"incoming_price"`
	StandardPrice int64  `json:"standard_price" db:"standard_price"`
	TotalCount    int    `json:"total_count" db:"total_count"`
	CreatedBy     string `json:"created_by" db:"created_by"`
	CreatedAt     string `json:"created_at" db:"created_at"`
}

// ProductList contains a list of products.
type ProductList struct {
	Products []Product `json:"products"`
}

// CountProductReq represents a request to update product quantity.
type CountProductReq struct {
	ID    string `json:"id" db:"id"`
	Count int    `json:"count" db:"count"`
}

// ProductNumber represents the total count of a product.
type ProductNumber struct {
	ID         string `json:"id" db:"id"`
	TotalCount int    `json:"total_count" db:"total_count"`
}

// Message represents a simple message.
type Message struct {
	Message string `json:"message"`
}

// Purchase related structs

// PurchaseUpdate represents data for updating a purchase.
type PurchaseUpdate struct {
	ID            string `json:"id" db:"id"`
	SupplierID    string `json:"supplier_id" db:"supplier_id"`
	CompanyID     string `json:"company_id" db:"company_id"`
	Description   string `json:"description" db:"description"`
	PaymentMethod string `json:"payment_method" db:"payment_method"`
}

// PurchaseResponse represents a response after a purchase.
type PurchaseResponse struct {
	ID            string            `json:"id" db:"id"`
	SupplierID    string            `json:"supplier_id" db:"supplier_id"`
	PurchasedBy   string            `json:"purchased_by" db:"purchased_by"`
	TotalCost     int64             `json:"total_cost" db:"total_cost"`
	Description   string            `json:"description" db:"description"`
	PaymentMethod string            `json:"payment_method" db:"payment_method"`
	CreatedAt     string            `json:"created_at" db:"created_at"`
	PurchaseItems []PurchaseItemReq `json:"purchase_items" db:"purchase_items"`
}

// PurchaseRequest is used for creating a new purchase.
type PurchaseRequest struct {
	SupplierID    string            `json:"supplier_id" db:"supplier_id"`
	PurchasedBy   string            `json:"purchased_by" db:"purchased_by"`
	TotalCost     int64             `json:"total_cost" db:"total_cost"`
	CompanyID     string            `json:"company_id" db:"company_id"`
	Description   string            `json:"description" db:"description"`
	PaymentMethod string            `json:"payment_method" db:"payment_method"`
	PurchaseItems []PurchaseItemReq `json:"purchase_items" db:"purchase_items"`
}

// PurchaseItemReq represents an item in a purchase request.
type PurchaseItemReq struct {
	ProductID     string `json:"product_id" db:"product_id"`
	Quantity      int    `json:"quantity" db:"quantity"`
	PurchasePrice int64  `json:"purchase_price" db:"purchase_price"`
	TotalPrice    int64  `json:"total_price" db:"total_price"`
}

// Purchase represents a complete purchase entity.
type Purchase struct {
	SupplierID    string         `json:"supplier_id" db:"supplier_id"`
	PurchasedBy   string         `json:"purchased_by" db:"purchased_by"`
	Description   string         `json:"description" db:"description"`
	PaymentMethod string         `json:"payment_method" db:"payment_method"`
	PurchaseItems []PurchaseItem `json:"purchase_items" db:"purchase_items"`
}

// PurchaseItem represents an item within a purchase.
type PurchaseItem struct {
	ProductID     string `json:"product_id" db:"product_id"`
	Quantity      int    `json:"quantity" db:"quantity"`
	PurchasePrice int64  `json:"purchase_price" db:"purchase_price"`
}

// PurchaseID represents a purchase ID.
type PurchaseID struct {
	ID        string `json:"id" db:"id"`
	CompanyID string `json:"company_id" db:"company_id"`
}

// FilterPurchase provides filtering options for purchases.
type FilterPurchase struct {
	ProductID   string `json:"product_id" db:"product_id"`
	SupplierID  string `json:"supplier_id" db:"supplier_id"`
	PurchasedBy string `json:"purchased_by" db:"purchased_by"`
	CompanyID   string `json:"company_id" db:"company_id"`
	CreatedAt   string `json:"created_at" db:"created_at"`
}

// PurchaseList represents a list of purchases.
type PurchaseList struct {
	Purchases []PurchaseResponse `json:"purchases"`
}

// Sales related structs

// SaleRequest represents the request to create a sale.
type SaleRequest struct {
	ClientID      string      `json:"client_id" db:"client_id"`
	SoldBy        string      `json:"sold_by" db:"sold_by"`
	PaymentMethod string      `json:"payment_method" db:"payment_method"`
	SoldProducts  []SalesItem `json:"products" db:"products"`
}

// SalesItemRequest represents an item in a sale request.
type SalesItemRequest struct {
	SaleID    string `json:"sale_id" db:"sale_id"`
	ProductID string `json:"product_id" db:"product_id"`
	Quantity  int    `json:"quantity" db:"quantity"`
	SalePrice int64  `json:"sale_price" db:"sale_price"`
}

// SalesTotal represents the total sale details.
type SalesTotal struct {
	ClientID       string      `json:"client_id" db:"client_id"`
	CompanyID      string      `json:"company_id" db:"company_id"`
	SoldBy         string      `json:"sold_by" db:"sold_by"`
	TotalSalePrice int64       `json:"total_sale_price" db:"total_sale_price"`
	PaymentMethod  string      `json:"payment_method" db:"payment_method"`
	SoldProducts   []SalesItem `json:"products" db:"products"`
}

// SalesItemTotal represents an item with total sales info.
type SalesItemTotal struct {
	SaleID     string `json:"sale_id" db:"sale_id"`
	ProductID  string `json:"product_id" db:"product_id"`
	Quantity   int    `json:"quantity" db:"quantity"`
	SalePrice  int64  `json:"sale_price" db:"sale_price"`
	TotalPrice int64  `json:"total_price" db:"total_price"`
}

// SaleResponse represents the response for a sale.
type SaleResponse struct {
	ID             string      `json:"id" db:"id"`
	ClientID       string      `json:"client_id" db:"client_id"`
	SoldBy         string      `json:"sold_by" db:"sold_by"`
	TotalSalePrice int64       `json:"total_sale_price" db:"total_sale_price"`
	PaymentMethod  string      `json:"payment_method" db:"payment_method"`
	CreatedAt      string      `json:"created_at" db:"created_at"`
	SoldProducts   []SalesItem `json:"products" db:"products"`
}

// SalesItem represents a sold product in a sale.
type SalesItem struct {
	ID         string `json:"id" db:"id"`
	SaleID     string `json:"sale_id" db:"sale_id"`
	ProductID  string `json:"product_id" db:"product_id"`
	Quantity   int64  `json:"quantity" db:"quantity"`
	SalePrice  int64  `json:"sale_price" db:"sale_price"`
	TotalPrice int64  `json:"total_price" db:"total_price"`
}

// SaleUpdate represents data to update a sale.
type SaleUpdate struct {
	ID            string `json:"id" db:"id"`
	ClientID      string `json:"client_id" db:"client_id"`
	CompanyID     string `json:"company_id" db:"company_id"`
	PaymentMethod string `json:"payment_method" db:"payment_method"`
}

// SaleList contains a list of sales.
type SaleList struct {
	Sales []SaleResponse `json:"sales"`
}

// SaleID represents a sale's ID.
type SaleID struct {
	ID        string `json:"id" db:"id"`
	CompanyID string `json:"company_id" db:"company_id"`
}

// SaleFilter provides filtering options for sales.
type SaleFilter struct {
	StartDate string `json:"start_date" db:"start_date"`
	EndDate   string `json:"end_date" db:"end_date"`
	ClientID  string `json:"client_id" db:"client_id"`
	SoldBy    string `json:"sold_by" db:"sold_by"`
	CompanyID string `json:"company_id" db:"company_id"`
}
