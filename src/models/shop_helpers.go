package models

import (
	"time"

	"golang.org/x/crypto/bcrypt"
)

// CheckShopPassword 检查店铺密码是否正确
func CheckShopPassword(shop *Shop, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(shop.OwnerPassword), []byte(password))
}

// HashShopPassword 对店铺密码进行哈希
func HashShopPassword(shop *Shop) error {
	hashed, err := bcrypt.GenerateFromPassword([]byte(shop.OwnerPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	shop.OwnerPassword = string(hashed)
	return nil
}

// IsShopExpired 判断店铺是否到期
func IsShopExpired(shop *Shop) bool {
	now := time.Now().UTC()
	return shop.ValidUntil.Before(now)
}

// GetShopRemainingDays 获取剩余有效天数（负数表示已过期）
func GetShopRemainingDays(shop *Shop) int {
	hours := time.Until(shop.ValidUntil.UTC()).Hours()
	return int(hours / 24) // 向下取整
}
