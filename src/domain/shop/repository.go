package shop

type ShopRepository interface {
	Save(shop *Shop) error
	FindByID(id uint64) (*Shop, error)
	FindByName(name string) (*Shop, error)
	FindByOwnerUsername(username string) (*Shop, error)
	FindAll(page, pageSize int, search string) ([]Shop, int64, error)
	Delete(id uint64) error
	Update(shop *Shop) error
	Exists(id uint64) (bool, error)
}

type TagRepository interface {
	Save(tag *Tag) error
	FindByID(id int) (*Tag, error)
	FindByShopID(shopID uint64) ([]Tag, error)
	Delete(id int) error
	Update(tag *Tag) error
}
