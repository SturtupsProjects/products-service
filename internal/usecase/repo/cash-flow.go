package repo

import (
	pb "crm-admin/internal/generated/products"
	"crm-admin/internal/usecase"
	"fmt"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type cashFlow struct {
	db *sqlx.DB
}

func NewCashFlow(db *sqlx.DB) usecase.CashFlowRepo {
	return &cashFlow{db: db}
}

// CreateIncome создает запись о доходе
func (c *cashFlow) CreateIncome(in *pb.CashFlowRequest) (*pb.CashFlow, error) {
	id := uuid.NewString()

	// SQL-запрос для записи о доходе
	query := `
		INSERT INTO cash_flow (id, user_id, amount, transaction_type, description, payment_method, company_id, branch_id)
		VALUES ($1, $2, $3, 'income', $4, $5, $6, $7)
		RETURNING id, user_id, transaction_date, amount, transaction_type, description, payment_method, company_id, branch_id
	`

	var cashFlow pb.CashFlow
	err := c.db.QueryRowx(query, id, in.UserId, in.Amount, in.Description, in.PaymentMethod, in.CompanyId, in.BranchId).
		Scan(&cashFlow.Id, &cashFlow.UserId, &cashFlow.TransactionDate, &cashFlow.Amount, &cashFlow.TransactionType, &cashFlow.Description, &cashFlow.PaymentMethod, &cashFlow.CompanyId, &cashFlow.BranchId)
	if err != nil {
		return nil, fmt.Errorf("failed to create income: %w", err)
	}

	return &cashFlow, nil
}

// CreateExpense создает запись о расходе
func (c *cashFlow) CreateExpense(in *pb.CashFlowRequest) (*pb.CashFlow, error) {
	id := uuid.NewString()

	// SQL-запрос для записи о расходе
	query := `
		INSERT INTO cash_flow (id, user_id, amount, transaction_type, description, payment_method, company_id, branch_id)
		VALUES ($1, $2, $3, 'expense', $4, $5, $6, $7)
		RETURNING id, user_id, transaction_date, amount, transaction_type, description, payment_method, company_id, branch_id
	`

	var cashFlow pb.CashFlow
	err := c.db.QueryRowx(query, id, in.UserId, in.Amount, in.Description, in.PaymentMethod, in.CompanyId, in.BranchId).
		Scan(&cashFlow.Id, &cashFlow.UserId, &cashFlow.TransactionDate, &cashFlow.Amount, &cashFlow.TransactionType, &cashFlow.Description, &cashFlow.PaymentMethod, &cashFlow.CompanyId, &cashFlow.BranchId)
	if err != nil {
		return nil, fmt.Errorf("failed to create expense: %w", err)
	}

	return &cashFlow, nil
}

func (c *cashFlow) Get(in *pb.StatisticReq) (*pb.ListCashFlow, error) {

	query := `
		SELECT id, user_id, transaction_date, amount, transaction_type, description, payment_method, company_id, branch_id
		FROM cash_flow
		WHERE company_id = $1 AND branch_id = $2
		AND transaction_date BETWEEN $3 AND $4 ORDER BY transaction_date DESC
		LIMIT $5 OFFSET $6
	`

	var cashFlows []*pb.CashFlow
	rows, err := c.db.Queryx(query, in.CompanyId, in.BranchId, in.StartDate, in.EndDate, in.Limit, (in.Page-1)*in.Limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var cashFlow pb.CashFlow
		if err := rows.Scan(&cashFlow.Id, &cashFlow.UserId, &cashFlow.TransactionDate, &cashFlow.Amount, &cashFlow.TransactionType, &cashFlow.Description, &cashFlow.PaymentMethod, &cashFlow.CompanyId, &cashFlow.BranchId); err != nil {
			return nil, err
		}
		cashFlows = append(cashFlows, &cashFlow)
	}

	if len(cashFlows) == 0 {
		return nil, fmt.Errorf("no cash flows found for the given parameters")
	}

	return &pb.ListCashFlow{Cash: cashFlows}, nil
}

// GetTotalIncome возвращает общую сумму доходов за указанный период, сгруппированную по типу денег
func (cf *cashFlow) GetTotalIncome(req *pb.StatisticReq) (*pb.PriceProducts, error) {
	query := `
		SELECT 
			payment_method AS many_type, 
			SUM(amount) AS total_price
		FROM 
			cash_flow
		WHERE 
			transaction_type = 'income'
			AND transaction_date BETWEEN $1 AND $2
			AND company_id = $3
			AND branch_id = $4
		GROUP BY 
			payment_method;
	`

	rows, err := cf.db.Query(query, req.StartDate, req.EndDate, req.CompanyId, req.BranchId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var prices []*pb.Price
	for rows.Next() {
		var price pb.Price
		if err := rows.Scan(&price.ManyType, &price.TotalPrice); err != nil {
			return nil, err
		}
		prices = append(prices, &price)
	}

	return &pb.PriceProducts{
		CompanyId: req.CompanyId,
		BranchId:  req.BranchId,
		Sum:       prices,
	}, nil
}

// GetTotalExpense возвращает общую сумму расходов за указанный период, сгруппированную по типу денег
func (cf *cashFlow) GetTotalExpense(req *pb.StatisticReq) (*pb.PriceProducts, error) {
	query := `
		SELECT 
			payment_method AS many_type, 
			SUM(amount) AS total_price
		FROM 
			cash_flow
		WHERE 
			transaction_type = 'expense'
			AND transaction_date BETWEEN $1 AND $2
			AND company_id = $3
			AND branch_id = $4
		GROUP BY 
			payment_method;
	`

	rows, err := cf.db.Query(query, req.StartDate, req.EndDate, req.CompanyId, req.BranchId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var prices []*pb.Price
	for rows.Next() {
		var price pb.Price
		if err := rows.Scan(&price.ManyType, &price.TotalPrice); err != nil {
			return nil, err
		}
		prices = append(prices, &price)
	}

	return &pb.PriceProducts{
		CompanyId: req.CompanyId,
		BranchId:  req.BranchId,
		Sum:       prices,
	}, nil
}

// GetNetProfit возвращает чистую прибыль за указанный период, сгруппированную по типу денег
func (cf *cashFlow) GetNetProfit(req *pb.StatisticReq) (*pb.PriceProducts, error) {
	query := `
		SELECT 
			payment_method AS many_type, 
			SUM(CASE WHEN transaction_type = 'income' THEN amount ELSE 0 END) -
			SUM(CASE WHEN transaction_type = 'expense' THEN amount ELSE 0 END) AS total_price
		FROM 
			cash_flow
		WHERE 
			transaction_date BETWEEN $1 AND $2
			AND company_id = $3
			AND branch_id = $4
		GROUP BY 
			payment_method;
	`

	rows, err := cf.db.Query(query, req.StartDate, req.EndDate, req.CompanyId, req.BranchId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var prices []*pb.Price
	for rows.Next() {
		var price pb.Price
		if err := rows.Scan(&price.ManyType, &price.TotalPrice); err != nil {
			return nil, err
		}
		prices = append(prices, &price)
	}

	return &pb.PriceProducts{
		CompanyId: req.CompanyId,
		BranchId:  req.BranchId,
		Sum:       prices,
	}, nil
}
