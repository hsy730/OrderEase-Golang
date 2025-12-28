package handlers

import (
	"fmt"
	"net/http"
	"orderease/models"
	"orderease/utils"
	"orderease/utils/log2"
	"strconv"
	"strings"
	"time"

	"github.com/bwmarrin/snowflake"
	"github.com/gin-gonic/gin"
)

// 高级查询订单请求

type AdvanceSearchOrderRequest struct {
	Page      int    `json:"page"`
	PageSize  int    `json:"pageSize"`
	UserID    string `json:"user_id"`
	Status    string `json:"status"`
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

	var orderItems []models.OrderItem
	for _, itemReq := range req.Items {
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

	order := models.Order{
		ID:     req.ID,
		UserID: req.UserID,
		ShopID: req.ShopID,
		Items:  orderItems,
		Remark: req.Remark,
		Status: 0,
	}
	utils.SanitizeOrder(&order)

	// 验证订单数据
	if err := validateOrder(&order); err != nil {
		errorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	// 假设存在一个 IsValidUserID 函数来验证用户ID的合法性
	if !h.IsValidUserID(order.UserID) {
		log2.Errorf("创建订单失败: 非法用户")
		errorResponse(c, http.StatusBadRequest, "创建订单失败")
		return
	}

	validShopID, err := h.validAndReturnShopID(c, order.ShopID)
	if err != nil {
		errorResponse(c, http.StatusBadRequest, err.Error())
		return
	}
	order.ShopID = validShopID // 更新订单的shopID

	tx := h.DB.Begin()

	totalPrice := float64(0.0)
	// 更新商品库存并保存商品快照
	for i := range order.Items {
		var product models.Product
		if err := tx.First(&product, order.Items[i].ProductID).Error; err != nil {
			tx.Rollback()
			h.logger.Errorf("商品不存在, ID: %d, 错误: %v", order.Items[i].ProductID, err)
			errorResponse(c, http.StatusBadRequest, "商品不存在")
			return
		}

		if product.Stock < order.Items[i].Quantity {
			tx.Rollback()
			h.logger.Errorf("商品库存不足, ID: %d, 当前库存: %d, 需求数量: %d",
				order.Items[i].ProductID, product.Stock, order.Items[i].Quantity)
			errorResponse(c, http.StatusBadRequest, fmt.Sprintf("商品 %s 库存不足", product.Name))
			return
		}

		// 保存商品快照信息
		order.Items[i].ProductName = product.Name
		order.Items[i].ProductDescription = product.Description
		order.Items[i].ProductImageURL = product.ImageURL
		order.Items[i].Price = models.Price(product.Price) // 使用当前价格

		// 处理订单项参数选项
		itemTotalPrice := float64(order.Items[i].Quantity) * product.Price
		for j := range order.Items[i].Options {
			// 获取参数选项信息
			var option models.ProductOption
			if err := tx.First(&option, order.Items[i].Options[j].OptionID).Error; err != nil {
				tx.Rollback()
				h.logger.Errorf("商品参数选项不存在, ID: %d, 错误: %v", order.Items[i].Options[j].OptionID, err)
				errorResponse(c, http.StatusBadRequest, "无效的商品参数选项")
				return
			}

			// 获取参数类别信息
			var category models.ProductOptionCategory
			if err := tx.First(&category, option.CategoryID).Error; err != nil {
				tx.Rollback()
				h.logger.Errorf("商品参数类别不存在, ID: %d, 错误: %v", option.CategoryID, err)
				errorResponse(c, http.StatusBadRequest, "无效的商品参数类别")
				return
			}

			// 验证参数所属商品
			if category.ProductID != product.ID {
				tx.Rollback()
				errorResponse(c, http.StatusBadRequest, "参数选项不属于指定商品")
				return
			}

			// 保存参数选项快照
			order.Items[i].Options[j].CategoryID = category.ID
			order.Items[i].Options[j].OptionName = option.Name
			order.Items[i].Options[j].CategoryName = category.Name
			order.Items[i].Options[j].PriceAdjustment = option.PriceAdjustment

			// 计算参数选项对总价的影响
			itemTotalPrice += float64(order.Items[i].Quantity) * option.PriceAdjustment
		}

		// 设置订单项总价
		order.Items[i].TotalPrice = models.Price(itemTotalPrice)

		// 更新库存
		product.Stock -= order.Items[i].Quantity
		totalPrice += itemTotalPrice
		if err := tx.Save(&product).Error; err != nil {
			tx.Rollback()
			h.logger.Errorf("更新商品库存失败: %v", err)
			errorResponse(c, http.StatusInternalServerError, "更新商品库存失败")
			return
		}
	}

	order.TotalPrice = models.Price(totalPrice)
	// 雪花ID生成逻辑
	order.ID = utils.GenerateSnowflakeID()
	// 设置订单初始状态
	order.Status = models.OrderStatusPending

	// 数据库写入
	if err := tx.Create(&order).Error; err != nil {
		tx.Rollback()
		log2.Errorf("创建订单失败: %v", err)
		errorResponse(c, http.StatusInternalServerError, "创建订单失败")
		return
	}

	// 更新订单项选项的OrderItemID
	for i := range order.Items {
		for j := range order.Items[i].Options {
			order.Items[i].Options[j].OrderItemID = order.Items[i].ID
			if err := tx.Save(&order.Items[i].Options[j]).Error; err != nil {
				tx.Rollback()
				log2.Errorf("更新订单项选项失败: %v", err)
				errorResponse(c, http.StatusInternalServerError, "创建订单失败")
				return
			}
		}
	}

	// 添加日志，打印创建的订单信息
	log2.Infof("创建的订单信息: %+v", order)
	for _, item := range order.Items {
		log2.Infof("订单项ID: %s, 选项数量: %d", item.ID, len(item.Options))
		for _, option := range item.Options {
			log2.Infof("选项ID: %s, 名称: %s", option.ID, option.OptionName)
		}
	}

	// 创建订单状态日志
	statusLog := models.OrderStatusLog{
		OrderID:     order.ID,
		OldStatus:   0,
		NewStatus:   order.Status,
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
	go h.NotifyNewOrder(order)

	successResponse(c, gin.H{
		"order_id":    order.ID,
		"total_price": order.TotalPrice,
		"created_at":  order.CreatedAt,
		"status":      order.Status,
	})
}

// 获取订单列表
func (h *Handler) GetOrders(c *gin.Context) {
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

	offset := (page - 1) * pageSize

	var total int64
	if err := h.DB.Model(&models.Order{}).Where("shop_id = ?", validShopID).Count(&total).Error; err != nil {
		h.logger.Errorf("获取订单总数失败: %v", err)
		errorResponse(c, http.StatusInternalServerError, "获取订单列表失败")
		return
	}

	var orders []models.Order
	// 预加载Items和Items.Options
	if err := h.DB.Where("shop_id = ?", validShopID).Offset(offset).Limit(pageSize).
		Order("created_at DESC").
		Find(&orders).Error; err != nil {
		h.logger.Errorf("查询订单列表失败: %v", err)
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

	var order models.Order
	// 预加载Items和Items.Options
	if err := h.DB.Preload("Items").
		Preload("Items.Options").
		Where("shop_id = ?", validShopID).
		Joins("User").
		First(&order, id).Error; err != nil {
		h.logger.Errorf("查询订单失败, ID: %s, 错误: %v", id, err)
		errorResponse(c, http.StatusNotFound, "订单未找到")
		return
	}

	// 添加日志，打印查询到的订单信息
	h.logger.Infof("查询到的订单信息: %+v", order)
	for _, item := range order.Items {
		h.logger.Infof("订单项ID: %s, 选项数量: %d", item.ID, len(item.Options))
		for _, option := range item.Options {
			h.logger.Infof("选项ID: %s, 名称: %s", option.ID, option.OptionName)
		}
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
	offset := (page - 1) * pageSize

	var orders []models.Order
	// 预加载Items和Items.Options
	query := h.DB.Where("user_id = ?", userID).Where("shop_id = ?", validShopID)

	// 获取总数
	var total int64
	if err := query.Model(&models.Order{}).Count(&total).Error; err != nil {
		log2.Errorf("查询订单总数失败, 用户ID: %s, 错误: %v", userID, err)
		errorResponse(c, http.StatusInternalServerError, "查询订单总数失败")
		return
	}

	// 分页查询订单
	if err := query.Order("created_at DESC").
		Offset(offset).Limit(pageSize).
		// Preload("Items").
		// Preload("Items.Options").
		Find(&orders).Error; err != nil {
		log2.Errorf("查询用户订单失败, 用户ID: %s, 错误: %v", userID, err)
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

	// 开启事务
	tx := h.DB.Begin()

	// 更新订单基本信息
	order.ShopID = updateData.ShopID
	order.Remark = updateData.Remark
	order.Status = updateData.Status

	// 删除原有的订单项和选项
	if err := tx.Where("order_id = ?", id).Delete(&models.OrderItem{}).Error; err != nil {
		tx.Rollback()
		h.logger.Errorf("删除原有订单项失败: %v", err)
		errorResponse(c, http.StatusInternalServerError, "更新订单失败")
		return
	}

	totalPrice := float64(0.0)

	// 创建新的订单项
	var orderItems []models.OrderItem
	for _, itemReq := range updateData.Items {
		var product models.Product
		if err := tx.First(&product, itemReq.ProductID).Error; err != nil {
			tx.Rollback()
			h.logger.Errorf("商品不存在, ID: %d, 错误: %v", itemReq.ProductID, err)
			errorResponse(c, http.StatusBadRequest, "商品不存在")
			return
		}

		orderItem := models.OrderItem{
			OrderID:   order.ID,
			ProductID: itemReq.ProductID,
			Quantity:  itemReq.Quantity,
			Price:     models.Price(product.Price), // 获取当前商品价格
		}

		// 处理选中的选项
		var options []models.OrderItemOption

		itemTotalPrice := float64(orderItem.Quantity) * product.Price

		for _, optionReq := range itemReq.Options {
			// 获取参数选项信息
			var option models.ProductOption
			if err := tx.First(&option, optionReq.OptionID).Error; err != nil {
				tx.Rollback()
				h.logger.Errorf("商品参数选项不存在, ID: %d, 错误: %v", optionReq.OptionID, err)
				errorResponse(c, http.StatusBadRequest, "无效的商品参数选项")
				return
			}

			// 获取参数类别信息
			var category models.ProductOptionCategory
			if err := tx.First(&category, option.CategoryID).Error; err != nil {
				tx.Rollback()
				h.logger.Errorf("商品参数类别不存在, ID: %d, 错误: %v", option.CategoryID, err)
				errorResponse(c, http.StatusBadRequest, "无效的商品参数类别")
				return
			}
			// 计算参数选项对总价的影响
			// 保存参数选项快照
			options = append(options, models.OrderItemOption{
				OrderItemID:     orderItem.ID,
				OptionID:        optionReq.OptionID,
				CategoryID:      optionReq.CategoryID,
				OptionName:      option.Name,
				CategoryName:    category.Name,
				PriceAdjustment: option.PriceAdjustment,
			})
			itemTotalPrice += float64(orderItem.Quantity) * option.PriceAdjustment
		}
		// 设置订单项总价
		orderItem.Options = options
		orderItem.TotalPrice = models.Price(itemTotalPrice)
		orderItems = append(orderItems, orderItem)
		totalPrice += itemTotalPrice
	}

	// 保存新的订单项
	if err := tx.Create(&orderItems).Error; err != nil {
		tx.Rollback()
		h.logger.Errorf("创建新订单项失败: %v", err)
		errorResponse(c, http.StatusInternalServerError, "更新订单失败")
		return
	}

	order.TotalPrice = models.Price(totalPrice)

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
	if err := h.DB.Where("shop_id = ?", validShopID).First(&order, id).Error; err != nil {
		h.logger.Errorf("删除订单失败, ID: %s, 错误: %v", id, err)
		errorResponse(c, http.StatusNotFound, "订单不存在")
		return
	}

	// 开启事务
	tx := h.DB.Begin()

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
	order, err := h.productRepo.GetOrderByIDAndShopID(orderID, validShopID)
	if err != nil {
		errorResponse(c, http.StatusNotFound, err.Error())
		return
	}

	// 验证请求的next_status是否在店铺的订单流转定义中允许
	if err := validateNextStatus(order.Status, req.NextStatus, shop.OrderStatusFlow); err != nil {
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

// validateNextStatus 验证请求的next_status是否在店铺的订单流转定义中允许
func validateNextStatus(currentStatus int, nextStatus int, flow models.OrderStatusFlow) error {
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

// 添加错误响应辅助函数
func errorResponse(c *gin.Context, code int, message string) {
	log2.Errorf("错误响应: %d - %s", code, message)
	c.JSON(code, gin.H{"error": message})
}

// 添加成功响应辅助函数
func successResponse(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, data)
}

// 添加订单验证函数
func validateOrder(order *models.Order) error {
	if order.UserID == 0 {
		return fmt.Errorf("用户ID不能为空")
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

	// 添加状态筛选（支持多个状态，用逗号分隔）
	if req.Status != "" {
		statuses := strings.Split(req.Status, ",")
		query = query.Where("status IN (?)", statuses)
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
	var shop models.Shop
	if err := h.DB.Select("order_status_flow").First(&shop, validShopID).Error; err != nil {
		h.logger.Errorf("获取店铺订单状态流转配置失败: %v", err)
		errorResponse(c, http.StatusInternalServerError, "获取订单状态流转配置失败")
		return
	}

	successResponse(c, gin.H{
		"shop_id":           validShopID,
		"order_status_flow": shop.OrderStatusFlow,
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

	offset := (page - 1) * pageSize

	var total int64
	// 查询未完成订单总数，status != 10
	if err := h.DB.Model(&models.Order{}).Where("shop_id = ? AND status != ?", validShopID, models.OrderStatusComplete).Count(&total).Error; err != nil {
		h.logger.Errorf("获取未完成订单总数失败: %v", err)
		errorResponse(c, http.StatusInternalServerError, "获取未完成订单列表失败")
		return
	}

	var orders []models.Order
	// 预加载Items和Items.Options
	if err := h.DB.Where("shop_id = ? AND status != ?", validShopID, models.OrderStatusComplete).Offset(offset).Limit(pageSize).
		Order("created_at DESC").
		Find(&orders).Error; err != nil {
		h.logger.Errorf("查询未完成订单列表失败: %v", err)
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
