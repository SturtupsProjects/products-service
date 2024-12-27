package entity

type ProductID struct {
	ID string `json:"id" db:"id"`
}

type CountProductReq struct {
	ID    string `json:"id" db:"id"`
	Count int    `json:"count" db:"count"`
}

type ProductNumber struct {
	ID         string `json:"id" db:"id"`
	TotalCount int    `json:"total_count" db:"total_count"`
}

type PurchaseRequest struct {
	SupplierID    string            `json:"supplier_id" db:"supplier_id"`
	PurchasedBy   string            `json:"purchased_by" db:"purchased_by"`
	TotalCost     float64           `json:"total_cost" db:"total_cost"`
	CompanyID     string            `json:"company_id" db:"company_id"`
	Description   string            `json:"description" db:"description"`
	PaymentMethod string            `json:"payment_method" db:"payment_method"`
	PurchaseItems []PurchaseItemReq `json:"purchase_items" db:"purchase_items"`
}

type PurchaseItemReq struct {
	ProductID     string  `json:"product_id" db:"product_id"`
	Quantity      int     `json:"quantity" db:"quantity"`
	PurchasePrice float64 `json:"purchase_price" db:"purchase_price"`
	TotalPrice    float64 `json:"total_price" db:"total_price"`
}

type Purchase struct {
	SupplierID    string         `json:"supplier_id" db:"supplier_id"`
	PurchasedBy   string         `json:"purchased_by" db:"purchased_by"`
	Description   string         `json:"description" db:"description"`
	PaymentMethod string         `json:"payment_method" db:"payment_method"`
	CompanyID     string         `json:"company_id" db:"company_id"`
	PurchaseItems []PurchaseItem `json:"purchase_items" db:"purchase_items"`
}

type PurchaseItem struct {
	ProductID     string  `json:"product_id" db:"product_id"`
	Quantity      int     `json:"quantity" db:"quantity"`
	PurchasePrice float64 `json:"purchase_price" db:"purchase_price"`
}

type SaleRequest struct {
	ClientID      string      `json:"client_id" db:"client_id"`
	SoldBy        string      `json:"sold_by" db:"sold_by"`
	PaymentMethod string      `json:"payment_method" db:"payment_method"`
	CompanyID     string      `json:"company_id" db:"company_id"`
	SoldProducts  []SalesItem `json:"products" db:"products"`
}

type SalesTotal struct {
	ClientID       string      `json:"client_id" db:"client_id"`
	CompanyID      string      `json:"company_id" db:"company_id"`
	SoldBy         string      `json:"sold_by" db:"sold_by"`
	TotalSalePrice float64     `json:"total_sale_price" db:"total_sale_price"`
	PaymentMethod  string      `json:"payment_method" db:"payment_method"`
	SoldProducts   []SalesItem `json:"products" db:"products"`
}

type SalesItem struct {
	ID         string  `json:"id" db:"id"`
	SaleID     string  `json:"sale_id" db:"sale_id"`
	ProductID  string  `json:"product_id" db:"product_id"`
	Quantity   int64   `json:"quantity" db:"quantity"`
	SalePrice  float64 `json:"sale_price" db:"sale_price"`
	TotalPrice float64 `json:"total_price" db:"total_price"`
}
