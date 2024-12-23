package usecase

import (
	"crm-admin/internal/entity"
	pb "crm-admin/internal/generated/products"
	"fmt"
	"log/slog"
	"sync"
)

type PurchaseUseCase struct {
	repo    PurchasesRepo
	product ProductQuantity
	log     *slog.Logger
}

func NewPurchaseUseCase(repo PurchasesRepo, pr ProductQuantity, log *slog.Logger) *PurchaseUseCase {
	return &PurchaseUseCase{
		repo:    repo,
		product: pr,
		log:     log,
	}
}

func (p *PurchaseUseCase) CalculateTotalPurchases(in *entity.Purchase) (*entity.PurchaseRequest, error) {
	var result entity.PurchaseRequest
	var totalSum int64
	var purchaseList []entity.PurchaseItemReq

	for _, pr := range in.PurchaseItems {
		if pr.Quantity == 0 {
			continue
		}

		purchase := entity.PurchaseItemReq{
			PurchasePrice: pr.PurchasePrice,
			ProductID:     pr.ProductID,
			Quantity:      pr.Quantity,
			TotalPrice:    int64(pr.Quantity) * pr.PurchasePrice,
		}

		purchaseList = append(purchaseList, purchase)
		totalSum += purchase.TotalPrice
	}

	result.PurchasedBy = in.PurchasedBy
	result.SupplierID = in.SupplierID
	result.PurchaseItems = purchaseList
	result.TotalCost = totalSum
	result.PaymentMethod = in.PaymentMethod
	result.Description = in.Description
	result.CompanyID = in.CompanyID

	return &result, nil
}

func (p *PurchaseUseCase) CreatePurchase(in *entity.Purchase) (*pb.PurchaseResponse, error) {
	req, err := p.CalculateTotalPurchases(in)
	if err != nil {
		p.log.Error("Error calculating total purchase cost", "error", err)
		return nil, fmt.Errorf("error calculating total purchase cost: %w", err)
	}

	res, err := p.repo.CreatePurchase(req)
	if err != nil {
		p.log.Error("Error creating purchase", "error", err)
		return nil, fmt.Errorf("error creating purchase: %w", err)
	}

	var wg sync.WaitGroup
	semaphore := make(chan struct{}, 10) // limit to 10 goroutines

	for _, item := range in.PurchaseItems {
		wg.Add(1)
		go func(item entity.PurchaseItem) {
			defer wg.Done()

			select {
			case semaphore <- struct{}{}:
				// Continue with the operation
			default:
				p.log.Error("Semaphore channel is full, skipping operation", "productID", item.ProductID)
				return
			}

			defer func() { <-semaphore }()

			productQuantityReq := &entity.CountProductReq{
				ID:    item.ProductID,
				Count: item.Quantity,
			}

			if _, err := p.product.AddProduct(productQuantityReq); err != nil {
				p.log.Error("Error adding product quantity", "productID", item.ProductID, "error", err)
			}
		}(item)
	}

	wg.Wait()
	return res, nil
}

func (p *PurchaseUseCase) UpdatePurchase(in *pb.PurchaseUpdate) (*pb.PurchaseResponse, error) {
	res, err := p.repo.UpdatePurchase(in)
	if err != nil {
		p.log.Error("Error updating purchase", "error", err)
		return nil, fmt.Errorf("error updating purchase: %w", err)
	}

	return res, nil
}

func (p *PurchaseUseCase) GetPurchase(req *pb.PurchaseID) (*pb.PurchaseResponse, error) {
	res, err := p.repo.GetPurchase(req)
	if err != nil {
		p.log.Error("Error fetching purchase data", "error", err)
		return nil, fmt.Errorf("error fetching purchase data: %w", err)
	}

	return res, nil
}

func (p *PurchaseUseCase) GetListPurchase(req *pb.FilterPurchase) (*pb.PurchaseList, error) {
	res, err := p.repo.GetPurchaseList(req)
	if err != nil {
		p.log.Error("Error fetching purchase list", "error", err)
		return nil, fmt.Errorf("error fetching purchase list: %w", err)
	}

	return res, nil
}

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
			p.log.Error("Error checking product quantity", "productID", item.ProductId, "error", err)
			return fmt.Errorf("error checking product quantity: %w", err)
		}

		if !check {
			return fmt.Errorf("insufficient product quantity to proceed with deletion")
		}
	}
	return nil
}

func (p *PurchaseUseCase) DeletePurchase(req *pb.PurchaseID) (*pb.Message, error) {
	purchase, err := p.repo.GetPurchase(req)
	if err != nil {
		p.log.Error("Error fetching purchase data", "error", err)
		return nil, fmt.Errorf("error fetching purchase data: %w", err)
	}

	if err := p.validatePurchaseItems(purchase); err != nil {
		p.log.Error("Purchase validation failed before deletion", "error", err)
		return nil, err
	}

	res, err := p.repo.DeletePurchase(req)
	if err != nil {
		p.log.Error("Error deleting purchase", "error", err)
		return nil, fmt.Errorf("error deleting purchase: %w", err)
	}

	go func() {
		for _, item := range purchase.Items {
			productQuantityReq := &entity.CountProductReq{
				ID:    item.ProductId,
				Count: int(item.Quantity),
			}

			if _, err := p.product.RemoveProduct(productQuantityReq); err != nil {
				p.log.Error("Error removing product in DeletePurchase", "productID", item.ProductId, "error", err)
			}
		}
	}()

	return res, nil
}
