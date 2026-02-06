// Package order (service) 提供订单领域服务。
//
// 领域服务职责：
//   - 处理跨实体的复杂业务逻辑
//   - 编排订单创建/更新的完整流程
//   - 管理订单状态流转验证
//   - 协调商品库存操作
//
// 与应用服务的区别：
//   - 领域服务：处理纯业务逻辑，不涉及 HTTP、事务管理
//   - 应用服务（Handler）：处理请求、事务、调用领域服务
//
// 事务边界：
//   - 领域服务不管理事务
//   - 事务由 Handler 层通过 Unit of Work 或手动管理
//
// 依赖：
//   - *gorm.DB: 数据库连接（用于库存查询和扣减）
package order

import (
	"fmt"

	"github.com/bwmarrin/snowflake"
	"gorm.io/gorm"
	"orderease/models"
)

// Service 订单领域服务
//
// 职责边界：
//   - 处理需要访问多个聚合根的业务逻辑
//   - 执行订单创建的核心流程（验证、快照、价格、库存）
//   - 提供状态流转验证
//
// 不处理：
//   - HTTP 请求/响应
//   - 事务管理
//   - 用户/店铺存在性验证
//
// 使用示例：
//
//	service := order.NewService(db)
//	orderModel, totalPrice, err := service.CreateOrder(dto)
type Service struct {
	db *gorm.DB
}

// NewService 创建订单领域服务
func NewService(db *gorm.DB) *Service {
	return &Service{db: db}
}

// CreateOrderResult 创建订单结果
type CreateOrderResult struct {
	Order      *models.Order
	TotalPrice float64
	Error      error
}

// CreateOrderDTO 创建订单 DTO
type CreateOrderDTO struct {
	UserID snowflake.ID
	ShopID snowflake.ID
	Items  []CreateOrderItemDTO
	Remark string
}

// CreateOrderItemDTO 创建订单项 DTO
type CreateOrderItemDTO struct {
	ProductID snowflake.ID
	Quantity  int
	Options   []CreateOrderItemOptionDTO
}

// CreateOrderItemOptionDTO 创建订单项选项 DTO
type CreateOrderItemOptionDTO struct {
	OptionID   snowflake.ID
	CategoryID snowflake.ID
}

// CreateOrder 创建订单（领域服务核心方法）
//
// 执行流程：
//   1. 构建订单项模型
//   2. 处理订单项选项
//   3. 调用 processOrderItems 执行核心逻辑
//
// 核心逻辑（processOrderItems）：
//   1. 验证商品存在
//   2. 验证库存充足
//   3. 保存商品快照（名称、描述、图片）
//   4. 处理参数选项（验证、快照、价格调整）
//   5. 计算订单项总价
//   6. 扣减商品库存
//
// 参数：
//   - dto: 创建订单数据传输对象
//
// 返回：
//   - *models.Order: 构建完成的订单模型（未持久化）
//   - float64: 订单总价
//   - error: 处理过程中的错误
//
// 注意：
//   - 订单 ID 为 0，需在 Handler 层生成
//   - 此方法不处理事务，需在 Handler 层管理
//   - 不验证用户/店铺存在性
func (s *Service) CreateOrder(dto CreateOrderDTO) (*models.Order, float64, error) {
	// 构建订单项
	var orderItems []models.OrderItem
	for _, itemReq := range dto.Items {
		orderItem := models.OrderItem{
			ProductID: itemReq.ProductID,
			Quantity:  itemReq.Quantity,
		}

		// 处理选中的选项
		var options []models.OrderItemOption
		for _, optionReq := range itemReq.Options {
			option := models.OrderItemOption{
				OptionID:   optionReq.OptionID,
				CategoryID: optionReq.CategoryID,
			}
			options = append(options, option)
		}
		orderItem.Options = options
		orderItems = append(orderItems, orderItem)
	}

	order := &models.Order{
		UserID: dto.UserID,
		ShopID: dto.ShopID,
		Items:  orderItems,
		Remark: dto.Remark,
		Status: models.OrderStatusPending,
	}

	// 执行订单创建的核心逻辑
	totalPrice, err := s.processOrderItems(s.db, order)
	if err != nil {
		return nil, 0, err
	}

	order.TotalPrice = models.Price(totalPrice)
	order.ID = snowflake.ID(0) // 将在 Handler 中生成

	return order, totalPrice, nil
}

// processOrderItems 处理订单项核心逻辑
//
// 处理流程：
//   1. 验证商品存在性
//   2. 验证库存充足性
//   3. 保存商品快照（防后续修改影响历史订单）
//   4. 处理参数选项（验证、快照、价格计算）
//   5. 计算订单项总价
//   6. 扣减商品库存
//
// 快照机制：
//   - 保存商品当前信息到订单项
//   - 即使后续商品修改，历史订单保持不变
//   - 包含：名称、描述、图片、选项价格
//
// 参数：
//   - db:    数据库连接
//   - order: 订单模型（会被修改）
//
// 返回：
//   - float64: 订单总价
//   - error:   处理错误
//
// 错误场景：
//   - 商品不存在
//   - 库存不足
//   - 参数选项不存在
//   - 参数选项不属于指定商品
//   - 库存更新失败
func (s *Service) processOrderItems(db *gorm.DB, order *models.Order) (float64, error) {
	totalPrice := float64(0.0)

	for i := range order.Items {
		// 1. 验证商品存在
		var product models.Product
		if err := db.First(&product, order.Items[i].ProductID).Error; err != nil {
			return 0, fmt.Errorf("商品不存在, ID: %d", order.Items[i].ProductID)
		}

		// 2. 验证库存充足
		if product.Stock < order.Items[i].Quantity {
			return 0, fmt.Errorf("商品 %s 库存不足, 当前库存: %d, 需求数量: %d",
				product.Name, product.Stock, order.Items[i].Quantity)
		}

		// 3. 保存商品快照信息
		order.Items[i].ProductName = product.Name
		order.Items[i].ProductDescription = product.Description
		order.Items[i].ProductImageURL = product.ImageURL
		order.Items[i].Price = models.Price(product.Price)

		// 4. 处理订单项参数选项
		itemTotalPrice := float64(order.Items[i].Quantity) * product.Price
		for j := range order.Items[i].Options {
			// 获取参数选项信息
			var option models.ProductOption
			if err := db.First(&option, order.Items[i].Options[j].OptionID).Error; err != nil {
				return 0, fmt.Errorf("商品参数选项不存在, ID: %d", order.Items[i].Options[j].OptionID)
			}

			// 获取参数类别信息
			var category models.ProductOptionCategory
			if err := db.First(&category, option.CategoryID).Error; err != nil {
				return 0, fmt.Errorf("商品参数类别不存在, ID: %d", order.Items[i].Options[j].CategoryID)
			}

			// 验证参数所属商品
			if category.ProductID != product.ID {
				return 0, fmt.Errorf("参数选项不属于指定商品")
			}

			// 保存参数选项快照
			order.Items[i].Options[j].CategoryID = category.ID
			order.Items[i].Options[j].OptionName = option.Name
			order.Items[i].Options[j].CategoryName = category.Name
			order.Items[i].Options[j].PriceAdjustment = option.PriceAdjustment

			// 计算参数选项对总价的影响
			itemTotalPrice += float64(order.Items[i].Quantity) * option.PriceAdjustment
		}

		// 5. 设置订单项总价
		order.Items[i].TotalPrice = models.Price(itemTotalPrice)

		// 6. 扣减库存
		product.Stock -= order.Items[i].Quantity
		totalPrice += itemTotalPrice
		if err := db.Save(&product).Error; err != nil {
			return 0, fmt.Errorf("更新商品库存失败: %v", err)
		}
	}

	return totalPrice, nil
}

// ValidateOrder 验证订单基础数据
func (s *Service) ValidateOrder(order *models.Order) error {
	if order.UserID == 0 {
		return fmt.Errorf("用户ID不能为空")
	}

	if order.ShopID == 0 {
		return fmt.Errorf("店铺ID不能为空")
	}

	if len(order.Items) == 0 {
		return fmt.Errorf("订单项不能为空")
	}

	for _, item := range order.Items {
		if item.ProductID == 0 {
			return fmt.Errorf("商品ID不能为空")
		}
		if item.Quantity <= 0 {
			return fmt.Errorf("商品数量必须大于0")
		}
	}

	return nil
}

// CalculateTotal 计算订单总价
func (s *Service) CalculateTotal(order *models.Order) float64 {
	total := float64(0)
	for _, item := range order.Items {
		itemTotal := float64(item.Quantity) * float64(item.Price)
		for _, opt := range item.Options {
			itemTotal += float64(item.Quantity) * opt.PriceAdjustment
		}
		total += itemTotal
	}
	return total
}

// RestoreStock 恢复商品库存（用于订单取消时）
func (s *Service) RestoreStock(tx *gorm.DB, order models.Order) error {
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

// UpdateOrderDTO 更新订单 DTO
type UpdateOrderDTO struct {
	OrderID snowflake.ID
	ShopID  snowflake.ID
	Items   []CreateOrderItemDTO
	Remark  string
	Status  int
}

// UpdateOrder 更新订单（领域服务方法）
//
// 执行流程：
//   1. 构建新订单项模型
//   2. 调用 processOrderItems 处理订单项
//   3. 计算新的订单总价
//
// 与 CreateOrder 的区别：
//   - 保留原订单 ID
//   - 支持更新状态
//
// 注意：
//   - 原订单项将被完全替换
//   - 库存已扣减的商品会被恢复（由 Handler 处理）
//   - 此方法不处理事务
func (s *Service) UpdateOrder(dto UpdateOrderDTO) (*models.Order, float64, error) {
	// 构建订单项
	var orderItems []models.OrderItem
	for _, itemReq := range dto.Items {
		orderItem := models.OrderItem{
			ProductID: itemReq.ProductID,
			Quantity:  itemReq.Quantity,
		}

		// 处理选中的选项
		var options []models.OrderItemOption
		for _, optionReq := range itemReq.Options {
			option := models.OrderItemOption{
				OptionID:   optionReq.OptionID,
				CategoryID: optionReq.CategoryID,
			}
			options = append(options, option)
		}
		orderItem.Options = options
		orderItems = append(orderItems, orderItem)
	}

	order := &models.Order{
		ID:     dto.OrderID,
		ShopID: dto.ShopID,
		Items:  orderItems,
		Remark: dto.Remark,
		Status: dto.Status,
	}

	// 执行订单项处理的核心逻辑（复用 processOrderItems）
	totalPrice, err := s.processOrderItems(s.db, order)
	if err != nil {
		return nil, 0, err
	}

	order.TotalPrice = models.Price(totalPrice)

	return order, totalPrice, nil
}

// ValidateStatusTransition 验证订单状态流转合法性
//
// 验证逻辑：
//   1. 在店铺配置中查找当前状态
//   2. 检查是否为终态（终态不允许流转）
//   3. 检查目标状态是否在当前状态的允许动作列表中
//
// 参数：
//   - currentStatus: 当前订单状态值
//   - nextStatus:    目标状态值
//   - flow:          店铺的订单状态流转配置
//
// 返回：
//   - nil:   流转合法
//   - error: 流转不合法（终态/不允许的流转）
//
// 使用场景：
//   - 订单状态翻转前验证
//   - 批量状态更新前验证
func (s *Service) ValidateStatusTransition(currentStatus int, nextStatus int, flow models.OrderStatusFlow) error {
	// 查找当前状态在店铺流转定义中的配置
	for _, status := range flow.Statuses {
		if status.Value == currentStatus {
			// 检查是否为终态
			if status.IsFinal {
				return fmt.Errorf("当前状态为终态，不允许转换")
			}

			// 检查请求的next_status是否在当前状态允许的动作列表中
			for _, action := range status.Actions {
				if action.NextStatus == nextStatus {
					// 找到匹配的动作，允许转换
					return nil
				}
			}

			// 没有找到匹配的动作
			return fmt.Errorf("当前状态不允许转换到指定的下一个状态")
		}
	}

	// 如果在店铺流转定义中找不到当前状态
	return fmt.Errorf("当前状态不允许转换")
}
