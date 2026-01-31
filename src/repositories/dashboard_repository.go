package repositories

import (
	"errors"
	"orderease/models"
	"orderease/utils/log2"
	"time"

	"github.com/bwmarrin/snowflake"
	"gorm.io/gorm"
)

type DashboardRepository struct {
	DB *gorm.DB
}

func NewDashboardRepository(db *gorm.DB) *DashboardRepository {
	return &DashboardRepository{DB: db}
}

// OrderStatsResponse 订单统计响应
type OrderStatsResponse struct {
	TodayOrders      int     `json:"todayOrders"`
	YesterdayOrders  int     `json:"yesterdayOrders"`
	TodayRevenue     float64 `json:"todayRevenue"`
	YesterdayRevenue float64 `json:"yesterdayRevenue"`
}

// ProductStatsResponse 商品统计响应
type ProductStatsResponse struct {
	ActiveProducts int `json:"activeProducts"`
	TotalProducts  int `json:"totalProducts"`
}

// UserStatsResponse 用户统计响应
type UserStatsResponse struct {
	TodayUsers int `json:"todayUsers"`
	TotalUsers int `json:"totalUsers"`
}

// OrderEfficiencyResponse 订单效率响应
type OrderEfficiencyResponse struct {
	AvgAcceptTime       float64 `json:"avgAcceptTime"`
	AvgCompleteTime     float64 `json:"avgCompleteTime"`
	TodayCompletionRate float64 `json:"todayCompletionRate"`
}

// SalesTrendResponse 销售趋势响应
type SalesTrendResponse struct {
	Period string           `json:"period"` // week/month/year
	Data   []SalesTrendData `json:"data"`
}

// SalesTrendData 销售趋势数据点
type SalesTrendData struct {
	Date    string  `json:"date"`    // 日期标签：周"01-25"，月"01-25"，年"2026-01"
	Orders  int     `json:"orders"`  // 订单数
	Revenue float64 `json:"revenue"` // 销售额
}

// HotProduct 热销商品
type HotProduct struct {
	ID       snowflake.ID `json:"id"`
	Name     string       `json:"name"`
	ImageURL string       `json:"image_url"`
	Price    float64      `json:"price"`
	Sales    int          `json:"sales"`
}

// RecentOrder 最近订单
type RecentOrder struct {
	ID         snowflake.ID `json:"id"`
	TotalPrice float64      `json:"total_price"`
	Status     int          `json:"status"`
	CreatedAt  time.Time    `json:"created_at"`
}

// GetOrderStats 获取订单统计（今日和昨日）
func (r *DashboardRepository) GetOrderStats(shopID snowflake.ID, todayStart, todayEnd, yesterdayStart, yesterdayEnd time.Time) (*OrderStatsResponse, error) {
	var todayCount, yesterdayCount int64
	var todayRevenue, yesterdayRevenue float64

	// 今日订单统计
	todayQuery := r.DB.Model(&models.Order{}).Where("shop_id = ? AND created_at >= ? AND created_at < ?", shopID, todayStart, todayEnd)
	if err := todayQuery.Count(&todayCount).Error; err != nil {
		log2.Errorf("GetOrderStats today count failed: %v", err)
		return nil, errors.New("获取今日订单数失败")
	}

	// 今日销售额
	todayQuery.Select("COALESCE(SUM(total_price), 0)").Scan(&todayRevenue)

	// 昨日订单统计
	yesterdayQuery := r.DB.Model(&models.Order{}).Where("shop_id = ? AND created_at >= ? AND created_at < ?", shopID, yesterdayStart, yesterdayEnd)
	if err := yesterdayQuery.Count(&yesterdayCount).Error; err != nil {
		log2.Errorf("GetOrderStats yesterday count failed: %v", err)
		return nil, errors.New("获取昨日订单数失败")
	}

	// 昨日销售额
	yesterdayQuery.Select("COALESCE(SUM(total_price), 0)").Scan(&yesterdayRevenue)

	return &OrderStatsResponse{
		TodayOrders:      int(todayCount),
		YesterdayOrders:  int(yesterdayCount),
		TodayRevenue:     todayRevenue,
		YesterdayRevenue: yesterdayRevenue,
	}, nil
}

// GetProductStats 获取商品统计
func (r *DashboardRepository) GetProductStats(shopID snowflake.ID) (*ProductStatsResponse, error) {
	var activeCount, totalCount int64

	// 在售商品数
	if err := r.DB.Model(&models.Product{}).Where("shop_id = ? AND status = ?", shopID, models.ProductStatusOnline).Count(&activeCount).Error; err != nil {
		log2.Errorf("GetProductStats active count failed: %v", err)
		return nil, errors.New("获取在售商品数失败")
	}

	// 总商品数
	if err := r.DB.Model(&models.Product{}).Where("shop_id = ?", shopID).Count(&totalCount).Error; err != nil {
		log2.Errorf("GetProductStats total count failed: %v", err)
		return nil, errors.New("获取总商品数失败")
	}

	return &ProductStatsResponse{
		ActiveProducts: int(activeCount),
		TotalProducts:  int(totalCount),
	}, nil
}

// GetUserStats 获取用户统计（按店铺过滤，基于订单用户去重）
func (r *DashboardRepository) GetUserStats(shopID snowflake.ID, todayStart time.Time) (*UserStatsResponse, error) {
	var todayCount, totalCount int64

	// 今日有订单的用户（去重）
	r.DB.Model(&models.Order{}).
		Where("shop_id = ? AND created_at >= ?", shopID, todayStart).
		Distinct("user_id").
		Count(&todayCount)

	// 该店铺所有有订单的用户（去重）
	r.DB.Model(&models.Order{}).
		Where("shop_id = ?", shopID).
		Distinct("user_id").
		Count(&totalCount)

	return &UserStatsResponse{
		TodayUsers: int(todayCount),
		TotalUsers: int(totalCount),
	}, nil
}

// GetOrderEfficiency 获取订单效率指标
func (r *DashboardRepository) GetOrderEfficiency(shopID snowflake.ID, dateStart, dateEnd time.Time) (*OrderEfficiencyResponse, error) {
	// 今日完成率计算
	var completedCount, cancelledCount int64

	// 已完成订单数（status = 9）
	r.DB.Model(&models.Order{}).Where("shop_id = ? AND created_at >= ? AND created_at < ? AND status = ?", shopID, dateStart, dateEnd, 9).Count(&completedCount)

	// 已取消订单数（status = 10）
	r.DB.Model(&models.Order{}).Where("shop_id = ? AND created_at >= ? AND created_at < ? AND status = ?", shopID, dateStart, dateEnd, 10).Count(&cancelledCount)

	finishedTotal := completedCount + cancelledCount
	completionRate := 0.0
	if finishedTotal > 0 {
		completionRate = (float64(completedCount) / float64(finishedTotal)) * 100
	}

	// 平均接单时间（从创建到状态1）
	type AcceptTimeResult struct {
		AvgTime float64
	}
	var acceptResult AcceptTimeResult
	r.DB.Raw(`
		SELECT AVG(TIMESTAMPDIFF(MINUTE, o.created_at, l.changed_time)) as avg_time
		FROM order_status_logs l
		JOIN orders o ON l.order_id = o.id
		WHERE l.new_status = 1 AND o.shop_id = ? AND l.changed_time >= ? AND l.changed_time < ?
	`, shopID, dateStart, dateEnd).Scan(&acceptResult)

	// 平均完成时间（从状态1到状态9）
	type CompleteTimeResult struct {
		AvgTime float64
	}
	var completeResult CompleteTimeResult
	r.DB.Raw(`
		SELECT AVG(TIMESTAMPDIFF(MINUTE,
			(SELECT changed_time FROM order_status_logs WHERE order_id = l.order_id AND new_status = 1 LIMIT 1),
			l.changed_time
		)) as avg_time
		FROM order_status_logs l
		JOIN orders o ON l.order_id = o.id
		WHERE l.new_status = 9 AND o.shop_id = ? AND l.changed_time >= ? AND l.changed_time < ?
	`, shopID, dateStart, dateEnd).Scan(&completeResult)

	return &OrderEfficiencyResponse{
		AvgAcceptTime:       acceptResult.AvgTime,
		AvgCompleteTime:     completeResult.AvgTime,
		TodayCompletionRate: completionRate,
	}, nil
}

// GetSalesTrend 获取销售趋势
func (r *DashboardRepository) GetSalesTrend(shopID snowflake.ID, period string) (*SalesTrendResponse, error) {
	now := time.Now()
	var startDate time.Time
	var groupByFormat string

	switch period {
	case "month":
		// 本月：从本月1号到今天
		startDate = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
		groupByFormat = "%Y-%m-%d"
	case "year":
		// 全年：从今年1月1号到今天
		startDate = time.Date(now.Year(), 1, 1, 0, 0, 0, 0, now.Location())
		groupByFormat = "%Y-%m"
	default: // "week"
		// 本周：从本周一到今天
		weekday := int(now.Weekday())
		if weekday == 0 {
			weekday = 7 // 周日改为7
		}
		startDate = time.Date(now.Year(), now.Month(), now.Day()-weekday+1, 0, 0, 0, 0, now.Location())
		groupByFormat = "%Y-%m-%d"
		period = "week"
	}

	type TrendResult struct {
		Date    string
		Orders  int64
		Revenue float64
	}

	var results []TrendResult

	// 根据周期使用不同的SQL查询
	if period == "year" {
		// 按月分组
		r.DB.Raw(`
			SELECT DATE_FORMAT(created_at, ?) as date,
			       COUNT(*) as orders,
			       COALESCE(SUM(total_price), 0) as revenue
			FROM orders
			WHERE shop_id = ? AND created_at >= ? AND created_at <= ?
			GROUP BY DATE_FORMAT(created_at, ?)
			ORDER BY created_at
		`, groupByFormat, shopID, startDate, now, groupByFormat).Scan(&results)
	} else {
		// 按天分组
		r.DB.Raw(`
			SELECT DATE_FORMAT(created_at, ?) as date,
			       COUNT(*) as orders,
			       COALESCE(SUM(total_price), 0) as revenue
			FROM orders
			WHERE shop_id = ? AND created_at >= ? AND created_at <= ?
			GROUP BY DATE_FORMAT(created_at, ?)
			ORDER BY created_at
		`, groupByFormat, shopID, startDate, now, groupByFormat).Scan(&results)
	}

	// 转换为响应格式
	trendData := make([]SalesTrendData, len(results))
	for i, r := range results {
		// 格式化日期
		dateStr := r.Date
		if len(dateStr) == 10 { // YYYY-MM-DD格式
			// 转换为 MM-DD
			dateStr = dateStr[5:] // MM-DD
		} else if len(dateStr) == 7 { // YYYY-MM格式
			// 保持 YYYY-MM
		}

		trendData[i] = SalesTrendData{
			Date:    dateStr,
			Orders:  int(r.Orders),
			Revenue: r.Revenue,
		}
	}

	return &SalesTrendResponse{
		Period: period,
		Data:   trendData,
	}, nil
}

// GetHotProducts 获取热销商品
func (r *DashboardRepository) GetHotProducts(shopID snowflake.ID, limit int) ([]HotProduct, error) {
	type ProductSales struct {
		ID       snowflake.ID
		Name     string
		ImageURL string
		Price    float64
		Sales    int
	}

	var results []ProductSales
	r.DB.Raw(`
		SELECT p.id, p.name, p.image_url, p.price, COUNT(oi.id) as sales
		FROM products p
		LEFT JOIN order_items oi ON p.id = oi.product_id
		WHERE p.shop_id = ?
		GROUP BY p.id, p.name, p.image_url, p.price
		ORDER BY sales DESC
		LIMIT ?
	`, shopID, limit).Scan(&results)

	hotProducts := make([]HotProduct, len(results))
	for i, r := range results {
		hotProducts[i] = HotProduct{
			ID:       r.ID,
			Name:     r.Name,
			ImageURL: r.ImageURL,
			Price:    r.Price,
			Sales:    r.Sales,
		}
	}

	return hotProducts, nil
}

// GetRecentOrders 获取最近订单
func (r *DashboardRepository) GetRecentOrders(shopID snowflake.ID, limit int) ([]RecentOrder, error) {
	var orders []models.Order
	if err := r.DB.Where("shop_id = ?", shopID).
		Order("created_at DESC").
		Limit(limit).
		Find(&orders).Error; err != nil {
		log2.Errorf("GetRecentOrders failed: %v", err)
		return nil, errors.New("获取最近订单失败")
	}

	recentOrders := make([]RecentOrder, len(orders))
	for i, o := range orders {
		recentOrders[i] = RecentOrder{
			ID:         o.ID,
			TotalPrice: float64(o.TotalPrice),
			Status:     o.Status,
			CreatedAt:  o.CreatedAt,
		}
	}

	return recentOrders, nil
}
