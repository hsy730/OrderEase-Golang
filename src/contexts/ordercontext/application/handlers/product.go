package handlers

import (
	"fmt"
	"net/http"
	productdomain "orderease/contexts/ordercontext/domain/product"
	"orderease/models"
	"orderease/utils"
	"orderease/utils/log2"
	"os"
	"strconv"
	"strings"

	"github.com/bwmarrin/snowflake"
	"github.com/gin-gonic/gin"
)

// 商品图片最大大小 512KB
const maxProductImageSize = 2048 * 1024
const maxProductImageZipSize = 512 * 1024

// 创建商品
// 修改商品结构体以支持参数类别
func (h *Handler) CreateProduct(c *gin.Context) {
	var request struct {
		ShopID          snowflake.ID                       `json:"shop_id" binding:"required"`
		Name            string                             `json:"name" binding:"required,min=1,max=200"`
		Description     string                             `json:"description" binding:"max=5000"`
		Price           float64                            `json:"price" binding:"required,gt=0"`
		Stock           int                                `json:"stock" binding:"required,min=0"`
		ImageURL        string                             `json:"image_url"`
		OptionCategories []models.ProductOptionCategory `json:"option_categories"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的商品数据: "+err.Error())
		return
	}

	validShopID, err := h.validAndReturnShopID(c, request.ShopID)
	if err != nil {
		errorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	// 验证店铺是否存在
	shopExists, err := h.shopRepo.ExistsById(validShopID)
	if err != nil {
		errorResponse(c, http.StatusInternalServerError, "验证店铺失败")
		return
	}
	if !shopExists {
		errorResponse(c, http.StatusNotFound, "店铺不存在")
		return
	}

	// 使用领域实体创建商品（封装完整创建逻辑）
	productDomain := productdomain.NewProductWithDefaults(
		validShopID,
		request.Name,
		request.Price,
		request.Stock,
		request.Description,
		request.ImageURL,
		request.OptionCategories,
	)

	// 清理商品数据（防止XSS攻击）
	productDomain.Sanitize()

	// 转换为持久化模型
	productModel := productDomain.ToModel()
	productModel.ID = utils.GenerateSnowflakeID()

	// 为参数类别生成 ID
	for i := range request.OptionCategories {
		request.OptionCategories[i].ID = utils.GenerateSnowflakeID()
	}

	// 使用 Repository 创建商品（包含参数类别）
	if err := h.productRepo.CreateWithCategories(productModel, request.OptionCategories); err != nil {
		h.logger.Errorf("创建商品失败: %v", err)
		errorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	// 查询创建后的商品，包含参数信息（使用 Repository）
	createdProduct, err := h.productRepo.GetProductByID(uint64(productModel.ID), validShopID)
	if err != nil {
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
		ID     snowflake.ID `json:"id" binding:"required"`
		Status string       `json:"status" binding:"required,oneof=pending online offline"`
		ShopID snowflake.ID `json:"shop_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		log2.Errorf("解析请求参数失败: %v", err)
		errorResponse(c, http.StatusBadRequest, "无效的请求参数")
		return
	}

	validShopID, err := h.validAndReturnShopID(c, req.ShopID)
	if err != nil {
		errorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	// 获取当前商品信息
	productModel, err := h.productRepo.GetProductByID(uint64(req.ID), validShopID)
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
	if err := h.productRepo.UpdateStatus(uint64(req.ID), validShopID, req.Status); err != nil {
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
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	requestShopID, err := utils.StringToSnowflakeID(c.Query("shop_id"))
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

	// 获取搜索关键词
	search := c.Query("search")

	// 判断是否为管理端请求：客户端只查询已上架商品，管理端查询所有状态商品
	onlyOnline := !strings.HasPrefix(c.Request.URL.Path, "/api/order-ease/v1/shopOwner/") &&
		!strings.HasPrefix(c.Request.URL.Path, "/api/order-ease/v1/admin/")

	// 使用 Repository 获取商品列表
	result, err := h.productRepo.GetProductsByShop(validShopID, page, pageSize, search, onlyOnline)
	if err != nil {
		errorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	successResponse(c, gin.H{
		"total":    result.Total,
		"page":     page,
		"pageSize": pageSize,
		"data":     result.Products,
	})
}

// 获取单个商品详情
func (h *Handler) GetProduct(c *gin.Context) {
	id, err := strconv.ParseUint(c.Query("id"), 10, 64)
	if err != nil {
		errorResponse(c, http.StatusBadRequest, "缺少商品ID")
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

	productModel, err := h.productRepo.GetProductByID(id, validShopID)
	if err != nil {
		errorResponse(c, http.StatusNotFound, err.Error())
		return
	}

	// 转换为领域实体（可以使用业务方法进行验证）
	productDomain := productdomain.ProductFromModel(productModel)

	var request struct {
		Name            string                             `json:"name" binding:"omitempty,min=1,max=200"`
		Description     string                             `json:"description" binding:"omitempty,max=5000"`
		Price           float64                            `json:"price" binding:"omitempty,gt=0"`
		Stock           *int                               `json:"stock" binding:"omitempty,min=0"`
		ImageURL        string                             `json:"image_url"`
		OptionCategories []models.ProductOptionCategory `json:"option_categories"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的更新数据: "+err.Error())
		return
	}

	// 验证至少有一个字段要更新
	allEmpty := request.Name == "" && request.Description == "" &&
		request.ImageURL == "" && len(request.OptionCategories) == 0
	noNumericUpdate := request.Price == 0 && (request.Stock == nil || *request.Stock == productDomain.Stock())
	if allEmpty && noNumericUpdate {
		errorResponse(c, http.StatusBadRequest, "至少需要提供一个要更新的字段")
		return
	}

	// 验证shop_id是否匹配
	if validShopID != productDomain.ShopID() {
		errorResponse(c, http.StatusForbidden, "无权操作此商品")
		return
	}

	// 更新领域实体字段
	if request.Name != "" {
		productDomain.SetName(request.Name)
	}
	if request.Description != "" {
		productDomain.SetDescription(request.Description)
	}
	if request.Price > 0 {
		productDomain.SetPrice(request.Price)
	}
	if request.ImageURL != "" {
		productDomain.SetImageURL(request.ImageURL)
	}
	// 库存验证（使用指针类型区分"未传值"和"传了0值"）
	if request.Stock != nil && *request.Stock != productDomain.Stock() {
		productDomain.SetStock(*request.Stock)
	}

	// 清理商品数据（防止XSS攻击）
	productDomain.Sanitize()

	// 转换回持久化模型
	productModel = productDomain.ToModel()

	// 为参数类别生成 ID
	for i := range request.OptionCategories {
		request.OptionCategories[i].ID = utils.GenerateSnowflakeID()
	}

	// 使用 Repository 更新商品（包含参数类别）
	if err := h.productRepo.UpdateWithCategories(productModel, request.OptionCategories); err != nil {
		log2.Errorf("更新商品失败: %v", err)
		errorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

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

	// 删除商品图片（文件系统操作保留在 handler）
	if product.ImageURL != "" {
		imagePath := strings.TrimPrefix(product.ImageURL, "/")
		if err := os.Remove(imagePath); err != nil && !os.IsNotExist(err) {
			log2.Errorf("删除商品图片失败: %v", err)
		}
	}

	// 使用 Repository 删除商品及其关联数据
	if err := h.productRepo.DeleteWithDependencies(id, validShopID); err != nil {
		errorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

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

	// 验证目录是否可写
	testFile := uploadDir + "/.write_test"
	if f, err := os.Create(testFile); err != nil {
		log2.Errorf("上传目录不可写: %v, 路径: %s", err, uploadDir)
		errorResponse(c, http.StatusInternalServerError, "上传目录不可写")
		return
	} else {
		f.Close()
		os.Remove(testFile)
	}

	// 使用 Media Service 删除旧图片
	if err := h.mediaService.RemoveOldImage(product.ImageURL); err != nil {
		log2.Errorf("删除旧图片失败: %v", err)
	}

	// 使用 Media Service 生成文件名
	filename := h.mediaService.GenerateUniqueFileName("product", id, file.Filename)

	// 使用 Media Service 构建文件路径
	filePath := h.mediaService.BuildFilePath(uploadDir, filename)

	// 打开源文件
	src, err := file.Open()
	if err != nil {
		log2.Errorf("打开上传文件失败: %v", err)
		errorResponse(c, http.StatusInternalServerError, "打开上传文件失败")
		return
	}
	defer src.Close()

	// 创建目标文件
	dst, err := os.Create(filePath)
	if err != nil {
		log2.Errorf("创建目标文件失败: %v, 路径: %s", err, filePath)
		errorResponse(c, http.StatusInternalServerError, "创建目标文件失败")
		return
	}
	defer dst.Close()

	// 复制文件内容
	if _, err := dst.ReadFrom(src); err != nil {
		log2.Errorf("写入文件失败: %v, 路径: %s", err, filePath)
		errorResponse(c, http.StatusInternalServerError, "写入文件失败")
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

	if err := h.productRepo.UpdateImageURL(id, validShopID, filename); err != nil {
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
