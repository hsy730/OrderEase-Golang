package http

import (
	"net/http"
	"orderease/application/dto"
	"orderease/application/services"
	"orderease/domain/order"
	"orderease/domain/shared"
	"orderease/utils/log2"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type ShopHandler struct {
	shopService *services.ShopService
}

func NewShopHandler(shopService *services.ShopService) *ShopHandler {
	return &ShopHandler{
		shopService: shopService,
	}
}

func (h *ShopHandler) CreateShop(c *gin.Context) {
	var req dto.CreateShopRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log2.Errorf("Bind Json failed: %v", err)
		errorResponse(c, http.StatusBadRequest, "无效的请求数据")
		return
	}

	if req.ValidUntil.IsZero() {
		req.ValidUntil = time.Now().AddDate(1, 0, 0)
	}

	response, err := h.shopService.CreateShop(&req)
	if err != nil {
		log2.Errorf("create shop failed: %v", err)
		errorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	log2.Infof("create shop success, ID: %s", response.ID.String())

	successResponse(c, gin.H{
		"code": 200,
		"data": response,
	})
}

func (h *ShopHandler) GetShopInfo(c *gin.Context) {
	shopIDStr := c.Query("shop_id")
	shopID, err := shared.ParseIDFromString(shopIDStr)
	if err != nil || shopID.IsZero() {
		errorResponse(c, http.StatusBadRequest, "无效的店铺ID")
		return
	}

	response, err := h.shopService.GetShop(shopID)
	if err != nil {
		if err.Error() == "店铺不存在" {
			errorResponse(c, http.StatusNotFound, "店铺不存在")
			return
		}
		log2.Errorf("查询店铺失败，ID: %s，错误: %v", shopID.String(), err)
		errorResponse(c, http.StatusInternalServerError, "查询失败")
		return
	}

	successResponse(c, response)
}

func (h *ShopHandler) GetShopList(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	search := c.Query("search")

	if page < 1 || pageSize < 1 || pageSize > 100 {
		errorResponse(c, http.StatusBadRequest, "无效的分页参数")
		return
	}

	response, err := h.shopService.GetShops(page, pageSize, search)
	if err != nil {
		log2.Errorf("查询店铺列表失败: %v", err)
		errorResponse(c, http.StatusInternalServerError, "查询失败")
		return
	}

	successResponse(c, response)
}

func (h *ShopHandler) UpdateShop(c *gin.Context) {
	var req dto.UpdateShopRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的请求数据: "+err.Error())
		return
	}

	response, err := h.shopService.UpdateShop(&req)
	if err != nil {
		log2.Errorf("更新店铺失败: %v", err)
		errorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	successResponse(c, gin.H{
		"code": 200,
		"data": response,
	})
}

func (h *ShopHandler) DeleteShop(c *gin.Context) {
	shopIDStr := c.Query("shop_id")
	shopID, err := shared.ParseIDFromString(shopIDStr)
	if err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的店铺ID")
		return
	}

	if err := h.shopService.DeleteShop(shopID); err != nil {
		log2.Errorf("删除店铺失败: %v", err)
		errorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	successResponse(c, gin.H{"message": "店铺删除成功"})
}

func (h *ShopHandler) CheckShopNameExists(c *gin.Context) {
	shopName := c.Query("name")
	if shopName == "" {
		errorResponse(c, http.StatusBadRequest, "商店名称不能为空")
		return
	}

	exists, err := h.shopService.CheckShopNameExists(shopName)
	if err != nil {
		log2.Errorf("检查商店名称失败: %v", err)
		errorResponse(c, http.StatusInternalServerError, "检查商店名称失败")
		return
	}

	successResponse(c, gin.H{
		"exists": exists,
	})
}

func (h *ShopHandler) UpdateOrderStatusFlow(c *gin.Context) {
	var req struct {
		ShopID          shared.ID             `json:"shop_id" binding:"required"`
		OrderStatusFlow order.OrderStatusFlow `json:"order_status_flow" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的请求数据")
		return
	}

	if err := h.shopService.UpdateOrderStatusFlow(req.ShopID, req.OrderStatusFlow); err != nil {
		log2.Errorf("更新店铺订单流转状态配置失败: %v", err)
		errorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	successResponse(c, gin.H{
		"code":    200,
		"message": "店铺订单流转状态配置更新成功",
	})
}

func (h *ShopHandler) GetShopTags(c *gin.Context) {
	shopIDStr := c.Param("shop_id")
	if shopIDStr == "" {
		shopIDStr = c.Query("shop_id")
	}
	shopID, err := shared.ParseIDFromString(shopIDStr)
	if err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的店铺ID")
		return
	}

	response, err := h.shopService.GetShopTags(shopID)
	if err != nil {
		log2.Errorf("查询店铺标签失败，ID: %s，错误: %v", shopID.String(), err)
		errorResponse(c, http.StatusInternalServerError, "查询失败")
		return
	}

	successResponse(c, response)
}

func (h *ShopHandler) CreateTag(c *gin.Context) {
	var req dto.CreateTagRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的请求数据: "+err.Error())
		return
	}

	response, err := h.shopService.CreateTag(&req)
	if err != nil {
		log2.Errorf("创建标签失败: %v", err)
		errorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	successResponse(c, response)
}

func (h *ShopHandler) UpdateTag(c *gin.Context) {
	idStr := c.Query("id")
	if idStr == "" {
		errorResponse(c, http.StatusBadRequest, "缺少标签ID")
		return
	}
	id, err := strconv.Atoi(idStr)
	if err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的标签ID")
		return
	}

	var req dto.CreateTagRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的请求数据: "+err.Error())
		return
	}

	response, err := h.shopService.UpdateTag(id, &req)
	if err != nil {
		log2.Errorf("更新标签失败: %v", err)
		errorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	successResponse(c, response)
}

func (h *ShopHandler) DeleteTag(c *gin.Context) {
	idStr := c.Query("id")
	if idStr == "" {
		errorResponse(c, http.StatusBadRequest, "缺少标签ID")
		return
	}
	id, err := strconv.Atoi(idStr)
	if err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的标签ID")
		return
	}

	if err := h.shopService.DeleteTag(id); err != nil {
		log2.Errorf("删除标签失败: %v", err)
		errorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	successResponse(c, gin.H{"message": "标签删除成功"})
}

func (h *ShopHandler) GetTag(c *gin.Context) {
	idStr := c.Query("id")
	if idStr == "" {
		errorResponse(c, http.StatusBadRequest, "缺少标签ID")
		return
	}
	id, err := strconv.Atoi(idStr)
	if err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的标签ID")
		return
	}

	response, err := h.shopService.GetTag(id)
	if err != nil {
		log2.Errorf("查询标签失败: %v", err)
		errorResponse(c, http.StatusNotFound, err.Error())
		return
	}

	successResponse(c, response)
}

// GetShopImage 获取店铺图片
// @Summary 获取店铺图片
// @Description 获取指定店铺的图片
// @Tags 店铺管理
// @Accept json
// @Produce image/*
// @Param path query string true "图片路径"
// @Success 200 {file} file "图片文件"
// @Security BearerAuth
// @Router /shopOwner/shop/upload-image [get]
func (h *ShopHandler) GetShopImage(c *gin.Context) {
	fileName := c.Query("path")
	if fileName == "" {
		errorResponse(c, http.StatusBadRequest, "缺少图片路径")
		return
	}

	imagePath := "./uploads/shops/" + fileName

	if _, err := os.Stat(imagePath); os.IsNotExist(err) {
		log2.Errorf("图片文件不存在: %s", imagePath)
		errorResponse(c, http.StatusNotFound, "图片不存在")
		return
	}

	c.File(imagePath)
}

// UploadShopImage 上传店铺图片
// @Summary 上传店铺图片
// @Description 上传店铺图片
// @Tags 店铺管理
// @Accept multipart/form-data
// @Produce json
// @Param id query string true "店铺ID"
// @Param image formData file true "店铺图片"
// @Success 200 {object} map[string]interface{} "上传成功"
// @Security BearerAuth
// @Router /shopOwner/shop/upload-image [post]
// @Router /admin/shop/upload-image [post]
func (h *ShopHandler) UploadShopImage(c *gin.Context) {
	const maxFileSize = 2 * 1024 * 1024 // 2MB
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxFileSize)

	shopIDStr := c.Query("id")
	if shopIDStr == "" {
		errorResponse(c, http.StatusBadRequest, "缺少店铺ID")
		return
	}

	shopID, err := shared.ParseIDFromString(shopIDStr)
	if err != nil || shopID.IsZero() {
		errorResponse(c, http.StatusBadRequest, "无效的店铺ID")
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
	filename, err := h.shopService.UploadShopImage(shopID, file, fileHeader.Filename)
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

// GetBoundTags 获取商品已绑定的标签
// @Summary 获取商品已绑定的标签
// @Description 获取指定商品已绑定的标签列表
// @Tags 标签管理
// @Accept json
// @Produce json
// @Param product_id query string true "商品ID"
// @Success 200 {object} map[string]interface{} "查询成功"
// @Security BearerAuth
// @Router /admin/tag/bound-tags [get]
// @Router /shopOwner/tag/bound-tags [get]
func (h *ShopHandler) GetBoundTags(c *gin.Context) {
	productIDStr := c.Query("product_id")
	if productIDStr == "" {
		errorResponse(c, http.StatusBadRequest, "缺少商品ID")
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

	tags, err := h.shopService.GetBoundTags(productIDStr, validShopID.ToUint64())
	if err != nil {
		log2.Errorf("查询已绑定标签失败: %v", err)
		errorResponse(c, http.StatusInternalServerError, "查询失败")
		return
	}

	successResponse(c, gin.H{
		"product_id": productIDStr,
		"tags":       tags,
	})
}

// GetUnboundTags 获取商品未绑定的标签
// @Summary 获取商品未绑定的标签
// @Description 获取指定商品未绑定的标签列表
// @Tags 标签管理
// @Accept json
// @Produce json
// @Param product_id query string true "商品ID"
// @Success 200 {object} map[string]interface{} "查询成功"
// @Security BearerAuth
// @Router /admin/tag/unbound-tags [get]
// @Router /shopOwner/tag/unbound-tags [get]
func (h *ShopHandler) GetUnboundTags(c *gin.Context) {
	productIDStr := c.Query("product_id")
	if productIDStr == "" {
		errorResponse(c, http.StatusBadRequest, "缺少商品ID")
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

	tags, err := h.shopService.GetUnboundTags(productIDStr, validShopID.ToUint64())
	if err != nil {
		log2.Errorf("查询未绑定标签失败: %v", err)
		errorResponse(c, http.StatusInternalServerError, "查询失败")
		return
	}

	successResponse(c, gin.H{
		"product_id": productIDStr,
		"tags":       tags,
	})
}

// BatchTagProducts 批量打标签
// @Summary 批量打标签
// @Description 批量给商品打标签
// @Tags 标签管理
// @Accept json
// @Produce json
// @Param request body dto.BatchTagRequest true "批量打标签信息"
// @Success 200 {object} map[string]interface{} "操作成功"
// @Security BearerAuth
// @Router /admin/tag/batch-tag [post]
// @Router /shopOwner/tag/batch-tag [post]
func (h *ShopHandler) BatchTagProducts(c *gin.Context) {
	var req struct {
		TagID      int      `json:"tag_id" binding:"required"`
		ProductIDs []string `json:"product_ids" binding:"required"`
		ShopID     string   `json:"shop_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的请求参数:"+err.Error())
		return
	}

	shopID, err := shared.ParseIDFromString(req.ShopID)
	if err != nil || shopID.IsZero() {
		errorResponse(c, http.StatusBadRequest, "无效的店铺ID")
		return
	}

	validShopID, err := h.validateShopID(c, shopID)
	if err != nil {
		errorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	if err := h.shopService.BatchTagProducts(req.TagID, req.ProductIDs, validShopID.ToUint64()); err != nil {
		log2.Errorf("批量打标签失败: %v", err)
		errorResponse(c, http.StatusInternalServerError, "批量打标签失败:"+err.Error())
		return
	}

	successResponse(c, gin.H{
		"message": "批量打标签成功",
	})
}

// BatchUntagProducts 批量解绑商品标签
// @Summary 批量解绑商品标签
// @Description 批量解绑商品的标签
// @Tags 标签管理
// @Accept json
// @Produce json
// @Param request body dto.BatchUntagRequest true "批量解绑标签信息"
// @Success 200 {object} map[string]interface{} "操作成功"
// @Security BearerAuth
// @Router /admin/tag/batch-untag [delete]
// @Router /shopOwner/tag/batch-untag [delete]
func (h *ShopHandler) BatchUntagProducts(c *gin.Context) {
	var req struct {
		TagID      int      `json:"tag_id" binding:"required"`
		ProductIDs []string `json:"product_ids" binding:"required"`
		ShopID     string   `json:"shop_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的请求参数")
		return
	}

	shopID, err := shared.ParseIDFromString(req.ShopID)
	if err != nil || shopID.IsZero() {
		errorResponse(c, http.StatusBadRequest, "无效的店铺ID")
		return
	}

	validShopID, err := h.validateShopID(c, shopID)
	if err != nil {
		errorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	if err := h.shopService.BatchUntagProducts(req.TagID, req.ProductIDs, validShopID.ToUint64()); err != nil {
		log2.Errorf("批量解绑标签失败: %v", err)
		errorResponse(c, http.StatusInternalServerError, "批量解绑标签失败")
		return
	}

	successResponse(c, gin.H{
		"message": "批量解绑标签成功",
	})
}

// BatchTagProduct 批量设置商品标签
// @Summary 批量设置商品标签
// @Description 批量设置商品的标签
// @Tags 标签管理
// @Accept json
// @Produce json
// @Param request body dto.BatchTagProductRequest true "批量设置标签信息"
// @Success 200 {object} map[string]interface{} "操作成功"
// @Security BearerAuth
// @Router /admin/tag/batch-tag-product [post]
// @Router /shopOwner/tag/batch-tag-product [post]
func (h *ShopHandler) BatchTagProduct(c *gin.Context) {
	var req struct {
		ProductID string   `json:"product_id" binding:"required"`
		TagIDs    []string `json:"tag_ids" binding:"required"`
		ShopID    string   `json:"shop_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的请求参数")
		return
	}

	shopID, err := shared.ParseIDFromString(req.ShopID)
	if err != nil || shopID.IsZero() {
		errorResponse(c, http.StatusBadRequest, "无效的店铺ID")
		return
	}

	validShopID, err := h.validateShopID(c, shopID)
	if err != nil {
		errorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	if err := h.shopService.BatchTagProduct(req.ProductID, req.TagIDs, validShopID.ToUint64()); err != nil {
		log2.Errorf("批量设置商品标签失败: %v", err)
		errorResponse(c, http.StatusInternalServerError, "批量设置商品标签失败")
		return
	}

	successResponse(c, gin.H{
		"message": "批量设置商品标签成功",
	})
}

// GetTagBoundProducts 获取标签已绑定的商品列表
// @Summary 获取标签已绑定的商品列表
// @Description 获取指定标签已绑定的商品列表
// @Tags 标签管理
// @Accept json
// @Produce json
// @Param tag_id query string true "标签ID"
// @Param shop_id query string true "店铺ID"
// @Param page query int false "页码"
// @Param pageSize query int false "每页数量"
// @Success 200 {object} map[string]interface{} "查询成功"
// @Security BearerAuth
// @Router /admin/tag/bound-products [get]
// @Router /shopOwner/tag/bound-products [get]
// @Router /front/tag/bound-products [get]
func (h *ShopHandler) GetTagBoundProducts(c *gin.Context) {
	tagID := c.Query("tag_id")
	if tagID == "" {
		errorResponse(c, http.StatusBadRequest, "缺少标签ID")
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

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))

	response, err := h.shopService.GetTagBoundProducts(tagID, validShopID.ToUint64(), page, pageSize)
	if err != nil {
		log2.Errorf("查询标签已绑定商品失败: %v", err)
		errorResponse(c, http.StatusInternalServerError, "查询失败")
		return
	}

	successResponse(c, response)
}

// GetUnboundProductsForTag 获取标签未绑定的商品列表
// @Summary 获取标签未绑定的商品列表
// @Description 获取指定标签未绑定的商品列表
// @Tags 标签管理
// @Accept json
// @Produce json
// @Param tag_id query string true "标签ID"
// @Param shop_id query string true "店铺ID"
// @Param page query int false "页码"
// @Param pageSize query int false "每页数量"
// @Success 200 {object} map[string]interface{} "查询成功"
// @Security BearerAuth
// @Router /admin/tag/unbound-products [get]
// @Router /shopOwner/tag/unbound-products [get]
func (h *ShopHandler) GetUnboundProductsForTag(c *gin.Context) {
	tagID := c.Query("tag_id")
	if tagID == "" {
		errorResponse(c, http.StatusBadRequest, "缺少标签ID")
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

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))

	response, err := h.shopService.GetUnboundProductsForTag(tagID, validShopID.ToUint64(), page, pageSize)
	if err != nil {
		log2.Errorf("查询标签未绑定商品失败: %v", err)
		errorResponse(c, http.StatusInternalServerError, "查询失败")
		return
	}

	successResponse(c, response)
}

// GetUnboundTagsList 获取没有绑定商品的标签列表
// @Summary 获取没有绑定商品的标签列表
// @Description 获取没有绑定任何商品的标签列表
// @Tags 标签管理
// @Accept json
// @Produce json
// @Param shop_id query string true "店铺ID"
// @Success 200 {object} map[string]interface{} "查询成功"
// @Security BearerAuth
// @Router /admin/tag/unbound-list [get]
// @Router /shopOwner/tag/unbound-list [get]
func (h *ShopHandler) GetUnboundTagsList(c *gin.Context) {
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

	tags, err := h.shopService.GetUnboundTagsList(validShopID.ToUint64())
	if err != nil {
		log2.Errorf("查询未绑定标签列表失败: %v", err)
		errorResponse(c, http.StatusInternalServerError, "查询失败")
		return
	}

	successResponse(c, gin.H{
		"tags": tags,
	})
}

// GetTagOnlineProducts 获取标签关联的已上架商品
// @Summary 获取标签关联的已上架商品
// @Description 获取指定标签关联的已上架商品列表
// @Tags 标签管理
// @Accept json
// @Produce json
// @Param tag_id query string true "标签ID"
// @Param shop_id query string true "店铺ID"
// @Success 200 {object} map[string]interface{} "查询成功"
// @Security BearerAuth
// @Router /admin/tag/online-products [get]
// @Router /shopOwner/tag/online-products [get]
func (h *ShopHandler) GetTagOnlineProducts(c *gin.Context) {
	tagID := c.Query("tag_id")
	if tagID == "" {
		errorResponse(c, http.StatusBadRequest, "缺少标签ID")
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

	products, err := h.shopService.GetTagOnlineProducts(tagID, validShopID.ToUint64())
	if err != nil {
		log2.Errorf("查询标签已上架商品失败: %v", err)
		errorResponse(c, http.StatusInternalServerError, "查询失败")
		return
	}

	successResponse(c, gin.H{
		"tag_id":   tagID,
		"products": products,
	})
}

func (h *ShopHandler) validateShopID(c *gin.Context, shopID shared.ID) (shared.ID, error) {
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
