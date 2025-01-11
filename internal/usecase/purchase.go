package usecase

import (
	"crm-admin/internal/entity"
	pb "crm-admin/internal/generated/products"
	"fmt"
	"github.com/shopspring/decimal"
	"log"
	"log/slog"
	"math"
	"sync"
)

type PurchaseUseCase struct {
	repo    PurchasesRepo
	product ProductQuantity
	cash    CashFlowRepo // добавляем репозиторий для работы с cash_flow
	log     *slog.Logger
}

// NewPurchaseUseCase создает новый экземпляр PurchaseUseCase
func NewPurchaseUseCase(repo PurchasesRepo, pr ProductQuantity, log *slog.Logger, cash CashFlowRepo) *PurchaseUseCase {
	return &PurchaseUseCase{
		repo:    repo,
		product: pr,
		cash:    cash,
		log:     log,
	}
}

func (p *PurchaseUseCase) CalculateTotalPurchases(in *entity.Purchase) (*entity.PurchaseRequest, error) {

	if in == nil {
		return nil, fmt.Errorf("input purchase is nil")
	}

	if in.PurchaseItems == nil || len(in.PurchaseItems) == 0 {
		return nil, fmt.Errorf("purchase items are empty")
	}

	var result entity.PurchaseRequest
	totalSum := decimal.NewFromFloat(0)
	var purchaseList []entity.PurchaseItemReq

	for _, pr := range in.PurchaseItems {
		if pr.Quantity == 0 {
			p.log.Warn("Skipping item with zero quantity", "productID", pr.ProductID, "SupplierID", in.SupplierID)
			continue
		}

		quantity := decimal.NewFromFloat(float64(pr.Quantity))
		purchasePrice := decimal.NewFromFloat(float64(pr.PurchasePrice))
		totalPrice := quantity.Mul(purchasePrice)

		total, _ := totalPrice.Float64()

		purchase := entity.PurchaseItemReq{
			PurchasePrice: pr.PurchasePrice,
			ProductID:     pr.ProductID,
			Quantity:      pr.Quantity,
			TotalPrice:    total,
		}

		purchaseList = append(purchaseList, purchase)
		totalSum = totalSum.Add(totalPrice)
	}

	// Преобразуем totalSum в float64 и округляем до двух знаков после запятой
	totalCost := math.Round(totalSum.InexactFloat64()*100) / 100

	log.Println(totalCost)

	result = entity.PurchaseRequest{
		PurchasedBy:   in.PurchasedBy,
		SupplierID:    in.SupplierID,
		PurchaseItems: purchaseList,
		TotalCost:     totalCost, // Применяем округление
		PaymentMethod: in.PaymentMethod,
		Description:   in.Description,
		CompanyID:     in.CompanyID,
		BranchID:      in.BranchID,
	}

	return &result, nil
}

func (p *PurchaseUseCase) CreatePurchase(in *entity.Purchase) (*pb.PurchaseResponse, error) {
	req, err := p.CalculateTotalPurchases(in)
	if err != nil {
		p.log.Error("Failed to calculate total purchase cost", "error", err)
		return nil, fmt.Errorf("error calculating total purchase cost: %w", err)
	}

	// Создаем покупку в репозитории
	res, err := p.repo.CreatePurchase(req)
	if err != nil {
		p.log.Error("Failed to create purchase", "error", err)
		return nil, fmt.Errorf("error creating purchase: %w", err)
	}

	// Добавляем запись в cash flow для покупки
	cashFlowRequest := &pb.CashFlowRequest{
		UserId:        in.PurchasedBy,
		Amount:        req.TotalCost,
		Description:   fmt.Sprintf("Purchase from supplier %v", in.SupplierID),
		PaymentMethod: in.PaymentMethod,
		CompanyId:     in.CompanyID,
	}

	_, err = p.cash.CreateExpense(cashFlowRequest) // предполагаем наличие метода Create в CashFlowRepo
	if err != nil {
		p.log.Error("Failed to create cash flow entry for purchase", "error", err)
		return nil, fmt.Errorf("error creating cash flow entry: %w", err)
	}

	// Обрабатываем товары, добавляем их в инвентарь
	var wg sync.WaitGroup
	semaphore := make(chan struct{}, 10) // ограничиваем до 10 горутин

	for _, item := range in.PurchaseItems {
		wg.Add(1)
		go func(item entity.PurchaseItem) {
			defer wg.Done()
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			productQuantityReq := &entity.CountProductReq{
				ID:    item.ProductID,
				Count: item.Quantity,
			}

			if _, err := p.product.AddProduct(productQuantityReq); err != nil {
				p.log.Error("Failed to add product quantity", "productID", item.ProductID, "error", err)
			}
		}(item)
	}

	wg.Wait()
	return res, nil
}

// UpdatePurchase обновляет данные покупки
func (p *PurchaseUseCase) UpdatePurchase(in *pb.PurchaseUpdate) (*pb.PurchaseResponse, error) {
	if in == nil {
		return nil, fmt.Errorf("update request is nil")
	}

	res, err := p.repo.UpdatePurchase(in)
	if err != nil {
		p.log.Error("Failed to update purchase", "error", err)
		return nil, fmt.Errorf("error updating purchase: %w", err)
	}

	return res, nil
}

// GetPurchase получает данные о покупке по ID
func (p *PurchaseUseCase) GetPurchase(req *pb.PurchaseID) (*pb.PurchaseResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("purchase ID request is nil")
	}

	res, err := p.repo.GetPurchase(req)
	if err != nil {
		p.log.Error("Failed to fetch purchase data", "error", err)
		return nil, fmt.Errorf("error fetching purchase data: %w", err)
	}

	return res, nil
}

// GetListPurchase получает список покупок
func (p *PurchaseUseCase) GetListPurchase(req *pb.FilterPurchase) (*pb.PurchaseList, error) {
	if req == nil {
		return nil, fmt.Errorf("filter purchase request is nil")
	}

	res, err := p.repo.GetPurchaseList(req)
	if err != nil {
		p.log.Error("Failed to fetch purchase list", "error", err)
		return nil, fmt.Errorf("error fetching purchase list: %w", err)
	}

	return res, nil
}

// DeletePurchase удаляет покупку
func (p *PurchaseUseCase) DeletePurchase(req *pb.PurchaseID) (*pb.Message, error) {
	if req == nil {
		return nil, fmt.Errorf("purchase ID request is nil")
	}

	purchase, err := p.repo.GetPurchase(req)
	if err != nil {
		p.log.Error("Failed to fetch purchase data", "error", err)
		return nil, fmt.Errorf("error fetching purchase data: %w", err)
	}

	// Удаляем запись о cash flow для покупки
	cashFlowRequest := &pb.CashFlowRequest{
		UserId:        purchase.PurchasedBy,
		Amount:        purchase.TotalCost,
		Description:   fmt.Sprintf("Refund for purchase ID %v", req.Id),
		PaymentMethod: purchase.PaymentMethod,
		CompanyId:     purchase.CompanyId,
	}

	_, err = p.cash.CreateIncome(cashFlowRequest) // предполагаем наличие метода Delete в CashFlowRepo
	if err != nil {
		p.log.Error("Failed to delete cash flow entry", "error", err)
		return nil, fmt.Errorf("error deleting cash flow entry: %w", err)
	}

	// Удаляем товары из инвентаря
	var wg sync.WaitGroup
	semaphore := make(chan struct{}, 10)

	for _, item := range purchase.Items {
		wg.Add(1)
		go func(item pb.PurchaseItemResponse) {
			defer wg.Done()
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			productQuantityReq := &entity.CountProductReq{
				ID:    item.ProductId,
				Count: int(item.Quantity),
			}

			if _, err := p.product.RemoveProduct(productQuantityReq); err != nil {
				p.log.Error("Failed to remove product quantity", "productID", item.ProductId, "error", err)
			}
		}(*item)
	}

	wg.Wait()

	res, err := p.repo.DeletePurchase(req)
	if err != nil {
		p.log.Error("Failed to delete purchase", "error", err)
		return nil, fmt.Errorf("error deleting purchase: %w", err)
	}

	return res, nil
}

// validatePurchaseItems проверяет корректность элементов покупки
func (p *PurchaseUseCase) validatePurchaseItems(purchase *pb.PurchaseResponse) error {
	for _, item := range purchase.Items {
		if item.Quantity == 0 {
			return fmt.Errorf("item quantity cannot be zero for product %v", item.ProductId)
		}

		productQuantityReq := &entity.CountProductReq{
			ID:    item.ProductId,
			Count: int(item.Quantity),
		}

		check, err := p.product.ProductCountChecker(productQuantityReq)
		if err != nil {
			p.log.Error("Failed to check product quantity", "productID", item.ProductId, "error", err)
			return fmt.Errorf("error checking product quantity: %w", err)
		}

		if !check {
			return fmt.Errorf("insufficient product quantity to proceed with deletion")
		}
	}
	return nil
}
