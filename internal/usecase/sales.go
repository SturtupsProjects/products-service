package usecase

import (
	"crm-admin/internal/entity"
	pb "crm-admin/internal/generated/products"
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

	// Calculate total price and fill sold products
	for _, item := range in.SoldProducts {
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
	}, nil
}

// CreateSales creates a sale record.
func (s *SalesUseCase) CreateSales(in *entity.SaleRequest) (*pb.SaleResponse, error) {
	// Calculate total sale cost
	total, err := s.CalculateTotalSales(in)
	if err != nil {
		s.log.Error("Error calculating total sale cost", "error", err.Error())
		return nil, fmt.Errorf("error calculating total sale cost: %w", err)
	}

	// Create sale in the database
	res, err := s.repo.CreateSale(total)
	if err != nil {
		s.log.Error("Error creating sale", "error", err.Error())
		return nil, fmt.Errorf("error creating sale: %w", err)
	}

	// Update product stock
	var wg sync.WaitGroup
	semaphore := make(chan struct{}, 10) // limit to 10 goroutines for product updates

	for _, item := range res.SoldProducts {
		wg.Add(1)
		go func(item *pb.SalesItem) {
			defer wg.Done()
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			productQuantityReq := &entity.CountProductReq{
				Id:    item.ProductId,
				Count: int(item.Quantity),
			}
			if _, err := s.product.RemoveProduct(productQuantityReq); err != nil {
				s.log.Error("Error removing product quantity during sale", "error", err.Error())
			}
		}(item)
	}

	wg.Wait()

	return res, nil
}

// UpdateSales updates the details of an existing sale.
func (s *SalesUseCase) UpdateSales(in *entity.SaleUpdate) (*pb.SaleResponse, error) {
	// Update sale in the database
	res, err := s.repo.UpdateSale(in)
	if err != nil {
		s.log.Error("Error updating sale", "error", err.Error())
		return nil, fmt.Errorf("error updating sale: %w", err)
	}

	return res, nil
}

// GetSales retrieves a specific sale based on the ID.
func (s *SalesUseCase) GetSales(req *entity.SaleID) (*pb.SaleResponse, error) {
	// Fetch sale from the database
	res, err := s.repo.GetSale(req)
	if err != nil {
		s.log.Error("Error fetching sale", "error", err.Error())
		return nil, fmt.Errorf("error fetching sale: %w", err)
	}

	return res, nil
}

// GetListSales retrieves a list of sales based on filter criteria.
func (s *SalesUseCase) GetListSales(req *entity.SaleFilter) (*pb.SaleList, error) {
	// Fetch sales list from the database
	res, err := s.repo.GetSaleList(req)
	if err != nil {
		s.log.Error("Error fetching sales list", "error", err.Error())
		return nil, fmt.Errorf("error fetching sales list: %w", err)
	}

	return res, nil
}

// DeleteSales deletes a sale record from the system.
func (s *SalesUseCase) DeleteSales(req *entity.SaleID) (*pb.Message, error) {
	// Fetch the sale to be deleted
	sale, err := s.repo.GetSale(req)
	if err != nil {
		s.log.Error("Error fetching sale for deletion", "error", err.Error())
		return nil, fmt.Errorf("error fetching sale for deletion: %w", err)
	}

	// Add logic to restore product stock if needed
	var wg sync.WaitGroup
	semaphore := make(chan struct{}, 10)

	for _, item := range sale.SoldProducts {
		wg.Add(1)
		go func(item *pb.SalesItem) {
			defer wg.Done()
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			productQuantityReq := &entity.CountProductReq{
				Id:    item.ProductId,
				Count: int(item.Quantity),
			}

			if _, err := s.product.AddProduct(productQuantityReq); err != nil {
				s.log.Error("Error restoring product stock after sale deletion", "error", err.Error())
			}
		}(item)
	}

	wg.Wait()

	// Delete the sale from the database
	res, err := s.repo.DeleteSale(req)
	if err != nil {
		s.log.Error("Error deleting sale", "error", err.Error())
		return nil, fmt.Errorf("error deleting sale: %w", err)
	}

	return res, nil
}
