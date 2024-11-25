package controller

import (
	"crm-admin/internal/usecase"
	"crm-admin/internal/usecase/repo"
	"github.com/jmoiron/sqlx"
	"log/slog"
)

type Controller struct {
	Product  *usecase.ProductsUseCase
	Purchase *usecase.PurchaseUseCase
	Sales    *usecase.SalesUseCase
}

func NewController(db *sqlx.DB, log *slog.Logger) *Controller {

	productRepo := repo.NewProductRepo(db)
	purchaseRepo := repo.NewPurchasesRepo(db)
	salesRepo := repo.NewSalesRepo(db)
	productQuantityRepo := repo.NewProductQuantity(db)

	ctr := &Controller{
		Product:  usecase.NewProductsUseCase(productRepo, log),
		Purchase: usecase.NewPurchaseUseCase(purchaseRepo, productQuantityRepo, log),
		Sales:    usecase.NewSalesUseCase(salesRepo, productQuantityRepo, log),
	}

	return ctr
}
