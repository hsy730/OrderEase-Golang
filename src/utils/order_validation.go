package utils

import (
	"fmt"

	"orderease/models"
	"gorm.io/gorm"
)

// ValidateOrderItems 验证订单项的基础数据
// 不涉及数据库查询，只验证数据完整性
func ValidateOrderItems(items []models.OrderItem) error {
	if len(items) == 0 {
		return fmt.Errorf("订单项不能为空")
	}

	for _, item := range items {
		if item.ProductID == 0 {
			return fmt.Errorf("商品ID不能为空")
		}
		if item.Quantity <= 0 {
			return fmt.Errorf("商品数量必须大于0")
		}
	}

	return nil
}

// ValidateProductStock 验证商品库存是否充足
// 需要传入数据库事务来查询商品库存
func ValidateProductStock(tx *gorm.DB, items []models.OrderItem) error {
	for _, item := range items {
		var product models.Product
		if err := tx.First(&product, item.ProductID).Error; err != nil {
			return fmt.Errorf("商品不存在, ID: %d", item.ProductID)
		}

		if product.Stock < item.Quantity {
			return fmt.Errorf("商品 %s 库存不足, 当前库存: %d, 需求数量: %d",
				product.Name, product.Stock, item.Quantity)
		}
	}

	return nil
}

// DeductProductStock 扣减商品库存
// 需要传入数据库事务来更新商品库存
func DeductProductStock(tx *gorm.DB, items []models.OrderItem) error {
	for _, item := range items {
		var product models.Product
		if err := tx.First(&product, item.ProductID).Error; err != nil {
			return fmt.Errorf("商品不存在, ID: %d", item.ProductID)
		}

		// 扣减库存
		product.Stock -= item.Quantity
		if err := tx.Save(&product).Error; err != nil {
			return fmt.Errorf("更新商品库存失败: %v", err)
		}
	}

	return nil
}

// RestoreProductStock 恢复商品库存（用于订单取消时）
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

// CalculateOrderTotal 计算订单总价
// 注意：这个函数不包含商品选项的价格调整
// 完整的价格计算需要在 CreateOrder 中处理商品选项
func CalculateOrderTotal(items []models.OrderItem) float64 {
	total := float64(0)
	for _, item := range items {
		// 使用订单项中已经保存的价格（如果已设置）
		if item.Price != 0 {
			total += float64(item.Quantity) * float64(item.Price)
		}
	}
	return total
}

// ValidateOrder 验证订单基础数据
func ValidateOrder(order *models.Order) error {
	if order.UserID == 0 {
		return fmt.Errorf("用户ID不能为空")
	}

	if order.ShopID == 0 {
		return fmt.Errorf("店铺ID不能为空")
	}

	return ValidateOrderItems(order.Items)
}
