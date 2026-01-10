package product

import (
	"orderease/domain/shared"
)

type ProductRepository interface {
	Save(product *Product) error
	FindByID(id shared.ID) (*Product, error)
	FindByIDAndShopID(id shared.ID, shopID uint64) (*Product, error)
	FindByShopID(shopID uint64, page, pageSize int, search string, excludeOffline bool) ([]Product, int64, error)
	FindByIDs(ids []shared.ID) ([]Product, error)
	Delete(id shared.ID) error
	Update(product *Product) error
	CountByProductID(productID shared.ID) (int64, error)
	FindOptionByID(id shared.ID) (*ProductOption, error)
	FindOptionCategoryByID(id shared.ID) (*ProductOptionCategory, error)
}

type ProductOptionCategoryRepository interface {
	Save(category *ProductOptionCategory) error
	FindByID(id shared.ID) (*ProductOptionCategory, error)
	FindByProductID(productID shared.ID) ([]ProductOptionCategory, error)
	DeleteByProductID(productID shared.ID) error
}

type ProductOptionRepository interface {
	Save(option *ProductOption) error
	FindByID(id shared.ID) (*ProductOption, error)
	FindByCategoryID(categoryID shared.ID) ([]ProductOption, error)
	DeleteByCategoryID(categoryID shared.ID) error
}

type ProductTagRepository interface {
	Save(productID shared.ID, tagID int) error
	FindByProductID(productID shared.ID) ([]int, error)
	FindByTagID(tagID int) ([]shared.ID, error)
	DeleteByProductID(productID shared.ID) error
}
