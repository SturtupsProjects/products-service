package usecase

import (
	"crm-admin/internal/entity"
	pb "crm-admin/internal/generated/products"
)

type ProductsRepo interface {
	CreateProductCategory(in *pb.CreateCategoryRequest) (*pb.Category, error)
	UpdateProductCategory(in *pb.UpdateCategoryRequest) (*pb.Category, error)
	DeleteProductCategory(in *pb.GetCategoryRequest) (*pb.Message, error)
	GetProductCategory(in *pb.GetCategoryRequest) (*pb.Category, error)
	GetListProductCategory(in *pb.CategoryName) (*pb.CategoryList, error)

	CreateProduct(in *pb.CreateProductRequest) (*pb.Product, error)
	UpdateProduct(in *pb.UpdateProductRequest) (*pb.Product, error)
	DeleteProduct(in *pb.GetProductRequest) (*pb.Message, error)
	GetProduct(in *pb.GetProductRequest) (*pb.Product, error)
	GetProductList(in *pb.ProductFilter) (*pb.ProductList, error)
}

type ProductQuantity interface {
	AddProduct(in *entity.CountProductReq) (*entity.ProductNumber, error)
	RemoveProduct(in *entity.CountProductReq) (*entity.ProductNumber, error)
	GetProductCount(in *entity.ProductID) (*entity.ProductNumber, error)
	ProductCountChecker(in *entity.CountProductReq) (bool, error)
}

type PurchasesRepo interface {
	CreatePurchase(in *entity.PurchaseRequest) (*pb.PurchaseResponse, error)
	UpdatePurchase(in *entity.PurchaseUpdate) (*pb.PurchaseResponse, error)
	GetPurchase(in *entity.PurchaseID) (*pb.PurchaseResponse, error)
	GetPurchaseList(in *entity.FilterPurchase) (*pb.PurchaseList, error)
	DeletePurchase(in *entity.PurchaseID) (*pb.Message, error)
}

type SalesRepo interface {
	CreateSale(in *entity.SalesTotal) (*pb.SaleResponse, error)
	UpdateSale(in *entity.SaleUpdate) (*pb.SaleResponse, error)
	GetSale(in *entity.SaleID) (*pb.SaleResponse, error)
	GetSaleList(filter *entity.SaleFilter) (*pb.SaleList, error)
	DeleteSale(in *entity.SaleID) (*pb.Message, error)
}

type ReturnedProductsRepo interface {
	CreateReturnedProducts() error
	UpdateReturnedProducts() error
	GetReturnedProducts() error
	GetReturnedProductsList() error
	DeleteReturnedProducts() error
}
