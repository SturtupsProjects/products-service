package grpc

import (
	"context"
	"crm-admin/internal/controller"
	pb "crm-admin/internal/generated/products"
	"crm-admin/internal/usecase"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ProductsGrpc struct {
	statistics usecase.StatisticsRepo
	cashFlow   usecase.CashFlowRepo
	product    *usecase.ProductsUseCase
	purchase   *usecase.PurchaseUseCase
	sales      *usecase.SalesUseCase

	pb.UnimplementedProductsServer
}

func NewProductGrpc(ctrl *controller.Controller, statistics usecase.StatisticsRepo, cash usecase.CashFlowRepo) *ProductsGrpc {
	return &ProductsGrpc{
		product:    ctrl.Product,
		statistics: statistics,
		purchase:   ctrl.Purchase,
		sales:      ctrl.Sales,
		cashFlow:   cash,
	}
}

// ---------------------------------- Product Category --------------------------------------------------------------

func (p *ProductsGrpc) CreateCategory(ctx context.Context, in *pb.CreateCategoryRequest) (*pb.Category, error) {

	category, err := p.product.CreateCategory(in)

	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to create category: %v", err)
	}

	return category, nil
}

func (p *ProductsGrpc) UpdateCategory(ctx context.Context, in *pb.UpdateCategoryRequest) (*pb.Category, error) {

	category, err := p.product.UpdateCategory(in)

	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to update category: %v", err)
	}

	return category, nil
}

func (p *ProductsGrpc) DeleteCategory(ctx context.Context, in *pb.GetCategoryRequest) (*pb.Message, error) {

	message, err := p.product.DeleteCategory(in)

	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to delete category: %v", err)
	}

	return message, nil
}

func (p *ProductsGrpc) GetCategory(ctx context.Context, in *pb.GetCategoryRequest) (*pb.Category, error) {

	category, err := p.product.GetCategory(in)

	if err != nil {
		return nil, status.Errorf(codes.NotFound, "Category not found: %v", err)
	}

	return category, nil
}

func (p *ProductsGrpc) GetListCategory(ctx context.Context, in *pb.CategoryName) (*pb.CategoryList, error) {

	categoryList, err := p.product.GetListCategory(in)

	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to retrieve category list: %v", err)
	}

	return categoryList, nil
}

// --------------------------------------- Products --------------------------------------------------------------

func (p *ProductsGrpc) CreateProduct(ctx context.Context, in *pb.CreateProductRequest) (*pb.Product, error) {

	product, err := p.product.CreateProduct(in)

	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to create product: %v", err)
	}

	return product, nil
}
func (p *ProductsGrpc) CreateBulkProducts(ctx context.Context, in *pb.CreateBulkProductsRequest) (*pb.BulkCreateResponse, error) {
	products, err := p.product.CreateBulkProducts(in)
	if err != nil {
		return nil, err
	}
	return products, nil
}

func (p *ProductsGrpc) UpdateProduct(ctx context.Context, in *pb.UpdateProductRequest) (*pb.Product, error) {

	product, err := p.product.UpdateProduct(in)

	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to update product: %v", err)
	}

	return product, nil
}

func (p *ProductsGrpc) DeleteProduct(ctx context.Context, in *pb.GetProductRequest) (*pb.Message, error) {

	message, err := p.product.DeleteProduct(in)

	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to delete product: %v", err)
	}

	return message, nil
}

func (p *ProductsGrpc) GetProduct(ctx context.Context, in *pb.GetProductRequest) (*pb.Product, error) {

	product, err := p.product.GetProduct(in)

	if err != nil {
		return nil, status.Errorf(codes.NotFound, "Product not found: %v", err)
	}

	return product, nil
}

func (p *ProductsGrpc) GetProductList(ctx context.Context, in *pb.ProductFilter) (*pb.ProductList, error) {

	productList, err := p.product.GetProductList(in)

	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to retrieve product list: %v", err)
	}

	return productList, nil
}

func (p *ProductsGrpc) GetProductDashboard(ctx context.Context, in *pb.GetProductsDashboardReq) (*pb.GetProductsDashboardRes, error) {

	res, err := p.product.GetProductDashboard(in)

	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to retrieve product dashboard: %v", err)
	}

	return res, nil
}
