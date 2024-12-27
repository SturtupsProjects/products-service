package repo

import (
	"crm-admin/internal/generated/products"
	"crm-admin/internal/usecase"
	"fmt"
	"github.com/jmoiron/sqlx"
	"github.com/shopspring/decimal"
)

type statisticsRepo struct {
	db *sqlx.DB
}

func NewStatisticsRepo(db *sqlx.DB) usecase.StatisticsRepo {
	return &statisticsRepo{db: db}
}

// TotalPriceOfProducts calculates the total price of all products in the inventory.
func (s *statisticsRepo) TotalPriceOfProducts(companyID *products.CompanyID) (*products.PriceProducts, error) {
	var result products.PriceProducts

	query := `
		SELECT COALESCE(SUM(standard_price * total_count), 0) AS total_price
		FROM products
		WHERE company_id = $1;
	`

	var totalPrice float64

	if err := s.db.Get(&totalPrice, query, companyID.GetId()); err != nil {
		return nil, fmt.Errorf("failed to query total price of products: %w", err)
	}

	totalPriceDecimal := decimal.NewFromFloat(totalPrice)

	result.TotalPrice = totalPriceDecimal.InexactFloat64()

	return &result, nil
}

// TotalSoldProducts calculates the total revenue from sold products.
func (s *statisticsRepo) TotalSoldProducts(companyID *products.CompanyID) (*products.PriceProducts, error) {
	query := `
		SELECT COALESCE(SUM(total_price), 0) AS total_price
		FROM sales_items
		WHERE company_id = $1;
	`

	var totalPrice decimal.Decimal

	if err := s.db.Get(&totalPrice, query, companyID.GetId()); err != nil {
		return nil, fmt.Errorf("failed to calculate total sold products: %w", err)
	}

	result := &products.PriceProducts{
		TotalPrice: totalPrice.InexactFloat64(),
	}

	return result, nil
}

// TotalPurchaseProducts calculates the total expenditure on purchased products.
func (s *statisticsRepo) TotalPurchaseProducts(companyID *products.CompanyID) (*products.PriceProducts, error) {
	query := `
		SELECT COALESCE(SUM(total_price), 0) AS total_price
		FROM purchase_items
		WHERE company_id = $1;
	`

	var totalPrice decimal.Decimal

	if err := s.db.Get(&totalPrice, query, companyID.GetId()); err != nil {
		return nil, fmt.Errorf("failed to calculate total purchase products: %w", err)
	}

	result := &products.PriceProducts{
		TotalPrice: totalPrice.InexactFloat64(),
	}

	return result, nil
}
