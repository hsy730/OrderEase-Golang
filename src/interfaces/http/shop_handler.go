package http

import (
	"net/http"
	"orderease/application/dto"
	"orderease/application/services"
	"orderease/domain/order"
	"orderease/utils/log2"
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
		log2.Errorf("创建店铺失败: %v", err)
		errorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	successResponse(c, gin.H{
		"code": 200,
		"data": response,
	})
}

func (h *ShopHandler) GetShopInfo(c *gin.Context) {
	shopIDStr := c.Query("shop_id")
	shopID, err := strconv.ParseUint(shopIDStr, 10, 64)
	if err != nil || shopID <= 0 {
		errorResponse(c, http.StatusBadRequest, "无效的店铺ID")
		return
	}

	response, err := h.shopService.GetShop(shopID)
	if err != nil {
		if err.Error() == "店铺不存在" {
			errorResponse(c, http.StatusNotFound, "店铺不存在")
			return
		}
		log2.Errorf("查询店铺失败，ID: %d，错误: %v", shopID, err)
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
	shopID, err := strconv.ParseUint(shopIDStr, 10, 64)
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
		ShopID          uint64                 `json:"shop_id" binding:"required"`
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
	shopID, err := strconv.ParseUint(shopIDStr, 10, 64)
	if err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的店铺ID")
		return
	}

	response, err := h.shopService.GetShopTags(shopID)
	if err != nil {
		log2.Errorf("查询店铺标签失败，ID: %d，错误: %v", shopID, err)
		errorResponse(c, http.StatusInternalServerError, "查询失败")
		return
	}

	successResponse(c, response)
}
