package order

import (
	"fmt"

	"github.com/bwmarrin/snowflake"
	"gorm.io/gorm"
	"orderease/models"
)

// Service 订单领域服务
// 负责跨实体的订单编排逻辑
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
	ShopID uint64
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

// CreateOrder 创建订单（领域服务方法）
// 负责订单创建的完整流程：
// 1. 验证订单数据
// 2. 验证商品库存
// 3. 保存商品快照
// 4. 处理商品选项
// 5. 计算订单总价
// 6. 扣减库存
// 注意：不包含事务管理、用户验证、店铺验证等 Handler 层职责
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

// processOrderItems 处理订单项（验证库存、保存快照、计算价格、扣减库存）
// 这是领域服务的核心方法
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
