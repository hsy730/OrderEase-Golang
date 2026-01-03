package repositories

import (
	"errors"
	"orderease/domain/order"
	"orderease/domain/shared"
	"orderease/infrastructure/persistence"
	"orderease/models"
	"orderease/utils/log2"
	"time"

	"gorm.io/gorm"
)

type OrderRepositoryImpl struct {
	db *gorm.DB
}

func NewOrderRepository(db *gorm.DB) order.OrderRepository {
	return &OrderRepositoryImpl{db: db}
}

func (r *OrderRepositoryImpl) Save(ord *order.Order) error {
	model := persistence.OrderToModel(ord)
	if err := r.db.Create(model).Error; err != nil {
		log2.Errorf("保存订单失败: %v", err)
		return errors.New("保存订单失败")
	}
	ord.ID = shared.ID(model.ID)
	return nil
}

func (r *OrderRepositoryImpl) FindByID(id shared.ID) (*order.Order, error) {
	var model models.Order
	if err := r.db.Preload("Items").Preload("Items.Options").First(&model, id.Value()).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("订单不存在")
		}
		log2.Errorf("查询订单失败: %v", err)
		return nil, errors.New("查询订单失败")
	}
	return persistence.OrderToDomain(model), nil
}

func (r *OrderRepositoryImpl) FindByIDAndShopID(id shared.ID, shopID uint64) (*order.Order, error) {
	var model models.Order
	if err := r.db.Where("shop_id = ?", shopID).Preload("Items").Preload("Items.Options").First(&model, id.Value()).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("订单不存在")
		}
		log2.Errorf("查询订单失败: %v", err)
		return nil, errors.New("查询订单失败")
	}
	return persistence.OrderToDomain(model), nil
}

func (r *OrderRepositoryImpl) FindByShopID(shopID uint64, page, pageSize int) ([]order.Order, int64, error) {
	var total int64
	if err := r.db.Model(&models.Order{}).Where("shop_id = ?", shopID).Count(&total).Error; err != nil {
		log2.Errorf("查询订单总数失败: %v", err)
		return nil, 0, errors.New("查询订单总数失败")
	}

	var modelsList []models.Order
	offset := (page - 1) * pageSize
	if err := r.db.Where("shop_id = ?", shopID).Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&modelsList).Error; err != nil {
		log2.Errorf("查询订单列表失败: %v", err)
		return nil, 0, errors.New("查询订单列表失败")
	}

	orders := make([]order.Order, len(modelsList))
	for i, m := range modelsList {
		orders[i] = *persistence.OrderToDomain(m)
	}
	return orders, total, nil
}

func (r *OrderRepositoryImpl) FindByUserID(userID shared.ID, shopID uint64, page, pageSize int) ([]order.Order, int64, error) {
	var total int64
	query := r.db.Model(&models.Order{}).Where("user_id = ? AND shop_id = ?", userID.Value(), shopID)
	if err := query.Count(&total).Error; err != nil {
		log2.Errorf("查询订单总数失败: %v", err)
		return nil, 0, errors.New("查询订单总数失败")
	}

	var modelsList []models.Order
	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&modelsList).Error; err != nil {
		log2.Errorf("查询用户订单失败: %v", err)
		return nil, 0, errors.New("查询用户订单失败")
	}

	orders := make([]order.Order, len(modelsList))
	for i, m := range modelsList {
		orders[i] = *persistence.OrderToDomain(m)
	}
	return orders, total, nil
}

func (r *OrderRepositoryImpl) FindUnfinishedByShopID(shopID uint64, flow order.OrderStatusFlow, page, pageSize int) ([]order.Order, int64, error) {
	unfinishedStatuses := flow.GetUnfinishedStatuses()
	statusInts := make([]int, len(unfinishedStatuses))
	for i, s := range unfinishedStatuses {
		statusInts[i] = int(s)
	}

	var total int64
	if err := r.db.Model(&models.Order{}).Where("shop_id = ? AND status IN (?)", shopID, statusInts).Count(&total).Error; err != nil {
		log2.Errorf("查询未完成订单总数失败: %v", err)
		return nil, 0, errors.New("查询未完成订单总数失败")
	}

	var modelsList []models.Order
	offset := (page - 1) * pageSize
	if err := r.db.Where("shop_id = ? AND status IN (?)", shopID, statusInts).Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&modelsList).Error; err != nil {
		log2.Errorf("查询未完成订单列表失败: %v", err)
		return nil, 0, errors.New("查询未完成订单列表失败")
	}

	orders := make([]order.Order, len(modelsList))
	for i, m := range modelsList {
		orders[i] = *persistence.OrderToDomain(m)
	}
	return orders, total, nil
}

func (r *OrderRepositoryImpl) Search(shopID uint64, userID string, statuses []order.OrderStatus, startTime, endTime time.Time, page, pageSize int) ([]order.Order, int64, error) {
	query := r.db.Model(&models.Order{}).Where("shop_id = ?", shopID)

	if userID != "" {
		query = query.Where("user_id = ?", userID)
	}

	if len(statuses) > 0 {
		statusInts := make([]int, len(statuses))
		for i, s := range statuses {
			statusInts[i] = int(s)
		}
		query = query.Where("status IN (?)", statusInts)
	}

	if !startTime.IsZero() {
		query = query.Where("created_at >= ?", startTime)
	}

	if !endTime.IsZero() {
		query = query.Where("created_at <= ?", endTime)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		log2.Errorf("查询订单总数失败: %v", err)
		return nil, 0, errors.New("查询订单总数失败")
	}

	var modelsList []models.Order
	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&modelsList).Error; err != nil {
		log2.Errorf("查询订单列表失败: %v", err)
		return nil, 0, errors.New("查询订单列表失败")
	}

	orders := make([]order.Order, len(modelsList))
	for i, m := range modelsList {
		orders[i] = *persistence.OrderToDomain(m)
	}
	return orders, total, nil
}

func (r *OrderRepositoryImpl) Delete(id shared.ID) error {
	if err := r.db.Delete(&models.Order{}, id.Value()).Error; err != nil {
		log2.Errorf("删除订单失败: %v", err)
		return errors.New("删除订单失败")
	}
	return nil
}

func (r *OrderRepositoryImpl) Update(ord *order.Order) error {
	model := persistence.OrderToModel(ord)
	if err := r.db.Save(model).Error; err != nil {
		log2.Errorf("更新订单失败: %v", err)
		return errors.New("更新订单失败")
	}
	return nil
}

type OrderItemRepositoryImpl struct {
	db *gorm.DB
}

func NewOrderItemRepository(db *gorm.DB) order.OrderItemRepository {
	return &OrderItemRepositoryImpl{db: db}
}

func (r *OrderItemRepositoryImpl) Save(item *order.OrderItem) error {
	model := persistence.OrderItemToModel(*item)
	if err := r.db.Create(&model).Error; err != nil {
		log2.Errorf("保存订单项失败: %v", err)
		return errors.New("保存订单项失败")
	}
	item.ID = shared.ID(model.ID)
	return nil
}

func (r *OrderItemRepositoryImpl) FindByOrderID(orderID shared.ID) ([]order.OrderItem, error) {
	var modelsList []models.OrderItem
	if err := r.db.Preload("Options").Where("order_id = ?", orderID.Value()).Find(&modelsList).Error; err != nil {
		log2.Errorf("查询订单项失败: %v", err)
		return nil, errors.New("查询订单项失败")
	}

	items := make([]order.OrderItem, len(modelsList))
	for i, m := range modelsList {
		items[i] = *persistence.OrderItemToDomain(m)
	}
	return items, nil
}

func (r *OrderItemRepositoryImpl) DeleteByOrderID(orderID shared.ID) error {
	if err := r.db.Where("order_id = ?", orderID.Value()).Delete(&models.OrderItem{}).Error; err != nil {
		log2.Errorf("删除订单项失败: %v", err)
		return errors.New("删除订单项失败")
	}
	return nil
}

type OrderItemOptionRepositoryImpl struct {
	db *gorm.DB
}

func NewOrderItemOptionRepository(db *gorm.DB) order.OrderItemOptionRepository {
	return &OrderItemOptionRepositoryImpl{db: db}
}

func (r *OrderItemOptionRepositoryImpl) Save(option *order.OrderItemOption) error {
	model := persistence.OrderItemOptionToModel(*option)
	if err := r.db.Create(&model).Error; err != nil {
		log2.Errorf("保存订单项选项失败: %v", err)
		return errors.New("保存订单项选项失败")
	}
	option.ID = shared.ID(model.ID)
	return nil
}

func (r *OrderItemOptionRepositoryImpl) FindByOrderItemID(orderItemID shared.ID) ([]order.OrderItemOption, error) {
	var modelsList []models.OrderItemOption
	if err := r.db.Where("order_item_id = ?", orderItemID.Value()).Find(&modelsList).Error; err != nil {
		log2.Errorf("查询订单项选项失败: %v", err)
		return nil, errors.New("查询订单项选项失败")
	}

	options := make([]order.OrderItemOption, len(modelsList))
	for i, m := range modelsList {
		options[i] = *persistence.OrderItemOptionToDomain(m)
	}
	return options, nil
}

func (r *OrderItemOptionRepositoryImpl) DeleteByOrderItemID(orderItemID shared.ID) error {
	if err := r.db.Where("order_item_id = ?", orderItemID.Value()).Delete(&models.OrderItemOption{}).Error; err != nil {
		log2.Errorf("删除订单项选项失败: %v", err)
		return errors.New("删除订单项选项失败")
	}
	return nil
}

type OrderStatusLogRepositoryImpl struct {
	db *gorm.DB
}

func NewOrderStatusLogRepository(db *gorm.DB) order.OrderStatusLogRepository {
	return &OrderStatusLogRepositoryImpl{db: db}
}

func (r *OrderStatusLogRepositoryImpl) Save(log *order.OrderStatusLog) error {
	model := persistence.OrderStatusLogToModel(log)
	if err := r.db.Create(model).Error; err != nil {
		log2.Errorf("保存订单状态日志失败: %v", err)
		return errors.New("保存订单状态日志失败")
	}
	log.ID = shared.ID(model.ID)
	return nil
}

func (r *OrderStatusLogRepositoryImpl) FindByOrderID(orderID shared.ID) ([]order.OrderStatusLog, error) {
	var modelsList []models.OrderStatusLog
	if err := r.db.Where("order_id = ?", orderID.Value()).Order("changed_time DESC").Find(&modelsList).Error; err != nil {
		log2.Errorf("查询订单状态日志失败: %v", err)
		return nil, errors.New("查询订单状态日志失败")
	}

	logs := make([]order.OrderStatusLog, len(modelsList))
	for i, m := range modelsList {
		logs[i] = *persistence.OrderStatusLogToDomain(m)
	}
	return logs, nil
}

func (r *OrderStatusLogRepositoryImpl) DeleteByOrderID(orderID shared.ID) error {
	if err := r.db.Where("order_id = ?", orderID.Value()).Delete(&models.OrderStatusLog{}).Error; err != nil {
		log2.Errorf("删除订单状态日志失败: %v", err)
		return errors.New("删除订单状态日志失败")
	}
	return nil
}
