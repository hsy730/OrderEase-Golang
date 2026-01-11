package handlers

import (
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
)

// 商品图片最大大小 512KB
const maxProductImageSize = 2048 * 1024
const maxProductImageZipSize = 512 * 1024

// CreateProductRequest 创建商品请求
type CreateProductRequest struct {
	Name             string                         `json:"name" binding:"required" example:"川菜套餐"`
	Description      string                         `json:"description" example:"美味川菜"`
	Price            int64                          `json:"price" binding:"required" example:"5000"`
	ShopID           uint64                         `json:"shop_id" binding:"required" example:"1"`
	OptionCategories []models.ProductOptionCategory `json:"option_categories"`
}

// UpdateProductRequest 更新商品请求
type UpdateProductRequest struct {
	ID          string `json:"id" binding:"required" example:"1"`
	Name        string `json:"name" example:"川菜套餐"`
	Description string `json:"description" example:"美味川菜"`
	Price       int64  `json:"price" example:"5000"`
	ShopID      uint64 `json:"shop_id" example:"1"`
}

// CreateProduct 创建商品
// @Summary 创建商品
// @Description 创建新商品
// @Tags 商品管理
// @Accept json
// @Produce json
// @Param product body CreateProductRequest true "商品信息"
// @Success 200 {object} map[string]interface{} "创建成功"
// @Security BearerAuth
// @Router /admin/product/create [post]
// @Router /shopOwner/product/create [post]
func (h *Handler) CreateProduct(c *gin.Context) {
	var request struct {
		models.Product
		OptionCategories []models.ProductOptionCategory `json:"option_categories"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的商品数据: "+err.Error())
		return
	}

	product := request.Product
	utils.SanitizeProduct(&product)
	product.Status = models.ProductStatusPending
	product.ID = utils.GenerateSnowflakeID()

	// 如果没有设置库存数量，设置默认值为100
	if product.Stock <= 0 {
		product.Stock = 100
	}

	validShopID, err := h.validAndReturnShopID(c, product.ShopID)
	if err != nil {
		errorResponse(c, http.StatusBadRequest, err.Error())
		return
	}
	product.ShopID = validShopID

	// 开启事务
	tx := h.DB.Begin()

	// 创建商品
	if err := tx.Create(&product).Error; err != nil {
		tx.Rollback()
		h.logger.Errorf("创建商品失败: %v", err)
		errorResponse(c, http.StatusInternalServerError, "创建商品失败")
		return
	}

	// 创建商品参数类别和选项
	for i := range request.OptionCategories {
		category := request.OptionCategories[i]
		category.ProductID = product.ID
		category.ID = utils.GenerateSnowflakeID()

		if err := tx.Create(&category).Error; err != nil {
			tx.Rollback()
			h.logger.Errorf("创建商品参数类别失败: %v", err)
			errorResponse(c, http.StatusInternalServerError, "创建商品参数失败")
			return
		}
	}

	tx.Commit()
	// 查询创建后的商品，包含参数信息
	var createdProduct models.Product
	if err := h.DB.First(&createdProduct, product.ID).Error; err != nil {
		h.logger.Errorf("获取创建后的商品失败: %v", err)
		successResponse(c, product)
		return
	}
	successResponse(c, createdProduct)
}

// ToggleProductStatus 更新商品状态
// @Summary 切换商品状态
// @Description 切换商品上架/下架状态
// @Tags 商品管理
// @Accept json
// @Produce json
// @Param productId query string true "商品ID"
// @Success 200 {object} map[string]interface{} "更新成功"
// @Security BearerAuth
// @Router /admin/product/toggle-status [put]
// @Router /shopOwner/product/toggle-status [put]
func (h *Handler) ToggleProductStatus(c *gin.Context) {
	// 解析请求参数
	var req struct {
		ID     string `json:"id" binding:"required"`
		Status string `json:"status" binding:"required,oneof=pending online offline"`
		ShopID int64  `json:"shop_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		log2.Errorf("解析请求参数失败: %v", err)
		errorResponse(c, http.StatusBadRequest, "无效的请求参数："+err.Error())
		return
	}

	validShopID, err := h.validAndReturnShopID(c, uint64(req.ShopID))
	if err != nil {
		errorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	productId, err := strconv.ParseUint(req.ID, 10, 64)
	if err != nil {
		errorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	// 获取当前商品信息
	product, err := h.productRepo.GetProductByID(productId, validShopID)
	if err != nil {
		errorResponse(c, http.StatusNotFound, err.Error())
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
		models.ProductStatusOffline: {models.ProductStatusOnline}, // 允许下架后重新上架
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

// GetProducts 获取商品列表
// @Summary 获取商品列表
// @Description 获取商品列表，支持分页和筛选
// @Tags 商品管理
// @Accept json
// @Produce json
// @Param page query int false "页码"
// @Param pageSize query int false "每页数量"
// @Param shopId query string false "店铺ID"
// @Param status query string false "商品状态"
// @Success 200 {object} map[string]interface{} "查询成功"
// @Security BearerAuth
// @Router /admin/product/list [get]
// @Router /shopOwner/product/list [get]
// @Router /product/list [get]
func (h *Handler) GetProducts(c *gin.Context) {
	var products []models.Product

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	requestShopID, err := strconv.ParseUint(c.Query("shop_id"), 10, 64)
	if err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的店铺ID")
		return
	}

	if page < 1 {
		errorResponse(c, http.StatusBadRequest, "页码必须大于0")
		return
	}

	if pageSize < 1 || pageSize > 100 {
		errorResponse(c, http.StatusBadRequest, "每页数量必须在1-100之间")
		return
	}

	validShopID, err := h.validAndReturnShopID(c, requestShopID)
	if err != nil {
		errorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	offset := (page - 1) * pageSize

	// 获取搜索关键词
	search := c.Query("search")

	// 只查询未下架的商品（待上架和已上架）
	query := h.DB.Where("status != ? and shop_id = ?", models.ProductStatusOffline, validShopID)

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

	// 获取分页数据，并预加载参数类别和选项信息
	if err := query.Offset(offset).Limit(pageSize).
		Order("created_at DESC").
		Preload("OptionCategories").
		Preload("OptionCategories.Options").
		Find(&products).Error; err != nil {
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

// GetProduct 获取单个商品详情
// @Summary 获取商品详情
// @Description 获取指定商品的详细信息
// @Tags 商品管理
// @Accept json
// @Produce json
// @Param shop_id query string true "店铺ID"
// @Param productId query string true "商品ID"
// @Success 200 {object} map[string]interface{} "查询成功"
// @Security BearerAuth
// @Router /admin/product/detail [get]
// @Router /shopOwner/product/detail [get]
// @Router /product/detail [get]
func (h *Handler) GetProduct(c *gin.Context) {
	id, err := strconv.ParseUint(c.Query("id"), 10, 64)
	if err != nil {
		errorResponse(c, http.StatusBadRequest, "缺少商品ID")
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

	product, err := h.productRepo.GetProductByID(id, validShopID)
	if err != nil {
		errorResponse(c, http.StatusNotFound, err.Error())
		return
	}
	successResponse(c, product)
}

// UpdateProduct 更新商品信息，支持参数类别更新
// @Summary 更新商品信息
// @Description 更新商品基本信息
// @Tags 商品管理
// @Accept json
// @Produce json
// @Param product body UpdateProductRequest true "商品信息"
// @Success 200 {object} map[string]interface{} "更新成功"
// @Security BearerAuth
// @Router /admin/product/update [put]
// @Router /shopOwner/product/update [put]
func (h *Handler) UpdateProduct(c *gin.Context) {
	id, err := strconv.ParseUint(c.Query("id"), 10, 64)
	if err != nil {
		errorResponse(c, http.StatusBadRequest, "缺少商品ID")
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

	product, err := h.productRepo.GetProductByID(id, validShopID)
	if err != nil {
		errorResponse(c, http.StatusNotFound, err.Error())
		return
	}

	var request struct {
		models.Product
		OptionCategories []models.ProductOptionCategory `json:"option_categories"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的更新数据: "+err.Error())
		return
	}

	utils.SanitizeProduct(&request.Product)

	// 开启事务
	tx := h.DB.Begin()

	// 更新商品基本信息
	if err := tx.Model(&product).Updates(request.Product).Error; err != nil {
		tx.Rollback()
		log2.Errorf("更新商品失败: %v", err)
		errorResponse(c, http.StatusInternalServerError, "更新商品失败")
		return
	}

	// 如果有提供参数类别，则先删除旧的参数类别和选项，再创建新的
	// if len(request.OptionCategories) > 0 {
	// 删除旧的参数类别
	if err := tx.Where("product_id = ?", product.ID).Delete(&models.ProductOptionCategory{}).Error; err != nil {
		tx.Rollback()
		log2.Errorf("删除旧商品参数类别失败: %v", err)
		errorResponse(c, http.StatusInternalServerError, "更新商品参数失败")
		return
	}

	// 创建新的参数类别和选项
	for i := range request.OptionCategories {
		category := request.OptionCategories[i]
		category.ProductID = product.ID
		category.ID = utils.GenerateSnowflakeID()

		if err := tx.Create(&category).Error; err != nil {
			tx.Rollback()
			log2.Errorf("创建商品参数类别失败: %v", err)
			errorResponse(c, http.StatusInternalServerError, "更新商品参数失败")
			return
		}
	}
	// }

	tx.Commit()
	// 重新获取更新后的商品信息
	updatedProduct, err := h.productRepo.GetProductByID(id, validShopID)
	if err != nil {
		errorResponse(c, http.StatusNotFound, err.Error())
		return
	}
	successResponse(c, updatedProduct)
}

// DeleteProduct 删除商品
// @Summary 删除商品
// @Description 删除指定商品
// @Tags 商品管理
// @Accept json
// @Produce json
// @Param productId query string true "商品ID"
// @Success 200 {object} map[string]interface{} "删除成功"
// @Security BearerAuth
// @Router /admin/product/delete [delete]
// @Router /shopOwner/product/delete [delete]
func (h *Handler) DeleteProduct(c *gin.Context) {
	id, err := strconv.ParseUint(c.Query("id"), 10, 64)
	if err != nil {
		errorResponse(c, http.StatusBadRequest, "缺少商品ID")
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

	product, err := h.productRepo.GetProductByID(id, validShopID)
	if err != nil {
		errorResponse(c, http.StatusNotFound, err.Error())
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

	// 删除商品参数选项 (先删除选项，再删除类别)
	if err := tx.Where(`category_id IN (
		SELECT id FROM product_option_categories WHERE product_id = ?
	)`, product.ID).Delete(&models.ProductOption{}).Error; err != nil {
		tx.Rollback()
		log2.Errorf("删除商品参数选项失败: %v", err)
		errorResponse(c, http.StatusInternalServerError, "删除商品失败: "+err.Error())
		return
	}

	// 删除商品参数类别
	if err := tx.Where("product_id = ?", product.ID).Delete(&models.ProductOptionCategory{}).Error; err != nil {
		tx.Rollback()
		log2.Errorf("删除商品参数类别失败: %v", err)
		errorResponse(c, http.StatusInternalServerError, "删除商品失败: "+err.Error())
		return
	}

	// 删除商品标签
	if err := tx.Where("product_id = ?", product.ID).Delete(&models.ProductTag{}).Error; err != nil {
		tx.Rollback()
		log2.Errorf("删除商品标签失败: %v", err)
		errorResponse(c, http.StatusInternalServerError, "删除商品失败: "+err.Error())
		return
	}

	// 删除商品记录
	if err := tx.Delete(&product).Error; err != nil {
		tx.Rollback()
		log2.Errorf("删除商品记录失败: %v", err)
		errorResponse(c, http.StatusInternalServerError, "删除商品失败: "+err.Error())
		return
	}

	tx.Commit()
	successResponse(c, gin.H{"message": "商品删除成功"})
}

// UploadProductImage 上传商品图片
// @Summary 上传商品图片
// @Description 上传商品图片
// @Tags 商品管理
// @Accept multipart/form-data
// @Produce json
// @Param image formData file true "商品图片"
// @Success 200 {object} map[string]interface{} "上传成功"
// @Security BearerAuth
// @Router /admin/product/upload-image [post]
// @Router /shopOwner/product/upload-image [post]
func (h *Handler) UploadProductImage(c *gin.Context) {
	// 限制文件大小
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxProductImageSize)

	id, err := strconv.ParseUint(c.Query("id"), 10, 64)
	if err != nil {
		errorResponse(c, http.StatusBadRequest, "缺少商品ID")
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

	product, err := h.productRepo.GetProductByID(id, validShopID)
	if err != nil {
		errorResponse(c, http.StatusNotFound, err.Error())
		return
	}

	file, err := c.FormFile("image")
	if err != nil {
		log2.Errorf("获取上传文件失败: %v", err)
		errorResponse(c, http.StatusBadRequest, "获取上传文件失败")
		return
	}

	// 检查文件类型
	if !utils.IsValidImageType(file.Header.Get("Content-Type")) {
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

	filename := fmt.Sprintf("product_%d_%d%s",
		id,
		time.Now().Unix(),
		filepath.Ext(file.Filename))

	// 修改：只保存文件名
	imageURL := filename
	filePath := fmt.Sprintf("%s/%s", uploadDir, filename)

	if err := c.SaveUploadedFile(file, filePath); err != nil {
		log2.Errorf("保存文件失败: %v", err)
		errorResponse(c, http.StatusInternalServerError, "保存文件失败")
		return
	}

	// 压缩图片
	compressedSize, err := utils.CompressImage(filePath, maxProductImageZipSize)
	if err != nil {
		log2.Errorf("压缩图片失败: %v", err)
		errorResponse(c, http.StatusInternalServerError, "压缩图片失败")
		return
	}

	if compressedSize > 0 {
		log2.Infof("图片压缩成功，原始大小: %d 字节，压缩后: %d 字节", file.Size, compressedSize)
	}

	if err := utils.ValidateImageURL(imageURL, "product"); err != nil {
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

// GetProductImage 获取商品图片
// @Summary 获取商品图片
// @Description 获取指定商品的图片
// @Tags 商品管理
// @Accept json
// @Produce json
// @Param productId query string true "商品ID"
// @Success 200 {object} map[string]interface{} "查询成功"
// @Security BearerAuth
// @Router /admin/product/image [get]
// @Router /shopOwner/product/image [get]
// @Router /product/image [get]
func (h *Handler) GetProductImage(c *gin.Context) {
	// 添加路径前缀
	fileName := c.Query("path")

	if err := utils.ValidateImageURL(fileName, "product"); err != nil {
		log2.Errorf("图片路径验证失败: %v", err)
		errorResponse(c, http.StatusBadRequest, "无效的图片路径")
		return
	}

	imagePath := fmt.Sprintf("./uploads/products/%s", fileName)

	if _, err := os.Stat(imagePath); os.IsNotExist(err) {
		log2.Errorf("图片文件不存在: %s", imagePath)
		errorResponse(c, http.StatusNotFound, "图片不存在")
		return
	}

	c.File(imagePath)
}
