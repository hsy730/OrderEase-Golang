package services

import (
	"errors"
	"orderease/application/dto"
	"orderease/domain/order"
	"orderease/domain/product"
	"orderease/domain/shared"
	"orderease/utils"
	"orderease/utils/log2"
	"time"

	"gorm.io/gorm"
)

type OrderService struct {
	db                        *gorm.DB
	productRepo               product.ProductRepository
	productOptionRepo         product.ProductOptionRepository
	productOptionCategoryRepo product.ProductOptionCategoryRepository
	orderRepo                 order.OrderRepository
	orderItemRepo             order.OrderItemRepository
	orderItemOptionRepo       order.OrderItemOptionRepository
	orderStatusLogRepo        order.OrderStatusLogRepository
}

// NewOrderService 创建 OrderService 实例
func NewOrderService(
	db *gorm.DB,
	productRepo product.ProductRepository,
	productOptionRepo product.ProductOptionRepository,
	productOptionCategoryRepo product.ProductOptionCategoryRepository,
	orderRepo order.OrderRepository,
	orderItemRepo order.OrderItemRepository,
	orderItemOptionRepo order.OrderItemOptionRepository,
	orderStatusLogRepo order.OrderStatusLogRepository,
) *OrderService {
	return &OrderService{
		db:                        db,
		productRepo:               productRepo,
		productOptionRepo:         productOptionRepo,
		productOptionCategoryRepo: productOptionCategoryRepo,
		orderRepo:                 orderRepo,
		orderItemRepo:             orderItemRepo,
		orderItemOptionRepo:       orderItemOptionRepo,
		orderStatusLogRepo:        orderStatusLogRepo,
	}
}

// buildOrderItems 从 DTO 构建订单项
func (s *OrderService) buildOrderItems(reqItems []dto.CreateOrderItemRequest) []order.OrderItem {
	items := make([]order.OrderItem, len(reqItems))
	for i, itemReq := range reqItems {
		options := make([]order.OrderItemOption, len(itemReq.Options))
		for j, optReq := range itemReq.Options {
			options[j] = order.OrderItemOption{
				CategoryID:      optReq.CategoryID,
				OptionID:        optReq.OptionID,
				OptionName:      "",
				CategoryName:    "",
				PriceAdjustment: 0,
			}
		}

		items[i] = order.OrderItem{
			ProductID: itemReq.ProductID,
			Quantity:  itemReq.Quantity,
			Price:     shared.Price(itemReq.Price),
			Options:   options,
		}
	}
	return items
}

// CreateOrder 创建订单（重构版本）
// 业务逻辑已迁移到领域层，应用层只负责编排
func (s *OrderService) CreateOrder(req *dto.CreateOrderRequest) (*dto.OrderResponse, error) {
	// 1. 构建 Order 对象
	items := s.buildOrderItems(req.Items)
	ord, err := order.NewOrder(req.UserID, req.ShopID.ToUint64(), items, req.Remark)
	if err != nil {
		return nil, err
	}

	// 2. 创建 finder 适配器
	finder := NewProductFinderAdapter(s.productRepo, s.productOptionRepo, s.productOptionCategoryRepo)

	// 3. 领域验证和价格计算（业务逻辑在领域层）
	if err := ord.ValidateItems(finder); err != nil {
		return nil, err
	}

	if err := ord.CalculateTotal(finder); err != nil {
		return nil, err
	}

	// 4. 执行事务（应用层职责）
	return s.executeCreateOrderTransaction(ord, finder)
}

// executeCreateOrderTransaction 执行订单创建的事务
func (s *OrderService) executeCreateOrderTransaction(ord *order.Order, finder order.ProductFinder) (*dto.OrderResponse, error) {
	var savedOrder *order.Order
	var err error

	// 使用事务模板
	err = WithTx(s.db, func(tx *gorm.DB) error {
		// 扣减库存
		for i := range ord.Items {
			prod, findErr := finder.FindProduct(ord.Items[i].ProductID)
			if findErr != nil {
				return findErr
			}

			if decreaseErr := prod.DecreaseStock(ord.Items[i].Quantity); decreaseErr != nil {
				return decreaseErr
			}

			if updateErr := s.productRepo.Update(prod); updateErr != nil {
				return errors.New("更新商品库存失败")
			}
		}

		// 设置订单ID
		ord.ID = shared.ID(utils.GenerateSnowflakeID())

		// 保存订单
		if err := s.orderRepo.Save(ord); err != nil {
			return errors.New("创建订单失败")
		}

		// 保存订单项
		for i := range ord.Items {
			ord.Items[i].OrderID = ord.ID
			if err := s.orderItemRepo.Save(&ord.Items[i]); err != nil {
				return errors.New("创建订单项失败")
			}

			for j := range ord.Items[i].Options {
				ord.Items[i].Options[j].OrderItemID = ord.Items[i].ID
				if err := s.orderItemOptionRepo.Save(&ord.Items[i].Options[j]); err != nil {
					return errors.New("创建订单项选项失败")
				}
			}
		}

		// 保存状态日志
		statusLog := &order.OrderStatusLog{
			OrderID:     ord.ID,
			OldStatus:   0,
			NewStatus:   ord.Status,
			ChangedTime: time.Now(),
		}
		if err := s.orderStatusLogRepo.Save(statusLog); err != nil {
			return errors.New("创建订单状态日志失败")
		}

		savedOrder = ord
		return nil
	})

	if err != nil {
		return nil, err
	}

	log2.Infof("订单创建成功: %+v", savedOrder)

	return &dto.OrderResponse{
		ID:         savedOrder.ID,
		UserID:     savedOrder.UserID,
		ShopID:     shared.ParseIDFromUint64(savedOrder.ShopID),
		TotalPrice: savedOrder.TotalPrice,
		Status:     savedOrder.Status,
		Remark:     savedOrder.Remark,
		CreatedAt:  savedOrder.CreatedAt,
		UpdatedAt:  savedOrder.UpdatedAt,
	}, nil
}
func (s *OrderService) GetOrder(id shared.ID, shopID shared.ID) (*dto.OrderDetailResponse, error) {
	ord, err := s.orderRepo.FindByIDAndShopID(id, shopID.ToUint64())
	if err != nil {
		return nil, err
	}

	items := make([]dto.OrderItemResponse, len(ord.Items))
	for i, item := range ord.Items {
		options := make([]dto.OrderItemOptionResponse, len(item.Options))
		for j, opt := range item.Options {
			options[j] = dto.OrderItemOptionResponse{
				ID:              opt.ID,
				CategoryID:      opt.CategoryID,
				OptionID:        opt.OptionID,
				OptionName:      opt.OptionName,
				CategoryName:    opt.CategoryName,
				PriceAdjustment: opt.PriceAdjustment,
			}
		}

		items[i] = dto.OrderItemResponse{
			ID:                 item.ID,
			ProductID:          item.ProductID,
			Quantity:           item.Quantity,
			Price:              item.Price,
			TotalPrice:         item.TotalPrice,
			ProductName:        item.ProductName,
			ProductDescription: item.ProductDescription,
			ProductImageURL:    item.ProductImageURL,
			Options:            options,
		}
	}

	return &dto.OrderDetailResponse{
		ID:         ord.ID,
		UserID:     ord.UserID,
		ShopID:     shared.ParseIDFromUint64(ord.ShopID),
		TotalPrice: ord.TotalPrice,
		Status:     ord.Status,
		Remark:     ord.Remark,
		CreatedAt:  ord.CreatedAt,
		UpdatedAt:  ord.UpdatedAt,
		Items:      items,
	}, nil
}

func (s *OrderService) GetOrders(shopID shared.ID, page, pageSize int) (*dto.OrderListResponse, error) {
	orders, total, err := s.orderRepo.FindByShopID(shopID.ToUint64(), page, pageSize)
	if err != nil {
		return nil, err
	}

	data := make([]dto.OrderResponse, len(orders))
	for i, ord := range orders {
		data[i] = dto.OrderResponse{
			ID:         ord.ID,
			UserID:     ord.UserID,
			ShopID:     shared.ParseIDFromUint64(ord.ShopID),
			TotalPrice: ord.TotalPrice,
			Status:     ord.Status,
			Remark:     ord.Remark,
			CreatedAt:  ord.CreatedAt,
			UpdatedAt:  ord.UpdatedAt,
		}
	}

	return &dto.OrderListResponse{
		Total:    total,
		Page:     page,
		PageSize: pageSize,
		Data:     data,
	}, nil
}

func (s *OrderService) GetOrdersByUser(userID shared.ID, shopID shared.ID, page, pageSize int) (*dto.OrderListResponse, error) {
	orders, total, err := s.orderRepo.FindByUserID(userID, shopID.ToUint64(), page, pageSize)
	if err != nil {
		return nil, err
	}

	data := make([]dto.OrderResponse, len(orders))
	for i, ord := range orders {
		data[i] = dto.OrderResponse{
			ID:         ord.ID,
			UserID:     ord.UserID,
			ShopID:     shared.ParseIDFromUint64(ord.ShopID),
			TotalPrice: ord.TotalPrice,
			Status:     ord.Status,
			Remark:     ord.Remark,
			CreatedAt:  ord.CreatedAt,
			UpdatedAt:  ord.UpdatedAt,
		}
	}

	return &dto.OrderListResponse{
		Total:    total,
		Page:     page,
		PageSize: pageSize,
		Data:     data,
	}, nil
}

func (s *OrderService) GetUnfinishedOrders(shopID shared.ID, flow order.OrderStatusFlow, page, pageSize int) (*dto.OrderListResponse, error) {
	orders, total, err := s.orderRepo.FindUnfinishedByShopID(shopID.ToUint64(), flow, page, pageSize)
	if err != nil {
		return nil, err
	}

	data := make([]dto.OrderResponse, len(orders))
	for i, ord := range orders {
		data[i] = dto.OrderResponse{
			ID:         ord.ID,
			UserID:     ord.UserID,
			ShopID:     shared.ParseIDFromUint64(ord.ShopID),
			TotalPrice: ord.TotalPrice,
			Status:     ord.Status,
			Remark:     ord.Remark,
			CreatedAt:  ord.CreatedAt,
			UpdatedAt:  ord.UpdatedAt,
		}
	}

	return &dto.OrderListResponse{
		Total:    total,
		Page:     page,
		PageSize: pageSize,
		Data:     data,
	}, nil
}

func (s *OrderService) SearchOrders(req *dto.SearchOrdersRequest) (*dto.OrderListResponse, error) {
	startTime := req.StartTime
	endTime := req.EndTime

	if req.StartTimeStr != "" {
		parsedTime, err := time.Parse(time.RFC3339, req.StartTimeStr)
		if err != nil {
			return nil, err
		}
		startTime = parsedTime
	}

	if req.EndTimeStr != "" {
		parsedTime, err := time.Parse(time.RFC3339, req.EndTimeStr)
		if err != nil {
			return nil, err
		}
		endTime = parsedTime
	}

	orders, total, err := s.orderRepo.Search(req.ShopID.ToUint64(), req.UserID, req.Statuses, startTime, endTime, req.Page, req.PageSize)
	if err != nil {
		return nil, err
	}

	data := make([]dto.OrderResponse, len(orders))
	for i, ord := range orders {
		data[i] = dto.OrderResponse{
			ID:         ord.ID,
			UserID:     ord.UserID,
			ShopID:     shared.ParseIDFromUint64(ord.ShopID),
			TotalPrice: ord.TotalPrice,
			Status:     ord.Status,
			Remark:     ord.Remark,
			CreatedAt:  ord.CreatedAt,
			UpdatedAt:  ord.UpdatedAt,
		}
	}

	return &dto.OrderListResponse{
		Total:    total,
		Page:     req.Page,
		PageSize: req.PageSize,
		Data:     data,
	}, nil
}

func (s *OrderService) UpdateOrderStatus(id shared.ID, shopID shared.ID, newStatus order.OrderStatus, flow order.OrderStatusFlow) error {
	ord, err := s.orderRepo.FindByIDAndShopID(id, shopID.ToUint64())
	if err != nil {
		return err
	}

	if err := ord.CanTransitionTo(newStatus, flow); err != nil {
		return err
	}

	oldStatus := ord.Status
	if err := ord.TransitionTo(newStatus, flow); err != nil {
		return err
	}

	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if err := s.orderRepo.Update(ord); err != nil {
		tx.Rollback()
		return errors.New("更新订单状态失败")
	}

	statusLog := &order.OrderStatusLog{
		OrderID:     ord.ID,
		OldStatus:   oldStatus,
		NewStatus:   newStatus,
		ChangedTime: time.Now(),
	}
	if err := s.orderStatusLogRepo.Save(statusLog); err != nil {
		tx.Rollback()
		return errors.New("记录状态变更失败")
	}

	tx.Commit()

	return nil
}

func (s *OrderService) DeleteOrder(id shared.ID, shopID shared.ID) error {
	ord, err := s.orderRepo.FindByIDAndShopID(id, shopID.ToUint64())
	if err != nil {
		return err
	}

	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if err := s.orderItemOptionRepo.DeleteByOrderItemID(id); err != nil {
		tx.Rollback()
		return errors.New("删除订单项选项失败")
	}

	if err := s.orderItemRepo.DeleteByOrderID(id); err != nil {
		tx.Rollback()
		return errors.New("删除订单项失败")
	}

	if err := s.orderStatusLogRepo.DeleteByOrderID(id); err != nil {
		tx.Rollback()
		return errors.New("删除订单状态日志失败")
	}

	if err := s.orderRepo.Delete(id); err != nil {
		tx.Rollback()
		return errors.New("删除订单失败")
	}

	tx.Commit()

	log2.Infof("订单删除成功: %+v", ord)

	return nil
}
