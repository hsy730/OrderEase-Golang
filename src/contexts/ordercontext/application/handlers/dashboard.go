package handlers

import (
	"orderease/contexts/ordercontext/infrastructure/repositories"
	"orderease/utils/log2"
	"sync"
	"time"

	"github.com/bwmarrin/snowflake"
	"github.com/gin-gonic/gin"
)

// DashboardStatsResponse 数据看板统计响应
type DashboardStatsResponse struct {
	OrderStats      repositories.OrderStatsResponse      `json:"orderStats"`
	ProductStats    repositories.ProductStatsResponse    `json:"productStats"`
	UserStats       repositories.UserStatsResponse       `json:"userStats"`
	OrderEfficiency repositories.OrderEfficiencyResponse `json:"orderEfficiency"`
	SalesTrend      repositories.SalesTrendResponse      `json:"salesTrend"`
	HotProducts     []repositories.HotProduct             `json:"hotProducts"`
	RecentOrders    []repositories.RecentOrder            `json:"recentOrders"`
}

// GetDashboardStats 获取数据看板统计
func (h *Handler) GetDashboardStats(c *gin.Context) {
	var shopID snowflake.ID
	var err error

	// 获取shop_id（管理员从query参数，店主从上下文）
	if shopIDStr := c.Query("shop_id"); shopIDStr != "" {
		// 管理员模式：从query参数获取
		shopID, err = snowflake.ParseString(shopIDStr)
		if err != nil {
			errorResponse(c, 400, "店铺ID格式错误")
			return
		}
	} else {
		// 店主模式：从用户上下文获取
		requestUser, err := h.getRequestUserInfo(c)
		if err != nil {
			errorResponse(c, 401, "获取用户信息失败")
			return
		}
		if requestUser.IsAdmin {
			errorResponse(c, 400, "管理员请指定shop_id参数")
			return
		}
		shopID = snowflake.ID(requestUser.UserID)
	}

	// 验证店铺是否存在
	exist, err := h.productRepo.CheckShopExists(shopID)
	if err != nil {
		log2.Errorf("CheckShopExists failed: %v", err)
		errorResponse(c, 500, "验证店铺失败")
		return
	}
	if !exist {
		errorResponse(c, 404, "店铺不存在")
		return
	}

	// 获取销售趋势周期参数
	period := c.DefaultQuery("period", "week")

	// 计算日期时间
	now := time.Now()
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	todayEnd := todayStart.AddDate(0, 0, 1)
	yesterdayStart := todayStart.AddDate(0, 0, -1)
	yesterdayEnd := todayStart

	// 并发获取所有统计数据
	var wg sync.WaitGroup
	var stats DashboardStatsResponse
	var orderStats *repositories.OrderStatsResponse
	var productStats *repositories.ProductStatsResponse
	var userStats *repositories.UserStatsResponse
	var orderEfficiency *repositories.OrderEfficiencyResponse
	var salesTrend *repositories.SalesTrendResponse
	var hotProducts []repositories.HotProduct
	var recentOrders []repositories.RecentOrder
	var statsErr error

	wg.Add(7)

	// 获取订单统计
	go func() {
		defer wg.Done()
		orderStats, statsErr = h.dashboardRepo.GetOrderStats(shopID, todayStart, todayEnd, yesterdayStart, yesterdayEnd)
		if statsErr != nil {
			log2.Errorf("GetOrderStats failed: %v", statsErr)
		}
	}()

	// 获取商品统计
	go func() {
		defer wg.Done()
		productStats, statsErr = h.dashboardRepo.GetProductStats(shopID)
		if statsErr != nil {
			log2.Errorf("GetProductStats failed: %v", statsErr)
		}
	}()

	// 获取用户统计
	go func() {
		defer wg.Done()
		userStats, statsErr = h.dashboardRepo.GetUserStats(shopID, todayStart)
		if statsErr != nil {
			log2.Errorf("GetUserStats failed: %v", statsErr)
		}
	}()

	// 获取订单效率
	go func() {
		defer wg.Done()
		orderEfficiency, statsErr = h.dashboardRepo.GetOrderEfficiency(shopID, todayStart, todayEnd)
		if statsErr != nil {
			log2.Errorf("GetOrderEfficiency failed: %v", statsErr)
		}
	}()

	// 获取销售趋势
	go func() {
		defer wg.Done()
		salesTrend, statsErr = h.dashboardRepo.GetSalesTrend(shopID, period)
		if statsErr != nil {
			log2.Errorf("GetSalesTrend failed: %v", statsErr)
		}
	}()

	// 获取热销商品
	go func() {
		defer wg.Done()
		hotProducts, statsErr = h.dashboardRepo.GetHotProducts(shopID, 5)
		if statsErr != nil {
			log2.Errorf("GetHotProducts failed: %v", statsErr)
		}
	}()

	// 获取最近订单
	go func() {
		defer wg.Done()
		recentOrders, statsErr = h.dashboardRepo.GetRecentOrders(shopID, 5)
		if statsErr != nil {
			log2.Errorf("GetRecentOrders failed: %v", statsErr)
		}
	}()

	wg.Wait()

	// 组装响应（解引用指针）
	if orderStats != nil {
		stats.OrderStats = *orderStats
	}
	if productStats != nil {
		stats.ProductStats = *productStats
	}
	if userStats != nil {
		stats.UserStats = *userStats
	}
	if orderEfficiency != nil {
		stats.OrderEfficiency = *orderEfficiency
	}
	if salesTrend != nil {
		stats.SalesTrend = *salesTrend
	}
	stats.HotProducts = hotProducts
	stats.RecentOrders = recentOrders

	successResponse(c, gin.H{
		"orderStats":      stats.OrderStats,
		"productStats":    stats.ProductStats,
		"userStats":       stats.UserStats,
		"orderEfficiency": stats.OrderEfficiency,
		"salesTrend":      stats.SalesTrend,
		"hotProducts":     stats.HotProducts,
		"recentOrders":    stats.RecentOrders,
	})
}
