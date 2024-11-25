package usecase

import "crm-admin/internal/entity"

type ProductsRepo interface {
	CreateProductCategory(in *entity.CategoryName) (*entity.Category, error)
	DeleteProductCategory(in *entity.CategoryID) (*entity.Message, error)
	GetProductCategory(in *entity.CategoryID) (*entity.Category, error)
	GetListProductCategory(in *entity.CategoryName) (*entity.CategoryList, error)

	CreateProduct(in *entity.ProductRequest) (*entity.Product, error)
	UpdateProduct(in *entity.ProductUpdate) (*entity.Product, error)
	DeleteProduct(in *entity.ProductID) (*entity.Message, error)
	GetProduct(in *entity.ProductID) (*entity.Product, error)
	GetProductList(in *entity.FilterProduct) (*entity.ProductList, error)
}

type ProductQuantity interface {
	AddProduct(in *entity.CountProductReq) (*entity.ProductNumber, error)
	RemoveProduct(in *entity.CountProductReq) (*entity.ProductNumber, error)
	GetProductCount(in *entity.ProductID) (*entity.ProductNumber, error)
	ProductCountChecker(in *entity.CountProductReq) (bool, error)
}

type PurchasesRepo interface {
	CreatePurchase(in *entity.PurchaseRequest) (*entity.PurchaseResponse, error)
	UpdatePurchase(in *entity.PurchaseUpdate) (*entity.PurchaseResponse, error)
	GetPurchase(in *entity.PurchaseID) (*entity.PurchaseResponse, error)
	GetPurchaseList(in *entity.FilterPurchase) (*entity.PurchaseList, error)
	DeletePurchase(in *entity.PurchaseID) (*entity.Message, error)
}

type SalesRepo interface {
	CreateSale(in *entity.SalesTotal) (*entity.SaleResponse, error)
	UpdateSale(in *entity.SaleUpdate) (*entity.SaleResponse, error)
	GetSale(in *entity.SaleID) (*entity.SaleResponse, error)
	GetSaleList(filter *entity.SaleFilter) (*entity.SaleList, error)
	DeleteSale(in *entity.SaleID) (*entity.Message, error)
}

type ReturnedProductsRepo interface {
	CreateReturnedProducts() error
	UpdateReturnedProducts() error
	GetReturnedProducts() error
	GetReturnedProductsList() error
	DeleteReturnedProducts() error
}
