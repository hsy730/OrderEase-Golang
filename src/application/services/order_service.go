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
	ord, err := s.orderRepo.FindByID(id)
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
	ord, err := s.orderRepo.FindByID(id)
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
	ord, err := s.orderRepo.FindByID(id)
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

func (s *OrderService) AdvanceSearchOrders(req *dto.AdvanceSearchOrderRequest) (*dto.OrderListResponse, error) {
	query := s.db.Model(&order.Order{}).Where("shop_id = ?", req.ShopID.ToUint64())

	// 添加用户ID筛选
	if req.UserID != "" {
		query = query.Where("user_id = ?", req.UserID)
	}

	// 添加状态筛选（支持多个状态）
	if len(req.Status) > 0 {
		query = query.Where("status IN ?", req.Status)
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
		return nil, errors.New("获取订单总数失败")
	}

	// 分页查询
	offset := (req.Page - 1) * req.PageSize
	var orders []order.Order
	if err := query.Offset(offset).Limit(req.PageSize).
		Order("created_at DESC").
		Find(&orders).Error; err != nil {
		return nil, errors.New("查询订单列表失败")
	}

	// 转换为响应格式
	data := make([]dto.OrderResponse, len(orders))
	for i, ord := range orders {
		data[i] = dto.OrderResponse{
			ID:         ord.ID,
			UserID:     ord.UserID,
			ShopID:     shared.ParseIDFromUint64(uint64(ord.ShopID)),
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

// UpdateOrder 更新订单
func (s *OrderService) UpdateOrder(req *dto.UpdateOrderRequest) (*dto.OrderDetailResponse, error) {
	ord, err := s.orderRepo.FindByID(req.ID)
	if err != nil {
		return nil, err
	}

	// 开启事务
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 删除原有的订单项和选项
	if err := s.orderItemOptionRepo.DeleteByOrderItemID(req.ID); err != nil {
		tx.Rollback()
		return nil, errors.New("删除原有订单项选项失败")
	}

	if err := s.orderItemRepo.DeleteByOrderID(req.ID); err != nil {
		tx.Rollback()
		return nil, errors.New("删除原有订单项失败")
	}

	totalPrice := float64(0.0)

	// 创建新的订单项
	for _, itemReq := range req.Items {
		prod, err := s.productRepo.FindByID(itemReq.ProductID)
		if err != nil {
			tx.Rollback()
			return nil, errors.New("商品不存在")
		}

		orderItem := &order.OrderItem{
			OrderID:   ord.ID,
			ProductID: itemReq.ProductID,
			Quantity:  itemReq.Quantity,
			Price:     prod.Price,
		}

		// 处理选中的选项
		var options []order.OrderItemOption
		itemTotalPrice := float64(orderItem.Quantity) * prod.Price.ToFloat64()

		for _, optionReq := range itemReq.Options {
			// 获取参数选项信息
			opt, err := s.productOptionRepo.FindByID(optionReq.OptionID)
			if err != nil {
				tx.Rollback()
				return nil, errors.New("商品参数选项不存在")
			}

			// 获取参数类别信息
			cat, err := s.productOptionCategoryRepo.FindByID(optionReq.CategoryID)
			if err != nil {
				tx.Rollback()
				return nil, errors.New("商品参数类别不存在")
			}

			// 保存参数选项快照
			options = append(options, order.OrderItemOption{
				OrderItemID:     orderItem.ID,
				OptionID:        optionReq.OptionID,
				CategoryID:      optionReq.CategoryID,
				OptionName:      opt.Name,
				CategoryName:    cat.Name,
				PriceAdjustment: opt.PriceAdjustment,
			})
			itemTotalPrice += float64(orderItem.Quantity) * opt.PriceAdjustment
		}

		// 设置订单项总价
		orderItem.Options = options
		orderItem.TotalPrice = shared.Price(itemTotalPrice)

		if err := s.orderItemRepo.Save(orderItem); err != nil {
			tx.Rollback()
			return nil, errors.New("创建新订单项失败")
		}
		totalPrice += itemTotalPrice
	}

	// 更新订单信息
	ord.ShopID = req.ShopID.ToUint64()
	ord.Remark = req.Remark
	ord.Status = req.Status
	ord.TotalPrice = shared.Price(totalPrice)

	if err := s.orderRepo.Update(ord); err != nil {
		tx.Rollback()
		return nil, errors.New("更新订单信息失败")
	}

	tx.Commit()

	// 重新获取更新后的订单信息
	ord, err = s.orderRepo.FindByID(req.ID)
	if err != nil {
		return nil, errors.New("获取更新后的订单信息失败")
	}

	return s.toOrderDetailResponse(ord), nil
}

// toOrderDetailResponse 转换为订单详情响应
func (s *OrderService) toOrderDetailResponse(ord *order.Order) *dto.OrderDetailResponse {
	items := make([]dto.OrderItemResponse, len(ord.Items))
	for i, item := range ord.Items {
		options := make([]dto.OrderItemOptionResponse, len(item.Options))
		for j, opt := range item.Options {
			options[j] = dto.OrderItemOptionResponse{
				CategoryID:      opt.CategoryID,
				OptionID:        opt.OptionID,
				CategoryName:    opt.CategoryName,
				OptionName:      opt.OptionName,
				PriceAdjustment: opt.PriceAdjustment,
			}
		}
		items[i] = dto.OrderItemResponse{
			ProductID:       item.ProductID,
			ProductName:     "", // 需要从商品获取
			Quantity:        item.Quantity,
			Price:           item.Price,
			TotalPrice:      item.TotalPrice,
			Options:         options,
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
	}
}
