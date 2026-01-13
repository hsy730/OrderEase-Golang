package shop

import "orderease/domain/shared"

type ShopRepository interface {
	Save(shop *Shop) error
	FindByID(id shared.ID) (*Shop, error)
	FindByName(name string) (*Shop, error)
	FindByOwnerUsername(username string) (*Shop, error)
	FindAll(page, pageSize int, search string) ([]Shop, int64, error)
	Delete(id shared.ID) error
	Update(shop *Shop) error
	Exists(id shared.ID) (bool, error)
}

type TagRepository interface {
	Save(tag *Tag) error
	FindByID(id int) (*Tag, error)
	FindByShopID(shopID shared.ID) ([]Tag, error)
	Delete(id int) error
	Update(tag *Tag) error
}
