package models

import (
	"golang.org/x/crypto/bcrypt"
)

// HashShopPassword 对店铺密码进行哈希
//
// DEPRECATED: 此函数保留是因为 Handler 层仍直接使用 models.Shop
// 未来应该使用 domain/shop 实体的 ToModel() 方法
// 注意：CheckShopPassword 和 IsShopExpired 已被删除，请使用 domain/shop 实体方法
func HashShopPassword(shop *Shop) error {
	hashed, err := bcrypt.GenerateFromPassword([]byte(shop.OwnerPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	shop.OwnerPassword = string(hashed)
	return nil
}

// 以下函数已被移除：
// - CheckShopPassword (已迁移到 shop.CheckPassword)
// - IsShopExpired (已迁移到 shop.IsExpired)
// - GetShopRemainingDays (未被使用，已删除)

