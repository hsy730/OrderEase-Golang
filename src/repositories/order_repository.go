package repositories

import (
	"errors"
	"orderease/models"
	"orderease/utils/log2"

	"gorm.io/gorm"
)

// OrderRepository 订单数据访问层
type OrderRepository struct {
	DB *gorm.DB
}

// NewOrderRepository 创建OrderRepository实例
func NewOrderRepository(db *gorm.DB) *OrderRepository {
	return &OrderRepository{DB: db}
}

// GetOrderByIDAndShopID 根据订单ID和店铺ID获取订单（预加载Items和Options）
func (r *OrderRepository) GetOrderByIDAndShopID(orderID uint64, shopID uint64) (*models.Order, error) {
	var order models.Order
	// 预加载Items和Items.Options
	err := r.DB.Preload("Items").
		Preload("Items.Options").
		Where("shop_id = ?", shopID).
		Joins("User").
		First(&order, orderID).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.New("订单不存在")
	}
	if err != nil {
		log2.Errorf("GetOrderByIDAndShopID failed: %v", err)
		return nil, errors.New("查询订单失败")
	}
	return &order, nil
}

// GetOrderByIDAndShopIDStr 根据订单ID（字符串）和店铺ID获取订单（预加载Items和Options）
func (r *OrderRepository) GetOrderByIDAndShopIDStr(orderID string, shopID uint64) (*models.Order, error) {
	var order models.Order
	// 预加载Items和Items.Options
	err := r.DB.Preload("Items").
		Preload("Items.Options").
		Where("shop_id = ?", shopID).
		Joins("User").
		First(&order, orderID).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.New("订单不存在")
	}
	if err != nil {
		log2.Errorf("GetOrderByIDAndShopIDStr failed: %v", err)
		return nil, errors.New("查询订单失败")
	}
	return &order, nil
}

// GetOrdersByShop 获取店铺的订单列表（分页）
func (r *OrderRepository) GetOrdersByShop(shopID uint64, page, pageSize int) ([]models.Order, int64, error) {
	var orders []models.Order
	var total int64

	baseQuery := r.DB.Model(&models.Order{}).Where("shop_id = ?", shopID)

	// 获取总数
	if err := baseQuery.Count(&total).Error; err != nil {
		log2.Errorf("GetOrdersByShop count failed: %v", err)
		return nil, 0, errors.New("获取订单总数失败")
	}

	// 分页查询，预加载关联数据
	offset := (page - 1) * pageSize
	if err := baseQuery.Preload("User").
		Preload("Items").
		Offset(offset).
		Limit(pageSize).
		Order("created_at DESC").
		Find(&orders).Error; err != nil {
		log2.Errorf("GetOrdersByShop query failed: %v", err)
		return nil, 0, errors.New("查询订单列表失败")
	}

	return orders, total, nil
}

// GetOrdersByUser 获取用户的订单列表（分页）
func (r *OrderRepository) GetOrdersByUser(userID string, shopID uint64, page, pageSize int) ([]models.Order, int64, error) {
	var orders []models.Order
	var total int64

	baseQuery := r.DB.Model(&models.Order{}).Where("user_id = ? AND shop_id = ?", userID, shopID)

	// 获取总数
	if err := baseQuery.Count(&total).Error; err != nil {
		log2.Errorf("GetOrdersByUser count failed: %v", err)
		return nil, 0, errors.New("获取订单总数失败")
	}

	// 分页查询
	offset := (page - 1) * pageSize
	if err := baseQuery.Preload("Items").
		Preload("Items.Options").
		Offset(offset).
		Limit(pageSize).
		Order("created_at DESC").
		Find(&orders).Error; err != nil {
		log2.Errorf("GetOrdersByUser query failed: %v", err)
		return nil, 0, errors.New("查询订单列表失败")
	}

	return orders, total, nil
}

// GetUnfinishedOrders 获取未完成的订单列表
func (r *OrderRepository) GetUnfinishedOrders(shopID uint64, unfinishedStatuses []int, page, pageSize int) ([]models.Order, int64, error) {
	var orders []models.Order
	var total int64

	baseQuery := r.DB.Model(&models.Order{}).Where("shop_id = ? AND status IN ?", shopID, unfinishedStatuses)

	// 获取总数
	if err := baseQuery.Count(&total).Error; err != nil {
		log2.Errorf("GetUnfinishedOrders count failed: %v", err)
		return nil, 0, errors.New("获取订单总数失败")
	}

	// 分页查询
	offset := (page - 1) * pageSize
	if err := baseQuery.Preload("User").
		Preload("Items").
		Offset(offset).
		Limit(pageSize).
		Order("created_at DESC").
		Find(&orders).Error; err != nil {
		log2.Errorf("GetUnfinishedOrders query failed: %v", err)
		return nil, 0, errors.New("查询订单列表失败")
	}

	return orders, total, nil
}
