package handlers

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"orderease/models"
	"orderease/utils"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"encoding/csv"
	"io"

	"archive/zip"

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
	product.Status = models.ProductStatusPending

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

	// 只查询未下架的商品（待上架和已上架）
	query := h.DB.Where("status != ?", models.ProductStatusOffline)

	// 获取总数
	var total int64
	if err := query.Model(&models.Product{}).Count(&total).Error; err != nil {
		h.logger.Printf("获取商品总数失败: %v", err)
		errorResponse(c, http.StatusInternalServerError, "获取商品列表失败")
		return
	}

	// 获取分页数据
	if err := query.Offset(offset).Limit(pageSize).Find(&products).Error; err != nil {
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

	// 获取商品信息
	var product models.Product
	if err := h.DB.First(&product, id).Error; err != nil {
		h.logger.Printf("删除商品失败, ID: %s, 错误: %v", id, err)
		errorResponse(c, http.StatusNotFound, "商品不存在")
		return
	}

	// 检查是否存在关联订单
	var orderCount int64
	if err := h.DB.Model(&models.OrderItem{}).
		Where("product_id = ?", id).
		Count(&orderCount).Error; err != nil {
		h.logger.Printf("检查商品订单关联失败: %v", err)
		errorResponse(c, http.StatusInternalServerError, "系统错误")
		return
	}

	// 如果存在关联订单，不允许删除
	if orderCount > 0 {
		errorResponse(c, http.StatusBadRequest,
			fmt.Sprintf("该商品有 %d 个关联订单，不能删除。建议将商品下架而不是删除", orderCount))
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

// 导出数据
func (h *Handler) ExportData(c *gin.Context) {
	// 生成带时间戳的文件名
	timestamp := time.Now().Format("20060102_150405")
	zipFilename := fmt.Sprintf("export_%s.zip", timestamp)

	// 创建一个缓冲区来保存 ZIP 文件
	buf := new(bytes.Buffer)
	zipWriter := zip.NewWriter(buf)

	// 导出用户数据
	if err := exportTableToCSV(h.DB, zipWriter, "users.csv", &[]models.User{}); err != nil {
		h.logger.Printf("导出用户数据失败: %v", err)
		errorResponse(c, http.StatusInternalServerError, "导出失败")
		return
	}

	// 导出商品数据
	if err := exportTableToCSV(h.DB, zipWriter, "products.csv", &[]models.Product{}); err != nil {
		h.logger.Printf("导出商品数据失败: %v", err)
		errorResponse(c, http.StatusInternalServerError, "导出失败")
		return
	}

	// 导出订单数据
	if err := exportTableToCSV(h.DB, zipWriter, "orders.csv", &[]models.Order{}); err != nil {
		h.logger.Printf("导出订单数据失败: %v", err)
		errorResponse(c, http.StatusInternalServerError, "导出失败")
		return
	}

	// 导出订单项数据
	if err := exportTableToCSV(h.DB, zipWriter, "order_items.csv", &[]models.OrderItem{}); err != nil {
		h.logger.Printf("导出订单项数据失败: %v", err)
		errorResponse(c, http.StatusInternalServerError, "导出失败")
		return
	}

	// 关闭 ZIP writer
	if err := zipWriter.Close(); err != nil {
		h.logger.Printf("关闭 ZIP writer 失败: %v", err)
		errorResponse(c, http.StatusInternalServerError, "导出失败")
		return
	}

	// 设置响应头为 ZIP
	c.Header("Content-Type", "application/zip")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", zipFilename))

	// 发送 ZIP 文件
	c.Writer.Write(buf.Bytes())
}

// 辅助函数：导出表数据到 CSV 文件
func exportTableToCSV(db *gorm.DB, zipWriter *zip.Writer, filename string, model interface{}) error {
	// 创建 CSV 文件
	w, err := zipWriter.Create(filename)
	if err != nil {
		return err
	}

	// 创建 CSV writer
	csvWriter := csv.NewWriter(w)
	defer csvWriter.Flush()

	// 查询数据
	if err := db.Find(model).Error; err != nil {
		return err
	}

	// 获取表头
	headers, err := getCSVHeaders(model)
	if err != nil {
		return err
	}

	// 写入表头
	if err := csvWriter.Write(headers); err != nil {
		return err
	}

	// 写入数据
	records, err := getCSVRecords(model)
	if err != nil {
		return err
	}

	for _, record := range records {
		if err := csvWriter.Write(record); err != nil {
			return err
		}
	}

	return nil
}

// 辅助函数：获取 CSV 表头
func getCSVHeaders(model interface{}) ([]string, error) {
	// 根据模型类型返回表头
	switch model.(type) {
	case *[]models.User:
		return []string{"id", "name", "phone", "address", "type", "created_at", "updated_at"}, nil
	case *[]models.Product:
		return []string{"id", "name", "description", "price", "stock", "image_url", "created_at", "updated_at"}, nil
	case *[]models.Order:
		return []string{"id", "user_id", "total_price", "status", "remark", "created_at", "updated_at"}, nil
	case *[]models.OrderItem:
		return []string{"id", "order_id", "product_id", "quantity", "price"}, nil
	default:
		return nil, fmt.Errorf("unsupported model type")
	}
}

// 辅助函数：获取 CSV 记录
func getCSVRecords(model interface{}) ([][]string, error) {
	var records [][]string

	switch v := model.(type) {
	case *[]models.User:
		for _, u := range *v {
			record := []string{
				strconv.FormatUint(uint64(u.ID), 10),
				u.Name,
				u.Phone,
				u.Address,
				u.Type,
				u.CreatedAt.Format(time.RFC3339),
				u.UpdatedAt.Format(time.RFC3339),
			}
			records = append(records, record)
		}
	case *[]models.Product:
		for _, p := range *v {
			record := []string{
				strconv.FormatUint(uint64(p.ID), 10),
				p.Name,
				p.Description,
				strconv.FormatFloat(p.Price, 'f', 2, 64),
				strconv.Itoa(p.Stock),
				p.ImageURL,
				p.CreatedAt.Format(time.RFC3339),
				p.UpdatedAt.Format(time.RFC3339),
			}
			records = append(records, record)
		}
	case *[]models.Order:
		for _, o := range *v {
			record := []string{
				strconv.FormatUint(uint64(o.ID), 10),
				strconv.FormatUint(uint64(o.UserID), 10),
				strconv.FormatFloat(float64(o.TotalPrice), 'f', 2, 64),
				o.Status,
				o.Remark,
				o.CreatedAt.Format(time.RFC3339),
				o.UpdatedAt.Format(time.RFC3339),
			}
			records = append(records, record)
		}
	case *[]models.OrderItem:
		for _, item := range *v {
			record := []string{
				strconv.FormatUint(uint64(item.ID), 10),
				strconv.FormatUint(uint64(item.OrderID), 10),
				strconv.FormatUint(uint64(item.ProductID), 10),
				strconv.Itoa(item.Quantity),
				strconv.FormatFloat(float64(item.Price), 'f', 2, 64),
			}
			records = append(records, record)
		}
	default:
		return nil, fmt.Errorf("unsupported model type")
	}

	return records, nil
}

// 导入数据
func (h *Handler) ImportData(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		errorResponse(c, http.StatusBadRequest, "请上传ZIP文件")
		return
	}

	if !strings.HasSuffix(file.Filename, ".zip") {
		errorResponse(c, http.StatusBadRequest, "只支持ZIP文件")
		return
	}

	f, err := file.Open()
	if err != nil {
		h.logger.Printf("打开上传文件失败: %v", err)
		errorResponse(c, http.StatusInternalServerError, "文件处理失败")
		return
	}
	defer f.Close()

	// 读取 ZIP 文件
	zipReader, err := zip.NewReader(f, file.Size)
	if err != nil {
		h.logger.Printf("读取ZIP文件失败: %v", err)
		errorResponse(c, http.StatusInternalServerError, "文件处理失败")
		return
	}

	tx := h.DB.Begin()

	// 逐个处理 CSV 文件
	for _, zipFile := range zipReader.File {
		if err := importCSVFile(tx, zipFile); err != nil {
			tx.Rollback()
			h.logger.Printf("导入数据失败: %v", err)
			errorResponse(c, http.StatusInternalServerError, "导入失败")
			return
		}
	}

	tx.Commit()
	successResponse(c, gin.H{"message": "数据导入成功"})
}

// 辅助函数：导入 CSV 文件
func importCSVFile(tx *gorm.DB, zipFile *zip.File) error {
	f, err := zipFile.Open()
	if err != nil {
		return err
	}
	defer f.Close()

	reader := csv.NewReader(f)

	// 读取表头
	headers, err := reader.Read()
	if err != nil {
		return fmt.Errorf("读取CSV表头失败: %v", err)
	}

	// 逐行读取并导入数据
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			continue // 跳过空行
		}
		if len(record) != len(headers) {
			continue // 跳过格式不匹配的行
		}

		// 根据文件名处理不同的数据
		switch filepath.Base(zipFile.Name) {
		case "users.csv":
			if err := importUserRecord(tx, record); err != nil {
				return err
			}
		case "products.csv":
			if err := importProductRecord(tx, record); err != nil {
				return err
			}
		case "orders.csv":
			if err := importOrderRecord(tx, record); err != nil {
				return err
			}
		case "order_items.csv":
			if err := importOrderItemRecord(tx, record); err != nil {
				return err
			}
		default:
			return fmt.Errorf("未知的CSV文件: %s", zipFile.Name)
		}
	}

	return nil
}

// 辅助函数：导入用户记录
func importUserRecord(tx *gorm.DB, record []string) error {
	user := models.User{
		Name:    record[1],
		Phone:   record[2],
		Address: record[3],
		Type:    record[4],
	}
	return tx.Create(&user).Error
}

// 辅助函数：导入商品记录
func importProductRecord(tx *gorm.DB, record []string) error {
	product := models.Product{
		Name:        record[1],
		Description: record[2],
		Price:       parseFloat(record[3]),
		Stock:       parseInt(record[4]),
		ImageURL:    record[5],
	}
	return tx.Create(&product).Error
}

// 辅助函数：导入订单记录
func importOrderRecord(tx *gorm.DB, record []string) error {
	order := models.Order{
		UserID:     uint(parseInt(record[1])),
		TotalPrice: models.Price(parseFloat(record[2])),
		Status:     record[3],
		Remark:     record[4],
	}
	return tx.Create(&order).Error
}

// 辅助函数：导入订单项记录
func importOrderItemRecord(tx *gorm.DB, record []string) error {
	orderItem := models.OrderItem{
		OrderID:   uint(parseInt(record[1])),
		ProductID: uint(parseInt(record[2])),
		Quantity:  parseInt(record[3]),
		Price:     models.Price(parseFloat(record[4])),
	}
	return tx.Create(&orderItem).Error
}

// 辅助函数：解析浮点数
func parseFloat(s string) float64 {
	f, _ := strconv.ParseFloat(s, 64)
	return f
}

// 辅助函数：解析整数
func parseInt(s string) int {
	i, _ := strconv.Atoi(s)
	return i
}
