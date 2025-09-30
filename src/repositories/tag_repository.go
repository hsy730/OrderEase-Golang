package repositories

import (
	"orderease/models"
	"orderease/utils/log2"

	"github.com/bwmarrin/snowflake"
)

// 在 ProductRepository 结构体新增方法
func (r *ProductRepository) GetCurrentProductTags(productID snowflake.ID, shopID uint64) ([]models.Tag, error) {
	var tags []models.Tag
	err := r.DB.Joins("JOIN product_tags ON product_tags.tag_id = tags.id").
		Where("product_tags.product_id = ? AND product_tags.shop_id = ?", productID, shopID).
		Find(&tags).Error
	if err != nil {
		log2.Errorf("GetCurrentProductTags failed: %v", err)
		return nil, err
	}
	return tags, err
}

func (r *ProductRepository) CheckProductsBelongToShop(productIDs []uint, shopID uint64) ([]uint, error) {
	var validIDs []uint
	err := r.DB.Model(&models.Product{}).
		Where("id IN (?) AND shop_id = ?", productIDs, shopID).
		Pluck("id", &validIDs).Error
	return validIDs, err
}

func (r *ProductRepository) GetShopTagsByID(shopID uint64) ([]models.Tag, error) {
	tags := make([]models.Tag, 0)
	err := r.DB.Where("shop_id = ?", shopID).Find(&tags).Error
	if err != nil {
		return nil, err
	}
	return tags, nil
}
