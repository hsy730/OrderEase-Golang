package http

import (
	"net/http"
	"os"
	"orderease/application/dto"
	"orderease/application/services"
	"orderease/domain/shared"
	"orderease/utils/log2"
	"strconv"

	"github.com/gin-gonic/gin"
)

type ProductHandler struct {
	productService *services.ProductService
	shopService   *services.ShopService
}

func NewProductHandler(
	productService *services.ProductService,
	shopService *services.ShopService,
) *ProductHandler {
	return &ProductHandler{
		productService: productService,
		shopService:   shopService,
	}
}

func (h *ProductHandler) CreateProduct(c *gin.Context) {
	var req dto.CreateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的商品数据: "+err.Error())
		return
	}

	shopID, err := h.validateShopID(c, req.ShopID)
	if err != nil {
		errorResponse(c, http.StatusBadRequest, err.Error())
		return
	}
	req.ShopID = shopID

	response, err := h.productService.CreateProduct(&req)
	if err != nil {
		log2.Errorf("创建商品失败: %v", err)
		errorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	successResponse(c, response)
}

func (h *ProductHandler) GetProduct(c *gin.Context) {
	idStr := c.Query("id")
	id, err := shared.ParseIDFromString(idStr)
	if err != nil {
		errorResponse(c, http.StatusBadRequest, "缺少商品ID")
		return
	}

	shopIDStr := c.Query("shop_id")
	shopID, err := shared.ParseIDFromString(shopIDStr)
	if err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的店铺ID")
		return
	}

	validShopID, err := h.validateShopID(c, shopID)
	if err != nil {
		errorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	response, err := h.productService.GetProduct(id, validShopID)
	if err != nil {
		log2.Errorf("查询商品失败: %v", err)
		errorResponse(c, http.StatusNotFound, err.Error())
		return
	}

	successResponse(c, response)
}

func (h *ProductHandler) GetProducts(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	search := c.Query("search")

	if page < 1 {
		errorResponse(c, http.StatusBadRequest, "页码必须大于0")
		return
	}

	if pageSize < 1 || pageSize > 100 {
		errorResponse(c, http.StatusBadRequest, "每页数量必须在1-100之间")
		return
	}

	shopIDStr := c.Query("shop_id")
	shopID, err := shared.ParseIDFromString(shopIDStr)
	if err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的店铺ID")
		return
	}

	validShopID, err := h.validateShopID(c, shopID)
	if err != nil {
		errorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	response, err := h.productService.GetProducts(validShopID, page, pageSize, search)
	if err != nil {
		log2.Errorf("查询商品列表失败: %v", err)
		errorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	successResponse(c, response)
}

func (h *ProductHandler) UpdateProduct(c *gin.Context) {
	idStr := c.Query("id")
	id, err := shared.ParseIDFromString(idStr)
	if err != nil {
		errorResponse(c, http.StatusBadRequest, "缺少商品ID")
		return
	}

	shopIDStr := c.Query("shop_id")
	shopID, err := shared.ParseIDFromString(shopIDStr)
	if err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的店铺ID")
		return
	}

	validShopID, err := h.validateShopID(c, shopID)
	if err != nil {
		errorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	var req dto.CreateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的更新数据: "+err.Error())
		return
	}

	response, err := h.productService.UpdateProduct(id, validShopID, &req)
	if err != nil {
		log2.Errorf("更新商品失败: %v", err)
		errorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	successResponse(c, response)
}

func (h *ProductHandler) DeleteProduct(c *gin.Context) {
	idStr := c.Query("id")
	id, err := shared.ParseIDFromString(idStr)
	if err != nil {
		errorResponse(c, http.StatusBadRequest, "缺少商品ID")
		return
	}

	shopIDStr := c.Query("shop_id")
	shopID, err := shared.ParseIDFromString(shopIDStr)
	if err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的店铺ID")
		return
	}

	validShopID, err := h.validateShopID(c, shopID)
	if err != nil {
		errorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	if err := h.productService.DeleteProduct(id, validShopID); err != nil {
		log2.Errorf("删除商品失败: %v", err)
		errorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	successResponse(c, gin.H{"message": "商品删除成功"})
}

func (h *ProductHandler) UpdateProductStatus(c *gin.Context) {
	var req dto.UpdateProductStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log2.Errorf("解析请求参数失败: %v", err)
		errorResponse(c, http.StatusBadRequest, "无效的请求参数")
		return
	}

	shopIDStr := c.Query("shop_id")
	shopID, err := shared.ParseIDFromString(shopIDStr)
	if err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的店铺ID")
		return
	}

	validShopID, err := h.validateShopID(c, shopID)
	if err != nil {
		errorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	if err := h.productService.UpdateProductStatus(&req, validShopID); err != nil {
		log2.Errorf("更新商品状态失败: %v", err)
		errorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	successResponse(c, gin.H{
		"message": "商品状态更新成功",
		"product": gin.H{
			"id":     req.ID,
			"status": req.Status,
		},
	})
}

func (h *ProductHandler) validateShopID(c *gin.Context, shopID shared.ID) (shared.ID, error) {
	requestUser, exists := c.Get("userInfo")
	if !exists {
		return shopID, nil
	}

	userInfo := requestUser.(interface {
		IsAdminUser() bool
		GetUserID() uint64
	})

	if !userInfo.IsAdminUser() {
		return shared.ParseIDFromUint64(userInfo.GetUserID()), nil
	}

	shop, err := h.shopService.GetShop(shopID)
	if err != nil {
		return shared.ID(0), err
	}

	return shop.ID, nil
}

// UploadProductImage 上传商品图片
// @Summary 上传商品图片
// @Description 上传商品图片
// @Tags 商品管理
// @Accept multipart/form-data
// @Produce json
// @Param id query string true "商品ID"
// @Param shop_id query string true "店铺ID"
// @Param image formData file true "商品图片"
// @Success 200 {object} map[string]interface{} "上传成功"
// @Security BearerAuth
// @Router /shopOwner/product/upload-image [post]
// @Router /admin/product/upload-image [post]
func (h *ProductHandler) UploadProductImage(c *gin.Context) {
	const maxFileSize = 2 * 1024 * 1024 // 2MB
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxFileSize)

	idStr := c.Query("id")
	if idStr == "" {
		errorResponse(c, http.StatusBadRequest, "缺少商品ID")
		return
	}

	id, err := shared.ParseIDFromString(idStr)
	if err != nil || id.IsZero() {
		errorResponse(c, http.StatusBadRequest, "无效的商品ID")
		return
	}

	shopIDStr := c.Query("shop_id")
	if shopIDStr == "" {
		errorResponse(c, http.StatusBadRequest, "缺少店铺ID")
		return
	}

	shopID, err := shared.ParseIDFromString(shopIDStr)
	if err != nil || shopID.IsZero() {
		errorResponse(c, http.StatusBadRequest, "无效的店铺ID")
		return
	}

	validShopID, err := h.validateShopID(c, shopID)
	if err != nil {
		errorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	// 获取上传文件
	fileHeader, err := c.FormFile("image")
	if err != nil {
		log2.Errorf("获取上传文件失败: %v", err)
		errorResponse(c, http.StatusBadRequest, "获取上传文件失败")
		return
	}

	// 检查文件类型
	contentType := fileHeader.Header.Get("Content-Type")
	validTypes := map[string]bool{
		"image/jpeg": true,
		"image/png":  true,
		"image/gif":  true,
		"image/webp": true,
	}
	if !validTypes[contentType] {
		errorResponse(c, http.StatusBadRequest, "不支持的文件类型")
		return
	}

	// 打开上传的文件
	file, err := fileHeader.Open()
	if err != nil {
		log2.Errorf("打开文件失败: %v", err)
		errorResponse(c, http.StatusBadRequest, "打开文件失败")
		return
	}
	defer file.Close()

	// 调用 service 层上传图片
	filename, err := h.productService.UploadProductImage(id, validShopID, file, fileHeader.Filename)
	if err != nil {
		log2.Errorf("上传图片失败: %v", err)
		errorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	successResponse(c, gin.H{
		"message": "图片上传成功",
		"url":     filename,
	})
}

// GetProductImage 获取商品图片
// @Summary 获取商品图片
// @Description 获取指定商品的图片
// @Tags 商品管理
// @Accept json
// @Produce image/*
// @Param path query string true "图片路径"
// @Success 200 {file} file "图片文件"
// @Security BearerAuth
// @Router /shopOwner/product/image [get]
// @Router /admin/product/image [get]
// @Router /product/image [get]
func (h *ProductHandler) GetProductImage(c *gin.Context) {
	fileName := c.Query("path")
	if fileName == "" {
		errorResponse(c, http.StatusBadRequest, "缺少图片路径")
		return
	}

	imagePath := "./uploads/products/" + fileName

	if _, err := os.Stat(imagePath); os.IsNotExist(err) {
		log2.Errorf("图片文件不存在: %s", imagePath)
		errorResponse(c, http.StatusNotFound, "图片不存在")
		return
	}

	c.File(imagePath)
}

// ToggleProductStatus 切换商品状态
// @Summary 切换商品状态
// @Description 切换商品上架/下架状态
// @Tags 商品管理
// @Accept json
// @Produce json
// @Param request body dto.UpdateProductStatusRequest true "商品状态信息"
// @Success 200 {object} map[string]interface{} "操作成功"
// @Security BearerAuth
// @Router /admin/product/toggle-status [put]
// @Router /shopOwner/product/toggle-status [put]
func (h *ProductHandler) ToggleProductStatus(c *gin.Context) {
	h.UpdateProductStatus(c)
}
