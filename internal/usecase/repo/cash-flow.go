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
		INSERT INTO cash_flow (id, user_id, amount, transaction_type, description, payment_method, company_id)
		VALUES ($1, $2, $3, 'income', $4, $5, $6)
		RETURNING id, user_id, transaction_date, amount, transaction_type, description, payment_method, company_id
	`

	var cashFlow pb.CashFlow
	err := c.db.QueryRowx(query, id, in.UserId, in.Amount, in.Description, in.PaymentMethod, in.CompanyId).
		Scan(&cashFlow.Id, &cashFlow.UserId, &cashFlow.TransactionDate, &cashFlow.Amount, &cashFlow.TransactionType, &cashFlow.Description, &cashFlow.PaymentMethod, &cashFlow.CompanyId)
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
		INSERT INTO cash_flow (id, user_id, amount, transaction_type, description, payment_method, company_id)
		VALUES ($1, $2, $3, 'expense', $4, $5, $6)
		RETURNING id, user_id, transaction_date, amount, transaction_type, description, payment_method, company_id
	`

	var cashFlow pb.CashFlow
	err := c.db.QueryRowx(query, id, in.UserId, in.Amount, in.Description, in.PaymentMethod, in.CompanyId).
		Scan(&cashFlow.Id, &cashFlow.UserId, &cashFlow.TransactionDate, &cashFlow.Amount, &cashFlow.TransactionType, &cashFlow.Description, &cashFlow.PaymentMethod, &cashFlow.CompanyId)
	if err != nil {
		return nil, fmt.Errorf("failed to create expense: %w", err)
	}

	return &cashFlow, nil
}

func (c *cashFlow) Get(in *pb.StatisticReq) (*pb.ListCashFlow, error) {

	query := `
		SELECT id, user_id, transaction_date, amount, transaction_type, description, payment_method, company_id
		FROM cash_flow
		WHERE company_id = $1
		AND transaction_date BETWEEN $2 AND $3
	`

	var cashFlows []*pb.CashFlow
	rows, err := c.db.Queryx(query, in.CompanyId, in.StartDate, in.EndDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var cashFlow pb.CashFlow
		if err := rows.Scan(&cashFlow.Id, &cashFlow.UserId, &cashFlow.TransactionDate, &cashFlow.Amount, &cashFlow.TransactionType, &cashFlow.Description, &cashFlow.PaymentMethod, &cashFlow.CompanyId); err != nil {
			return nil, err
		}
		cashFlows = append(cashFlows, &cashFlow)
	}

	if len(cashFlows) == 0 {
		return nil, fmt.Errorf("no cash flows found for the given parameters")
	}

	return &pb.ListCashFlow{Cash: cashFlows}, nil
}
