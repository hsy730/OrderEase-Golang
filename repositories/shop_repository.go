package repositories

import (
	"errors"
	"orderease/models"
	"orderease/utils/log2"

	"gorm.io/gorm"
)

// 在 ShopRepository 结构体新增方法
func (r *ProductRepository) GetShopByID(shopID uint64) (*models.Shop, error) {
	var shop models.Shop
	err := r.DB.First(&shop, shopID).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.New("店铺不存在")
		}
		log2.Errorf("GetShopByID failed: %v", err)
		return nil, errors.New("服务器内部错误")
	}
	return &shop, nil
}
