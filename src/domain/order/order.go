package order

import (
	"errors"
	"fmt"
	"time"

	"orderease/domain/shared"
	"orderease/domain/product"
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

// ============ 领域层增强方法 ============

// SetProductSnapshot 设置商品快照信息
func (oi *OrderItem) SetProductSnapshot(prod *product.Product) {
	oi.ProductName = prod.Name
	oi.ProductDescription = prod.Description
	oi.ProductImageURL = prod.ImageURL
	oi.Price = prod.Price
}

// CalculatePrice 计算订单项总价（含选项价格调整）
func (oi *OrderItem) CalculatePrice(finder ProductFinder) (shared.Price, error) {
	itemTotal := oi.Price.Multiply(oi.Quantity)

	for j := range oi.Options {
		opt, err := finder.FindOption(oi.Options[j].OptionID)
		if err != nil {
			return 0, fmt.Errorf("商品参数选项不存在: %w", err)
		}

		cat, err := finder.FindOptionCategory(opt.CategoryID)
		if err != nil {
			return 0, fmt.Errorf("商品参数类别不存在: %w", err)
		}

		// 设置选项快照
		oi.Options[j].CategoryID = cat.ID
		oi.Options[j].OptionName = opt.Name
		oi.Options[j].CategoryName = cat.Name
		oi.Options[j].PriceAdjustment = opt.PriceAdjustment

		// 累加选项价格
		itemTotal = itemTotal.Add(shared.Price(opt.PriceAdjustment * float64(oi.Quantity)))
	}

	oi.TotalPrice = itemTotal
	return itemTotal, nil
}

// ValidateProduct 验证商品（库存、归属）
func (oi *OrderItem) ValidateProduct(prod *product.Product, shopID uint64) error {
	if prod.ShopID != shopID {
		return fmt.Errorf("商品不属于该店铺")
	}
	if !prod.HasStock(oi.Quantity) {
		return fmt.Errorf("商品 %s 库存不足", prod.Name)
	}
	return nil
}

// ValidateItems 验证所有订单项
func (o *Order) ValidateItems(finder ProductFinder) error {
	for i := range o.Items {
		prod, err := finder.FindProduct(o.Items[i].ProductID)
		if err != nil {
			return fmt.Errorf("商品不存在: %w", err)
		}

		if err := o.Items[i].ValidateProduct(prod, o.ShopID); err != nil {
			return err
		}

		o.Items[i].SetProductSnapshot(prod)
	}
	return nil
}

// CalculateTotal 计算订单总价
func (o *Order) CalculateTotal(finder ProductFinder) error {
	totalPrice := shared.Price(0)

	for i := range o.Items {
		itemTotal, err := o.Items[i].CalculatePrice(finder)
		if err != nil {
			return err
		}
		totalPrice = totalPrice.Add(itemTotal)
	}

	o.TotalPrice = totalPrice
	return nil
}
