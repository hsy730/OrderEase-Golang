package repositories

import (
	"errors"
	"orderease/models"
	"orderease/utils/cache"
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

type OrderStatsResponse struct {
	TodayOrders      int     `json:"todayOrders"`
	YesterdayOrders  int     `json:"yesterdayOrders"`
	TodayRevenue     float64 `json:"todayRevenue"`
	YesterdayRevenue float64 `json:"yesterdayRevenue"`
}

type ProductStatsResponse struct {
	ActiveProducts int `json:"activeProducts"`
	TotalProducts  int `json:"totalProducts"`
}

type UserStatsResponse struct {
	TodayUsers int `json:"todayUsers"`
	TotalUsers int `json:"totalUsers"`
}

type OrderEfficiencyResponse struct {
	AvgAcceptTime       float64 `json:"avgAcceptTime"`
	AvgCompleteTime     float64 `json:"avgCompleteTime"`
	TodayCompletionRate float64 `json:"todayCompletionRate"`
}

type SalesTrendResponse struct {
	Period string           `json:"period"`
	Data   []SalesTrendData `json:"data"`
}

type SalesTrendData struct {
	Date    string  `json:"date"`
	Orders  int     `json:"orders"`
	Revenue float64 `json:"revenue"`
}

type HotProduct struct {
	ID       snowflake.ID `json:"id,string"`
	Name     string       `json:"name"`
	ImageURL string       `json:"image_url"`
	Price    float64      `json:"price"`
	Sales    int          `json:"sales"`
}

type RecentOrder struct {
	ID         snowflake.ID `json:"id,string"`
	TotalPrice float64      `json:"total_price"`
	Status     int          `json:"status"`
	CreatedAt  time.Time    `json:"created_at"`
}

func (r *DashboardRepository) GetOrderStats(shopID snowflake.ID, todayStart, todayEnd, yesterdayStart, yesterdayEnd time.Time) (*OrderStatsResponse, error) {
	cacheKey := cache.BuildCacheKey(cache.CacheKeyOrderStats, int64(shopID), todayStart.Format("2006-01-02"))
	c := cache.GetCache()

	if cached, found := c.Get(cacheKey); found {
		if result, ok := cached.(*OrderStatsResponse); ok {
			return result, nil
		}
	}

	type StatsResult struct {
		TodayOrders      int64
		YesterdayOrders  int64
		TodayRevenue     float64
		YesterdayRevenue float64
	}

	var result StatsResult
	err := r.DB.Model(&models.Order{}).
		Select(`
			COALESCE(SUM(CASE WHEN created_at >= ? AND created_at < ? THEN 1 ELSE 0 END), 0) as today_orders,
			COALESCE(SUM(CASE WHEN created_at >= ? AND created_at < ? THEN 1 ELSE 0 END), 0) as yesterday_orders,
			COALESCE(SUM(CASE WHEN created_at >= ? AND created_at < ? THEN total_price ELSE 0 END), 0) as today_revenue,
			COALESCE(SUM(CASE WHEN created_at >= ? AND created_at < ? THEN total_price ELSE 0 END), 0) as yesterday_revenue
		`, todayStart, todayEnd, yesterdayStart, yesterdayEnd, todayStart, todayEnd, yesterdayStart, yesterdayEnd).
		Where("shop_id = ?", shopID).
		Where("created_at >= ? AND created_at < ?", yesterdayStart, todayEnd).
		Scan(&result).Error

	if err != nil {
		log2.Errorf("GetOrderStats failed: %v", err)
		return nil, errors.New("获取订单统计失败")
	}

	response := &OrderStatsResponse{
		TodayOrders:      int(result.TodayOrders),
		YesterdayOrders:  int(result.YesterdayOrders),
		TodayRevenue:     result.TodayRevenue,
		YesterdayRevenue: result.YesterdayRevenue,
	}

	c.Set(cacheKey, response, 2*time.Minute)
	return response, nil
}

func (r *DashboardRepository) GetProductStats(shopID snowflake.ID) (*ProductStatsResponse, error) {
	cacheKey := cache.BuildCacheKey(cache.CacheKeyProductStats, int64(shopID))
	c := cache.GetCache()

	if cached, found := c.Get(cacheKey); found {
		if result, ok := cached.(*ProductStatsResponse); ok {
			return result, nil
		}
	}

	var activeCount, totalCount int64

	if err := r.DB.Model(&models.Product{}).Where("shop_id = ? AND status = ?", shopID, models.ProductStatusOnline).Count(&activeCount).Error; err != nil {
		log2.Errorf("GetProductStats active count failed: %v", err)
		return nil, errors.New("获取在售商品数失败")
	}

	if err := r.DB.Model(&models.Product{}).Where("shop_id = ?", shopID).Count(&totalCount).Error; err != nil {
		log2.Errorf("GetProductStats total count failed: %v", err)
		return nil, errors.New("获取总商品数失败")
	}

	response := &ProductStatsResponse{
		ActiveProducts: int(activeCount),
		TotalProducts:  int(totalCount),
	}

	c.Set(cacheKey, response, 10*time.Minute)
	return response, nil
}

func (r *DashboardRepository) GetUserStats(shopID snowflake.ID, todayStart time.Time) (*UserStatsResponse, error) {
	cacheKey := cache.BuildCacheKey(cache.CacheKeyUserStats, int64(shopID), todayStart.Format("2006-01-02"))
	c := cache.GetCache()

	if cached, found := c.Get(cacheKey); found {
		if result, ok := cached.(*UserStatsResponse); ok {
			return result, nil
		}
	}

	var todayCount, totalCount int64

	r.DB.Model(&models.Order{}).
		Where("shop_id = ? AND created_at >= ?", shopID, todayStart).
		Distinct("user_id").
		Count(&todayCount)

	r.DB.Model(&models.Order{}).
		Where("shop_id = ?", shopID).
		Distinct("user_id").
		Count(&totalCount)

	response := &UserStatsResponse{
		TodayUsers: int(todayCount),
		TotalUsers: int(totalCount),
	}

	c.Set(cacheKey, response, 5*time.Minute)
	return response, nil
}

func (r *DashboardRepository) GetOrderEfficiency(shopID snowflake.ID, dateStart, dateEnd time.Time) (*OrderEfficiencyResponse, error) {
	cacheKey := cache.BuildCacheKey(cache.CacheKeyOrderEfficiency, int64(shopID), dateStart.Format("2006-01-02"))
	c := cache.GetCache()

	if cached, found := c.Get(cacheKey); found {
		if result, ok := cached.(*OrderEfficiencyResponse); ok {
			return result, nil
		}
	}

	var completedCount, cancelledCount int64

	r.DB.Model(&models.Order{}).Where("shop_id = ? AND created_at >= ? AND created_at < ? AND status = ?", shopID, dateStart, dateEnd, 9).Count(&completedCount)
	r.DB.Model(&models.Order{}).Where("shop_id = ? AND created_at >= ? AND created_at < ? AND status = ?", shopID, dateStart, dateEnd, 10).Count(&cancelledCount)

	finishedTotal := completedCount + cancelledCount
	completionRate := 0.0
	if finishedTotal > 0 {
		completionRate = (float64(completedCount) / float64(finishedTotal)) * 100
	}

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

	response := &OrderEfficiencyResponse{
		AvgAcceptTime:       acceptResult.AvgTime,
		AvgCompleteTime:     completeResult.AvgTime,
		TodayCompletionRate: completionRate,
	}

	c.Set(cacheKey, response, 5*time.Minute)
	return response, nil
}

func (r *DashboardRepository) GetSalesTrend(shopID snowflake.ID, period string) (*SalesTrendResponse, error) {
	cacheKey := cache.BuildCacheKey(cache.CacheKeySalesTrend, int64(shopID), period)
	c := cache.GetCache()

	if cached, found := c.Get(cacheKey); found {
		if result, ok := cached.(*SalesTrendResponse); ok {
			return result, nil
		}
	}

	now := time.Now()
	var startDate time.Time
	var groupByFormat string

	switch period {
	case "month":
		startDate = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
		groupByFormat = "%Y-%m-%d"
	case "year":
		startDate = time.Date(now.Year(), 1, 1, 0, 0, 0, 0, now.Location())
		groupByFormat = "%Y-%m"
	default:
		weekday := int(now.Weekday())
		if weekday == 0 {
			weekday = 7
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

	if period == "year" {
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

	trendData := make([]SalesTrendData, len(results))
	for i, r := range results {
		dateStr := r.Date
		if len(dateStr) == 10 {
			dateStr = dateStr[5:]
		} else if len(dateStr) == 7 {
		}

		trendData[i] = SalesTrendData{
			Date:    dateStr,
			Orders:  int(r.Orders),
			Revenue: r.Revenue,
		}
	}

	response := &SalesTrendResponse{
		Period: period,
		Data:   trendData,
	}

	c.Set(cacheKey, response, 10*time.Minute)
	return response, nil
}

func (r *DashboardRepository) GetHotProducts(shopID snowflake.ID, limit int) ([]HotProduct, error) {
	return r.GetHotProductsInRange(shopID, limit, time.Now().AddDate(0, 0, -30), time.Now())
}

func (r *DashboardRepository) GetHotProductsInRange(shopID snowflake.ID, limit int, startTime, endTime time.Time) ([]HotProduct, error) {
	cacheKey := cache.BuildCacheKey(cache.CacheKeyHotProducts, int64(shopID), startTime.Format("2006-01-02"), endTime.Format("2006-01-02"))
	c := cache.GetCache()

	if cached, found := c.Get(cacheKey); found {
		if result, ok := cached.([]HotProduct); ok {
			return result, nil
		}
	}

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
		LEFT JOIN orders o ON oi.order_id = o.id AND o.shop_id = ? AND o.created_at >= ? AND o.created_at <= ?
		WHERE p.shop_id = ?
		GROUP BY p.id, p.name, p.image_url, p.price
		ORDER BY sales DESC
		LIMIT ?
	`, shopID, startTime, endTime, shopID, limit).Scan(&results)

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

	c.Set(cacheKey, hotProducts, 15*time.Minute)
	return hotProducts, nil
}

func (r *DashboardRepository) GetRecentOrders(shopID snowflake.ID, limit int) ([]RecentOrder, error) {
	cacheKey := cache.BuildCacheKey(cache.CacheKeyRecentOrders, int64(shopID))
	c := cache.GetCache()

	if cached, found := c.Get(cacheKey); found {
		if result, ok := cached.([]RecentOrder); ok {
			return result, nil
		}
	}

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

	c.Set(cacheKey, recentOrders, 30*time.Second)
	return recentOrders, nil
}
