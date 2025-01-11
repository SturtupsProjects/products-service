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

// TotalPriceOfProducts calculates the total price of all products in the inventory within a date range.
func (s *statisticsRepo) TotalPriceOfProducts(req *products.StatisticReq) (*products.PriceProducts, error) {
	query := `
		SELECT COALESCE(SUM(standard_price * total_count), 0) AS total_price
		FROM products
		WHERE company_id = $1 AND branch_id = $2 AND created_at BETWEEN $3 AND $4;
	`
	var totalPrice decimal.Decimal
	if err := s.db.Get(&totalPrice, query, req.GetCompanyId(), req.GetBranchId(), req.GetStartDate(), req.GetEndDate()); err != nil {
		return nil, fmt.Errorf("failed to query total price of products: %w", err)
	}

	result := &products.PriceProducts{
		CompanyId: req.GetCompanyId(),
		BranchId:  req.GetBranchId(),
		Sum: []*products.Price{
			{
				ManyType:   "N/A", // No payment method here
				TotalPrice: totalPrice.InexactFloat64(),
			},
		},
	}

	return result, nil
}

// TotalSoldProducts calculates the total revenue from sold products within a date range.
func (s *statisticsRepo) TotalSoldProducts(req *products.StatisticReq) (*products.PriceProducts, error) {
	query := `
		SELECT s.payment_method AS money_type, COALESCE(SUM(si.total_price), 0) AS total_price
		FROM sales_items si
		JOIN sales s ON si.sale_id = s.id
		WHERE s.company_id = $1 AND s.branch_id = $2 AND s.created_at BETWEEN $3 AND $4
		GROUP BY s.payment_method;
	`
	type tempResult struct {
		MoneyType  string          `db:"money_type"`
		TotalPrice decimal.Decimal `db:"total_price"`
	}

	var tempResults []tempResult

	if err := s.db.Select(&tempResults, query, req.GetCompanyId(), req.GetBranchId(), req.GetStartDate(), req.GetEndDate()); err != nil {
		return nil, fmt.Errorf("failed to calculate total sold products: %w", err)
	}

	var prices []*products.Price
	for _, r := range tempResults {
		prices = append(prices, &products.Price{
			ManyType:   r.MoneyType,
			TotalPrice: r.TotalPrice.InexactFloat64(),
		})
	}

	result := &products.PriceProducts{
		CompanyId: req.GetCompanyId(),
		BranchId:  req.GetBranchId(),
		Sum:       prices,
	}

	return result, nil
}

// TotalPurchaseProducts calculates the total expenditure on purchased products within a date range.
func (s *statisticsRepo) TotalPurchaseProducts(req *products.StatisticReq) (*products.PriceProducts, error) {
	query := `
		SELECT p.payment_method AS money_type, COALESCE(SUM(pi.total_price), 0) AS total_price
		FROM purchase_items pi
		JOIN purchases p ON pi.purchase_id = p.id
		WHERE p.company_id = $1 AND p.branch_id = $2 AND p.created_at BETWEEN $3 AND $4
		GROUP BY p.payment_method;
	`
	type tempResult struct {
		MoneyType  string          `db:"money_type"`
		TotalPrice decimal.Decimal `db:"total_price"`
	}

	var tempResults []tempResult

	if err := s.db.Select(&tempResults, query, req.GetCompanyId(), req.GetBranchId(), req.GetStartDate(), req.GetEndDate()); err != nil {
		return nil, fmt.Errorf("failed to calculate total purchase products: %w", err)
	}

	var prices []*products.Price
	for _, r := range tempResults {
		prices = append(prices, &products.Price{
			ManyType:   r.MoneyType,
			TotalPrice: r.TotalPrice.InexactFloat64(),
		})
	}

	result := &products.PriceProducts{
		CompanyId: req.GetCompanyId(),
		BranchId:  req.GetBranchId(),
		Sum:       prices,
	}

	return result, nil
}
