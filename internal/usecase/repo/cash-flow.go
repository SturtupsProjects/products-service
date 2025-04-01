package repo

import (
	pb "crm-admin/internal/generated/products"
	"crm-admin/internal/usecase"
	"fmt"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"strings"
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

func (c *cashFlow) Get(in *pb.CashFlowReq) (*pb.ListCashFlow, error) {
	// Базовый запрос с обязательными фильтрами и оконной функцией для подсчёта общего количества
	query := `
		SELECT 
			id, user_id, transaction_date, amount, transaction_type, description, payment_method, company_id, branch_id,
			COUNT(*) OVER() AS total_count
		FROM cash_flow
		WHERE company_id = $1 
		  AND branch_id = $2
		  AND transaction_date BETWEEN $3 AND $4
	`
	args := []interface{}{in.CompanyId, in.BranchId, in.StartDate, in.EndDate}
	index := 5

	// Дополнительные фильтры
	if in.Description != "" {
		query += fmt.Sprintf(" AND description ILIKE $%d", index)
		args = append(args, "%"+in.Description+"%")
		index++
	}
	if in.TransactionType != "" {
		query += fmt.Sprintf(" AND transaction_type = $%d", index)
		args = append(args, in.TransactionType)
		index++
	}
	if in.PaymentMethod != "" {
		query += fmt.Sprintf(" AND payment_method = $%d", index)
		args = append(args, in.PaymentMethod)
		index++
	}

	query += " ORDER BY transaction_date DESC"
	if in.Limit > 0 && in.Page > 0 {
		query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", index, index+1)
		args = append(args, in.Limit, (in.Page-1)*in.Limit)
		index += 2
	}

	rows, err := c.db.Queryx(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}
	defer rows.Close()

	var cashFlows []*pb.CashFlow
	var totalCount int64

	for rows.Next() {
		var cashFlow pb.CashFlow
		var cnt int64
		if err := rows.Scan(
			&cashFlow.Id,
			&cashFlow.UserId,
			&cashFlow.TransactionDate,
			&cashFlow.Amount,
			&cashFlow.TransactionType,
			&cashFlow.Description,
			&cashFlow.PaymentMethod,
			&cashFlow.CompanyId,
			&cashFlow.BranchId,
			&cnt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		totalCount = cnt
		cashFlows = append(cashFlows, &cashFlow)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return &pb.ListCashFlow{
		Cash:       cashFlows,
		TotalCount: totalCount,
	}, nil
}

func (cf *cashFlow) GetTotalIncome(req *pb.StatisticReq) (*pb.PriceProducts, error) {
	var args []interface{}
	argIndex := 1

	// Основной запрос
	queryBuilder := strings.Builder{}
	queryBuilder.WriteString(`
		SELECT 
			COALESCE(payment_method, 'uzs') AS many_type, 
			SUM(amount) AS total_price
		FROM 
			cash_flow
		WHERE 
			transaction_type = 'income'
			AND company_id = $1
			AND branch_id = $2
	`)

	args = append(args, req.CompanyId, req.BranchId)
	argIndex += 2

	// Фильтр по датам
	if req.StartDate != "" {
		queryBuilder.WriteString(fmt.Sprintf(" AND transaction_date >= $%d", argIndex))
		args = append(args, req.StartDate)
		argIndex++
	}
	if req.EndDate != "" {
		queryBuilder.WriteString(fmt.Sprintf(" AND transaction_date <= $%d", argIndex))
		args = append(args, req.EndDate)
		argIndex++
	}

	// Группировка
	queryBuilder.WriteString(`
		GROUP BY 
			payment_method
		ORDER BY 
			total_price DESC
	`)

	// Пагинация
	if req.Limit > 0 && req.Page > 0 {
		queryBuilder.WriteString(fmt.Sprintf(" LIMIT $%d OFFSET $%d", argIndex, argIndex+1))
		args = append(args, req.Limit, (req.Page-1)*req.Limit)
	}

	query := queryBuilder.String()

	// Выполнение запроса
	rows, err := cf.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}
	defer rows.Close()

	// Обработка результатов
	var prices []*pb.Price
	for rows.Next() {
		var price pb.Price
		if err := rows.Scan(&price.ManyType, &price.TotalPrice); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		prices = append(prices, &price)
	}

	// Возврат результата
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

func (r *cashFlow) GetBranchIncome(in *pb.BranchIncomeReq) (*pb.BranchIncomeRes, error) {
	query := `
        SELECT 
            branch_id,
            payment_method,
            SUM(amount) AS total_income
        FROM 
            cash_flow
        WHERE 
            transaction_date BETWEEN $1 AND $2
            AND company_id = $3
            AND transaction_type = 'income'
        GROUP BY 
            branch_id, payment_method
        ORDER BY 
            branch_id;
    `

	rows, err := r.db.Query(query, in.StartDate, in.EndDate, in.CompanyId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Для хранения данных по филиалам
	branchMap := make(map[string]map[string]float64) // branch_id -> payment_method -> total_income
	totalIncome := 0.0

	for rows.Next() {
		var branchID string
		var paymentMethod string
		var totalIncomeForMethod float64

		if err := rows.Scan(&branchID, &paymentMethod, &totalIncomeForMethod); err != nil {
			return nil, err
		}

		if _, exists := branchMap[branchID]; !exists {
			branchMap[branchID] = make(map[string]float64)
		}

		branchMap[branchID][paymentMethod] += totalIncomeForMethod
		totalIncome += totalIncomeForMethod
	}

	// Формирование результата
	result := &pb.BranchIncomeRes{
		Total: totalIncome,
	}

	for branchID, paymentMethods := range branchMap {
		branchIncomeData := &pb.BranchIncomeData{
			BranchId: branchID,
		}

		for method, amount := range paymentMethods {
			branchIncomeData.Values = append(branchIncomeData.Values, &pb.Price{
				ManyType:   method,
				TotalPrice: amount,
			})
		}

		result.Data = append(result.Data, branchIncomeData)
	}

	return result, nil
}
