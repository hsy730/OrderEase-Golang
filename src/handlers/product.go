package handlers

import (
	"fmt"
	"net/http"
	productdomain "orderease/domain/product"
	"orderease/models"
	"orderease/utils"
	"orderease/utils/log2"
	"os"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

// 商品图片最大大小 512KB
const maxProductImageSize = 2048 * 1024
const maxProductImageZipSize = 512 * 1024

// 创建商品
// 修改商品结构体以支持参数类别
func (h *Handler) CreateProduct(c *gin.Context) {
	var request struct {
		models.Product
		OptionCategories []models.ProductOptionCategory `json:"option_categories"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的商品数据: "+err.Error())
		return
	}

	utils.SanitizeProduct(&request.Product)

	validShopID, err := h.validAndReturnShopID(c, request.Product.ShopID)
	if err != nil {
		errorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	// 使用领域实体创建商品（设置基础字段和初始状态）
	productDomain := productdomain.NewProduct(validShopID, request.Product.Name, request.Product.Price, request.Product.Stock)
	productDomain.SetDescription(request.Product.Description)
	productDomain.SetImageURL(request.Product.ImageURL)
	productDomain.SetOptionCategories(request.OptionCategories)

	// 转换为持久化模型
	productModel := productDomain.ToModel()
	productModel.Status = models.ProductStatusPending
	productModel.ID = utils.GenerateSnowflakeID()

	// 开启事务
	tx := h.DB.Begin()

	// 创建商品
	if err := tx.Create(&productModel).Error; err != nil {
		tx.Rollback()
		h.logger.Errorf("创建商品失败: %v", err)
		errorResponse(c, http.StatusInternalServerError, "创建商品失败")
		return
	}

	// 创建商品参数类别和选项
	for i := range request.OptionCategories {
		category := request.OptionCategories[i]
		category.ProductID = productModel.ID
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
	if err := h.DB.First(&createdProduct, productModel.ID).Error; err != nil {
		h.logger.Errorf("获取创建后的商品失败: %v", err)
		successResponse(c, productModel)
		return
	}
	successResponse(c, createdProduct)
}

// ToggleProductStatus 更新商品状态
func (h *Handler) ToggleProductStatus(c *gin.Context) {
	// 解析请求参数
	var req struct {
		ID     string `json:"id" binding:"required"`
		Status string `json:"status" binding:"required,oneof=pending online offline"`
		ShopID int64  `json:"shop_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		log2.Errorf("解析请求参数失败: %v", err)
		errorResponse(c, http.StatusBadRequest, "无效的请求参数")
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
	productModel, err := h.productRepo.GetProductByID(productId, validShopID)
	if err != nil {
		errorResponse(c, http.StatusNotFound, err.Error())
		return
	}

	if productModel.Status == "" {
		productModel.Status = models.ProductStatusPending
	}

	log2.Debugf("更新商品前状态: %s", productModel.Status)
	// 使用领域服务验证状态流转
	if !h.productService.CanTransitionTo(productModel.Status, req.Status) {
		errorResponse(c, http.StatusBadRequest, "无效的状态变更")
		return
	}

	// 更新商品状态
	if err := h.DB.Model(&productModel).Update("status", req.Status).Error; err != nil {
		log2.Errorf("更新商品状态失败: %v", err)
		errorResponse(c, http.StatusInternalServerError, "更新商品状态失败")
		return
	}

	// 返回成功响应
	successResponse(c, gin.H{
		"message": "商品状态更新成功",
		"product": gin.H{
			"id":         productModel.ID,
			"status":     req.Status,
			"updated_at": productModel.UpdatedAt,
		},
	})
}

// 获取商品列表
func (h *Handler) GetProducts(c *gin.Context) {
	var products []models.Product

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	requestShopID, err := strconv.ParseUint(c.Query("shop_id"), 10, 64)
	if err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的店铺ID")
		return
	}

	if err := ValidatePaginationParams(page, pageSize); err != nil {
		errorResponse(c, http.StatusBadRequest, err.Error())
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

// 获取单个商品详情
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

// 更新商品信息，支持参数类别更新
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

	productModel, err := h.productRepo.GetProductByID(id, validShopID)
	if err != nil {
		errorResponse(c, http.StatusNotFound, err.Error())
		return
	}

	// 转换为领域实体（可以使用业务方法进行验证）
	productDomain := productdomain.ProductFromModel(productModel)

	var request struct {
		models.Product
		OptionCategories []models.ProductOptionCategory `json:"option_categories"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的更新数据: "+err.Error())
		return
	}

	utils.SanitizeProduct(&request.Product)

	// 使用领域实体验证库存（如果有更新库存）
	if request.Product.Stock > 0 && request.Product.Stock != productDomain.Stock() {
		// 可以添加业务验证逻辑
		productDomain.SetStock(request.Product.Stock)
	}

	// 开启事务
	tx := h.DB.Begin()

	// 更新商品基本信息
	if err := tx.Model(&productModel).Updates(request.Product).Error; err != nil {
		tx.Rollback()
		log2.Errorf("更新商品失败: %v", err)
		errorResponse(c, http.StatusInternalServerError, "更新商品失败")
		return
	}

	// 如果有提供参数类别，则先删除旧的参数类别和选项，再创建新的
	// if len(request.OptionCategories) > 0 {
	// 删除旧的参数类别
	if err := tx.Where("product_id = ?", productModel.ID).Delete(&models.ProductOptionCategory{}).Error; err != nil {
		tx.Rollback()
		log2.Errorf("删除旧商品参数类别失败: %v", err)
		errorResponse(c, http.StatusInternalServerError, "更新商品参数失败")
		return
	}

	// 创建新的参数类别和选项
	for i := range request.OptionCategories {
		category := request.OptionCategories[i]
		category.ProductID = productModel.ID
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

// 删除商品
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

	// 使用领域服务验证是否可以删除（检查关联订单）
	if err := h.productService.ValidateForDeletion(id); err != nil {
		errorResponse(c, http.StatusBadRequest, err.Error())
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
		errorResponse(c, http.StatusInternalServerError, "删除商品失败")
		return
	}

	// 删除商品参数类别
	if err := tx.Where("product_id = ?", product.ID).Delete(&models.ProductOptionCategory{}).Error; err != nil {
		tx.Rollback()
		log2.Errorf("删除商品参数类别失败: %v", err)
		errorResponse(c, http.StatusInternalServerError, "删除商品失败")
		return
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

	// 使用 Media Service 验证文件类型
	if _, err := h.mediaService.ValidateImageType(file.Header.Get("Content-Type")); err != nil {
		errorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	uploadDir := "./uploads/products"
	if err := h.mediaService.CreateUploadDir(uploadDir); err != nil {
		log2.Errorf("创建上传目录失败: %v", err)
		errorResponse(c, http.StatusInternalServerError, "创建上传目录失败")
		return
	}

	// 使用 Media Service 删除旧图片
	if err := h.mediaService.RemoveOldImage(product.ImageURL); err != nil {
		log2.Errorf("删除旧图片失败: %v", err)
	}

	// 使用 Media Service 生成文件名
	filename := h.mediaService.GenerateUniqueFileName("product", id, file.Filename)

	// 使用 Media Service 构建文件路径
	filePath := h.mediaService.BuildFilePath(uploadDir, filename)

	if err := c.SaveUploadedFile(file, filePath); err != nil {
		log2.Errorf("保存文件失败: %v", err)
		errorResponse(c, http.StatusInternalServerError, "保存文件失败")
		return
	}

	// 压缩图片（继续使用 utils.CompressImage，未来可迁移到 Media Service）
	compressedSize, err := utils.CompressImage(filePath, maxProductImageZipSize)
	if err != nil {
		log2.Errorf("压缩图片失败: %v", err)
		errorResponse(c, http.StatusInternalServerError, "压缩图片失败")
		return
	}

	if compressedSize > 0 {
		log2.Infof("图片压缩成功，原始大小: %d 字节，压缩后: %d 字节", file.Size, compressedSize)
	}

	// 使用 Media Service 验证图片 URL
	// 注意：文件名格式是 product_xxx.jpg，folder 参数也需要是 "product"（单数）
	if err := h.mediaService.ValidateImageURL(filename, "product"); err != nil {
		log2.Errorf("图片URL验证失败: %v", err)
		errorResponse(c, http.StatusBadRequest, "无效的图片格式")
		return
	}

	if err := h.DB.Model(&product).Update("image_url", filename).Error; err != nil {
		log2.Errorf("更新商品图片失败: %v", err)
		errorResponse(c, http.StatusInternalServerError, "更新商品图片失败")
		return
	}

	// 使用 Media Service 获取消息和操作类型
	message := h.mediaService.GetUploadMessage(product.ImageURL == "" && filename == "")
	operationType := h.mediaService.GetOperationType(message)

	c.JSON(http.StatusOK, gin.H{
		"message": message,
		"url":     filename,
		"type":    operationType,
	})
}

// 获取商品图片
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
