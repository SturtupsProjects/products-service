package usecase

import (
	"crm-admin/internal/entity"
	pb "crm-admin/internal/generated/products"
	"fmt"
	"github.com/shopspring/decimal"
	"log/slog"
	"sync"
)

type PurchaseUseCase struct {
	repo    PurchasesRepo
	product ProductQuantity
	log     *slog.Logger
}

// NewPurchaseUseCase создает новый экземпляр PurchaseUseCase
func NewPurchaseUseCase(repo PurchasesRepo, pr ProductQuantity, log *slog.Logger) *PurchaseUseCase {
	return &PurchaseUseCase{
		repo:    repo,
		product: pr,
		log:     log,
	}
}

// CalculateTotalPurchases вычисляет общую стоимость покупки
func (p *PurchaseUseCase) CalculateTotalPurchases(in *entity.Purchase) (*entity.PurchaseRequest, error) {
	if in == nil {
		return nil, fmt.Errorf("input purchase is nil")
	}

	var result entity.PurchaseRequest
	totalSum := decimal.NewFromFloat(0)
	var purchaseList []entity.PurchaseItemReq

	for _, pr := range in.PurchaseItems {
		if pr.Quantity == 0 {
			p.log.Warn("Skipping item with zero quantity", "productID", pr.ProductID)
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

	total, _ := totalSum.Float64()

	result = entity.PurchaseRequest{
		PurchasedBy:   in.PurchasedBy,
		SupplierID:    in.SupplierID,
		PurchaseItems: purchaseList,
		TotalCost:     total, // Преобразуем итоговую сумму в int64
		PaymentMethod: in.PaymentMethod,
		Description:   in.Description,
		CompanyID:     in.CompanyID,
	}

	return &result, nil
}

// CreatePurchase создает новую покупку
func (p *PurchaseUseCase) CreatePurchase(in *entity.Purchase) (*pb.PurchaseResponse, error) {
	req, err := p.CalculateTotalPurchases(in)
	if err != nil {
		p.log.Error("Failed to calculate total purchase cost", "error", err)
		return nil, fmt.Errorf("error calculating total purchase cost: %w", err)
	}

	res, err := p.repo.CreatePurchase(req)
	if err != nil {
		p.log.Error("Failed to create purchase", "error", err)
		return nil, fmt.Errorf("error creating purchase: %w", err)
	}

	var wg sync.WaitGroup
	semaphore := make(chan struct{}, 10) // ограничиваем до 10 goroutine

	for _, item := range in.PurchaseItems {
		wg.Add(1)
		go func(item entity.PurchaseItem) {
			defer wg.Done()

			// Контролируем количество goroutine
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

	if err := p.validatePurchaseItems(purchase); err != nil {
		p.log.Error("Validation failed before deletion", "error", err)
		return nil, err
	}

	res, err := p.repo.DeletePurchase(req)
	if err != nil {
		p.log.Error("Failed to delete purchase", "error", err)
		return nil, fmt.Errorf("error deleting purchase: %w", err)
	}

	go func() {
		for _, item := range purchase.Items {
			productQuantityReq := &entity.CountProductReq{
				ID:    item.ProductId,
				Count: int(item.Quantity),
			}

			if _, err := p.product.RemoveProduct(productQuantityReq); err != nil {
				p.log.Error("Failed to remove product in DeletePurchase", "productID", item.ProductId, "error", err)
			}
		}
	}()

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
