package grpc

import (
	"context"
	pb "crm-admin/internal/generated/products"
)

func (p *ProductsGrpc) TotalPriceOfProducts(ctx context.Context, in *pb.StatisticReq) (*pb.PriceProducts, error) {

	res, err := p.statistics.TotalPriceOfProducts(in)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (p *ProductsGrpc) TotalSoldProducts(ctx context.Context, in *pb.StatisticReq) (*pb.PriceProducts, error) {

	res, err := p.statistics.TotalSoldProducts(in)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (p *ProductsGrpc) TotalPurchaseProducts(ctx context.Context, in *pb.StatisticReq) (*pb.PriceProducts, error) {

	res, err := p.statistics.TotalPurchaseProducts(in)
	if err != nil {
		return nil, err
	}

	return res, nil
}

// --------------------------------------------------- Cash Flow -------------------------------------------------------

func (p *ProductsGrpc) GetCashFlow(ctx context.Context, in *pb.StatisticReq) (*pb.ListCashFlow, error) {

	res, err := p.cashFlow.Get(in)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (p *ProductsGrpc) CreateIncome(ctx context.Context, in *pb.CashFlowRequest) (*pb.CashFlow, error) {

	res, err := p.cashFlow.CreateIncome(in)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (p *ProductsGrpc) CreateExpense(ctx context.Context, in *pb.CashFlowRequest) (*pb.CashFlow, error) {

	res, err := p.cashFlow.CreateExpense(in)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (p *ProductsGrpc) GetTotalIncome(ctx context.Context, in *pb.StatisticReq) (*pb.PriceProducts, error) {

	res, err := p.cashFlow.GetTotalIncome(in)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (p *ProductsGrpc) GetTotalExpense(ctx context.Context, in *pb.StatisticReq) (*pb.PriceProducts, error) {

	res, err := p.cashFlow.GetTotalExpense(in)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (p *ProductsGrpc) GetNetProfit(ctx context.Context, in *pb.StatisticReq) (*pb.PriceProducts, error) {

	res, err := p.cashFlow.GetNetProfit(in)
	if err != nil {
		return nil, err
	}

	return res, nil
}
