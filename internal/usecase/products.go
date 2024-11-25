package usecase

import (
	"crm-admin/internal/entity"
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

func (p *ProductsUseCase) CreateCategory(in *entity.CategoryName) (*entity.Category, error) {
	res, err := p.repo.CreateProductCategory(in)

	if err != nil {
		p.log.Error("CreateCategory", "error", err.Error())
		return nil, err
	}

	return res, nil
}

func (p *ProductsUseCase) DeleteCategory(in *entity.CategoryID) (*entity.Message, error) {
	res, err := p.repo.DeleteProductCategory(in)

	if err != nil {
		p.log.Error("DeleteCategory", "error", err.Error())
		return nil, err
	}

	return res, nil
}

func (p *ProductsUseCase) GetCategory(in *entity.CategoryID) (*entity.Category, error) {
	res, err := p.repo.GetProductCategory(in)

	if err != nil {
		p.log.Error("GetCategory", "error", err.Error())
		return nil, err
	}

	return res, nil
}

func (p *ProductsUseCase) GetListCategory(in *entity.CategoryName) (*entity.CategoryList, error) {
	res, err := p.repo.GetListProductCategory(in)

	if err != nil {
		p.log.Error("GetListCategory", "error", err.Error())
		return nil, err
	}

	return res, nil
}

// --------------------------- Products ----------------------------------------------------------------------------

func (p *ProductsUseCase) CreateProduct(in *entity.ProductRequest) (*entity.Product, error) {
	res, err := p.repo.CreateProduct(in)

	if err != nil {
		p.log.Error("CreateProduct", "error", err.Error())
		return nil, err
	}

	return res, nil
}

func (p *ProductsUseCase) UpdateProduct(in *entity.ProductUpdate) (*entity.Product, error) {
	res, err := p.repo.UpdateProduct(in)

	if err != nil {
		p.log.Error("UpdateProduct", "error", err.Error())
		return nil, err
	}

	return res, nil
}

func (p *ProductsUseCase) DeleteProduct(in *entity.ProductID) (*entity.Message, error) {
	res, err := p.repo.DeleteProduct(in)

	if err != nil {
		p.log.Error("DeleteProduct", "error", err.Error())
		return nil, err
	}

	return res, nil
}

func (p *ProductsUseCase) GetProduct(in *entity.ProductID) (*entity.Product, error) {
	res, err := p.repo.GetProduct(in)

	if err != nil {
		p.log.Error("GetProduct", "error", err.Error())
		return nil, err
	}

	return res, nil
}

func (p *ProductsUseCase) GetProductList(in *entity.FilterProduct) (*entity.ProductList, error) {
	res, err := p.repo.GetProductList(in)

	if err != nil {
		p.log.Error("GetProductList", "error", err.Error())
		return nil, err
	}

	return res, nil
}
