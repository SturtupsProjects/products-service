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

	for _, pr := range *in.PurchaseItem {
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
	result.PurchaseItem = &purchaseList
	result.TotalCost = totalSum
	result.PaymentMethod = in.PaymentMethod
	result.Description = in.Description

	return &result, nil
}

func (p *PurchaseUseCase) CreatePurchase(in *entity.Purchase) (*pb.PurchaseResponse, error) {

	req, err := p.CalculateTotalPurchases(in)
	if err != nil {
		p.log.Error("Error calculating total purchase cost", "error", err.Error())
		return nil, fmt.Errorf("error calculating total purchase cost: %w", err)
	}

	res, err := p.repo.CreatePurchase(req)
	if err != nil {
		p.log.Error("Error creating purchase", "error", err.Error())
		return nil, fmt.Errorf("error creating purchase: %w", err)
	}

	var wg sync.WaitGroup
	semaphore := make(chan struct{}, 10) // limit to 10 goroutines

	for _, item := range res.Items {
		wg.Add(1)
		go func(item *pb.PurchaseItemResponse) {
			defer wg.Done()
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			productQuantityReq := &entity.CountProductReq{
				Id:    item.ProductId,
				Count: int(item.Quantity),
			}
			if _, err := p.product.AddProduct(productQuantityReq); err != nil {
				p.log.Error("Error adding product quantity", "error", err.Error())
			}
		}(item)
	}

	wg.Wait()
	return res, nil
}

func (p *PurchaseUseCase) UpdatePurchase(in *entity.PurchaseUpdate) (*pb.PurchaseResponse, error) {
	res, err := p.repo.UpdatePurchase(in)
	if err != nil {
		p.log.Error("Error updating purchase", "error", err.Error())
		return nil, fmt.Errorf("error updating purchase: %w", err)
	}

	return res, nil
}

func (p *PurchaseUseCase) GetPurchase(req *entity.PurchaseID) (*pb.PurchaseResponse, error) {
	res, err := p.repo.GetPurchase(req)
	if err != nil {
		p.log.Error("Error fetching purchase data", "error", err.Error())
		return nil, fmt.Errorf("error fetching purchase data: %w", err)
	}

	return res, nil
}

func (p *PurchaseUseCase) GetListPurchase(req *entity.FilterPurchase) (*pb.PurchaseList, error) {
	res, err := p.repo.GetPurchaseList(req)
	if err != nil {
		p.log.Error("Error fetching purchase list", "error", err.Error())
		return nil, fmt.Errorf("error fetching purchase list: %w", err)
	}

	return res, nil
}

func (p *PurchaseUseCase) validatePurchaseItems(purchase *pb.PurchaseResponse) error {
	for _, item := range purchase.Items {
		if item.Quantity == 0 {
			item.Quantity = 1
		}

		productQuantityReq := &entity.CountProductReq{
			Id:    item.ProductId,
			Count: int(item.Quantity),
		}

		check, err := p.product.ProductCountChecker(productQuantityReq)
		if err != nil {
			p.log.Error("Error checking product quantity", "error", err.Error())
			return fmt.Errorf("error checking product quantity: %w", err)
		}

		if !check {
			return fmt.Errorf("insufficient product quantity to proceed with deletion")
		}
	}
	return nil
}

func (p *PurchaseUseCase) DeletePurchase(req *entity.PurchaseID) (*pb.Message, error) {
	purchase, err := p.repo.GetPurchase(req)
	if err != nil {
		p.log.Error("Error fetching purchase data", "error", err.Error())
		return nil, fmt.Errorf("error fetching purchase data: %w", err)
	}

	if err := p.validatePurchaseItems(purchase); err != nil {
		p.log.Error("Purchase validation failed before deletion", "error", err.Error())
		return nil, err
	}

	res, err := p.repo.DeletePurchase(req)
	if err != nil {
		p.log.Error("Error deleting purchase", "error", err.Error())
		return nil, fmt.Errorf("error deleting purchase: %w", err)
	}

	go func() {
		for _, item := range purchase.Items {
			if item.Quantity == 0 {
				item.Quantity = 1
			}

			productQuantityReq := &entity.CountProductReq{
				Id:    item.ProductId,
				Count: int(item.Quantity),
			}

			if _, err := p.product.RemoveProduct(productQuantityReq); err != nil {
				p.log.Error("Error removing product in DeletePurchase", "error", err.Error())
			}
		}
	}()

	return res, nil
}
