package services

import (
	"errors"
	"fmt"
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
	orderRepo           order.OrderRepository
	orderItemRepo       order.OrderItemRepository
	orderItemOptionRepo order.OrderItemOptionRepository
	orderStatusLogRepo  order.OrderStatusLogRepository
	productRepo         product.ProductRepository
	productOptionRepo   product.ProductOptionRepository
	productCategoryRepo product.ProductOptionCategoryRepository
	userRepo            order.OrderRepository
	db                  *gorm.DB
}

func NewOrderService(
	orderRepo order.OrderRepository,
	orderItemRepo order.OrderItemRepository,
	orderItemOptionRepo order.OrderItemOptionRepository,
	orderStatusLogRepo order.OrderStatusLogRepository,
	productRepo product.ProductRepository,
	productOptionRepo product.ProductOptionRepository,
	productCategoryRepo product.ProductOptionCategoryRepository,
	db *gorm.DB,
) *OrderService {
	return &OrderService{
		orderRepo:           orderRepo,
		orderItemRepo:       orderItemRepo,
		orderItemOptionRepo: orderItemOptionRepo,
		orderStatusLogRepo:  orderStatusLogRepo,
		productRepo:         productRepo,
		productOptionRepo:   productOptionRepo,
		productCategoryRepo: productCategoryRepo,
		db:                  db,
	}
}

func (s *OrderService) CreateOrder(req *dto.CreateOrderRequest) (*dto.OrderResponse, error) {
	items := make([]order.OrderItem, len(req.Items))
	for i, itemReq := range req.Items {
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

	ord, err := order.NewOrder(req.UserID, req.ShopID, items, req.Remark)
	if err != nil {
		return nil, err
	}

	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	totalPrice := shared.Price(0)

	for i := range ord.Items {
		prod, err := s.productRepo.FindByID(ord.Items[i].ProductID)
		if err != nil {
			tx.Rollback()
			return nil, errors.New("商品不存在")
		}

		if prod.ShopID != ord.ShopID {
			tx.Rollback()
			return nil, errors.New("商品不属于该店铺")
		}

		if !prod.HasStock(ord.Items[i].Quantity) {
			tx.Rollback()
			return nil, fmt.Errorf("商品 %s 库存不足", prod.Name)
		}

		ord.Items[i].ProductName = prod.Name
		ord.Items[i].ProductDescription = prod.Description
		ord.Items[i].ProductImageURL = prod.ImageURL
		ord.Items[i].Price = prod.Price

		itemTotalPrice := prod.Price.Multiply(ord.Items[i].Quantity)

		for j := range ord.Items[i].Options {
			opt, err := s.productOptionRepo.FindByID(ord.Items[i].Options[j].OptionID)
			if err != nil {
				tx.Rollback()
				return nil, errors.New("商品参数选项不存在")
			}

			cat, err := s.productCategoryRepo.FindByID(opt.CategoryID)
			if err != nil {
				tx.Rollback()
				return nil, errors.New("商品参数类别不存在")
			}

			if cat.ProductID != prod.ID {
				tx.Rollback()
				return nil, errors.New("参数选项不属于指定商品")
			}

			ord.Items[i].Options[j].CategoryID = cat.ID
			ord.Items[i].Options[j].OptionName = opt.Name
			ord.Items[i].Options[j].CategoryName = cat.Name
			ord.Items[i].Options[j].PriceAdjustment = opt.PriceAdjustment

			itemTotalPrice = itemTotalPrice.Add(shared.Price(opt.PriceAdjustment * float64(ord.Items[i].Quantity)))
		}

		ord.Items[i].TotalPrice = itemTotalPrice
		totalPrice = totalPrice.Add(itemTotalPrice)

		if err := prod.DecreaseStock(ord.Items[i].Quantity); err != nil {
			tx.Rollback()
			return nil, err
		}

		if err := s.productRepo.Update(prod); err != nil {
			tx.Rollback()
			return nil, errors.New("更新商品库存失败")
		}
	}

	ord.TotalPrice = totalPrice
	ord.ID = shared.ID(utils.GenerateSnowflakeID())

	if err := s.orderRepo.Save(ord); err != nil {
		tx.Rollback()
		return nil, errors.New("创建订单失败")
	}

	for i := range ord.Items {
		ord.Items[i].OrderID = ord.ID
		if err := s.orderItemRepo.Save(&ord.Items[i]); err != nil {
			tx.Rollback()
			return nil, errors.New("创建订单项失败")
		}

		for j := range ord.Items[i].Options {
			ord.Items[i].Options[j].OrderItemID = ord.Items[i].ID
			if err := s.orderItemOptionRepo.Save(&ord.Items[i].Options[j]); err != nil {
				tx.Rollback()
				return nil, errors.New("创建订单项选项失败")
			}
		}
	}

	statusLog := &order.OrderStatusLog{
		OrderID:     ord.ID,
		OldStatus:   0,
		NewStatus:   ord.Status,
		ChangedTime: time.Now(),
	}
	if err := s.orderStatusLogRepo.Save(statusLog); err != nil {
		tx.Rollback()
		return nil, errors.New("创建订单状态日志失败")
	}

	tx.Commit()

	log2.Infof("订单创建成功: %+v", ord)

	return &dto.OrderResponse{
		ID:         ord.ID,
		UserID:     ord.UserID,
		ShopID:     ord.ShopID,
		TotalPrice: ord.TotalPrice,
		Status:     ord.Status,
		Remark:     ord.Remark,
		CreatedAt:  ord.CreatedAt,
		UpdatedAt:  ord.UpdatedAt,
	}, nil
}

func (s *OrderService) GetOrder(id shared.ID, shopID uint64) (*dto.OrderDetailResponse, error) {
	ord, err := s.orderRepo.FindByIDAndShopID(id, shopID)
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
		ShopID:     ord.ShopID,
		TotalPrice: ord.TotalPrice,
		Status:     ord.Status,
		Remark:     ord.Remark,
		CreatedAt:  ord.CreatedAt,
		UpdatedAt:  ord.UpdatedAt,
		Items:      items,
	}, nil
}

func (s *OrderService) GetOrders(shopID uint64, page, pageSize int) (*dto.OrderListResponse, error) {
	orders, total, err := s.orderRepo.FindByShopID(shopID, page, pageSize)
	if err != nil {
		return nil, err
	}

	data := make([]dto.OrderResponse, len(orders))
	for i, ord := range orders {
		data[i] = dto.OrderResponse{
			ID:         ord.ID,
			UserID:     ord.UserID,
			ShopID:     ord.ShopID,
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

func (s *OrderService) GetOrdersByUser(userID shared.ID, shopID uint64, page, pageSize int) (*dto.OrderListResponse, error) {
	orders, total, err := s.orderRepo.FindByUserID(userID, shopID, page, pageSize)
	if err != nil {
		return nil, err
	}

	data := make([]dto.OrderResponse, len(orders))
	for i, ord := range orders {
		data[i] = dto.OrderResponse{
			ID:         ord.ID,
			UserID:     ord.UserID,
			ShopID:     ord.ShopID,
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

func (s *OrderService) GetUnfinishedOrders(shopID uint64, flow order.OrderStatusFlow, page, pageSize int) (*dto.OrderListResponse, error) {
	orders, total, err := s.orderRepo.FindUnfinishedByShopID(shopID, flow, page, pageSize)
	if err != nil {
		return nil, err
	}

	data := make([]dto.OrderResponse, len(orders))
	for i, ord := range orders {
		data[i] = dto.OrderResponse{
			ID:         ord.ID,
			UserID:     ord.UserID,
			ShopID:     ord.ShopID,
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

	orders, total, err := s.orderRepo.Search(req.ShopID, req.UserID, req.Statuses, startTime, endTime, req.Page, req.PageSize)
	if err != nil {
		return nil, err
	}

	data := make([]dto.OrderResponse, len(orders))
	for i, ord := range orders {
		data[i] = dto.OrderResponse{
			ID:         ord.ID,
			UserID:     ord.UserID,
			ShopID:     ord.ShopID,
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

func (s *OrderService) UpdateOrderStatus(id shared.ID, shopID uint64, newStatus order.OrderStatus, flow order.OrderStatusFlow) error {
	ord, err := s.orderRepo.FindByIDAndShopID(id, shopID)
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

func (s *OrderService) DeleteOrder(id shared.ID, shopID uint64) error {
	ord, err := s.orderRepo.FindByIDAndShopID(id, shopID)
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
