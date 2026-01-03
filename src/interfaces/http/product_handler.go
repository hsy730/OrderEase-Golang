package http

import (
	"net/http"
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
	shopID, err := strconv.ParseUint(shopIDStr, 10, 64)
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
	shopID, err := strconv.ParseUint(shopIDStr, 10, 64)
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
	shopID, err := strconv.ParseUint(shopIDStr, 10, 64)
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
	shopID, err := strconv.ParseUint(shopIDStr, 10, 64)
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
	shopID, err := strconv.ParseUint(shopIDStr, 10, 64)
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

func (h *ProductHandler) validateShopID(c *gin.Context, shopID uint64) (uint64, error) {
	requestUser, exists := c.Get("userInfo")
	if !exists {
		return shopID, nil
	}

	userInfo := requestUser.(interface {
		IsAdminUser() bool
		GetUserID() uint64
	})

	if !userInfo.IsAdminUser() {
		return userInfo.GetUserID(), nil
	}

	shop, err := h.shopService.GetShop(shopID)
	if err != nil {
		return 0, err
	}

	return shop.ID, nil
}
