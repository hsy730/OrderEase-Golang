package models

import (
	"time"

	"golang.org/x/crypto/bcrypt"
)

// CheckShopPassword 检查店铺密码是否正确
//
// DEPRECATED: 此函数保留是因为 Handler 层仍直接使用 models.Shop
// 未来应该使用 domain/shop 实体的方法
func CheckShopPassword(shop *Shop, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(shop.OwnerPassword), []byte(password))
}

// HashShopPassword 对店铺密码进行哈希
//
// DEPRECATED: 此函数保留是因为 Handler 层仍直接使用 models.Shop
// 未来应该使用 domain/shop 实体的方法
func HashShopPassword(shop *Shop) error {
	hashed, err := bcrypt.GenerateFromPassword([]byte(shop.OwnerPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	shop.OwnerPassword = string(hashed)
	return nil
}

// IsShopExpired 判断店铺是否到期
//
// DEPRECATED: 此函数保留是因为 Handler 层仍直接使用 models.Shop
// 未来应该使用 domain/shop 实体的方法
func IsShopExpired(shop *Shop) bool {
	now := time.Now().UTC()
	return shop.ValidUntil.Before(now)
}

// 以下函数已被移除：
// - GetShopRemainingDays (未被使用，已删除)

