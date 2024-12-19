package grpc

import (
	"context"
	"crm-admin/internal/entity"
	pb "crm-admin/internal/generated/products"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// CreatePurchase creates a purchase.
func (p *ProductsGrpc) CreatePurchase(ctx context.Context, in *pb.PurchaseRequest) (*pb.PurchaseResponse, error) {
	// Map incoming gRPC request to entity struct
	purchaseReq := &entity.Purchase{
		SupplierID:    in.GetSupplierId(),
		PurchasedBy:   in.GetPurchasedBy(),
		Description:   in.GetDescription(),
		PaymentMethod: in.GetPaymentMethod(),
		PurchaseItem:  mapPbPurchaseItemToEntity(in.GetItems()), // Map items
	}

	// Create purchase via usecase
	purchase, err := p.purchase.CreatePurchase(purchaseReq)
	if err != nil {
		p.log.Error("Failed to create purchase", "error", err.Error())
		return nil, status.Errorf(codes.Internal, "Failed to create purchase: %v", err)
	}

	return purchase, nil
}

// GetPurchase retrieves a specific purchase.
func (p *ProductsGrpc) GetPurchase(ctx context.Context, in *pb.PurchaseID) (*pb.PurchaseResponse, error) {
	purchaseID := &entity.PurchaseID{
		ID: in.GetId(),
	}

	// Get purchase via usecase
	purchase, err := p.purchase.GetPurchase(purchaseID)
	if err != nil {
		p.log.Error("Failed to retrieve purchase", "error", err.Error())
		return nil, status.Errorf(codes.NotFound, "Purchase not found: %v", err)
	}

	return purchase, nil
}

// GetListPurchase retrieves a list of purchases.
func (p *ProductsGrpc) GetListPurchase(ctx context.Context, in *pb.FilterPurchase) (*pb.PurchaseList, error) {
	filter := &entity.FilterPurchase{
		ProductID:   in.GetProductId(),
		SupplierID:  in.GetSupplierId(),
		PurchasedBy: in.GetPurchasedBy(),
		CreatedAt:   in.GetCreatedAt(),
	}

	// Get list of purchases via usecase
	purchaseList, err := p.purchase.GetListPurchase(filter)
	if err != nil {
		p.log.Error("Failed to retrieve purchase list", "error", err.Error())
		return nil, status.Errorf(codes.Internal, "Failed to retrieve purchase list: %v", err)
	}

	return purchaseList, nil
}

// UpdatePurchase updates an existing purchase.
func (p *ProductsGrpc) UpdatePurchase(ctx context.Context, in *pb.PurchaseUpdate) (*pb.PurchaseResponse, error) {
	purchaseUpdate := &entity.PurchaseUpdate{
		ID:            in.GetId(),
		SupplierID:    in.GetSupplierId(),
		Description:   in.GetDescription(),
		PaymentMethod: in.GetPaymentMethod(),
	}

	// Update purchase via usecase
	purchase, err := p.purchase.UpdatePurchase(purchaseUpdate)
	if err != nil {
		p.log.Error("Failed to update purchase", "error", err.Error())
		return nil, status.Errorf(codes.Internal, "Failed to update purchase: %v", err)
	}

	return purchase, nil
}

// DeletePurchase deletes a purchase.
func (p *ProductsGrpc) DeletePurchase(ctx context.Context, in *pb.PurchaseID) (*pb.Message, error) {
	purchaseID := &entity.PurchaseID{
		ID: in.GetId(),
	}

	// Delete purchase via usecase
	message, err := p.purchase.DeletePurchase(purchaseID)
	if err != nil {
		p.log.Error("Failed to delete purchase", "error", err.Error())
		return nil, status.Errorf(codes.Internal, "Failed to delete purchase: %v", err)
	}

	return &pb.Message{Message: message.Message}, nil
}

// Helper function to map pb PurchaseItemRequest to entity PurchaseItem
func mapPbPurchaseItemToEntity(items []*pb.PurchaseItem) *[]entity.PurchaseItem {
	var purchaseItems []entity.PurchaseItem
	for _, item := range items {
		purchaseItems = append(purchaseItems, entity.PurchaseItem{
			ProductID:     item.GetProductId(),
			Quantity:      int(item.GetQuantity()),
			PurchasePrice: item.GetPurchasePrice(),
		})
	}
	return &purchaseItems // Return a pointer to the slice
}
