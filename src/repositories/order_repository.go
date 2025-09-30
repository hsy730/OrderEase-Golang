package repositories

import (
	"errors"
	"orderease/models"
	"orderease/utils/log2"
)

// 在 OrderRepository 结构体新增方法
func (r *ProductRepository) GetOrderByIDAndShopID(orderID uint64, shopID uint64) (*models.Order, error) {
	var order models.Order
	err := r.DB.Where("shop_id = ?", shopID).First(&order, orderID).Error
	if err != nil {
		if err.Error() == "record not found" {
			return nil, errors.New("订单不存在")
		}
		log2.Errorf("GetOrderByIDAndShopID failed: %v", err)
		return nil, errors.New("服务器内部错误")
	}
	return &order, nil
}
