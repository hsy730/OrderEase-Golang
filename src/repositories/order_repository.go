package repositories

import (
	"errors"
	"orderease/models"
	"orderease/utils/log2"

	"github.com/bwmarrin/snowflake"
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
func (r *OrderRepository) GetOrderByIDAndShopID(orderID uint64, shopID snowflake.ID) (*models.Order, error) {
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
func (r *OrderRepository) GetOrderByIDAndShopIDStr(orderID string, shopID snowflake.ID) (*models.Order, error) {
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
func (r *OrderRepository) GetOrdersByShop(shopID snowflake.ID, page, pageSize int) ([]models.Order, int64, error) {
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
func (r *OrderRepository) GetOrdersByUser(userID string, shopID snowflake.ID, page, pageSize int) ([]models.Order, int64, error) {
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
func (r *OrderRepository) GetUnfinishedOrders(shopID snowflake.ID, unfinishedStatuses []int, page, pageSize int) ([]models.Order, int64, error) {
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

// GetByIDStr 根据订单ID（字符串）获取订单
func (r *OrderRepository) GetByIDStr(orderID string) (*models.Order, error) {
	var order models.Order
	err := r.DB.First(&order, orderID).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.New("订单不存在")
	}
	if err != nil {
		log2.Errorf("GetByIDStr failed: %v", err)
		return nil, errors.New("查询订单失败")
	}
	return &order, nil
}

// GetByIDStrWithItems 根据订单ID（字符串）获取订单（预加载Items和Options）
func (r *OrderRepository) GetByIDStrWithItems(orderID string) (*models.Order, error) {
	var order models.Order
	err := r.DB.Preload("Items").
		Preload("Items.Options").
		First(&order, orderID).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.New("订单不存在")
	}
	if err != nil {
		log2.Errorf("GetByIDStrWithItems failed: %v", err)
		return nil, errors.New("查询订单失败")
	}
	return &order, nil
}

// AdvanceSearchOrderRequest 高级搜索请求参数
type AdvanceSearchOrderRequest struct {
	Page      int
	PageSize  int
	UserID    string
	Status    []int
	StartTime string
	EndTime   string
	ShopID    snowflake.ID
}

// AdvanceSearchResult 高级搜索结果
type AdvanceSearchResult struct {
	Orders []models.Order
	Total  int64
}

// AdvanceSearch 订单高级搜索（支持多条件筛选和分页）
func (r *OrderRepository) AdvanceSearch(req AdvanceSearchOrderRequest) (*AdvanceSearchResult, error) {
	// 构建查询
	query := r.DB.Model(&models.Order{}).Where("shop_id = ?", req.ShopID)

	// 添加用户ID筛选
	if req.UserID != "" {
		query = query.Where("user_id = ?", req.UserID)
	}

	// 添加状态筛选（支持多个状态）
	if len(req.Status) > 0 {
		query = query.Where("status IN (?)", req.Status)
	}

	// 添加时间范围筛选
	if req.StartTime != "" {
		query = query.Where("created_at >= ?", req.StartTime)
	}
	if req.EndTime != "" {
		query = query.Where("created_at <= ?", req.EndTime)
	}

	// 获取总数
	var total int64
	if err := query.Count(&total).Error; err != nil {
		log2.Errorf("AdvanceSearch count failed: %v", err)
		return nil, errors.New("获取订单总数失败")
	}

	// 分页查询
	offset := (req.Page - 1) * req.PageSize
	var orders []models.Order
	if err := query.Offset(offset).Limit(req.PageSize).
		Order("created_at DESC").
		Find(&orders).Error; err != nil {
		log2.Errorf("AdvanceSearch query failed: %v", err)
		return nil, errors.New("查询订单列表失败")
	}

	return &AdvanceSearchResult{
		Orders: orders,
		Total:  total,
	}, nil
}

// CreateOrder 创建订单及其关联数据（事务）
func (r *OrderRepository) CreateOrder(order *models.Order) error {
	tx := r.DB.Begin()

	// 创建订单
	if err := tx.Create(order).Error; err != nil {
		tx.Rollback()
		log2.Errorf("CreateOrder create order failed: %v", err)
		return errors.New("创建订单失败")
	}

	// 更新订单项选项的 OrderItemID
	for i := range order.Items {
		for j := range order.Items[i].Options {
			order.Items[i].Options[j].OrderItemID = order.Items[i].ID
			if err := tx.Save(&order.Items[i].Options[j]).Error; err != nil {
				tx.Rollback()
				log2.Errorf("CreateOrder update option failed: %v", err)
				return errors.New("创建订单失败")
			}
		}
	}

	if err := tx.Commit().Error; err != nil {
		log2.Errorf("CreateOrder commit failed: %v", err)
		return errors.New("创建订单失败")
	}

	return nil
}

// CreateOrderStatusLog 创建订单状态日志
func (r *OrderRepository) CreateOrderStatusLog(statusLog *models.OrderStatusLog) error {
	if err := r.DB.Create(statusLog).Error; err != nil {
		log2.Errorf("CreateOrderStatusLog failed: %v", err)
		return errors.New("创建订单状态日志失败")
	}
	return nil
}

// UpdateOrder 更新订单及其订单项（事务）
func (r *OrderRepository) UpdateOrder(order *models.Order, newItems []models.OrderItem) error {
	tx := r.DB.Begin()

	// 删除原有的订单项
	if err := tx.Where("order_id = ?", order.ID).Delete(&models.OrderItem{}).Error; err != nil {
		tx.Rollback()
		log2.Errorf("UpdateOrder delete old items failed: %v", err)
		return errors.New("删除原有订单项失败")
	}

	// 保存新的订单项
	for i := range newItems {
		newItems[i].OrderID = order.ID
	}
	if err := tx.Create(&newItems).Error; err != nil {
		tx.Rollback()
		log2.Errorf("UpdateOrder create new items failed: %v", err)
		return errors.New("创建新订单项失败")
	}

	// 更新订单信息
	if err := tx.Save(order).Error; err != nil {
		tx.Rollback()
		log2.Errorf("UpdateOrder save order failed: %v", err)
		return errors.New("更新订单信息失败")
	}

	if err := tx.Commit().Error; err != nil {
		log2.Errorf("UpdateOrder commit failed: %v", err)
		return errors.New("更新订单失败")
	}

	return nil
}

// DeleteOrder 删除订单及其关联数据（事务）
func (r *OrderRepository) DeleteOrder(orderID string, shopID snowflake.ID) error {
	tx := r.DB.Begin()

	// 删除订单项
	if err := tx.Where("order_id = ?", orderID).Delete(&models.OrderItem{}).Error; err != nil {
		tx.Rollback()
		log2.Errorf("DeleteOrder delete items failed: %v", err)
		return errors.New("删除订单项失败")
	}

	// 删除订单状态日志
	if err := tx.Where("order_id = ?", orderID).Delete(&models.OrderStatusLog{}).Error; err != nil {
		tx.Rollback()
		log2.Errorf("DeleteOrder delete status logs failed: %v", err)
		return errors.New("删除订单状态日志失败")
	}

	// 删除订单记录
	result := tx.Where("id = ? AND shop_id = ?", orderID, shopID).Delete(&models.Order{})
	if result.Error != nil {
		tx.Rollback()
		log2.Errorf("DeleteOrder delete order failed: %v", result.Error)
		return errors.New("删除订单记录失败")
	}
	if result.RowsAffected == 0 {
		tx.Rollback()
		return errors.New("订单不存在")
	}

	if err := tx.Commit().Error; err != nil {
		log2.Errorf("DeleteOrder commit failed: %v", err)
		return errors.New("删除订单失败")
	}

	return nil
}

// DeleteOrderInTx 在给定事务中删除订单及其关联数据（不提交事务）
func (r *OrderRepository) DeleteOrderInTx(tx *gorm.DB, orderID string, shopID snowflake.ID) error {
	// 删除订单项
	if err := tx.Where("order_id = ?", orderID).Delete(&models.OrderItem{}).Error; err != nil {
		log2.Errorf("DeleteOrderInTx delete items failed: %v", err)
		return errors.New("删除订单项失败")
	}

	// 删除订单状态日志
	if err := tx.Where("order_id = ?", orderID).Delete(&models.OrderStatusLog{}).Error; err != nil {
		log2.Errorf("DeleteOrderInTx delete status logs failed: %v", err)
		return errors.New("删除订单状态日志失败")
	}

	// 删除订单记录
	result := tx.Where("id = ? AND shop_id = ?", orderID, shopID).Delete(&models.Order{})
	if result.Error != nil {
		log2.Errorf("DeleteOrderInTx delete order failed: %v", result.Error)
		return errors.New("删除订单记录失败")
	}
	if result.RowsAffected == 0 {
		return errors.New("订单不存在")
	}

	return nil
}

// UpdateOrderStatusInTx 在给定事务中更新订单状态并创建状态日志
func (r *OrderRepository) UpdateOrderStatusInTx(tx *gorm.DB, order *models.Order, newStatus int) error {
	// 更新订单状态
	if err := tx.Model(order).Update("status", newStatus).Error; err != nil {
		log2.Errorf("UpdateOrderStatusInTx update status failed: %v", err)
		return errors.New("更新订单状态失败")
	}

	// 记录状态变更
	statusLog := models.OrderStatusLog{
		OrderID:     order.ID,
		OldStatus:   order.Status,
		NewStatus:   newStatus,
		ChangedTime: tx.NowFunc(),
	}
	if err := tx.Create(&statusLog).Error; err != nil {
		log2.Errorf("UpdateOrderStatusInTx create status log failed: %v", err)
		return errors.New("记录状态变更失败")
	}

	return nil
}
