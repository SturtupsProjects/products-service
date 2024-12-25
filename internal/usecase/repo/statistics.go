package repo

import (
	"crm-admin/internal/generated/products"
	"crm-admin/internal/usecase"
	"fmt"
	"github.com/jmoiron/sqlx"
	"strconv"
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
		SELECT SUM(standard_price * total_count) AS total_price
		FROM products
		WHERE company_id = $1;
	`

	// Используем строку для считывания результата
	var total []struct {
		TotalPrice string `db:"total_price"`
	}

	if err := s.db.Select(&total, query, companyID.GetId()); err != nil {
		return nil, err
	}

	var totalPrice int64

	for _, price := range total {
		// Преобразуем строку в int64
		val, err := strconv.ParseInt(price.TotalPrice, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("failed to parse total_price: %w", err)
		}
		totalPrice += val
	}

	result.TotalPrice = totalPrice

	return &result, nil
}

// TotalSoldProducts calculates the total revenue from sold products.
func (s *statisticsRepo) TotalSoldProducts(companyID *products.CompanyID) (*products.PriceProducts, error) {
	query := `
		SELECT SUM(total_price) AS total_price
		FROM sales_items
		WHERE company_id = $1;
	`
	var result products.PriceProducts
	if err := s.db.Get(&result.TotalPrice, query, companyID.GetId()); err != nil {
		return nil, err
	}
	return &result, nil
}

// TotalPurchaseProducts calculates the total expenditure on purchased products.
func (s *statisticsRepo) TotalPurchaseProducts(companyID *products.CompanyID) (*products.PriceProducts, error) {
	query := `
		SELECT SUM(total_price) AS total_price
		FROM purchase_items
		WHERE company_id = $1;
	`
	var result products.PriceProducts
	if err := s.db.Get(&result.TotalPrice, query, companyID.GetId()); err != nil {
		return nil, err
	}
	return &result, nil
}
