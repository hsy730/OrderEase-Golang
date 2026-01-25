package order

import (
	"errors"
	"fmt"
	"time"

	"github.com/bwmarrin/snowflake"
	"orderease/domain/shared/value_objects"
	"orderease/models"
)

// Order 订单聚合根
type Order struct {
	id         snowflake.ID
	userID     snowflake.ID
	shopID     uint64
	totalPrice models.Price
	status     value_objects.OrderStatus
	remark     string
	items      []OrderItem
	createdAt  time.Time
	updatedAt  time.Time
}

// NewOrder 创建新订单
func NewOrder(userID snowflake.ID, shopID uint64) *Order {
	return &Order{
		id:        snowflake.ID(0), // 将在持久化时生成
		userID:    userID,
		shopID:    shopID,
		status:    value_objects.OrderStatusPending,
		createdAt: time.Now(),
		updatedAt: time.Now(),
	}
}

// Getters
func (o *Order) ID() snowflake.ID {
	return o.id
}

func (o *Order) UserID() snowflake.ID {
	return o.userID
}

func (o *Order) ShopID() uint64 {
	return o.shopID
}

func (o *Order) TotalPrice() models.Price {
	return o.totalPrice
}

func (o *Order) Status() value_objects.OrderStatus {
	return o.status
}

func (o *Order) Remark() string {
	return o.remark
}

func (o *Order) Items() []OrderItem {
	return o.items
}

func (o *Order) CreatedAt() time.Time {
	return o.createdAt
}

func (o *Order) UpdatedAt() time.Time {
	return o.updatedAt
}

// Setters
func (o *Order) SetID(id snowflake.ID) {
	o.id = id
}

func (o *Order) SetTotalPrice(price models.Price) {
	o.totalPrice = price
}

func (o *Order) SetStatus(status value_objects.OrderStatus) {
	o.status = status
}

func (o *Order) SetRemark(remark string) {
	o.remark = remark
}

func (o *Order) SetItems(items []OrderItem) {
	o.items = items
}

func (o *Order) SetCreatedAt(t time.Time) {
	o.createdAt = t
}

func (o *Order) SetUpdatedAt(t time.Time) {
	o.updatedAt = t
}

// AddItem 添加订单项
func (o *Order) AddItem(item OrderItem) {
	o.items = append(o.items, item)
}

// ToModel 转换为持久化模型
func (o *Order) ToModel() *models.Order {
	modelItems := make([]models.OrderItem, len(o.items))
	for i, item := range o.items {
		modelItems[i] = *item.ToModel(o.id)
	}

	return &models.Order{
		ID:         o.id,
		UserID:     o.userID,
		ShopID:     o.shopID,
		TotalPrice: o.totalPrice,
		Status:     int(o.status),
		Remark:     o.remark,
		CreatedAt:  o.createdAt,
		UpdatedAt:  o.updatedAt,
		Items:      modelItems,
	}
}

// OrderFromModel 从持久化模型创建领域实体
func OrderFromModel(model *models.Order) *Order {
	items := make([]OrderItem, len(model.Items))
	for i, item := range model.Items {
		items[i] = *OrderItemFromModel(&item)
	}

	return &Order{
		id:         model.ID,
		userID:     model.UserID,
		shopID:     model.ShopID,
		totalPrice: model.TotalPrice,
		status:     value_objects.OrderStatusFromInt(model.Status),
		remark:     model.Remark,
		items:      items,
		createdAt:  model.CreatedAt,
		updatedAt:  model.UpdatedAt,
	}
}

// ==================== 业务方法 ====================

// ValidateItems 验证订单项
func (o *Order) ValidateItems() error {
	if len(o.items) == 0 {
		return errors.New("订单项不能为空")
	}

	for _, item := range o.items {
		if item.productID == 0 {
			return errors.New("商品ID不能为空")
		}
		if item.quantity <= 0 {
			return fmt.Errorf("商品数量必须大于0，当前值: %d", item.quantity)
		}
	}

	return nil
}

// CalculateTotal 计算订单总价
// 根据订单项的数量、价格和选项计算总价
func (o *Order) CalculateTotal() models.Price {
	total := float64(0)

	for _, item := range o.items {
		// 基础价格：商品单价 × 数量
		itemTotal := float64(item.quantity) * float64(item.price)

		// 加上选项价格调整
		for _, opt := range item.options {
			itemTotal += float64(item.quantity) * opt.PriceAdjustment
		}

		total += itemTotal
	}

	return models.Price(total)
}

// CanTransitionTo 判断是否可以转换到目标状态
func (o *Order) CanTransitionTo(to value_objects.OrderStatus) bool {
	return o.status.CanTransitionTo(to)
}

// IsFinal 判断订单是否处于最终状态
func (o *Order) IsFinal() bool {
	return o.status.IsFinal()
}

// GetItemCount 获取订单项数量
func (o *Order) GetItemCount() int {
	return len(o.items)
}

// GetTotalQuantity 获取商品总数量
func (o *Order) GetTotalQuantity() int {
	total := 0
	for _, item := range o.items {
		total += item.quantity
	}
	return total
}

// IsPending 检查订单是否为待处理状态
func (o *Order) IsPending() bool {
	return o.status == value_objects.OrderStatusPending
}

// CanBeDeleted 检查订单是否可以删除
// 订单可以删除的条件：未取消且未完成
func (o *Order) CanBeDeleted() bool {
	return o.status != value_objects.OrderStatusCanceled &&
		o.status != value_objects.OrderStatusComplete
}

// HasItems 检查订单是否包含商品项
func (o *Order) HasItems() bool {
	return len(o.items) > 0
}

// IsEmpty 检查订单是否为空（无商品项）
func (o *Order) IsEmpty() bool {
	return len(o.items) == 0
}
