package usecase

import (
	pb "crm-admin/internal/generated/products"
	"crm-admin/internal/webapi"
	"github.com/shopspring/decimal"
	"log"

	"log/slog"
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
		p.log.Error("GetProductDashboard", "error", err.Error())
		return nil, err
	}

	usdRate, err := webapi.GetUSDCourse()
	if err != nil {
		p.log.Error("GetProductDashboard", "error", err.Error())
		return nil, err
	}
	usdRateDec := decimal.NewFromFloat(usdRate)

	purchasePrice := decimal.Zero
	salePrice := decimal.Zero

	convertPrice := func(price decimal.Decimal, srcCurrency, targetCurrency string) decimal.Decimal {
		switch targetCurrency {
		case "usd":
			if srcCurrency == "uzs" || srcCurrency == "card" {
				return price.Div(usdRateDec)
			}
			return price
		case "uzs":
			if srcCurrency == "usd" {
				return price.Mul(usdRateDec)
			}
			return price
		default:
			return price
		}
	}

	for _, item := range dbRes.AmountDeliveryPrice {
		price := decimal.NewFromFloat(item.Price)
		purchasePrice = purchasePrice.Add(convertPrice(price, item.Currency, in.Currency))
	}

	for _, item := range dbRes.AmountSalePrice {
		price := decimal.NewFromFloat(item.Price)
		salePrice = salePrice.Add(convertPrice(price, item.Currency, in.Currency))
	}

	res := &pb.GetProductsDashboardRes{
		ProductItems:        dbRes.ProductItems,
		ProductUnits:        dbRes.ProductUnits,
		AmountDeliveryPrice: purchasePrice.Round(1).InexactFloat64(),
		AmountSalePrice:     salePrice.Round(1).InexactFloat64(),
	}

	return res, nil
}
