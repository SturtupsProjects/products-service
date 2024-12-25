package grpc

import (
	"context"
	"crm-admin/internal/entity"
	pb "crm-admin/internal/generated/products"
	"fmt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"log"
)

// CalculateTotalSales calculates the total sale price from the sale request.
func (p *ProductsGrpc) CalculateTotalSales(ctx context.Context, in *pb.SaleRequest) (*pb.SaleResponse, error) {
	// Map the incoming gRPC SaleRequest to entity SaleRequest
	saleReq := &entity.SaleRequest{
		ClientID:      in.GetClientId(),
		SoldBy:        in.GetSoldBy(),
		PaymentMethod: in.GetPaymentMethod(),
	}

	// Map SaleItems from pb to entity
	var soldProducts []entity.SalesItem
	for _, item := range in.GetSoldProducts() {
		soldProducts = append(soldProducts, entity.SalesItem{
			ProductID: item.GetProductId(),
			Quantity:  int64(item.GetQuantity()),
			SalePrice: item.GetSalePrice(),
		})
	}
	saleReq.SoldProducts = soldProducts

	// Calculate the total sale price
	salesTotal, err := p.sales.CalculateTotalSales(saleReq)
	if err != nil {
		p.log.Error("Failed to calculate total sale price", "error", err.Error())
		return nil, status.Errorf(codes.Internal, "Failed to calculate total sale price: %v", err)
	}

	// Map the SalesTotal entity back to a gRPC SaleResponse
	return mapSalesTotalToSaleResponse(salesTotal), nil
}

// CreateSales creates a sale record.
func (p *ProductsGrpc) CreateSales(ctx context.Context, in *pb.SaleRequest) (*pb.SaleResponse, error) {
	// Map incoming gRPC request to entity struct
	saleReq := &entity.SaleRequest{
		ClientID:      in.GetClientId(),
		SoldBy:        in.GetSoldBy(),
		PaymentMethod: in.GetPaymentMethod(),
		CompanyID:     in.GetCompanyId(),
	}

	// Map SaleItems
	var soldProducts []entity.SalesItem
	for _, item := range in.GetSoldProducts() {
		soldProducts = append(soldProducts, entity.SalesItem{
			ProductID: item.GetProductId(),
			Quantity:  int64(item.GetQuantity()),
			SalePrice: item.GetSalePrice(),
		})
	}
	saleReq.SoldProducts = soldProducts

	// Create sale
	saleResp, err := p.sales.CreateSales(saleReq)
	if err != nil {
		p.log.Error("Failed to create sale", "error", err.Error())
		return nil, status.Errorf(codes.Internal, "Failed to create sale: %v", err)
	}

	return saleResp, nil
}

// UpdateSales updates the details of an existing sale.
func (p *ProductsGrpc) UpdateSales(ctx context.Context, in *pb.SaleUpdate) (*pb.SaleResponse, error) {

	saleResp, err := p.sales.UpdateSales(in)
	if err != nil {
		p.log.Error("Failed to update sale", "error", err.Error())
		return nil, status.Errorf(codes.Internal, "Failed to update sale: %v", err)
	}

	return saleResp, nil
}

// GetSales retrieves a specific sale by its ID.
func (p *ProductsGrpc) GetSales(ctx context.Context, in *pb.SaleID) (*pb.SaleResponse, error) {

	saleResp, err := p.sales.GetSales(in)
	if err != nil {
		p.log.Error("Failed to retrieve sale", "error", err.Error())
		return nil, status.Errorf(codes.NotFound, "Sale not found: %v", err)
	}

	return saleResp, nil
}

// GetListSales retrieves a list of sales based on filter criteria.
func (p *ProductsGrpc) GetListSales(ctx context.Context, in *pb.SaleFilter) (*pb.SaleList, error) {

	salesList, err := p.sales.GetListSales(in)
	if err != nil {
		p.log.Error("Failed to retrieve sales list", "error", err.Error())
		return nil, status.Errorf(codes.Internal, "Failed to retrieve sales list: %v", err)
	}

	return salesList, nil
}

// DeleteSales deletes a sale record.
func (p *ProductsGrpc) DeleteSales(ctx context.Context, in *pb.SaleID) (*pb.Message, error) {

	message, err := p.sales.DeleteSales(in)
	if err != nil {
		p.log.Error("Failed to delete sale", "error", err.Error())
		return nil, status.Errorf(codes.Internal, "Failed to delete sale: %v", err)
	}

	return message, nil
}

// GetMostSoldProductsByDay retrieves the most sold products by day.
func (p *ProductsGrpc) GetMostSoldProductsByDay(ctx context.Context, in *pb.MostSoldProductsRequest) (*pb.MostSoldProductsResponse, error) {

	log.Println("Mana keldi")
	fmt.Println("Mana keldi")

	res, err := p.sales.GetMostSoldProductsByDay(in)
	if err != nil {
		p.log.Error("Failed to get most sold products by day", "error", err.Error())
		return nil, status.Errorf(codes.Internal, "Failed to get most sold products by day: %v", err)
	}
	return res, err
}

func (p *ProductsGrpc) GetTopClients(ctx context.Context, in *pb.GetTopEntitiesRequest) (*pb.GetTopEntitiesResponse, error) {

	res, err := p.sales.GetTopClients(in)
	if err != nil {
		p.log.Error("Failed to get top clients", "error", err.Error())
		return nil, status.Errorf(codes.Internal, "Failed to get top clients: %v", err)
	}

	return res, err
}
func (p *ProductsGrpc) GetTopSuppliers(ctx context.Context, in *pb.GetTopEntitiesRequest) (*pb.GetTopEntitiesResponse, error) {

	res, err := p.sales.GetTopSuppliers(in)
	if err != nil {
		p.log.Error("Failed to get top suppliers", "error", err.Error())
		return nil, status.Errorf(codes.Internal, "Failed to get top suppliers: %v", err)
	}

	return res, err
}

// Helper function to map SalesTotal entity to SaleResponse
func mapSalesTotalToSaleResponse(total *entity.SalesTotal) *pb.SaleResponse {
	var soldProducts []*pb.SalesItem
	for _, item := range total.SoldProducts {
		soldProducts = append(soldProducts, &pb.SalesItem{
			ProductId:  item.ProductID,
			Quantity:   int32(item.Quantity),
			SalePrice:  item.SalePrice,
			TotalPrice: item.TotalPrice,
		})
	}

	return &pb.SaleResponse{
		ClientId:       total.ClientID,
		SoldBy:         total.SoldBy,
		TotalSalePrice: total.TotalSalePrice,
		PaymentMethod:  total.PaymentMethod,
		SoldProducts:   soldProducts,
	}
}
