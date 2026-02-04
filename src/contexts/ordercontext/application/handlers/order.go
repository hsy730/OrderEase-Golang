package handlers

import (
	"fmt"
	"net/http"
	orderdomain "orderease/contexts/ordercontext/domain/order"
	"orderease/contexts/ordercontext/domain/user"
	"orderease/contexts/ordercontext/infrastructure/repositories"
	"orderease/models"
	"orderease/utils"
	"orderease/utils/log2"
	"strconv"
	"time"

	"github.com/bwmarrin/snowflake"
	"github.com/gin-gonic/gin"
)

// 使用 Domain 层的 DTO 定义
// - orderdomain.CreateOrderRequest
// - orderdomain.CreateOrderItemRequest
// - orderdomain.CreateOrderItemOption
// - orderdomain.AdvanceSearchOrderRequest
// - orderdomain.ToggleOrderStatusRequest

// 创建订单请求结构体，添加参数支持
func (h *Handler) CreateOrder(c *gin.Context) {
	var req orderdomain.CreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的订单数据: "+err.Error())
		return
	}

	// 使用 Domain DTO 的验证方法
	if err := req.Validate(); err != nil {
		errorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	// 假设存在一个 IsValidUserID 函数来验证用户ID的合法性
	if !h.IsValidUserID(req.UserID) {
		log2.Errorf("创建订单失败: 非法用户")
		errorResponse(c, http.StatusBadRequest, "创建订单失败")
		return
	}

	validShopID, err := h.validAndReturnShopID(c, snowflake.ID(req.ShopID))
	if err != nil {
		errorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

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
		h.logger.Errorf("创建订单失败: %v", err)
		errorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	orderModel.TotalPrice = models.Price(totalPrice)
	// 雪花ID生成逻辑
	orderModel.ID = utils.GenerateSnowflakeID()

	// 添加日志，打印创建的订单信息
	log2.Infof("创建的订单信息: %+v", orderModel)
	for _, item := range orderModel.Items {
		log2.Infof("订单项ID: %s, 选项数量: %d", item.ID, len(item.Options))
		for _, option := range item.Options {
			log2.Infof("选项ID: %s, 名称: %s", option.ID, option.OptionName)
		}
	}

	// 使用 Repository 创建订单（包含订单项和选项）
	if err := h.orderRepo.CreateOrder(orderModel); err != nil {
		log2.Errorf("创建订单失败: %v", err)
		errorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	// 创建订单状态日志
	statusLog := models.OrderStatusLog{
		OrderID:     orderModel.ID,
		OldStatus:   0,
		NewStatus:   orderModel.Status,
		ChangedTime: time.Now(),
	}
	if err := h.orderRepo.CreateOrderStatusLog(&statusLog); err != nil {
		log2.Errorf("创建订单状态日志失败: %v", err)
		// 订单已创建，但状态日志失败，仅记录错误
	}

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

	requestShopID, err := utils.StringToSnowflakeID(c.Query("shop_id"))
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

	// 使用 Domain 辅助函数转换订单列表
	simpleOrders := orderdomain.ToOrderElements(orders)

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

	requestShopID, err := utils.StringToSnowflakeID(c.Query("shop_id"))
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

	requestShopID, err := utils.StringToSnowflakeID(c.Query("shop_id"))
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

	// 使用 Domain 辅助函数转换订单列表
	simpleOrders := orderdomain.ToOrderElements(orders)

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
	idStr := c.Query("id")
	if idStr == "" {
		errorResponse(c, http.StatusBadRequest, "缺少订单ID")
		return
	}

	// 将字符串类型的 ID 转换为整数（与其他 Handler 保持一致）
	// 先验证 ID 格式是否正确
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		h.logger.Errorf("无效的订单ID格式: %s, 错误: %v", idStr, err)
		errorResponse(c, http.StatusBadRequest, "无效的订单ID")
		return
	}

	// 使用转换后的整数 ID 进行查询（转换为字符串）
	order, err := h.orderRepo.GetByIDStr(fmt.Sprint(id))
	if err != nil {
		h.logger.Errorf("更新订单失败, ID: %d, 错误: %v", id, err)
		errorResponse(c, http.StatusNotFound, "订单未找到")
		return
	}

	var updateData orderdomain.CreateOrderRequest
	if err := c.ShouldBindJSON(&updateData); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的更新数据: "+err.Error())
		return
	}

	// 使用 Domain DTO 的验证方法
	if err := updateData.Validate(); err != nil {
		errorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	// 验证 shop_id（与 CreateOrder 保持一致）
	validShopID, err := h.validAndReturnShopID(c, snowflake.ID(updateData.ShopID))
	if err != nil {
		errorResponse(c, http.StatusBadRequest, err.Error())
		return
	}
	updateData.ShopID = validShopID

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

	// 更新订单基本信息
	order.ShopID = updatedOrder.ShopID
	order.Remark = updatedOrder.Remark
	order.Status = updatedOrder.Status
	order.TotalPrice = updatedOrder.TotalPrice

	// 使用 Repository 更新订单（包含删除旧订单项、创建新订单项）
	if err := h.orderRepo.UpdateOrder(order, updatedOrder.Items); err != nil {
		h.logger.Errorf("更新订单失败: %v", err)
		errorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	// 重新获取更新后的订单信息
	order, err = h.orderRepo.GetByIDStrWithItems(fmt.Sprint(id))
	if err != nil {
		h.logger.Errorf("获取更新后的订单信息失败: %v", err)
		errorResponse(c, http.StatusInternalServerError, "获取更新后的订单信息失败")
		return
	}

	// 转换为领域实体并使用 ToCreateOrderRequest 方法
	orderEntity := orderdomain.OrderFromModel(order)
	response := orderEntity.ToCreateOrderRequest()

	successResponse(c, response)
}

// 删除订单
func (h *Handler) DeleteOrder(c *gin.Context) {
	id := c.Query("id")
	if id == "" {
		errorResponse(c, http.StatusBadRequest, "缺少订单ID")
		return
	}

	requestShopID, err := utils.StringToSnowflakeID(c.Query("shop_id"))
	if err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的店铺ID")
		return
	}

	validShopID, err := h.validAndReturnShopID(c, requestShopID)
	if err != nil {
		errorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	order, err := h.orderRepo.GetOrderByIDAndShopIDStr(id, validShopID)
	if err != nil {
		h.logger.Errorf("删除订单失败, ID: %s, 错误: %v", id, err)
		errorResponse(c, http.StatusNotFound, "订单不存在")
		return
	}

	// 开启事务
	tx := h.DB.Begin()

	// 使用领域实体验证是否可删除
	orderDomain := orderdomain.OrderFromModel(order)
	if !orderDomain.CanBeDeleted() {
		// 订单已取消或已完成，无需恢复库存
	} else {
		// 恢复商品库存（仅在订单未取消且未完成时）
		if err := h.orderService.RestoreStock(tx, *order); err != nil {
			tx.Rollback()
			h.logger.Errorf("恢复商品库存失败: %v", err)
			errorResponse(c, http.StatusInternalServerError, "删除订单失败")
			return
		}
	}

	// 使用 Repository 删除订单及其关联数据
	if err := h.orderRepo.DeleteOrderInTx(tx, id, validShopID); err != nil {
		tx.Rollback()
		h.logger.Errorf("删除订单失败: %v", err)
		errorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	tx.Commit()
	successResponse(c, gin.H{"message": "订单删除成功"})
}

// 翻转订单状态
func (h *Handler) ToggleOrderStatus(c *gin.Context) {
	// 使用 Domain DTO
	var req orderdomain.ToggleOrderStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log2.Errorf("无效的请求参数: %v", err)
		errorResponse(c, http.StatusBadRequest, "无效的请求参数")
		return
	}

	// 使用 Domain DTO 的验证方法
	if err := req.Validate(); err != nil {
		errorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	// 验证店铺ID
	validShopID, err := h.validAndReturnShopID(c, req.ShopID)
	if err != nil {
		errorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	// 获取店铺信息，包括OrderStatusFlow（使用 Repository）
	shop, err := h.shopRepo.GetShopByID(validShopID)
	if err != nil {
		log2.Errorf("获取店铺信息失败: %v", err)
		errorResponse(c, http.StatusInternalServerError, "获取店铺信息失败")
		return
	}

	// 获取订单信息
	order, err := h.orderRepo.GetOrderByIDAndShopID(uint64(req.ID), validShopID)
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

	// 使用 Repository 更新订单状态并记录状态变更
	if err := h.orderRepo.UpdateOrderStatusInTx(tx, order, req.NextStatus); err != nil {
		tx.Rollback()
		log2.Errorf("更新订单状态失败: %v", err)
		errorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	tx.Commit()

	// 返回更新后的订单信息
	oldStatus := order.Status
	order.Status = req.NextStatus
	successResponse(c, gin.H{
		"message":    "订单状态更新成功",
		"old_status": oldStatus,
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
	var req orderdomain.AdvanceSearchOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的查询参数: "+err.Error())
		return
	}

	// 使用 Domain DTO 的验证方法（包含分页参数和店铺ID验证）
	if err := req.Validate(); err != nil {
		errorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	// 验证并获取有效的店铺ID
	validShopID, err := h.validAndReturnShopID(c, snowflake.ID(req.ShopID))
	if err != nil {
		errorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	// 使用 Repository 执行高级搜索
	result, err := h.orderRepo.AdvanceSearch(repositories.AdvanceSearchOrderRequest{
		Page:      req.Page,
		PageSize:  req.PageSize,
		UserID:    req.UserID,
		Status:    req.Status,
		StartTime: req.StartTime,
		EndTime:   req.EndTime,
		ShopID:    validShopID,
	})
	if err != nil {
		h.logger.Errorf("查询订单列表失败: %v", err)
		errorResponse(c, http.StatusInternalServerError, "获取订单列表失败")
		return
	}

	// 使用 Domain 辅助函数转换订单列表
	simpleOrders := orderdomain.ToOrderElements(result.Orders)

	successResponse(c, gin.H{
		"total":    result.Total,
		"page":     req.Page,
		"pageSize": req.PageSize,
		"data":     simpleOrders,
	})
}

// 验证用户ID的合法性（使用 Domain Service）
func (h *Handler) IsValidUserID(userID snowflake.ID) bool {
	_, err := h.userDomain.GetByID(user.UserID(fmt.Sprintf("%d", userID)))
	return err == nil
}

// 获取订单状态流转配置
func (h *Handler) GetOrderStatusFlow(c *gin.Context) {
	// 获取并验证shop_id参数
	requestShopID, err := utils.StringToSnowflakeID(c.Query("shop_id"))
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

	requestShopID, err := utils.StringToSnowflakeID(c.Query("shop_id"))
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

	// 使用 Domain 辅助函数转换订单列表
	simpleOrders := orderdomain.ToOrderElements(orders)

	successResponse(c, gin.H{
		"total":    total,
		"page":     page,
		"pageSize": pageSize,
		"data":     simpleOrders,
	})
}
