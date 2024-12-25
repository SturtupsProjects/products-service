package grpc

import (
	"context"
	pb "crm-admin/internal/generated/products"
)

func (p *ProductsGrpc) TotalPriceOfProducts(ctx context.Context, in *pb.CompanyID) (*pb.PriceProducts, error) {
	p.log.Info("TotalPriceOfProducts called")

	res, err := p.statistics.TotalPriceOfProducts(in)
	if err != nil {
		p.log.Error("Error in TotalPriceOfProducts", "error", err)
		return nil, err
	}

	return res, nil
}

func (p *ProductsGrpc) TotalSoldProducts(ctx context.Context, in *pb.CompanyID) (*pb.PriceProducts, error) {
	p.log.Info("TotalSoldProducts called")

	res, err := p.statistics.TotalSoldProducts(in)
	if err != nil {
		p.log.Error("Error in TotalSoldProducts", "error", err)
		return nil, err
	}

	return res, nil
}

func (p *ProductsGrpc) TotalPurchaseProducts(ctx context.Context, in *pb.CompanyID) (*pb.PriceProducts, error) {
	p.log.Info("TotalPurchaseProducts called")

	res, err := p.statistics.TotalPurchaseProducts(in)
	if err != nil {
		p.log.Error("Error in TotalPurchaseProducts", "error", err)
		return nil, err
	}

	return res, nil
}
