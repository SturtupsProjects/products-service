package grpc

import (
	"context"
	"crm-admin/internal/controller"
	"crm-admin/internal/entity"
	"crm-admin/internal/usecase"
	pb "crm-admin/pkg/gednerated/products"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"log/slog"
)

type ProductsGrpc struct {
	product  *usecase.ProductsUseCase
	purchase *usecase.PurchaseUseCase
	sales    *usecase.SalesUseCase
	log      *slog.Logger
	pb.UnimplementedProductsServer
}

func NewProductGrpc(ctrl *controller.Controller, log *slog.Logger) *ProductsGrpc {
	return &ProductsGrpc{
		product:  ctrl.Product,
		purchase: ctrl.Purchase,
		sales:    ctrl.Sales,
		log:      log,
	}
}

// ---------------------------------- Product Category --------------------------------------------------------------

func (p *ProductsGrpc) CreateCategory(ctx context.Context, in *pb.CreateCategoryRequest) (*pb.Category, error) {
	// Map gRPC request to internal use case structure
	categoryName := &entity.CategoryName{
		Name:      in.GetName(),
		CreatedBy: in.GetCreatedBy(),
	}

	// Call the usecase layer to create the category
	category, err := p.product.CreateCategory(categoryName)
	if err != nil {
		// Log the error for debugging
		p.log.Error("Failed to create category", "error", err.Error())

		// Return a gRPC status error with internal code
		return nil, status.Errorf(codes.Internal, "Failed to create category: %v", err)
	}

	return category, nil
}

func (p *ProductsGrpc) DeleteCategory(ctx context.Context, in *pb.GetCategoryRequest) (*pb.Message, error) {
	categoryID := &entity.CategoryID{
		ID: in.GetId(),
	}

	// Call usecase to delete the category
	message, err := p.product.DeleteCategory(categoryID)
	if err != nil {
		// Log the error for debugging
		p.log.Error("Failed to delete category", "error", err.Error())

		// Return a gRPC status error with internal code
		return nil, status.Errorf(codes.Internal, "Failed to delete category: %v", err)
	}

	return message, nil
}

func (p *ProductsGrpc) GetCategory(ctx context.Context, in *pb.GetCategoryRequest) (*pb.Category, error) {
	categoryID := &entity.CategoryID{
		ID: in.GetId(),
	}

	// Call usecase to retrieve category details
	category, err := p.product.GetCategory(categoryID)
	if err != nil {
		// Log the error for debugging
		p.log.Error("Failed to retrieve category", "error", err.Error())

		// Return a gRPC status error with not found code if no category found
		return nil, status.Errorf(codes.NotFound, "Category not found: %v", err)
	}

	return category, nil
}

func (p *ProductsGrpc) GetListCategory(ctx context.Context, in *emptypb.Empty) (*pb.CategoryList, error) {
	// Call usecase to get the list of categories
	categoryList, err := p.product.GetListCategory(nil) // Adjust if GetListCategory requires parameters
	if err != nil {
		// Log the error for debugging
		p.log.Error("Failed to retrieve category list", "error", err.Error())

		// Return a gRPC status error with internal code
		return nil, status.Errorf(codes.Internal, "Failed to retrieve category list: %v", err)
	}

	return categoryList, nil
}

// --------------------------------------- Products --------------------------------------------------------------

func (p *ProductsGrpc) CreateProduct(ctx context.Context, in *pb.CreateProductRequest) (*pb.Product, error) {
	productReq := &entity.ProductRequest{
		CategoryID:    in.GetCategoryId(),
		Name:          in.GetName(),
		BillFormat:    in.GetBillFormat(),
		IncomingPrice: in.GetIncomingPrice(),
		StandardPrice: in.GetStandardPrice(),
		CreatedBy:     in.GetCreatedBy(),
	}

	// Call usecase to create product
	product, err := p.product.CreateProduct(productReq)
	if err != nil {
		// Log the error for debugging
		p.log.Error("Failed to create product", "error", err.Error())

		// Return a gRPC status error with internal code
		return nil, status.Errorf(codes.Internal, "Failed to create product: %v", err)
	}

	return product, nil
}

func (p *ProductsGrpc) UpdateProduct(ctx context.Context, in *pb.UpdateProductRequest) (*pb.Product, error) {
	productUpdate := &entity.ProductUpdate{
		ID:            in.GetId(),
		CategoryID:    in.GetCategoryId(),
		Name:          in.GetName(),
		BillFormat:    in.GetBillFormat(),
		IncomingPrice: in.GetIncomingPrice(),
		StandardPrice: in.GetStandardPrice(),
	}

	// Call usecase to update the product
	product, err := p.product.UpdateProduct(productUpdate)
	if err != nil {
		// Log the error for debugging
		p.log.Error("Failed to update product", "error", err.Error())

		// Return a gRPC status error with internal code
		return nil, status.Errorf(codes.Internal, "Failed to update product: %v", err)
	}

	return product, nil
}

func (p *ProductsGrpc) DeleteProduct(ctx context.Context, in *pb.GetProductRequest) (*pb.Message, error) {
	productID := &entity.ProductID{
		ID: in.GetId(),
	}

	// Call usecase to delete the product
	message, err := p.product.DeleteProduct(productID)
	if err != nil {
		// Log the error for debugging
		p.log.Error("Failed to delete product", "error", err.Error())

		// Return a gRPC status error with internal code
		return nil, status.Errorf(codes.Internal, "Failed to delete product: %v", err)
	}

	return message, nil
}

func (p *ProductsGrpc) GetProduct(ctx context.Context, in *pb.GetProductRequest) (*pb.Product, error) {
	productID := &entity.ProductID{
		ID: in.GetId(),
	}

	// Call usecase to retrieve product details
	product, err := p.product.GetProduct(productID)
	if err != nil {
		// Log the error for debugging
		p.log.Error("Failed to retrieve product", "error", err.Error())

		// Return a gRPC status error with not found code if product not found
		return nil, status.Errorf(codes.NotFound, "Product not found: %v", err)
	}

	return product, nil
}

// Continuing from GetProductList method

func (p *ProductsGrpc) GetProductList(ctx context.Context, in *pb.ProductFilter) (*pb.ProductList, error) {
	filter := &entity.FilterProduct{
		CategoryId: in.GetCategoryId(),
		Name:       in.GetName(),
		CreatedBy:  in.GetCreatedBy(),
	}

	// Call the usecase to retrieve the filtered product list
	productList, err := p.product.GetProductList(filter)
	if err != nil {
		// Log the error for debugging
		p.log.Error("Failed to retrieve product list", "error", err.Error())

		// Return a gRPC status error with internal code
		return nil, status.Errorf(codes.Internal, "Failed to retrieve product list: %v", err)
	}

	return productList, nil
}
