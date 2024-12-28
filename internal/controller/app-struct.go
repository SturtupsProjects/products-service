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
	cashFlowRepo := repo.NewCashFlow(db)

	ctr := &Controller{
		Product:  usecase.NewProductsUseCase(productRepo, log),
		Purchase: usecase.NewPurchaseUseCase(purchaseRepo, productQuantityRepo, log, cashFlowRepo),
		Sales:    usecase.NewSalesUseCase(salesRepo, productQuantityRepo, log, cashFlowRepo),
	}

	return ctr
}
