package order

import (
	"errors"
	"fmt"
	"time"

	"orderease/domain/shared"
)

type OrderStatus int

const (
	OrderStatusPending  OrderStatus = 1
	OrderStatusAccepted OrderStatus = 2
	OrderStatusRejected OrderStatus = 3
	OrderStatusShipped  OrderStatus = 4
	OrderStatusComplete OrderStatus = 10
	OrderStatusCanceled OrderStatus = -1
)

func (s OrderStatus) String() string {
	switch s {
	case OrderStatusPending:
		return "待处理"
	case OrderStatusAccepted:
		return "已接单"
	case OrderStatusRejected:
		return "已拒绝"
	case OrderStatusShipped:
		return "已发货"
	case OrderStatusComplete:
		return "已完成"
	case OrderStatusCanceled:
		return "已取消"
	default:
		return "未知状态"
	}
}

func (s OrderStatus) IsFinal() bool {
	return s == OrderStatusComplete || s == OrderStatusRejected || s == OrderStatusCanceled
}

type OrderStatusTransition struct {
	Name            string
	NextStatus      OrderStatus
	NextStatusLabel string
}

type OrderStatusConfig struct {
	Value   OrderStatus
	Label   string
	Type    string
	IsFinal bool
	Actions []OrderStatusTransition
}

type OrderStatusFlow struct {
	Statuses []OrderStatusConfig
}

func (flow *OrderStatusFlow) CanTransition(from, to OrderStatus) bool {
	for _, status := range flow.Statuses {
		if status.Value == from {
			if status.IsFinal {
				return false
			}
			for _, action := range status.Actions {
				if action.NextStatus == to {
					return true
				}
			}
		}
	}
	return false
}

func (flow *OrderStatusFlow) GetUnfinishedStatuses() []OrderStatus {
	var statuses []OrderStatus
	for _, status := range flow.Statuses {
		if !status.IsFinal {
			statuses = append(statuses, status.Value)
		}
	}
	return statuses
}

type Order struct {
	ID         shared.ID
	UserID     shared.ID
	ShopID     uint64
	TotalPrice shared.Price
	Status     OrderStatus
	Remark     string
	CreatedAt  time.Time
	UpdatedAt  time.Time
	Items      []OrderItem
}

type OrderItem struct {
	ID                shared.ID
	OrderID           shared.ID
	ProductID         shared.ID
	Quantity          int
	Price             shared.Price
	TotalPrice        shared.Price
	ProductName       string
	ProductDescription string
	ProductImageURL   string
	Options           []OrderItemOption
}

type OrderItemOption struct {
	ID              shared.ID
	OrderItemID     shared.ID
	CategoryID      shared.ID
	OptionID        shared.ID
	OptionName      string
	CategoryName    string
	PriceAdjustment float64
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

type OrderStatusLog struct {
	ID          shared.ID
	OrderID     shared.ID
	OldStatus   OrderStatus
	NewStatus   OrderStatus
	ChangedTime time.Time
}

func NewOrder(userID shared.ID, shopID uint64, items []OrderItem, remark string) (*Order, error) {
	if userID.IsZero() {
		return nil, errors.New("用户ID不能为空")
	}

	if shopID == 0 {
		return nil, errors.New("店铺ID不能为空")
	}

	if len(items) == 0 {
		return nil, errors.New("订单项不能为空")
	}

	for _, item := range items {
		if item.ProductID.IsZero() {
			return nil, errors.New("商品ID不能为空")
		}
		if item.Quantity <= 0 {
			return nil, errors.New("商品数量必须大于0")
		}
	}

	now := time.Now()

	for i := range items {
		items[i].OrderID = shared.ID(0)
		if items[i].TotalPrice.IsZero() {
			items[i].TotalPrice = items[i].Price.Multiply(items[i].Quantity)
		}
	}

	totalPrice := shared.Price(0)
	for _, item := range items {
		totalPrice = totalPrice.Add(item.TotalPrice)
	}

	return &Order{
		ID:         shared.ID(0),
		UserID:     userID,
		ShopID:     shopID,
		TotalPrice: totalPrice,
		Status:     OrderStatusPending,
		Remark:     remark,
		CreatedAt:  now,
		UpdatedAt:  now,
		Items:      items,
	}, nil
}

func (o *Order) CanTransitionTo(newStatus OrderStatus, flow OrderStatusFlow) error {
	if flow.CanTransition(o.Status, newStatus) {
		return nil
	}
	return fmt.Errorf("当前状态 %s 不允许转换到状态 %s", o.Status, newStatus)
}

func (o *Order) TransitionTo(newStatus OrderStatus, flow OrderStatusFlow) error {
	if err := o.CanTransitionTo(newStatus, flow); err != nil {
		return err
	}

	o.Status = newStatus
	o.UpdatedAt = time.Now()

	return nil
}

func (o *Order) IsFinal() bool {
	return o.Status.IsFinal()
}

func (o *Order) IsUnfinished(flow OrderStatusFlow) bool {
	unfinishedStatuses := flow.GetUnfinishedStatuses()
	for _, status := range unfinishedStatuses {
		if o.Status == status {
			return true
		}
	}
	return false
}

func NewOrderItem(productID shared.ID, quantity int, price shared.Price, options []OrderItemOption) OrderItem {
	totalPrice := price.Multiply(quantity)

	for _, option := range options {
		totalPrice = totalPrice.Add(shared.Price(option.PriceAdjustment * float64(quantity)))
	}

	return OrderItem{
		ID:         shared.ID(0),
		ProductID:  productID,
		Quantity:   quantity,
		Price:      price,
		TotalPrice: totalPrice,
		Options:    options,
	}
}

func NewOrderItemOption(categoryID, optionID shared.ID, optionName, categoryName string, priceAdjustment float64) OrderItemOption {
	return OrderItemOption{
		ID:              shared.ID(0),
		CategoryID:      categoryID,
		OptionID:        optionID,
		OptionName:      optionName,
		CategoryName:    categoryName,
		PriceAdjustment: priceAdjustment,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}
}
