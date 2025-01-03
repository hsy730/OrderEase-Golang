package handlers

import (
	"OrderEase/models"
	"OrderEase/utils"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

const (
	maxFileSize = 32 << 20 // 32MB
)

type Handler struct {
	DB     *gorm.DB
	logger *log.Logger
}

// 创建处理器实例
func NewHandler(db *gorm.DB) *Handler {
	return &Handler{
		DB:     db,
		logger: utils.Logger,
	}
}

// 创建商品
func (h *Handler) CreateProduct(c *gin.Context) {
	var product models.Product
	if err := c.ShouldBindJSON(&product); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的商品数据: "+err.Error())
		return
	}

	utils.SanitizeProduct(&product)

	if err := h.DB.Create(&product).Error; err != nil {
		h.logger.Printf("创建商品失败: %v", err)
		errorResponse(c, http.StatusInternalServerError, "创建商品失败")
		return
	}

	successResponse(c, product)
}

// 获取商品列表
func (h *Handler) GetProducts(c *gin.Context) {
	var products []models.Product

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

	offset := (page - 1) * pageSize

	var total int64
	if err := h.DB.Model(&models.Product{}).Count(&total).Error; err != nil {
		h.logger.Printf("获取商品总数失败: %v", err)
		errorResponse(c, http.StatusInternalServerError, "获取商品列表失败")
		return
	}

	if err := h.DB.Offset(offset).Limit(pageSize).Find(&products).Error; err != nil {
		h.logger.Printf("查询商品列表失败: %v", err)
		errorResponse(c, http.StatusInternalServerError, "获取商品列表失败")
		return
	}

	successResponse(c, gin.H{
		"total":    total,
		"page":     page,
		"pageSize": pageSize,
		"data":     products,
	})
}

// 获取单个商品详情
func (h *Handler) GetProduct(c *gin.Context) {
	id := c.Query("id")
	if id == "" {
		errorResponse(c, http.StatusBadRequest, "缺少商品ID")
		return
	}

	var product models.Product
	if err := h.DB.First(&product, id).Error; err != nil {
		h.logger.Printf("查询商品失败, ID: %s, 错误: %v", id, err)
		errorResponse(c, http.StatusNotFound, "商品未找到")
		return
	}

	successResponse(c, product)
}

// 更新商品信息
func (h *Handler) UpdateProduct(c *gin.Context) {
	id := c.Query("id")
	if id == "" {
		errorResponse(c, http.StatusBadRequest, "缺少商品ID")
		return
	}

	var product models.Product
	if err := h.DB.First(&product, id).Error; err != nil {
		h.logger.Printf("更新商品失败, ID: %s, 错误: %v", id, err)
		errorResponse(c, http.StatusNotFound, "商品未找到")
		return
	}

	var updateData models.Product
	if err := c.ShouldBindJSON(&updateData); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的更新数据: "+err.Error())
		return
	}

	utils.SanitizeProduct(&updateData)

	if err := h.DB.Model(&product).Updates(updateData).Error; err != nil {
		h.logger.Printf("更新商品失败: %v", err)
		errorResponse(c, http.StatusInternalServerError, "更新商品失败")
		return
	}

	// 重新获取更新后的商品信息
	if err := h.DB.First(&product, id).Error; err != nil {
		h.logger.Printf("获取更新后的商品信息失败: %v", err)
		errorResponse(c, http.StatusInternalServerError, "获取更新后的商品信息失败")
		return
	}

	successResponse(c, product)
}

// 删除商品
func (h *Handler) DeleteProduct(c *gin.Context) {
	id := c.Query("id")
	if id == "" {
		errorResponse(c, http.StatusBadRequest, "缺少商品ID")
		return
	}

	var product models.Product
	if err := h.DB.First(&product, id).Error; err != nil {
		h.logger.Printf("删除商品失败, ID: %s, 错误: %v", id, err)
		errorResponse(c, http.StatusNotFound, "商品不存在")
		return
	}

	// 开启事务
	tx := h.DB.Begin()

	// 删除商品图片
	if product.ImageURL != "" {
		imagePath := strings.TrimPrefix(product.ImageURL, "/")
		if err := os.Remove(imagePath); err != nil && !os.IsNotExist(err) {
			h.logger.Printf("删除商品图片失败: %v", err)
		}
	}

	// 删除商品记录
	if err := tx.Delete(&product).Error; err != nil {
		tx.Rollback()
		h.logger.Printf("删除商品记录失败: %v", err)
		errorResponse(c, http.StatusInternalServerError, "删除商品失败")
		return
	}

	tx.Commit()
	successResponse(c, gin.H{"message": "商品删除成功"})
}

// 上传商品图片
func (h *Handler) UploadProductImage(c *gin.Context) {
	// 限制文件大小
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxFileSize)

	id := c.Query("id")
	if id == "" {
		errorResponse(c, http.StatusBadRequest, "缺少商品ID")
		return
	}

	var product models.Product
	if err := h.DB.First(&product, id).Error; err != nil {
		utils.Logger.Printf("商品不存在, ID: %s, 错误: %v", id, err)
		errorResponse(c, http.StatusNotFound, "商品不存在")
		return
	}

	file, err := c.FormFile("image")
	if err != nil {
		utils.Logger.Printf("获取上传文件失败: %v", err)
		errorResponse(c, http.StatusBadRequest, "获取上传文件失败")
		return
	}

	// 检查文件类型
	if !isValidImageType(file.Header.Get("Content-Type")) {
		errorResponse(c, http.StatusBadRequest, "不支持的文件类型")
		return
	}

	uploadDir := "./uploads/products"
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		h.logger.Printf("创建上传目录失败: %v", err)
		errorResponse(c, http.StatusInternalServerError, "创建上传目录失败")
		return
	}

	// 如果已有图片，先删除旧图片
	if product.ImageURL != "" {
		oldImagePath := strings.TrimPrefix(product.ImageURL, "/")
		if err := os.Remove(oldImagePath); err != nil && !os.IsNotExist(err) {
			utils.Logger.Printf("删除旧图片失败: %v", err)
		}
	}

	filename := fmt.Sprintf("product_%s_%d%s",
		id,
		time.Now().Unix(),
		filepath.Ext(file.Filename))

	imageURL := fmt.Sprintf("/uploads/products/%s", filename)
	filePath := fmt.Sprintf("%s/%s", uploadDir, filename)

	if err := c.SaveUploadedFile(file, filePath); err != nil {
		h.logger.Printf("保存文件失败: %v", err)
		errorResponse(c, http.StatusInternalServerError, "保存文件失败")
		return
	}

	if err := utils.ValidateImageURL(imageURL); err != nil {
		h.logger.Printf("图片URL验证失败: %v", err)
		errorResponse(c, http.StatusBadRequest, "无效的图片格式")
		return
	}

	if err := h.DB.Model(&product).Update("image_url", imageURL).Error; err != nil {
		h.logger.Printf("更新商品图片失败: %v", err)
		errorResponse(c, http.StatusInternalServerError, "更新商品图片失败")
		return
	}

	message := "图片更新成功"
	if product.ImageURL == "" {
		message = "图片上传成功"
	}

	operationType := "update"
	if message == "图片上传成功" {
		operationType = "create"
	}

	c.JSON(http.StatusOK, gin.H{
		"message": message,
		"url":     imageURL,
		"type":    operationType,
	})
}

// 获取商品图片
func (h *Handler) GetProductImage(c *gin.Context) {
	imagePath := c.Query("path")
	if imagePath == "" {
		errorResponse(c, http.StatusBadRequest, "缺少图片路径")
		return
	}

	if !strings.HasPrefix(imagePath, "/") {
		imagePath = "/" + imagePath
	}

	if err := utils.ValidateImageURL(imagePath); err != nil {
		h.logger.Printf("图片路径验证失败: %v", err)
		errorResponse(c, http.StatusBadRequest, "无效的图片路径")
		return
	}

	imagePath = "." + imagePath

	if _, err := os.Stat(imagePath); os.IsNotExist(err) {
		h.logger.Printf("图片文件不存在: %s", imagePath)
		errorResponse(c, http.StatusNotFound, "图片不存在")
		return
	}

	c.File(imagePath)
}

// 创建订单
func (h *Handler) CreateOrder(c *gin.Context) {
	var order models.Order
	if err := c.ShouldBindJSON(&order); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的订单数据: "+err.Error())
		return
	}

	utils.SanitizeOrder(&order)

	// 验证订单数据
	if err := validateOrder(&order); err != nil {
		errorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	tx := h.DB.Begin()

	if err := tx.Create(&order).Error; err != nil {
		tx.Rollback()
		utils.Logger.Printf("创建订单失败: %v", err)
		errorResponse(c, http.StatusInternalServerError, "创建订单失败")
		return
	}

	// 更新商品库存
	for _, item := range order.Items {
		var product models.Product
		if err := tx.First(&product, item.ProductID).Error; err != nil {
			tx.Rollback()
			h.logger.Printf("商品不存在, ID: %d, 错误: %v", item.ProductID, err)
			errorResponse(c, http.StatusBadRequest, "商品不存在")
			return
		}

		if product.Stock < item.Quantity {
			tx.Rollback()
			h.logger.Printf("商品库存不足, ID: %d, 当前库存: %d, 需求数量: %d",
				item.ProductID, product.Stock, item.Quantity)
			errorResponse(c, http.StatusBadRequest, fmt.Sprintf("商品 %s 库存不足", product.Name))
			return
		}

		product.Stock -= item.Quantity
		if err := tx.Save(&product).Error; err != nil {
			tx.Rollback()
			h.logger.Printf("更新商品库存失败: %v", err)
			errorResponse(c, http.StatusInternalServerError, "更新商品库存失败")
			return
		}
	}

	tx.Commit()
	successResponse(c, order)
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

	// 开启事务
	tx := h.DB.Begin()

	if err := tx.Model(&order).Updates(updateData).Error; err != nil {
		tx.Rollback()
		h.logger.Printf("更新订单失败: %v", err)
		errorResponse(c, http.StatusInternalServerError, "更新订单失败")
		return
	}

	// 重新获取更新后的订单信息，包括关联数据
	if err := tx.Preload("Items").Preload("Items.Product").First(&order, id).Error; err != nil {
		tx.Rollback()
		h.logger.Printf("获取更新后的订单信息失败: %v", err)
		errorResponse(c, http.StatusInternalServerError, "获取更新后的订单信息失败")
		return
	}

	tx.Commit()
	successResponse(c, order)
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

	offset := (page - 1) * pageSize

	var total int64
	if err := h.DB.Model(&models.Order{}).Count(&total).Error; err != nil {
		h.logger.Printf("获取订单总数失败: %v", err)
		errorResponse(c, http.StatusInternalServerError, "获取订单列表失败")
		return
	}

	var orders []models.Order
	if err := h.DB.Offset(offset).Limit(pageSize).
		Preload("Items").
		Preload("Items.Product").
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

	var order models.Order
	if err := h.DB.Preload("Items").
		Preload("Items.Product").
		First(&order, id).Error; err != nil {
		h.logger.Printf("查询订单失败, ID: %s, 错误: %v", id, err)
		errorResponse(c, http.StatusNotFound, "订单未找到")
		return
	}

	successResponse(c, order)
}

// 删除订单
func (h *Handler) DeleteOrder(c *gin.Context) {
	id := c.Query("id")
	if id == "" {
		errorResponse(c, http.StatusBadRequest, "缺少订单ID")
		return
	}

	var order models.Order
	if err := h.DB.First(&order, id).Error; err != nil {
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
	id := c.Query("id")
	if id == "" {
		errorResponse(c, http.StatusBadRequest, "缺少订单ID")
		return
	}

	var order models.Order
	if err := h.DB.First(&order, id).Error; err != nil {
		utils.Logger.Printf("订单不存在, ID: %s, 错误: %v", id, err)
		errorResponse(c, http.StatusNotFound, "订单不存在")
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
		h.logger.Printf("更新订单状态失败: %v", err)
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
		h.logger.Printf("记录状态变更失败: %v", err)
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

// 检查文件类型是否为图片
func isValidImageType(contentType string) bool {
	validTypes := map[string]bool{
		"image/jpeg": true,
		"image/png":  true,
		"image/gif":  true,
	}
	return validTypes[contentType]
}

// 添加错误响应辅助函数
func errorResponse(c *gin.Context, code int, message string) {
	utils.Logger.Printf("错误响应: %d - %s", code, message)
	c.JSON(code, gin.H{"error": message})
}

// 添加成功响应辅助函数
func successResponse(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, data)
}
