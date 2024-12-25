package usecase

import (
	"context"
	"crm-admin/internal/entity"
	pb "crm-admin/internal/generated/products"
	"errors"
	"fmt"
	"log/slog"
	"sync"
)

type SalesUseCase struct {
	repo    SalesRepo
	product ProductQuantity
	log     *slog.Logger
}

func NewSalesUseCase(repo SalesRepo, pr ProductQuantity, log *slog.Logger) *SalesUseCase {
	return &SalesUseCase{
		repo:    repo,
		product: pr,
		log:     log,
	}
}

// CalculateTotalSales calculates the total sale price from the sale request.
func (s *SalesUseCase) CalculateTotalSales(in *entity.SaleRequest) (*entity.SalesTotal, error) {
	var totalPrice int64
	var soldProducts []entity.SalesItem

	for _, item := range in.SoldProducts {
		if item.Quantity <= 0 || item.SalePrice < 0 {
			return nil, fmt.Errorf("invalid item data: quantity or sale price is non-positive")
		}

		totalPrice += item.Quantity * item.SalePrice
		soldProducts = append(soldProducts, entity.SalesItem{
			ProductID:  item.ProductID,
			Quantity:   item.Quantity,
			SalePrice:  item.SalePrice,
			TotalPrice: item.Quantity * item.SalePrice,
		})
	}

	return &entity.SalesTotal{
		ClientID:       in.ClientID,
		SoldBy:         in.SoldBy,
		TotalSalePrice: totalPrice,
		PaymentMethod:  in.PaymentMethod,
		SoldProducts:   soldProducts,
		CompanyID:      in.CompanyID,
	}, nil
}

func (s *SalesUseCase) CreateSales(in *entity.SaleRequest) (*pb.SaleResponse, error) {
	total, err := s.CalculateTotalSales(in)
	if err != nil {
		s.log.Error("Error calculating total sale cost", "error", err)
		return nil, fmt.Errorf("error calculating total sale cost: %w", err)
	}

	res, err := s.repo.CreateSale(total)
	if err != nil {
		s.log.Error("Error creating sale", "error", err)
		return nil, fmt.Errorf("error creating sale: %w", err)
	}

	var wg sync.WaitGroup
	semaphore := make(chan struct{}, 10)

	for _, item := range res.SoldProducts {
		wg.Add(1)
		go func(item pb.SalesItem) {
			defer wg.Done()
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			productQuantityReq := &entity.CountProductReq{
				ID:    item.ProductId,
				Count: int(item.Quantity),
			}

			if _, err := s.product.RemoveProduct(productQuantityReq); err != nil {
				s.log.Error("Error removing product quantity during sale", "productID", item.ProductId, "error", err)
			}
		}(*item)
	}

	wg.Wait()
	return res, nil
}

// UpdateSales updates an existing sale record.
func (s *SalesUseCase) UpdateSales(in *pb.SaleUpdate) (*pb.SaleResponse, error) {
	res, err := s.repo.UpdateSale(in)
	if err != nil {
		s.log.Error("Error updating sale", "error", err)
		return nil, fmt.Errorf("error updating sale: %w", err)
	}
	return res, nil
}

// GetSales retrieves a specific sale by ID.
func (s *SalesUseCase) GetSales(req *pb.SaleID) (*pb.SaleResponse, error) {
	res, err := s.repo.GetSale(req)
	if err != nil {
		s.log.Error("Error fetching sale", "saleID", req.Id, "error", err)
		return nil, fmt.Errorf("error fetching sale: %w", err)
	}
	return res, nil
}

// GetListSales retrieves a list of sales based on filters.
func (s *SalesUseCase) GetListSales(req *pb.SaleFilter) (*pb.SaleList, error) {
	res, err := s.repo.GetSaleList(req)
	if err != nil {
		s.log.Error("Error fetching sales list", "filter", req, "error", err)
		return nil, fmt.Errorf("error fetching sales list: %w", err)
	}
	return res, nil
}

// DeleteSales deletes a sale record and restores product stock.
func (s *SalesUseCase) DeleteSales(req *pb.SaleID) (*pb.Message, error) {
	sale, err := s.repo.GetSale(req)
	if err != nil {
		s.log.Error("Error fetching sale for deletion", "saleID", req.Id, "error", err)
		return nil, fmt.Errorf("error fetching sale for deletion: %w", err)
	}

	var wg sync.WaitGroup
	semaphore := make(chan struct{}, 10)

	for _, item := range sale.SoldProducts {
		wg.Add(1)
		go func(item pb.SalesItem) {
			defer wg.Done()
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			productQuantityReq := &entity.CountProductReq{
				ID:    item.ProductId,
				Count: int(item.Quantity),
			}

			if _, err := s.product.AddProduct(productQuantityReq); err != nil {
				s.log.Error("Error restoring product stock after sale deletion", "productID", item.ProductId, "error", err)
			}
		}(*item)
	}

	wg.Wait()

	res, err := s.repo.DeleteSale(req)
	if err != nil {
		s.log.Error("Error deleting sale", "saleID", req.Id, "error", err)
		return nil, fmt.Errorf("error deleting sale: %w", err)
	}

	return res, nil
}

// GetMostSoldProductsByDay retrieves the most sold products grouped by day.
func (s *SalesUseCase) GetMostSoldProductsByDay(req *pb.MostSoldProductsRequest) (*pb.MostSoldProductsResponse, error) {
	// Validate request
	if req.GetCompanyId() == "" {
		return nil, errors.New("company_id is required")
	}
	if req.GetStartDate() == "" || req.GetEndDate() == "" {
		return nil, errors.New("start_date and end_date are required")
	}

	// Call the repository to get sales data
	results, err := s.repo.GetSalesByDay(req)
	if err != nil {
		return nil, err
	}

	// Prepare the response
	response := &pb.MostSoldProductsResponse{
		DailySales: results,
	}

	return response, nil
}
func (s *SalesUseCase) GetTopClients(ctx context.Context, req *pb.GetTopEntitiesRequest) (*pb.GetTopEntitiesResponse, error) {
	if req.CompanyId == "" || req.StartDate == "" || req.EndDate == "" {
		return nil, errors.New("company_id, start_date, and end_date are required")
	}

	entities, err := s.repo.GetTopClients(req)
	if err != nil {
		return nil, err
	}

	return &pb.GetTopEntitiesResponse{Entities: entities}, nil
}

func (s *SalesUseCase) GetTopSuppliers(ctx context.Context, req *pb.GetTopEntitiesRequest) (*pb.GetTopEntitiesResponse, error) {
	if req.CompanyId == "" || req.StartDate == "" || req.EndDate == "" {
		return nil, errors.New("company_id, start_date, and end_date are required")
	}

	entities, err := s.repo.GetTopSuppliers(req)
	if err != nil {
		return nil, err
	}

	return &pb.GetTopEntitiesResponse{Entities: entities}, nil
}
