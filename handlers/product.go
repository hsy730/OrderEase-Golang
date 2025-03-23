package handlers

import (
	"errors"
	"fmt"
	"net/http"
	"orderease/models"
	"orderease/utils"
	"orderease/utils/log2"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// 创建商品
func (h *Handler) CreateProduct(c *gin.Context) {
	c.Get("")
	var product models.Product
	if err := c.ShouldBindJSON(&product); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的商品数据: "+err.Error())
		return
	}

	utils.SanitizeProduct(&product)
	product.Status = models.ProductStatusPending
	product.ID = utils.GenerateSnowflakeID()

	if err := h.DB.Create(&product).Error; err != nil {
		h.logger.Printf("创建商品失败: %v", err)
		errorResponse(c, http.StatusInternalServerError, "创建商品失败")
		return
	}

	successResponse(c, product)
}

// ToggleProductStatus 更新商品状态
func (h *Handler) ToggleProductStatus(c *gin.Context) {
	// 解析请求参数
	var req struct {
		ID     uint   `json:"id" binding:"required"`
		Status string `json:"status" binding:"required,oneof=pending online offline"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		log2.Errorf("解析请求参数失败: %v", err)
		errorResponse(c, http.StatusBadRequest, "无效的请求参数")
		return
	}

	// 获取当前商品信息
	var product models.Product
	if err := h.DB.First(&product, req.ID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			errorResponse(c, http.StatusNotFound, "商品不存在")
			return
		}
		log2.Errorf("查询商品失败: %v", err)
		errorResponse(c, http.StatusInternalServerError, "服务器内部错误")
		return
	}

	if product.Status == "" {
		product.Status = models.ProductStatusPending
	}

	log2.Debugf("更新商品前状态: %s", product.Status)
	// 验证状态流转
	if !isValidProductStatusTransition(product.Status, req.Status) {
		errorResponse(c, http.StatusBadRequest, "无效的状态变更")
		return
	}

	// 更新商品状态
	if err := h.DB.Model(&product).Update("status", req.Status).Error; err != nil {
		log2.Errorf("更新商品状态失败: %v", err)
		errorResponse(c, http.StatusInternalServerError, "更新商品状态失败")
		return
	}

	// 返回成功响应
	successResponse(c, gin.H{
		"message": "商品状态更新成功",
		"product": gin.H{
			"id":         product.ID,
			"status":     req.Status,
			"updated_at": product.UpdatedAt,
		},
	})
}

// isValidProductStatusTransition 验证商品状态流转是否合法
func isValidProductStatusTransition(currentStatus, newStatus string) bool {
	// 定义状态流转规则
	transitions := map[string][]string{
		models.ProductStatusPending: {models.ProductStatusOnline},
		models.ProductStatusOnline:  {models.ProductStatusOffline},
		models.ProductStatusOffline: {}, // 已下架状态不能转换到其他状态
	}

	// 检查是否是允许的状态转换
	allowedStatus, exists := transitions[currentStatus]
	if !exists {
		return false
	}

	// 检查新状态是否在允许的转换列表中
	for _, status := range allowedStatus {
		if status == newStatus {
			return true
		}
	}

	return false
}

// 获取商品列表
func (h *Handler) GetProducts(c *gin.Context) {
	var products []models.Product

	v, _ := c.Get("username")
	log2.Debugf("用户名： %v", v)
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

	// 获取搜索关键词
	search := c.Query("search")

	// 只查询未下架的商品（待上架和已上架）
	query := h.DB.Where("status != ?", models.ProductStatusOffline)

	// 如果有搜索关键词，添加模糊搜索条件
	if search != "" {
		query = query.Where("name LIKE ?", "%"+search+"%")
	}

	// 获取总数
	var total int64
	if err := query.Model(&models.Product{}).Count(&total).Error; err != nil {
		log2.Errorf("获取商品总数失败: %v", err)
		errorResponse(c, http.StatusInternalServerError, "获取商品列表失败")
		return
	}

	// 获取分页数据
	if err := query.Offset(offset).Limit(pageSize).Find(&products).Error; err != nil {
		log2.Errorf("查询商品列表失败: %v", err)
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
		log2.Errorf("查询商品失败, ID: %s, 错误: %v", id, err)
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
		log2.Errorf("更新商品失败, ID: %s, 错误: %v", id, err)
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
		log2.Errorf("更新商品失败: %v", err)
		errorResponse(c, http.StatusInternalServerError, "更新商品失败")
		return
	}

	// 重新获取更新后的商品信息
	if err := h.DB.First(&product, id).Error; err != nil {
		log2.Errorf("获取更新后的商品信息失败: %v", err)
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
		log2.Errorf("删除商品失败, ID: %s, 错误: %v", id, err)
		errorResponse(c, http.StatusNotFound, "商品不存在")
		return
	}

	// 检查是否存在关联订单
	var orderCount int64
	if err := h.DB.Model(&models.OrderItem{}).
		Where("product_id = ?", id).
		Count(&orderCount).Error; err != nil {
		log2.Errorf("检查商品订单关联失败: %v", err)
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
			log2.Errorf("删除商品图片失败: %v", err)
		}
	}

	// 删除商品记录
	if err := tx.Delete(&product).Error; err != nil {
		tx.Rollback()
		log2.Errorf("删除商品记录失败: %v", err)
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
		log2.Errorf("商品不存在, ID: %s, 错误: %v", id, err)
		errorResponse(c, http.StatusNotFound, "商品不存在")
		return
	}

	file, err := c.FormFile("image")
	if err != nil {
		log2.Errorf("获取上传文件失败: %v", err)
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
		log2.Errorf("创建上传目录失败: %v", err)
		errorResponse(c, http.StatusInternalServerError, "创建上传目录失败")
		return
	}

	// 如果已有图片，先删除旧图片
	if product.ImageURL != "" {
		oldImagePath := strings.TrimPrefix(product.ImageURL, "/")
		if err := os.Remove(oldImagePath); err != nil && !os.IsNotExist(err) {
			log2.Errorf("删除旧图片失败: %v", err)
		}
	}

	filename := fmt.Sprintf("product_%s_%d%s",
		id,
		time.Now().Unix(),
		filepath.Ext(file.Filename))

	imageURL := fmt.Sprintf("/uploads/products/%s", filename)
	filePath := fmt.Sprintf("%s/%s", uploadDir, filename)

	if err := c.SaveUploadedFile(file, filePath); err != nil {
		log2.Errorf("保存文件失败: %v", err)
		errorResponse(c, http.StatusInternalServerError, "保存文件失败")
		return
	}

	if err := utils.ValidateImageURL(imageURL); err != nil {
		log2.Errorf("图片URL验证失败: %v", err)
		errorResponse(c, http.StatusBadRequest, "无效的图片格式")
		return
	}

	if err := h.DB.Model(&product).Update("image_url", imageURL).Error; err != nil {
		log2.Errorf("更新商品图片失败: %v", err)
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
		log2.Errorf("图片路径验证失败: %v", err)
		errorResponse(c, http.StatusBadRequest, "无效的图片路径")
		return
	}

	imagePath = "." + imagePath

	if _, err := os.Stat(imagePath); os.IsNotExist(err) {
		log2.Errorf("图片文件不存在: %s", imagePath)
		errorResponse(c, http.StatusNotFound, "图片不存在")
		return
	}

	c.File(imagePath)
}
