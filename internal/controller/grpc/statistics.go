package grpc

import (
	"context"
	pb "crm-admin/internal/generated/products"
)

func (p *ProductsGrpc) TotalPriceOfProducts(ctx context.Context, in *pb.StatisticReq) (*pb.PriceProducts, error) {
	p.log.Info("TotalPriceOfProducts called")

	res, err := p.statistics.TotalPriceOfProducts(in)
	if err != nil {
		p.log.Error("Error in TotalPriceOfProducts", "error", err)
		return nil, err
	}

	return res, nil
}

func (p *ProductsGrpc) TotalSoldProducts(ctx context.Context, in *pb.StatisticReq) (*pb.PriceProducts, error) {
	p.log.Info("TotalSoldProducts called")

	res, err := p.statistics.TotalSoldProducts(in)
	if err != nil {
		p.log.Error("Error in TotalSoldProducts", "error", err)
		return nil, err
	}

	return res, nil
}

func (p *ProductsGrpc) TotalPurchaseProducts(ctx context.Context, in *pb.StatisticReq) (*pb.PriceProducts, error) {
	p.log.Info("TotalPurchaseProducts called")

	res, err := p.statistics.TotalPurchaseProducts(in)
	if err != nil {
		p.log.Error("Error in TotalPurchaseProducts", "error", err)
		return nil, err
	}

	return res, nil
}

// --------------------------------------------------- Cash Flow -------------------------------------------------------

func (p *ProductsGrpc) GetCashFlow(ctx context.Context, in *pb.StatisticReq) (*pb.ListCashFlow, error) {
	p.log.Info("GetCashFlow called")

	res, err := p.cashFlow.Get(in)
	if err != nil {
		p.log.Error("Error in GetCashFlow", "error", err)
		return nil, err
	}

	return res, nil
}

func (p *ProductsGrpc) CreateIncome(ctx context.Context, in *pb.CashFlowRequest) (*pb.CashFlow, error) {
	p.log.Info("CreateIncome called")

	res, err := p.cashFlow.CreateIncome(in)
	if err != nil {
		p.log.Error("Error in CreateIncome", "error", err)
		return nil, err
	}

	return res, nil
}

func (p *ProductsGrpc) CreateExpense(ctx context.Context, in *pb.CashFlowRequest) (*pb.CashFlow, error) {
	p.log.Info("CreateExpense called")

	res, err := p.cashFlow.CreateExpense(in)
	if err != nil {
		p.log.Error("Error in CreateExpense", "error", err)
		return nil, err
	}

	return res, nil
}

func (p *ProductsGrpc) GetTotalIncome(ctx context.Context, in *pb.StatisticReq) (*pb.PriceProducts, error) {
	p.log.Info("GetTotalIncome called")

	res, err := p.cashFlow.GetTotalIncome(in)
	if err != nil {
		p.log.Error("Error in GetTotalIncome", "error", err)
		return nil, err
	}

	return res, nil
}

func (p *ProductsGrpc) GetTotalExpense(ctx context.Context, in *pb.StatisticReq) (*pb.PriceProducts, error) {
	p.log.Info("GetTotalExpense called")

	res, err := p.cashFlow.GetTotalExpense(in)
	if err != nil {
		p.log.Error("Error in GetTotalExpense", "error", err)
		return nil, err
	}

	return res, nil
}

func (p *ProductsGrpc) GetNetProfit(ctx context.Context, in *pb.StatisticReq) (*pb.PriceProducts, error) {
	p.log.Info("GetNetProfit called")

	res, err := p.cashFlow.GetNetProfit(in)
	if err != nil {
		p.log.Error("Error in GetNetProfit", "error", err)
		return nil, err
	}

	return res, nil
}
