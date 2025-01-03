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
	CreateBulkProducts(in *pb.CreateBulkProductsRequest) (*pb.BulkCreateResponse, error)
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
	UpdatePurchase(in *pb.PurchaseUpdate) (*pb.PurchaseResponse, error)
	GetPurchase(in *pb.PurchaseID) (*pb.PurchaseResponse, error)
	GetPurchaseList(in *pb.FilterPurchase) (*pb.PurchaseList, error)
	DeletePurchase(in *pb.PurchaseID) (*pb.Message, error)
}

type SalesRepo interface {
	CreateSale(in *entity.SalesTotal) (*pb.SaleResponse, error)
	UpdateSale(in *pb.SaleUpdate) (*pb.SaleResponse, error)
	GetSale(in *pb.SaleID) (*pb.SaleResponse, error)
	GetSaleList(filter *pb.SaleFilter) (*pb.SaleList, error)
	DeleteSale(in *pb.SaleID) (*pb.Message, error)
	GetSalesByDay(request *pb.MostSoldProductsRequest) ([]*pb.DailySales, error)
	GetTopClients(req *pb.GetTopEntitiesRequest) ([]*pb.TopEntity, error)
	GetTopSuppliers(req *pb.GetTopEntitiesRequest) ([]*pb.TopEntity, error)
}

type StatisticsRepo interface {
	TotalPriceOfProducts(req *pb.StatisticReq) (*pb.PriceProducts, error)
	TotalSoldProducts(req *pb.StatisticReq) (*pb.PriceProducts, error)
	TotalPurchaseProducts(id *pb.StatisticReq) (*pb.PriceProducts, error)
}

type CashFlowRepo interface {
	CreateIncome(in *pb.CashFlowRequest) (*pb.CashFlow, error)
	CreateExpense(in *pb.CashFlowRequest) (*pb.CashFlow, error)
	Get(in *pb.StatisticReq) (*pb.ListCashFlow, error)
}

type ReturnedProductsRepo interface {
	CreateReturnedProducts() error
	UpdateReturnedProducts() error
	GetReturnedProducts() error
	GetReturnedProductsList() error
	DeleteReturnedProducts() error
}
