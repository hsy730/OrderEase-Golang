package order

import (
	"github.com/bwmarrin/snowflake"
)

// Repository 订单仓储接口
type Repository interface {
	// Create 创建订单
	Create(order *Order) error

	// GetByID 根据ID获取订单
	GetByID(id snowflake.ID) (*Order, error)

	// GetByShopID 获取店铺的订单列表
	GetByShopID(shopID uint64, page int, pageSize int) ([]*Order, int64, error)

	// GetByUserID 获取用户的订单列表
	GetByUserID(userID snowflake.ID, page int, pageSize int) ([]*Order, int64, error)

	// Update 更新订单
	Update(order *Order) error

	// Delete 删除订单
	Delete(order *Order) error

	// Exists 检查订单是否存在
	Exists(id snowflake.ID) (bool, error)
}
