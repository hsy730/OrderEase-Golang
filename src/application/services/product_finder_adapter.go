package services

import (
	"orderease/domain/order"
	"orderease/domain/product"
	"orderease/domain/shared"
)

// ProductFinderAdapter 实现 order.ProductFinder 接口
// 将 Repository 接口适配为 Domain Service 需要的接口
type ProductFinderAdapter struct {
	productRepo   product.ProductRepository
	optionRepo    product.ProductOptionRepository
	categoryRepo  product.ProductOptionCategoryRepository
}

// NewProductFinderAdapter 创建 ProductFinder 适配器
func NewProductFinderAdapter(
	productRepo product.ProductRepository,
	optionRepo product.ProductOptionRepository,
	categoryRepo product.ProductOptionCategoryRepository,
) order.ProductFinder {
	return &ProductFinderAdapter{
		productRepo:  productRepo,
		optionRepo:   optionRepo,
		categoryRepo: categoryRepo,
	}
}

// FindProduct 查找商品
func (a *ProductFinderAdapter) FindProduct(id shared.ID) (*product.Product, error) {
	return a.productRepo.FindByID(id)
}

// FindOption 查找商品选项
func (a *ProductFinderAdapter) FindOption(id shared.ID) (*product.ProductOption, error) {
	return a.optionRepo.FindByID(id)
}

// FindOptionCategory 查找选项类别
func (a *ProductFinderAdapter) FindOptionCategory(id shared.ID) (*product.ProductOptionCategory, error) {
	return a.categoryRepo.FindByID(id)
}
