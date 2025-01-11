package usecase

import (
	"crm-admin/internal/entity"
	pb "crm-admin/internal/generated/products"
	"errors"
	"fmt"
	"github.com/shopspring/decimal"
	"log"
	"log/slog"
	"math"
	"sync"
)

type SalesUseCase struct {
	repo    SalesRepo
	product ProductQuantity
	cash    CashFlowRepo
	log     *slog.Logger
}

func NewSalesUseCase(repo SalesRepo, pr ProductQuantity, log *slog.Logger, cash CashFlowRepo) *SalesUseCase {
	return &SalesUseCase{
		repo:    repo,
		product: pr,
		cash:    cash,
		log:     log,
	}
}

// CalculateTotalSales calculates the total sale price from the sale request.
func (s *SalesUseCase) CalculateTotalSales(in *entity.SaleRequest) (*entity.SalesTotal, error) {
	if in == nil {
		return nil, errors.New("input sale request is nil")
	}

	if in.SoldProducts == nil || len(in.SoldProducts) == 0 {
		return nil, errors.New("sold products list is empty")
	}

	var totalPrice decimal.Decimal
	var soldProducts []entity.SalesItem

	for _, item := range in.SoldProducts {
		if item.Quantity <= 0 {
			s.log.Warn("Skipping item with non-positive quantity", "ProductID", item.ProductID)
			continue
		}
		if item.SalePrice < 0 {
			return nil, fmt.Errorf("invalid sale price for product %v: cannot be negative", item.ProductID)
		}

		quantity := decimal.NewFromFloat(float64(item.Quantity))
		salePrice := decimal.NewFromFloat(item.SalePrice)
		totalItemPrice := quantity.Mul(salePrice)

		totalPrice = totalPrice.Add(totalItemPrice)

		// Преобразуем и округляем itemTotalPrice до двух знаков после запятой
		itemTotalPriceRounded := math.Round(totalItemPrice.InexactFloat64()*100) / 100

		soldProducts = append(soldProducts, entity.SalesItem{
			ProductID:  item.ProductID,
			Quantity:   item.Quantity,
			SalePrice:  item.SalePrice,
			TotalPrice: itemTotalPriceRounded, // Округленная цена
		})
	}

	// Преобразуем и округляем totalPrice до двух знаков после запятой
	totalSalePrice := math.Round(totalPrice.InexactFloat64()*100) / 100

	return &entity.SalesTotal{
		ClientID:       in.ClientID,
		SoldBy:         in.SoldBy,
		TotalSalePrice: totalSalePrice, // Округленная итоговая сумма
		PaymentMethod:  in.PaymentMethod,
		SoldProducts:   soldProducts,
		CompanyID:      in.CompanyID,
		BranchID:       in.BranchID,
	}, nil
}

// CreateSales creates a new sale record and a cash flow record for the sale.
func (s *SalesUseCase) CreateSales(in *entity.SaleRequest) (*pb.SaleResponse, error) {
	total, err := s.CalculateTotalSales(in)
	if err != nil {
		s.log.Error("Error calculating total sale cost", "error", err)
		return nil, fmt.Errorf("error calculating total sale cost: %w", err)
	}

	// Create the sale record
	res, err := s.repo.CreateSale(total)
	if err != nil {
		s.log.Error("Error creating sale", "error", err)
		return nil, fmt.Errorf("error creating sale: %w", err)
	}

	// Create a cash flow record for this sale
	cashFlowRequest := &pb.CashFlowRequest{
		UserId:        in.SoldBy,
		Amount:        total.TotalSalePrice,
		Description:   fmt.Sprintf("Sale for client %s", in.ClientID),
		PaymentMethod: in.PaymentMethod,
		CompanyId:     in.CompanyID,
	}

	// Add to cash flow
	cashFlow, err := s.cash.CreateIncome(cashFlowRequest)
	if err != nil {
		s.log.Error("Error creating cash flow", "error", err)
		return nil, fmt.Errorf("error creating cash flow: %w", err)
	}

	// Log the successful creation of cash flow
	s.log.Info("Created cash flow record", "cashFlowID", cashFlow.Id)

	// Update stock quantities concurrently
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
	if in == nil {
		return nil, errors.New("input sale update is nil")
	}

	res, err := s.repo.UpdateSale(in)
	if err != nil {
		s.log.Error("Error updating sale", "error", err)
		return nil, fmt.Errorf("error updating sale: %w", err)
	}
	return res, nil
}

// GetSales retrieves a specific sale by ID.
func (s *SalesUseCase) GetSales(req *pb.SaleID) (*pb.SaleResponse, error) {
	if req == nil {
		return nil, errors.New("sale ID request is nil")
	}

	res, err := s.repo.GetSale(req)
	if err != nil {
		s.log.Error("Error fetching sale", "saleID", req.Id, "error", err)
		return nil, fmt.Errorf("error fetching sale: %w", err)
	}
	return res, nil
}

// GetListSales retrieves a list of sales based on filters.
func (s *SalesUseCase) GetListSales(req *pb.SaleFilter) (*pb.SaleList, error) {
	if req == nil {
		return nil, errors.New("sale filter request is nil")
	}

	res, err := s.repo.GetSaleList(req)
	if err != nil {
		s.log.Error("Error fetching sales list", "filter", req, "error", err)
		return nil, fmt.Errorf("error fetching sales list: %w", err)
	}
	return res, nil
}

// DeleteSales deletes a sale record and restores product stock.
// DeleteSales deletes a sale record and restores product stock, as well as removes the corresponding cash flow entry.
func (s *SalesUseCase) DeleteSales(req *pb.SaleID) (*pb.Message, error) {
	if req == nil {
		return nil, errors.New("sale ID request is nil")
	}

	// Fetch the sale to be deleted
	sale, err := s.repo.GetSale(req)
	if err != nil {
		s.log.Error("Error fetching sale for deletion", "saleID", req.Id, "error", err)
		return nil, fmt.Errorf("error fetching sale for deletion: %w", err)
	}

	// Restore the product stock
	var wg sync.WaitGroup
	semaphore := make(chan struct{}, 10)

	for _, item := range sale.SoldProducts {
		wg.Add(1)
		go func(item pb.SalesItem) {
			defer wg.Done()
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			// Restore product stock
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

	// Delete the sale record from the database
	res, err := s.repo.DeleteSale(req)
	if err != nil {
		s.log.Error("Error deleting sale", "saleID", req.Id, "error", err)
		return nil, fmt.Errorf("error deleting sale: %w", err)
	}

	// Now, handle the cash flow for the sale
	// We assume that the cash flow related to this sale needs to be reversed or deleted.
	cashFlowRequest := &pb.CashFlowRequest{
		UserId:        sale.SoldBy,
		Amount:        sale.TotalSalePrice,
		Description:   fmt.Sprintf("Refund for sale ID %s", req.Id),
		PaymentMethod: sale.PaymentMethod,
		CompanyId:     req.CompanyId,
	}

	// Delete the related cash flow entry
	_, err = s.cash.CreateExpense(cashFlowRequest) // Assuming you have a Delete method in CashFlowRepo
	if err != nil {
		s.log.Error("Error deleting cash flow for the sale", "saleID", req.Id, "error", err)
		return nil, fmt.Errorf("error deleting cash flow: %w", err)
	}

	// Log the successful deletion of sale and cash flow
	s.log.Info("Successfully deleted sale and corresponding cash flow", "saleID", req.Id)

	return res, nil
}

// GetMostSoldProductsByDay retrieves the most sold products grouped by day.
func (s *SalesUseCase) GetMostSoldProductsByDay(req *pb.MostSoldProductsRequest) (*pb.MostSoldProductsResponse, error) {

	log.Println(req.CompanyId)

	if req.GetCompanyId() == "" {
		return nil, errors.New("company_id is required")
	}
	if req.GetStartDate() == "" || req.GetEndDate() == "" {
		return nil, errors.New("start_date and end_date are required")
	}

	results, err := s.repo.GetSalesByDay(req)
	if err != nil {
		return nil, err
	}

	response := &pb.MostSoldProductsResponse{
		DailySales: results,
	}

	return response, nil
}
func (s *SalesUseCase) GetTopClients(req *pb.GetTopEntitiesRequest) (*pb.GetTopEntitiesResponse, error) {
	if req.CompanyId == "" || req.StartDate == "" || req.EndDate == "" {
		return nil, errors.New("company_id, start_date, and end_date are required")
	}

	entities, err := s.repo.GetTopClients(req)
	if err != nil {
		return nil, err
	}

	return &pb.GetTopEntitiesResponse{Entities: entities}, nil
}

func (s *SalesUseCase) GetTopSuppliers(req *pb.GetTopEntitiesRequest) (*pb.GetTopEntitiesResponse, error) {
	if req.CompanyId == "" || req.StartDate == "" || req.EndDate == "" {
		return nil, errors.New("company_id, start_date, and end_date are required")
	}

	entities, err := s.repo.GetTopSuppliers(req)
	if err != nil {
		return nil, err
	}

	return &pb.GetTopEntitiesResponse{Entities: entities}, nil
}
