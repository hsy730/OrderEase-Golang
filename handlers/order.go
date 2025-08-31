package handlers

import (
	"fmt"
	"net/http"
	"orderease/models"
	"orderease/utils"
	"orderease/utils/log2"
	"strconv"
	"time"

	"github.com/bwmarrin/snowflake"
	"github.com/gin-gonic/gin"
)

// 创建订单
type CreateOrderRequest struct {
	UserID snowflake.ID       `json:"user_id"`
	ShopID uint64             `json:"shop_id"`
	Items  []models.OrderItem `json:"items"`
	Remark string             `json:"remark"`
}

// 创建订单请求结构体，添加参数支持
func (h *Handler) CreateOrder(c *gin.Context) {
	var req CreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的订单数据: " + err.Error())
		return
	}

	order := models.Order{
		UserID: req.UserID,
		ShopID: req.ShopID,
		Items:  req.Items,
		Remark: req.Remark,
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
			h.logger.Printf("商品不存在, ID: %d, 错误: %v", order.Items[i].ProductID, err)
			errorResponse(c, http.StatusBadRequest, "商品不存在")
			return
		}

		if product.Stock < order.Items[i].Quantity {
			tx.Rollback()
			h.logger.Printf("商品库存不足, ID: %d, 当前库存: %d, 需求数量: %d",
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
				h.logger.Printf("商品参数选项不存在, ID: %d, 错误: %v", order.Items[i].Options[j].OptionID, err)
				errorResponse(c, http.StatusBadRequest, "无效的商品参数选项")
				return
			}

			// 获取参数类别信息
			var category models.ProductOptionCategory
			if err := tx.First(&category, option.CategoryID).Error; err != nil {
				tx.Rollback()
				h.logger.Printf("商品参数类别不存在, ID: %d, 错误: %v", option.CategoryID, err)
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
			h.logger.Printf("更新商品库存失败: %v", err)
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

	// 创建订单状态日志
	statusLog := models.OrderStatusLog{
		OrderID:     order.ID,
		OldStatus:   "",
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
		h.logger.Printf("获取订单总数失败: %v", err)
		errorResponse(c, http.StatusInternalServerError, "获取订单列表失败")
		return
	}

	var orders []models.Order
	// 预加载Items和Items.Options
	if err := h.DB.Where("shop_id = ?", validShopID).Offset(offset).Limit(pageSize).
		Preload("Items").
		Preload("Items.Options").
		Find(&orders).Error; err != nil {
		h.logger.Printf("查询订单列表失败: %v", err)
		errorResponse(c, http.StatusInternalServerError, "获取订单列表失败")
		return
	}

	successResponse(c, gin.H{
		"total":    total,
		"page":     page,
		"pageSize": pageSize,
		"data":     orders,
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
		h.logger.Printf("查询订单失败, ID: %s, 错误: %v", id, err)
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

	var orders []models.Order
	// 预加载Items和Items.Options
	if err := h.DB.Where("user_id = ?", userID).Where("shop_id = ?", validShopID).
		Preload("Items").
		Preload("Items.Options").
		Find(&orders).Error; err != nil {
		log2.Errorf("查询用户订单失败, 用户ID: %s, 错误: %v", userID, err)
		errorResponse(c, http.StatusInternalServerError, "查询用户订单失败")
		return
	}

	successResponse(c, gin.H{
		"code": 200,
		"data": orders,
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
		h.logger.Printf("更新订单失败, ID: %s, 错误: %v", id, err)
		errorResponse(c, http.StatusNotFound, "订单未找到")
		return
	}

	var updateData models.Order
	if err := c.ShouldBindJSON(&updateData); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的更新数据: "+err.Error())
		return
	}

	utils.SanitizeOrder(&updateData)

	validShopID, err := h.validAndReturnShopID(c, order.ShopID)
	if err != nil {
		errorResponse(c, http.StatusBadRequest, err.Error())
		return
	}
	order.ShopID = validShopID // 更新订单的shopID

	// 开启事务
	tx := h.DB.Begin()

	if err := tx.Model(&order).Updates(updateData).Error; err != nil {
		tx.Rollback()
		h.logger.Printf("更新订单失败: %v", err)
		errorResponse(c, http.StatusInternalServerError, "更新订单失败")
		return
	}

	// 重新获取更新后的订单信息
	if err := tx.Preload("Items").First(&order, id).Error; err != nil {
		tx.Rollback()
		h.logger.Printf("获取更新后的订单信息失败: %v", err)
		errorResponse(c, http.StatusInternalServerError, "获取更新后的订单信息失败")
		return
	}

	tx.Commit()
	successResponse(c, order)
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
		h.logger.Printf("删除订单失败, ID: %s, 错误: %v", id, err)
		errorResponse(c, http.StatusNotFound, "订单不存在")
		return
	}

	// 开启事务
	tx := h.DB.Begin()

	// 删除订单项
	if err := tx.Where("order_id = ?", id).Delete(&models.OrderItem{}).Error; err != nil {
		tx.Rollback()
		h.logger.Printf("删除订单项失败: %v", err)
		errorResponse(c, http.StatusInternalServerError, "删除订单失败")
		return
	}

	// 删除订单状态日志
	if err := tx.Where("order_id = ?", id).Delete(&models.OrderStatusLog{}).Error; err != nil {
		tx.Rollback()
		h.logger.Printf("删除订单状态日志失败: %v", err)
		errorResponse(c, http.StatusInternalServerError, "删除订单失败")
		return
	}

	// 删除订单
	if err := tx.Delete(&order).Error; err != nil {
		tx.Rollback()
		h.logger.Printf("删除订单记录失败: %v", err)
		errorResponse(c, http.StatusInternalServerError, "删除订单失败")
		return
	}

	tx.Commit()
	successResponse(c, gin.H{"message": "订单删除成功"})
}

// 翻转订单状态
func (h *Handler) ToggleOrderStatus(c *gin.Context) {
	id, err := strconv.ParseUint(c.Query("id"), 10, 64)
	if err != nil {
		log2.Errorf("无效的订单ID: %v", err)
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

	order, err := h.productRepo.GetOrderByIDAndShopID(id, validShopID)
	if err != nil {
		errorResponse(c, http.StatusNotFound, err.Error())
		return
	}

	// 检查当前状态是否允许转换
	if !isValidStatusTransition(order.Status) {
		errorResponse(c, http.StatusBadRequest, "当前状态不允许转换")
		return
	}

	// 获取下一个状态
	nextStatus, exists := models.OrderStatusTransitions[order.Status]
	if !exists {
		nextStatus = models.OrderStatusPending // 如果当前状态未定义转换，重置为待处理
	}

	// 开启事务
	tx := h.DB.Begin()

	// 更新订单状态
	if err := tx.Model(&order).Update("status", nextStatus).Error; err != nil {
		tx.Rollback()
		log2.Errorf("更新订单状态失败: %v", err)
		errorResponse(c, http.StatusInternalServerError, "更新订单状态失败")
		return
	}

	// 记录状态变更
	if err := tx.Create(&models.OrderStatusLog{
		OrderID:     order.ID,
		OldStatus:   order.Status,
		NewStatus:   nextStatus,
		ChangedTime: time.Now(),
	}).Error; err != nil {
		tx.Rollback()
		log2.Errorf("记录状态变更失败: %v", err)
		errorResponse(c, http.StatusInternalServerError, "记录状态变更失败")
		return
	}

	tx.Commit()

	// 返回更新后的订单信息
	order.Status = nextStatus
	successResponse(c, gin.H{
		"message":    "订单状态更新成功",
		"old_status": order.Status,
		"new_status": nextStatus,
		"order":      order,
	})
}

// 添加状态转换验证函数
func isValidStatusTransition(currentStatus string) bool {
	_, exists := models.OrderStatusTransitions[currentStatus]
	return exists
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

// 验证用户ID的合法性
func (h *Handler) IsValidUserID(userID snowflake.ID) bool {
	var user models.User
	err := h.DB.First(&user, userID).Error
	return err == nil
}
