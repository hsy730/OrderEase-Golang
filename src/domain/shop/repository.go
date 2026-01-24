package shop

import (
	"github.com/bwmarrin/snowflake"
)

// Repository 店铺仓储接口
type Repository interface {
	// Create 创建店铺
	Create(shop *Shop) error

	// GetByID 根据ID获取店铺
	GetByID(id uint64) (*Shop, error)

	// GetByOwnerUsername 根据店主用户名获取店铺
	GetByOwnerUsername(username string) (*Shop, error)

	// GetByName 根据店铺名称获取店铺
	GetByName(name string) (*Shop, error)

	// GetList 获取店铺列表
	GetList(page int, pageSize int) ([]*Shop, int64, error)

	// Update 更新店铺
	Update(shop *Shop) error

	// Delete 删除店铺
	Delete(shop *Shop) error

	// Exists 检查店铺是否存在
	Exists(id uint64) (bool, error)

	// NameExists 检查店铺名称是否存在
	NameExists(name string) (bool, error)

	// OwnerUsernameExists 检查店主用户名是否存在
	OwnerUsernameExists(username string) (bool, error)
}
