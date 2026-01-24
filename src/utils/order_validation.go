package utils

import (
	"fmt"

	"orderease/models"
	"gorm.io/gorm"
)

// RestoreProductStock 恢复商品库存（用于订单取消时）
//
// DEPRECATED: 此函数将被 domain/order/service.RestoreStock 替代
// 保留此函数是因为 Handler 层仍直接使用 models.Order
// 未来 Handler 迁移到使用领域实体后，应该使用 domain order service
//
// 需要传入数据库事务来更新商品库存
func RestoreProductStock(tx *gorm.DB, order models.Order) error {
	for _, item := range order.Items {
		var product models.Product
		if err := tx.First(&product, item.ProductID).Error; err != nil {
			return fmt.Errorf("商品不存在, ID: %d", item.ProductID)
		}

		// 恢复库存
		product.Stock += item.Quantity
		if err := tx.Save(&product).Error; err != nil {
			return fmt.Errorf("恢复商品库存失败: %v", err)
		}
	}

	return nil
}

// ValidateOrder 验证订单基础数据
//
// DEPRECATED: 简单验证逻辑，未来应该使用 domain/order/service.ValidateOrder
// 保留此函数是因为 Handler 层仍直接使用 models.Order
func ValidateOrder(order *models.Order) error {
	if order.UserID == 0 {
		return fmt.Errorf("用户ID不能为空")
	}

	if order.ShopID == 0 {
		return fmt.Errorf("店铺ID不能为空")
	}

	if len(order.Items) == 0 {
		return fmt.Errorf("订单项不能为空")
	}

	for _, item := range order.Items {
		if item.ProductID == 0 {
			return fmt.Errorf("商品ID不能为空")
		}
		if item.Quantity <= 0 {
			return fmt.Errorf("商品数量必须大于0")
		}
	}

	return nil
}

// 以下函数已被 domain/order/service 替代，不再需要：
// - ValidateOrderItems (逻辑已合并到 ValidateOrder)
// - ValidateProductStock (逻辑已迁移到 domain service)
// - DeductProductStock (逻辑已迁移到 domain service)
// - CalculateOrderTotal (逻辑已迁移到 domain service)

