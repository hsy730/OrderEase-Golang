// Package order 提供订单领域模型的核心业务逻辑。
//
// 职责：
//   - 订单生命周期管理（创建、更新、删除）
//   - 订单状态流转控制
//   - 订单项和价格计算
//   - 订单业务规则验证
//
// 业务规则：
//   - 订单必须包含至少一个订单项
//   - 已完成/已取消的订单不可删除
//   - 订单总价 = Σ(商品单价 × 数量 + 选项价格调整 × 数量)
//   - 订单状态流转必须遵循店铺配置的 OrderStatusFlow
//
// 使用示例：
//
//	// 从持久化模型创建领域实体
//	order := order.OrderFromModel(orderModel)
//
//	// 执行业务验证
//	if err := order.ValidateItems(); err != nil {
//	    return err
//	}
//
//	// 检查状态流转是否合法
//	if !order.CanTransitionTo(targetStatus) {
//	    return errors.New("状态流转不合法")
//	}
package order

import (
	"errors"
	"fmt"
	"time"

	"github.com/bwmarrin/snowflake"
	"orderease/contexts/ordercontext/domain/shared/value_objects"
	"orderease/models"
)

// Order 订单聚合根
//
// 作为聚合根，Order 负责：
//   - 管理订单自身的生命周期和状态
//   - 维护订单项 (OrderItem) 集合的一致性
//   - 提供价格计算和状态流转验证
//
// 约束：
//   - ID 为 0 表示未持久化的新订单
//   - items 切片在创建后可通过 AddItem 追加
//   - 所有修改必须通过业务方法进行，避免直接修改字段
type Order struct {
	id         snowflake.ID
	userID     snowflake.ID
	shopID     snowflake.ID
	totalPrice models.Price
	status     value_objects.OrderStatus
	remark     string
	items      []OrderItem
	createdAt  time.Time
	updatedAt  time.Time
}

// NewOrder 创建新订单
//
// 参数：
//   - userID: 下单用户ID，必须有效
//   - shopID: 所属店铺ID，必须有效
//
// 返回：
//   - 初始状态为 OrderStatusPending 的订单实体
//   - ID 为 0，表示未持久化，需在持久化时分配
//
// 注意：返回的订单不包含订单项，需通过 AddItem 添加
func NewOrder(userID snowflake.ID, shopID snowflake.ID) *Order {
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

func (o *Order) ShopID() snowflake.ID {
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

// ValidateItems 验证订单项的有效性
//
// 验证规则：
//   - 订单项不能为空（至少包含一个）
//   - 每个订单项的商品ID必须有效（非0）
//   - 每个订单项的数量必须大于0
//
// 使用时机：
//   - 创建订单前
//   - 更新订单前
//   - 从外部数据构建订单后
//
// 错误示例：
//   - "订单项不能为空"
//   - "商品ID不能为空"
//   - "商品数量必须大于0，当前值: -1"
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
//
// 计算公式：
//   总价 = Σ(商品单价 × 数量 + Σ(选项价格调整 × 数量))
//
// 示例：
//   商品A: 价格=100, 数量=2, 选项=[加价10, 加价5]
//   小计 = 100×2 + (10+5)×2 = 230
//
// 注意：此方法仅计算，不修改订单状态
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
//
// 使用 OrderStatus 值对象的规则进行验证。
// 状态流转规则由店铺配置的 OrderStatusFlow 决定。
//
// 常见流转：
//   Pending -> Accepted -> Preparing -> Ready -> Completed
//   Pending -> Canceled (任意状态可取消)
//
// 注意：终态（Completed/Canceled）不允许再流转
func (o *Order) CanTransitionTo(to value_objects.OrderStatus) bool {
	return o.status.CanTransitionTo(to)
}

// IsFinal 判断订单是否处于最终状态
//
// 终态定义：Completed（已完成）、Canceled（已取消）
// 终态订单不允许删除和状态流转。
//
// 使用场景：
//   - 删除订单前验证
//   - 状态流转前验证
//   - 订单列表的终态标记
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
//
// 可删除条件：订单状态不是终态（非 Completed 且非 Canceled）
//
// 业务逻辑：
//   - 终态订单具有业务追溯价值，不允许删除
//   - 非终态订单删除时需要恢复商品库存
//
// 相关方法：IsFinal() - 终态判断的底层方法
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

// ==================== 辅助函数 ====================

// ToOrderElements 批量转换订单为轻量级响应对象
//
// 用途：
//   - 订单列表查询（减少数据传输量）
//   - 排除订单项详情，只返回摘要信息
//
// 性能：O(n)，n 为订单数量
//
// 注意：返回的是 models.OrderElement，不含订单项
func ToOrderElements(orders []models.Order) []models.OrderElement {
	elements := make([]models.OrderElement, 0, len(orders))
	for _, order := range orders {
		elements = append(elements, models.OrderElement{
			ID:         order.ID,
			UserID:     order.UserID,
			ShopID:     order.ShopID,
			TotalPrice: order.TotalPrice,
			Status:     order.Status,
			Remark:     order.Remark,
			CreatedAt:  order.CreatedAt,
			UpdatedAt:  order.UpdatedAt,
		})
	}
	return elements
}

// ToCreateOrderRequest 转换为创建订单响应 DTO
//
// 用途：
//   - 订单更新后返回完整订单信息
//   - API 响应数据组装
//
// 包含字段：
//   - 订单基本信息（ID, UserID, ShopID, Status, Remark）
//   - 订单项列表（含选项详情）
//
// 注意：此方法用于响应构建，不处理业务逻辑
func (o *Order) ToCreateOrderRequest() CreateOrderRequest {
	responseItems := make([]CreateOrderItemRequest, len(o.items))
	for i, item := range o.items {
		responseOptions := make([]CreateOrderItemOption, len(item.options))
		for j, opt := range item.options {
			responseOptions[j] = CreateOrderItemOption{
				CategoryID: opt.CategoryID,
				OptionID:   opt.OptionID,
			}
		}

		responseItems[i] = CreateOrderItemRequest{
			ProductID: item.ProductID(),
			Quantity:  item.Quantity(),
			Price:     float64(item.Price()),
			Options:   responseOptions,
		}
	}

	return CreateOrderRequest{
		ID:     o.ID(),
		UserID: o.UserID(),
		ShopID: o.ShopID(),
		Items:  responseItems,
		Remark: o.Remark(),
		Status: int(o.Status()),
	}
}
