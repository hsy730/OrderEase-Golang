package http

import (
	"net/http"
	"orderease/application/dto"
	"orderease/application/services"
	"orderease/domain/order"
	"orderease/domain/shared"
	"orderease/utils/log2"
	"strconv"

	"github.com/gin-gonic/gin"
)

type OrderHandler struct {
	orderService *services.OrderService
	shopService  *services.ShopService
}

func NewOrderHandler(
	orderService *services.OrderService,
	shopService *services.ShopService,
) *OrderHandler {
	return &OrderHandler{
		orderService: orderService,
		shopService:  shopService,
	}
}

func (h *OrderHandler) CreateOrder(c *gin.Context) {
	var req dto.CreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的订单数据: "+err.Error())
		return
	}

	shopID, err := h.validateShopID(c, req.ShopID)
	if err != nil {
		errorResponse(c, http.StatusBadRequest, err.Error())
		return
	}
	req.ShopID = shopID

	response, err := h.orderService.CreateOrder(&req)
	if err != nil {
		log2.Errorf("创建订单失败: %v", err)
		errorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	successResponse(c, response)
}

func (h *OrderHandler) GetOrder(c *gin.Context) {
	idStr := c.Query("id")
	if idStr == "" {
		errorResponse(c, http.StatusBadRequest, "缺少订单ID")
		return
	}

	id, err := shared.ParseIDFromString(idStr)
	if err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的订单ID")
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

	response, err := h.orderService.GetOrder(id, validShopID)
	if err != nil {
		log2.Errorf("查询订单失败: %v", err)
		errorResponse(c, http.StatusNotFound, err.Error())
		return
	}

	successResponse(c, response)
}

func (h *OrderHandler) GetOrders(c *gin.Context) {
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

	response, err := h.orderService.GetOrders(validShopID, page, pageSize)
	if err != nil {
		log2.Errorf("查询订单列表失败: %v", err)
		errorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	successResponse(c, response)
}

func (h *OrderHandler) GetOrdersByUser(c *gin.Context) {
	userIDStr := c.Query("user_id")
	if userIDStr == "" {
		errorResponse(c, http.StatusBadRequest, "缺少用户ID")
		return
	}

	userID, err := shared.ParseIDFromString(userIDStr)
	if err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的用户ID")
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

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	if pageSize > 100 {
		pageSize = 100
	}

	response, err := h.orderService.GetOrdersByUser(userID, validShopID, page, pageSize)
	if err != nil {
		log2.Errorf("查询用户订单失败: %v", err)
		errorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	successResponse(c, response)
}

func (h *OrderHandler) GetUnfinishedOrders(c *gin.Context) {
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

	shop, err := h.shopService.GetShop(validShopID)
	if err != nil {
		errorResponse(c, http.StatusInternalServerError, "获取店铺信息失败")
		return
	}

	response, err := h.orderService.GetUnfinishedOrders(validShopID, shop.OrderStatusFlow, page, pageSize)
	if err != nil {
		log2.Errorf("查询未完成订单列表失败: %v", err)
		errorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	successResponse(c, response)
}

func (h *OrderHandler) SearchOrders(c *gin.Context) {
	var req dto.SearchOrdersRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的查询参数: "+err.Error())
		return
	}

	if req.Page < 1 {
		req.Page = 1
	}
	if req.PageSize < 1 || req.PageSize > 100 {
		req.PageSize = 10
	}

	validShopID, err := h.validateShopID(c, req.ShopID)
	if err != nil {
		errorResponse(c, http.StatusBadRequest, err.Error())
		return
	}
	req.ShopID = validShopID

	response, err := h.orderService.SearchOrders(&req)
	if err != nil {
		log2.Errorf("查询订单列表失败: %v", err)
		errorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	successResponse(c, response)
}

func (h *OrderHandler) UpdateOrderStatus(c *gin.Context) {
	type UpdateOrderStatusRequest struct {
		ID         string    `json:"id" binding:"required"`
		ShopID     shared.ID `json:"shop_id" binding:"required"`
		NextStatus int       `json:"next_status" binding:"required"`
	}

	var req UpdateOrderStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log2.Errorf("无效的请求参数: %v", err)
		errorResponse(c, http.StatusBadRequest, "无效的请求参数")
		return
	}

	id, err := shared.ParseIDFromString(req.ID)
	if err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的订单ID格式")
		return
	}

	validShopID, err := h.validateShopID(c, req.ShopID)
	if err != nil {
		errorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	shop, err := h.shopService.GetShop(validShopID)
	if err != nil {
		errorResponse(c, http.StatusInternalServerError, "获取店铺信息失败")
		return
	}

	newStatus := order.OrderStatus(req.NextStatus)
	if err := h.orderService.UpdateOrderStatus(id, validShopID, newStatus, shop.OrderStatusFlow); err != nil {
		log2.Errorf("更新订单状态失败: %v", err)
		errorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	successResponse(c, gin.H{
		"message": "订单状态更新成功",
	})
}

func (h *OrderHandler) DeleteOrder(c *gin.Context) {
	idStr := c.Query("id")
	if idStr == "" {
		errorResponse(c, http.StatusBadRequest, "缺少订单ID")
		return
	}

	id, err := shared.ParseIDFromString(idStr)
	if err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的订单ID")
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

	if err := h.orderService.DeleteOrder(id, validShopID); err != nil {
		log2.Errorf("删除订单失败: %v", err)
		errorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	successResponse(c, gin.H{"message": "订单删除成功"})
}

func (h *OrderHandler) validateShopID(c *gin.Context, shopID shared.ID) (shared.ID, error) {
	requestUser, exists := c.Get("userInfo")
	if !exists {
		return shared.ID(0), nil
	}

	userInfo := requestUser.(interface {
		IsAdminUser() bool
		GetUserID() uint64
	})

	if !userInfo.IsAdminUser() && c.Request.URL.Path != "" {
		return shared.ParseIDFromUint64(userInfo.GetUserID()), nil
	}

	shop, err := h.shopService.GetShop(shopID)
	if err != nil {
		return shared.ID(0), err
	}

	return shop.ID, nil
}

func errorResponse(c *gin.Context, code int, message string) {
	log2.Errorf("错误响应: %d - %s", code, message)
	c.JSON(code, gin.H{"error": message})
}

func successResponse(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, data)
}
