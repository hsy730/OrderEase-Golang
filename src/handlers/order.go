package handlers

import (
	"net/http"
	orderdomain "orderease/domain/order"
	"orderease/models"
	"orderease/utils"
	"orderease/utils/log2"
	"strconv"
	"time"

	"github.com/bwmarrin/snowflake"
	"github.com/gin-gonic/gin"
)

// 高级查询订单请求

type AdvanceSearchOrderRequest struct {
	Page      int    `json:"page"`
	PageSize  int    `json:"pageSize"`
	UserID    string `json:"user_id"`
	Status    []int  `json:"status"`
	StartTime string `json:"start_time"`
	EndTime   string `json:"end_time"`
	ShopID    uint64 `json:"shop_id"`
}

// 创建订单

type CreateOrderRequest struct {
	ID     snowflake.ID             `json:"id"`
	UserID snowflake.ID             `json:"user_id"`
	ShopID uint64                   `json:"shop_id"`
	Items  []CreateOrderItemRequest `json:"items"`
	Remark string                   `json:"remark"`
	Status int                      `json:"status"`
}

type CreateOrderItemRequest struct {
	ProductID snowflake.ID            `json:"product_id"`
	Quantity  int                     `json:"quantity"`
	Price     float64                 `json:"price"`
	Options   []CreateOrderItemOption `json:"options"`
}

type CreateOrderItemOption struct {
	CategoryID snowflake.ID `json:"category_id"`
	OptionID   snowflake.ID `json:"option_id"`
}

// 创建订单请求结构体，添加参数支持
func (h *Handler) CreateOrder(c *gin.Context) {
	var req CreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的订单数据: "+err.Error())
		return
	}

	// 假设存在一个 IsValidUserID 函数来验证用户ID的合法性
	if !h.IsValidUserID(req.UserID) {
		log2.Errorf("创建订单失败: 非法用户")
		errorResponse(c, http.StatusBadRequest, "创建订单失败")
		return
	}

	validShopID, err := h.validAndReturnShopID(c, req.ShopID)
	if err != nil {
		errorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	// 开启事务
	tx := h.DB.Begin()

	// 构建领域服务 DTO
	var itemsDTO []orderdomain.CreateOrderItemDTO
	for _, itemReq := range req.Items {
		var optionsDTO []orderdomain.CreateOrderItemOptionDTO
		for _, optionReq := range itemReq.Options {
			optionsDTO = append(optionsDTO, orderdomain.CreateOrderItemOptionDTO{
				OptionID:   optionReq.OptionID,
				CategoryID: optionReq.CategoryID,
			})
		}
		itemsDTO = append(itemsDTO, orderdomain.CreateOrderItemDTO{
			ProductID: itemReq.ProductID,
			Quantity:  itemReq.Quantity,
			Options:   optionsDTO,
		})
	}

	// 调用领域服务创建订单（处理库存验证、快照、价格计算、库存扣减）
	orderModel, totalPrice, err := h.orderService.CreateOrder(orderdomain.CreateOrderDTO{
		UserID: req.UserID,
		ShopID: validShopID,
		Items:  itemsDTO,
		Remark: req.Remark,
	})
	if err != nil {
		tx.Rollback()
		h.logger.Errorf("创建订单失败: %v", err)
		errorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	orderModel.TotalPrice = models.Price(totalPrice)
	// 雪花ID生成逻辑
	orderModel.ID = utils.GenerateSnowflakeID()

	// 数据库写入
	if err := tx.Create(&orderModel).Error; err != nil {
		tx.Rollback()
		log2.Errorf("创建订单失败: %v", err)
		errorResponse(c, http.StatusInternalServerError, "创建订单失败")
		return
	}

	// 更新订单项选项的OrderItemID
	for i := range orderModel.Items {
		for j := range orderModel.Items[i].Options {
			orderModel.Items[i].Options[j].OrderItemID = orderModel.Items[i].ID
			if err := tx.Save(&orderModel.Items[i].Options[j]).Error; err != nil {
				tx.Rollback()
				log2.Errorf("更新订单项选项失败: %v", err)
				errorResponse(c, http.StatusInternalServerError, "创建订单失败")
				return
			}
		}
	}

	// 添加日志，打印创建的订单信息
	log2.Infof("创建的订单信息: %+v", orderModel)
	for _, item := range orderModel.Items {
		log2.Infof("订单项ID: %s, 选项数量: %d", item.ID, len(item.Options))
		for _, option := range item.Options {
			log2.Infof("选项ID: %s, 名称: %s", option.ID, option.OptionName)
		}
	}

	// 创建订单状态日志
	statusLog := models.OrderStatusLog{
		OrderID:     orderModel.ID,
		OldStatus:   0,
		NewStatus:   orderModel.Status,
		ChangedTime: time.Now(),
	}
	if err := tx.Create(&statusLog).Error; err != nil {
		tx.Rollback()
		log2.Errorf("创建订单状态日志失败: %v", err)
		errorResponse(c, http.StatusInternalServerError, "创建订单失败")
		return
	}

	tx.Commit()

	// 触发SSE通知
	go h.NotifyNewOrder(*orderModel)

	successResponse(c, gin.H{
		"order_id":    orderModel.ID,
		"total_price": orderModel.TotalPrice,
		"created_at":  orderModel.CreatedAt,
		"status":      orderModel.Status,
	})
}

// 获取订单列表
func (h *Handler) GetOrders(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))

	if err := ValidatePaginationParams(page, pageSize); err != nil {
		errorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	requestShopID, err := strconv.ParseUint(c.Query("shop_id"), 10, 64)
	if err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的店铺ID")
		return
	}

	validShopID, err := h.validAndReturnShopID(c, requestShopID)
	if err != nil {
		errorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	// 使用 repository 查询订单
	orders, total, err := h.orderRepo.GetOrdersByShop(validShopID, page, pageSize)
	if err != nil {
		errorResponse(c, http.StatusInternalServerError, "获取订单列表失败")
		return
	}

	var simpleOrders []models.OrderElement
	// 初始化为空切片，确保即使没有订单也会返回[]
	simpleOrders = make([]models.OrderElement, 0)

	for _, order := range orders {
		simpleOrders = append(simpleOrders, models.OrderElement{
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

	successResponse(c, gin.H{
		"total":    total,
		"page":     page,
		"pageSize": pageSize,
		"data":     simpleOrders,
	})
}

// 获取订单详情
func (h *Handler) GetOrder(c *gin.Context) {
	id := c.Query("id")
	if id == "" {
		errorResponse(c, http.StatusBadRequest, "缺少订单ID")
		return
	}

	requestShopID, err := strconv.ParseUint(c.Query("shop_id"), 10, 64)
	if err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的店铺ID")
		return
	}

	validShopID, err := h.validAndReturnShopID(c, requestShopID)
	if err != nil {
		errorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	// 使用 repository 查询订单
	order, err := h.orderRepo.GetOrderByIDAndShopIDStr(id, validShopID)
	if err != nil {
		errorResponse(c, http.StatusNotFound, "订单未找到")
		return
	}

	successResponse(c, order)
}

// 查询某用户创建的所有订单
func (h *Handler) GetOrdersByUser(c *gin.Context) {
	userID := c.Query("user_id")
	if userID == "" {
		errorResponse(c, http.StatusBadRequest, "缺少用户ID")
		return
	}

	requestShopID, err := strconv.ParseUint(c.Query("shop_id"), 10, 64)
	if err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的店铺ID")
		return
	}

	validShopID, err := h.validAndReturnShopID(c, requestShopID)
	if err != nil {
		errorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	// 获取分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	if pageSize > 100 {
		pageSize = 100 // 限制最大页面大小
	}

	// 使用 repository 查询订单
	orders, total, err := h.orderRepo.GetOrdersByUser(userID, validShopID, page, pageSize)
	if err != nil {
		errorResponse(c, http.StatusInternalServerError, "查询用户订单失败")
		return
	}

	var simpleOrders []models.OrderElement
	// 初始化为空切片，确保即使没有订单也会返回[]
	simpleOrders = make([]models.OrderElement, 0)

	for _, order := range orders {
		simpleOrders = append(simpleOrders, models.OrderElement{
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

	successResponse(c, gin.H{
		"code":     200,
		"data":     simpleOrders,
		"total":    total,
		"page":     page,
		"pageSize": pageSize,
	})
}

// 更新订单
func (h *Handler) UpdateOrder(c *gin.Context) {
	id := c.Query("id")
	if id == "" {
		errorResponse(c, http.StatusBadRequest, "缺少订单ID")
		return
	}

	var order models.Order
	if err := h.DB.First(&order, id).Error; err != nil {
		h.logger.Errorf("更新订单失败, ID: %s, 错误: %v", id, err)
		errorResponse(c, http.StatusNotFound, "订单未找到")
		return
	}

	var updateData CreateOrderRequest
	if err := c.ShouldBindJSON(&updateData); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的更新数据: "+err.Error())
		return
	}

	// 构建领域服务 DTO
	var itemsDTO []orderdomain.CreateOrderItemDTO
	for _, item := range updateData.Items {
		var optionsDTO []orderdomain.CreateOrderItemOptionDTO
		for _, opt := range item.Options {
			optionsDTO = append(optionsDTO, orderdomain.CreateOrderItemOptionDTO{
				OptionID:   opt.OptionID,
				CategoryID: opt.CategoryID,
			})
		}
		itemsDTO = append(itemsDTO, orderdomain.CreateOrderItemDTO{
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
			Options:   optionsDTO,
		})
	}

	// 使用领域服务处理订单更新逻辑
	updatedOrder, _, err := h.orderService.UpdateOrder(orderdomain.UpdateOrderDTO{
		OrderID: order.ID,
		ShopID:  updateData.ShopID,
		Items:   itemsDTO,
		Remark:  updateData.Remark,
		Status:  updateData.Status,
	})
	if err != nil {
		h.logger.Errorf("更新订单失败: %v", err)
		errorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	// 开启事务
	tx := h.DB.Begin()

	// 删除原有的订单项和选项
	if err := tx.Where("order_id = ?", id).Delete(&models.OrderItem{}).Error; err != nil {
		tx.Rollback()
		h.logger.Errorf("删除原有订单项失败: %v", err)
		errorResponse(c, http.StatusInternalServerError, "更新订单失败")
		return
	}

	// 更新订单基本信息
	order.ShopID = updatedOrder.ShopID
	order.Remark = updatedOrder.Remark
	order.Status = updatedOrder.Status
	order.TotalPrice = updatedOrder.TotalPrice

	// 保存新的订单项（已由领域服务处理）
	for i := range updatedOrder.Items {
		updatedOrder.Items[i].OrderID = order.ID
	}
	if err := tx.Create(&updatedOrder.Items).Error; err != nil {
		tx.Rollback()
		h.logger.Errorf("创建新订单项失败: %v", err)
		errorResponse(c, http.StatusInternalServerError, "更新订单失败")
		return
	}

	// 更新订单信息
	if err := tx.Save(&order).Error; err != nil {
		tx.Rollback()
		h.logger.Errorf("更新订单信息失败: %v", err)
		errorResponse(c, http.StatusInternalServerError, "更新订单失败")
		return
	}

	tx.Commit()

	// 重新获取更新后的订单信息
	if err := h.DB.Preload("Items").Preload("Items.Options").First(&order, id).Error; err != nil {
		h.logger.Errorf("获取更新后的订单信息失败: %v", err)
		errorResponse(c, http.StatusInternalServerError, "获取更新后的订单信息失败")
		return
	}

	// 转换为响应格式
	var responseItems []CreateOrderItemRequest
	for _, item := range order.Items {
		var responseOptions []CreateOrderItemOption
		for _, option := range item.Options {
			responseOption := CreateOrderItemOption{
				CategoryID: option.CategoryID,
				OptionID:   option.OptionID,
			}
			responseOptions = append(responseOptions, responseOption)
		}

		responseItem := CreateOrderItemRequest{
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
			Price:     float64(item.Price),
			Options:   responseOptions,
		}
		responseItems = append(responseItems, responseItem)
	}

	response := CreateOrderRequest{
		ID:     order.ID,
		UserID: order.UserID,
		ShopID: order.ShopID,
		Items:  responseItems,
		Remark: order.Remark,
		Status: order.Status,
	}

	successResponse(c, response)
}

// 删除订单
func (h *Handler) DeleteOrder(c *gin.Context) {
	id := c.Query("id")
	if id == "" {
		errorResponse(c, http.StatusBadRequest, "缺少订单ID")
		return
	}

	requestShopID, err := strconv.ParseUint(c.Query("shop_id"), 10, 64)
	if err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的店铺ID")
		return
	}

	validShopID, err := h.validAndReturnShopID(c, requestShopID)
	if err != nil {
		errorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	var order models.Order
	if err := h.DB.Preload("Items").Where("shop_id = ?", validShopID).First(&order, id).Error; err != nil {
		h.logger.Errorf("删除订单失败, ID: %s, 错误: %v", id, err)
		errorResponse(c, http.StatusNotFound, "订单不存在")
		return
	}

	// 开启事务
	tx := h.DB.Begin()

	// 恢复商品库存（仅在订单未取消且未完成时）
	if order.Status != models.OrderStatusCanceled && order.Status != models.OrderStatusComplete {
		if err := h.orderService.RestoreStock(tx, order); err != nil {
			tx.Rollback()
			h.logger.Errorf("恢复商品库存失败: %v", err)
			errorResponse(c, http.StatusInternalServerError, "删除订单失败")
			return
		}
	}

	// 删除订单项
	if err := tx.Where("order_id = ?", id).Delete(&models.OrderItem{}).Error; err != nil {
		tx.Rollback()
		h.logger.Errorf("删除订单项失败: %v", err)
		errorResponse(c, http.StatusInternalServerError, "删除订单失败")
		return
	}

	// 删除订单状态日志
	if err := tx.Where("order_id = ?", id).Delete(&models.OrderStatusLog{}).Error; err != nil {
		tx.Rollback()
		h.logger.Errorf("删除订单状态日志失败: %v", err)
		errorResponse(c, http.StatusInternalServerError, "删除订单失败")
		return
	}

	// 删除订单
	if err := tx.Delete(&order).Error; err != nil {
		tx.Rollback()
		h.logger.Errorf("删除订单记录失败: %v", err)
		errorResponse(c, http.StatusInternalServerError, "删除订单失败")
		return
	}

	tx.Commit()
	successResponse(c, gin.H{"message": "订单删除成功"})
}

// 翻转订单状态
func (h *Handler) ToggleOrderStatus(c *gin.Context) {
	// 定义请求结构体
	type ToggleOrderStatusRequest struct {
		ID         string `json:"id" binding:"required"`
		ShopID     uint64 `json:"shop_id" binding:"required"`
		NextStatus int    `json:"next_status" binding:"required"`
	}

	// 绑定请求体
	var req ToggleOrderStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log2.Errorf("无效的请求参数: %v", err)
		errorResponse(c, http.StatusBadRequest, "无效的请求参数")
		return
	}

	// 将字符串ID转换为uint64
	orderID, err := strconv.ParseUint(req.ID, 10, 64)
	if err != nil {
		log2.Errorf("无效的订单ID格式: %v", err)
		errorResponse(c, http.StatusBadRequest, "无效的订单ID格式")
		return
	}

	// 验证店铺ID
	validShopID, err := h.validAndReturnShopID(c, req.ShopID)
	if err != nil {
		errorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	// 获取店铺信息，包括OrderStatusFlow
	var shop models.Shop
	if err := h.DB.First(&shop, validShopID).Error; err != nil {
		log2.Errorf("获取店铺信息失败: %v", err)
		errorResponse(c, http.StatusInternalServerError, "获取店铺信息失败")
		return
	}

	// 获取订单信息
	order, err := h.orderRepo.GetOrderByIDAndShopID(orderID, validShopID)
	if err != nil {
		errorResponse(c, http.StatusNotFound, err.Error())
		return
	}

	// 使用领域实体验证终态（基础验证）
	orderDomain := orderdomain.OrderFromModel(order)
	if orderDomain.IsFinal() {
		errorResponse(c, http.StatusBadRequest, "当前状态为终态，不允许转换")
		return
	}

	// 验证请求的next_status是否在店铺的订单流转定义中允许（店铺配置验证）
	if err := h.orderService.ValidateStatusTransition(order.Status, req.NextStatus, shop.OrderStatusFlow); err != nil {
		errorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	// 开启事务
	tx := h.DB.Begin()

	// 更新订单状态
	if err := tx.Model(&order).Update("status", req.NextStatus).Error; err != nil {
		tx.Rollback()
		log2.Errorf("更新订单状态失败: %v", err)
		errorResponse(c, http.StatusInternalServerError, "更新订单状态失败")
		return
	}

	// 记录状态变更
	if err := tx.Create(&models.OrderStatusLog{
		OrderID:     order.ID,
		OldStatus:   order.Status,
		NewStatus:   req.NextStatus,
		ChangedTime: time.Now(),
	}).Error; err != nil {
		tx.Rollback()
		log2.Errorf("记录状态变更失败: %v", err)
		errorResponse(c, http.StatusInternalServerError, "记录状态变更失败")
		return
	}

	tx.Commit()

	// 返回更新后的订单信息
	order.Status = req.NextStatus
	successResponse(c, gin.H{
		"message":    "订单状态更新成功",
		"old_status": order.Status,
		"new_status": req.NextStatus,
		"order":      order,
	})
}

// 添加错误响应辅助函数
func errorResponse(c *gin.Context, code int, message string) {
	log2.Errorf("错误响应: %d - %s", code, message)
	c.JSON(code, gin.H{"error": message})
}

// 添加成功响应辅助函数
func successResponse(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, data)
}

// 高级查询订单
func (h *Handler) GetAdvanceSearchOrders(c *gin.Context) {
	var req AdvanceSearchOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的查询参数: "+err.Error())
		return
	}

	// 验证分页参数
	if req.Page < 1 {
		req.Page = 1
	}
	if req.PageSize < 1 || req.PageSize > 100 {
		req.PageSize = 10
	}

	// 验证并获取有效的店铺ID
	validShopID, err := h.validAndReturnShopID(c, req.ShopID)
	if err != nil {
		errorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	// 构建查询
	query := h.DB.Model(&models.Order{}).Where("shop_id = ?", validShopID)

	// 添加用户ID筛选
	if req.UserID != "" {
		query = query.Where("user_id = ?", req.UserID)
	}

	// 添加状态筛选（支持多个状态，直接使用数组）
	if len(req.Status) > 0 {
		query = query.Where("status IN (?)", req.Status)
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
		h.logger.Errorf("获取订单总数失败: %v", err)
		errorResponse(c, http.StatusInternalServerError, "获取订单列表失败")
		return
	}

	// 分页查询
	offset := (req.Page - 1) * req.PageSize
	var orders []models.Order
	if err := query.Offset(offset).Limit(req.PageSize).
		Order("created_at DESC").
		Find(&orders).Error; err != nil {
		h.logger.Errorf("查询订单列表失败: %v", err)
		errorResponse(c, http.StatusInternalServerError, "获取订单列表失败")
		return
	}

	// 转换为响应格式
	var simpleOrders []models.OrderElement
	simpleOrders = make([]models.OrderElement, 0, len(orders))

	for _, order := range orders {
		simpleOrders = append(simpleOrders, models.OrderElement{
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

	successResponse(c, gin.H{
		"total":    total,
		"page":     req.Page,
		"pageSize": req.PageSize,
		"data":     simpleOrders,
	})
}

// 验证用户ID的合法性
func (h *Handler) IsValidUserID(userID snowflake.ID) bool {
	var user models.User
	err := h.DB.First(&user, userID).Error
	return err == nil
}

// 获取订单状态流转配置
func (h *Handler) GetOrderStatusFlow(c *gin.Context) {
	// 获取并验证shop_id参数
	requestShopID, err := strconv.ParseUint(c.Query("shop_id"), 10, 64)
	if err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的店铺ID")
		return
	}

	// 验证店铺ID是否有效
	validShopID, err := h.validAndReturnShopID(c, requestShopID)
	if err != nil {
		errorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	// 查询店铺信息，获取OrderStatusFlow
	orderStatusFlow, err := h.shopRepo.GetOrderStatusFlow(validShopID)
	if err != nil {
		h.logger.Errorf("获取店铺订单状态流转配置失败: %v", err)
		errorResponse(c, http.StatusInternalServerError, "获取订单状态流转配置失败")
		return
	}

	successResponse(c, gin.H{
		"shop_id":           validShopID,
		"order_status_flow": orderStatusFlow,
	})
}

// 获取未完成的订单列表
func (h *Handler) GetUnfinishedOrders(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))

	if page < 1 {
		errorResponse(c, http.StatusBadRequest, "页码必须大于0")
		return
	}

	if pageSize < 1 || pageSize > 100 {
		errorResponse(c, http.StatusBadRequest, "每页数量必须在1-100之间")
		return
	}

	requestShopID, err := strconv.ParseUint(c.Query("shop_id"), 10, 64)
	if err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的店铺ID")
		return
	}

	validShopID, err := h.validAndReturnShopID(c, requestShopID)
	if err != nil {
		errorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	// 获取店铺的订单状态流转配置
	orderStatusFlow, err := h.shopRepo.GetOrderStatusFlow(validShopID)
	if err != nil {
		h.logger.Errorf("获取店铺订单状态流转配置失败: %v", err)
		errorResponse(c, http.StatusInternalServerError, "获取未完成订单列表失败")
		return
	}

	// 收集所有isFinal为false的状态值
	var unfinishedStatuses []int
	for _, status := range orderStatusFlow.Statuses {
		if !status.IsFinal {
			unfinishedStatuses = append(unfinishedStatuses, status.Value)
		}
	}

	// 如果没有未完成状态，直接返回空列表
	if len(unfinishedStatuses) == 0 {
		successResponse(c, gin.H{
			"total":    0,
			"page":     page,
			"pageSize": pageSize,
			"data":     []models.OrderElement{},
		})
		return
	}

	// 使用 repository 查询未完成订单
	orders, total, err := h.orderRepo.GetUnfinishedOrders(validShopID, unfinishedStatuses, page, pageSize)
	if err != nil {
		errorResponse(c, http.StatusInternalServerError, "获取未完成订单列表失败")
		return
	}

	var simpleOrders []models.OrderElement
	// 初始化为空切片，确保即使没有订单也会返回[]
	simpleOrders = make([]models.OrderElement, 0)

	for _, order := range orders {
		simpleOrders = append(simpleOrders, models.OrderElement{
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

	successResponse(c, gin.H{
		"total":    total,
		"page":     page,
		"pageSize": pageSize,
		"data":     simpleOrders,
	})
}
