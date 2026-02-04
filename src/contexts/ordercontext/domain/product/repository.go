package product

import (
	"github.com/bwmarrin/snowflake"
)

// Repository 商品仓储接口
type Repository interface {
	// Create 创建商品
	Create(product *Product) error

	// GetByID 根据ID获取商品
	GetByID(id snowflake.ID) (*Product, error)

	// GetByShopID 获取店铺的商品列表
	GetByShopID(shopID uint64, page int, pageSize int) ([]*Product, int64, error)

	// GetByName 根据商品名称获取商品
	GetByName(shopID uint64, name string) (*Product, error)

	// Update 更新商品
	Update(product *Product) error

	// Delete 删除商品
	Delete(product *Product) error

	// Exists 检查商品是否存在
	Exists(id snowflake.ID) (bool, error)

	// UpdateStock 更新库存
	UpdateStock(id snowflake.ID, delta int) error

	// UpdateStatus 更新状态
	UpdateStatus(id snowflake.ID, status string) error
}
