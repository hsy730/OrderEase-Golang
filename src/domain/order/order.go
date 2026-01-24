package order

import (
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
