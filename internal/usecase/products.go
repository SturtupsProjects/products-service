package usecase

import (
	pb "crm-admin/internal/generated/products"
	"crm-admin/internal/webapi"
	"fmt"
	"github.com/shopspring/decimal"
	"log/slog"
	"strings"
)

type ProductsUseCase struct {
	repo ProductsRepo
	log  *slog.Logger
}

func NewProductsUseCase(repo ProductsRepo, log *slog.Logger) *ProductsUseCase {
	return &ProductsUseCase{repo: repo, log: log}
}

// --------------------  Product Category ----------------------------------------------------------------------

func (p *ProductsUseCase) CreateCategory(in *pb.CreateCategoryRequest) (*pb.Category, error) {
	res, err := p.repo.CreateProductCategory(in)

	if err != nil {
		p.log.Error("CreateCategory", "error", err.Error())
		return nil, err
	}

	return res, nil
}

func (p *ProductsUseCase) UpdateCategory(in *pb.UpdateCategoryRequest) (*pb.Category, error) {
	res, err := p.repo.UpdateProductCategory(in)

	if err != nil {
		p.log.Error("UpdateCategory", "error", err.Error())
		return nil, err
	}

	return res, nil
}

func (p *ProductsUseCase) DeleteCategory(in *pb.GetCategoryRequest) (*pb.Message, error) {
	res, err := p.repo.DeleteProductCategory(in)

	if err != nil {
		p.log.Error("DeleteCategory", "error", err.Error())
		return nil, err
	}

	return res, nil
}

func (p *ProductsUseCase) GetCategory(in *pb.GetCategoryRequest) (*pb.Category, error) {
	res, err := p.repo.GetProductCategory(in)

	if err != nil {
		p.log.Error("GetCategory", "error", err.Error())
		return nil, err
	}

	return res, nil
}

func (p *ProductsUseCase) GetListCategory(in *pb.CategoryName) (*pb.CategoryList, error) {
	res, err := p.repo.GetListProductCategory(in)

	if err != nil {
		p.log.Error("GetListCategory", "error", err.Error())
		return nil, err
	}

	return res, nil
}

// --------------------------- Products ----------------------------------------------------------------------------

func (p *ProductsUseCase) CreateProduct(in *pb.CreateProductRequest) (*pb.Product, error) {
	res, err := p.repo.CreateProduct(in)

	if err != nil {
		p.log.Error("CreateProduct", "error", err.Error())
		return nil, err
	}

	return res, nil
}
func (p *ProductsUseCase) CreateBulkProducts(in *pb.CreateBulkProductsRequest) (*pb.BulkCreateResponse, error) {
	res, err := p.repo.CreateBulkProducts(in)

	if err != nil {
		p.log.Error("CreateBulkProducts", "error", err.Error())
		return nil, err
	}

	return res, nil
}

func (p *ProductsUseCase) UpdateProduct(in *pb.UpdateProductRequest) (*pb.Product, error) {
	res, err := p.repo.UpdateProduct(in)

	if err != nil {
		p.log.Error("UpdateProduct", "error", err.Error())
		return nil, err
	}

	return res, nil
}

func (p *ProductsUseCase) DeleteProduct(in *pb.GetProductRequest) (*pb.Message, error) {
	res, err := p.repo.DeleteProduct(in)

	if err != nil {
		p.log.Error("DeleteProduct", "error", err.Error())
		return nil, err
	}

	return res, nil
}

func (p *ProductsUseCase) GetProduct(in *pb.GetProductRequest) (*pb.Product, error) {
	res, err := p.repo.GetProduct(in)

	if err != nil {
		p.log.Error("GetProduct", "error", err.Error())
		return nil, err
	}

	return res, nil
}

func (p *ProductsUseCase) GetProductList(in *pb.ProductFilter) (*pb.ProductList, error) {
	res, err := p.repo.GetProductList(in)

	if err != nil {
		p.log.Error("GetProductList", "error", err.Error())
		return nil, err
	}

	return res, nil
}

func (p *ProductsUseCase) GetProductDashboard(in *pb.GetProductsDashboardReq) (*pb.GetProductsDashboardRes, error) {

	dbRes, err := p.repo.GetProductDashboard(in)
	if err != nil {
		p.log.Error("GetProductDashboard repo error", "error", err)
		return nil, err
	}

	target := strings.ToUpper(in.Currency)
	if target != "USD" && target != "UZS" {
		return nil, fmt.Errorf("unsupported currency: %s", in.Currency)
	}

	usdRateFloat, err := webapi.GetUSDCourse()
	if err != nil {
		p.log.Error("GetProductDashboard rate fetch error", "error", err)
		return nil, err
	}

	rateUZSperUSD := decimal.NewFromFloat(usdRateFloat)

	normalize := func(c string) string {
		switch strings.ToUpper(c) {
		case "USD":
			return "USD"
		default:
			return "UZS"
		}
	}

	convert := func(amount decimal.Decimal, src string) decimal.Decimal {
		srcNorm := normalize(src)
		switch {
		case srcNorm == "UZS" && target == "USD":
			return amount.Div(rateUZSperUSD)
		case srcNorm == "USD" && target == "UZS":
			return amount.Mul(rateUZSperUSD)
		default:
			return amount
		}
	}

	sumDelivery := decimal.Zero
	for _, item := range dbRes.AmountDeliveryPrice {
		sumDelivery = sumDelivery.Add(convert(decimal.NewFromFloat(item.Price), item.Currency))
	}

	sumSale := decimal.Zero
	for _, item := range dbRes.AmountSalePrice {
		sumSale = sumSale.Add(convert(decimal.NewFromFloat(item.Price), item.Currency))
	}

	sumDelivery = sumDelivery.Round(2)
	sumSale = sumSale.Round(2)

	return &pb.GetProductsDashboardRes{
		ProductItems:        dbRes.ProductItems,
		ProductUnits:        dbRes.ProductUnits,
		AmountDeliveryPrice: sumDelivery.InexactFloat64(),
		AmountSalePrice:     sumSale.InexactFloat64(),
	}, nil
}
