package repositories

import (
	"database/sql"
	"fmt"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/bwmarrin/snowflake"
	"github.com/stretchr/testify/assert"
	"orderease/models"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func setupDashboardTestDB(t *testing.T) (*gorm.DB, sqlmock.Sqlmock, *sql.DB) {
	t.Helper()

	sqlDB, mock, err := sqlmock.New()
	assert.NoError(t, err)

	dialector := mysql.Config{
		Conn:                      sqlDB,
		SkipInitializeWithVersion: true,
	}

	db, err := gorm.Open(mysql.New(dialector), &gorm.Config{})
	assert.NoError(t, err)

	return db, mock, sqlDB
}

func TestNewDashboardRepository(t *testing.T) {
	db, _, sqlDB := setupDashboardTestDB(t)
	defer sqlDB.Close()

	repo := NewDashboardRepository(db)

	assert.NotNil(t, repo)
	assert.Equal(t, db, repo.DB)
}

func TestDashboardRepository_GetOrderStats_Success(t *testing.T) {
	db, mock, sqlDB := setupDashboardTestDB(t)
	defer sqlDB.Close()

	shopID := snowflake.ID(123)
	now := time.Now()
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	todayEnd := todayStart.Add(24 * time.Hour)
	yesterdayStart := todayStart.Add(-24 * time.Hour)
	yesterdayEnd := todayStart

	todayCountRows := sqlmock.NewRows([]string{"count"}).AddRow(10)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT count(*) FROM `orders`")).
		WithArgs(shopID, todayStart, todayEnd).
		WillReturnRows(todayCountRows)

	todayRevenueRows := sqlmock.NewRows([]string{"total_price"}).AddRow(1000.0)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT COALESCE(SUM(total_price), 0) FROM `orders`")).
		WillReturnRows(todayRevenueRows)

	yesterdayCountRows := sqlmock.NewRows([]string{"count"}).AddRow(8)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT count(*) FROM `orders`")).
		WithArgs(shopID, yesterdayStart, yesterdayEnd).
		WillReturnRows(yesterdayCountRows)

	yesterdayRevenueRows := sqlmock.NewRows([]string{"total_price"}).AddRow(800.0)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT COALESCE(SUM(total_price), 0) FROM `orders`")).
		WillReturnRows(yesterdayRevenueRows)

	repo := NewDashboardRepository(db)
	stats, err := repo.GetOrderStats(shopID, todayStart, todayEnd, yesterdayStart, yesterdayEnd)

	assert.NoError(t, err)
	assert.Equal(t, 10, stats.TodayOrders)
	assert.Equal(t, 8, stats.YesterdayOrders)
	assert.Equal(t, 1000.0, stats.TodayRevenue)
	assert.Equal(t, 800.0, stats.YesterdayRevenue)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDashboardRepository_GetOrderStats_DatabaseError(t *testing.T) {
	db, mock, sqlDB := setupDashboardTestDB(t)
	defer sqlDB.Close()

	shopID := snowflake.ID(123)
	now := time.Now()
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	todayEnd := todayStart.Add(24 * time.Hour)
	yesterdayStart := todayStart.Add(-24 * time.Hour)
	yesterdayEnd := todayStart

	mock.ExpectQuery(regexp.QuoteMeta("SELECT count(*) FROM `orders`")).
		WithArgs(shopID, todayStart, todayEnd).
		WillReturnError(fmt.Errorf("database error"))

	repo := NewDashboardRepository(db)
	stats, err := repo.GetOrderStats(shopID, todayStart, todayEnd, yesterdayStart, yesterdayEnd)

	assert.Error(t, err)
	assert.Equal(t, "获取今日订单数失败", err.Error())
	assert.Nil(t, stats)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDashboardRepository_GetProductStats_Success(t *testing.T) {
	db, mock, sqlDB := setupDashboardTestDB(t)
	defer sqlDB.Close()

	shopID := snowflake.ID(123)

	activeRows := sqlmock.NewRows([]string{"count"}).AddRow(15)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT count(*) FROM `products`")).
		WithArgs(shopID, models.ProductStatusOnline).
		WillReturnRows(activeRows)

	totalRows := sqlmock.NewRows([]string{"count"}).AddRow(20)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT count(*) FROM `products`")).
		WithArgs(shopID).
		WillReturnRows(totalRows)

	repo := NewDashboardRepository(db)
	stats, err := repo.GetProductStats(shopID)

	assert.NoError(t, err)
	assert.Equal(t, 15, stats.ActiveProducts)
	assert.Equal(t, 20, stats.TotalProducts)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDashboardRepository_GetProductStats_DatabaseError(t *testing.T) {
	db, mock, sqlDB := setupDashboardTestDB(t)
	defer sqlDB.Close()

	shopID := snowflake.ID(123)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT count(*) FROM `products`")).
		WithArgs(shopID, models.ProductStatusOnline).
		WillReturnError(fmt.Errorf("database error"))

	repo := NewDashboardRepository(db)
	stats, err := repo.GetProductStats(shopID)

	assert.Error(t, err)
	assert.Equal(t, "获取在售商品数失败", err.Error())
	assert.Nil(t, stats)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDashboardRepository_GetUserStats_Success(t *testing.T) {
	db, mock, sqlDB := setupDashboardTestDB(t)
	defer sqlDB.Close()

	shopID := snowflake.ID(123)
	now := time.Now()
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	todayRows := sqlmock.NewRows([]string{"count"}).AddRow(5)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT count(DISTINCT `user_id`) FROM `orders`")).
		WithArgs(shopID, todayStart).
		WillReturnRows(todayRows)

	totalRows := sqlmock.NewRows([]string{"count"}).AddRow(50)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT count(DISTINCT `user_id`) FROM `orders`")).
		WithArgs(shopID).
		WillReturnRows(totalRows)

	repo := NewDashboardRepository(db)
	stats, err := repo.GetUserStats(shopID, todayStart)

	assert.NoError(t, err)
	assert.Equal(t, 5, stats.TodayUsers)
	assert.Equal(t, 50, stats.TotalUsers)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDashboardRepository_GetRecentOrders_Success(t *testing.T) {
	db, mock, sqlDB := setupDashboardTestDB(t)
	defer sqlDB.Close()

	shopID := snowflake.ID(123)
	now := time.Now()

	rows := sqlmock.NewRows([]string{"id", "shop_id", "user_id", "total_price", "status", "created_at", "updated_at"}).
		AddRow(1, shopID, 100, 500.0, 1, now, now).
		AddRow(2, shopID, 101, 300.0, 2, now.Add(-1*time.Hour), now.Add(-1*time.Hour))

	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `orders`")).
		WithArgs(shopID).
		WillReturnRows(rows)

	repo := NewDashboardRepository(db)
	orders, err := repo.GetRecentOrders(shopID, 10)

	assert.NoError(t, err)
	assert.Len(t, orders, 2)
	assert.Equal(t, snowflake.ID(1), orders[0].ID)
	assert.Equal(t, 500.0, orders[0].TotalPrice)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDashboardRepository_GetRecentOrders_DatabaseError(t *testing.T) {
	db, mock, sqlDB := setupDashboardTestDB(t)
	defer sqlDB.Close()

	shopID := snowflake.ID(123)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `orders`")).
		WithArgs(shopID).
		WillReturnError(fmt.Errorf("database error"))

	repo := NewDashboardRepository(db)
	orders, err := repo.GetRecentOrders(shopID, 10)

	assert.Error(t, err)
	assert.Equal(t, "获取最近订单失败", err.Error())
	assert.Nil(t, orders)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDashboardRepository_GetHotProducts_Success(t *testing.T) {
	db, mock, sqlDB := setupDashboardTestDB(t)
	defer sqlDB.Close()

	shopID := snowflake.ID(123)

	rows := sqlmock.NewRows([]string{"id", "name", "image_url", "price", "sales"}).
		AddRow(1, "Product 1", "http://example.com/1.jpg", 100.0, 50).
		AddRow(2, "Product 2", "http://example.com/2.jpg", 200.0, 30)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT p.id, p.name, p.image_url, p.price, COUNT(oi.id) as sales")).
		WithArgs(shopID, 5).
		WillReturnRows(rows)

	repo := NewDashboardRepository(db)
	products, err := repo.GetHotProducts(shopID, 5)

	assert.NoError(t, err)
	assert.Len(t, products, 2)
	assert.Equal(t, "Product 1", products[0].Name)
	assert.Equal(t, 50, products[0].Sales)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDashboardRepository_GetOrderEfficiency_Success(t *testing.T) {
	db, mock, sqlDB := setupDashboardTestDB(t)
	defer sqlDB.Close()

	shopID := snowflake.ID(123)
	now := time.Now()
	dateStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	dateEnd := dateStart.Add(24 * time.Hour)

	completedRows := sqlmock.NewRows([]string{"count"}).AddRow(8)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT count(*) FROM `orders`")).
		WithArgs(shopID, dateStart, dateEnd, 9).
		WillReturnRows(completedRows)

	cancelledRows := sqlmock.NewRows([]string{"count"}).AddRow(2)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT count(*) FROM `orders`")).
		WithArgs(shopID, dateStart, dateEnd, 10).
		WillReturnRows(cancelledRows)

	acceptRows := sqlmock.NewRows([]string{"avg_time"}).AddRow(5.5)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT AVG(TIMESTAMPDIFF(MINUTE, o.created_at, l.changed_time)")).
		WillReturnRows(acceptRows)

	completeRows := sqlmock.NewRows([]string{"avg_time"}).AddRow(30.0)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT AVG(TIMESTAMPDIFF(MINUTE")).
		WillReturnRows(completeRows)

	repo := NewDashboardRepository(db)
	efficiency, err := repo.GetOrderEfficiency(shopID, dateStart, dateEnd)

	assert.NoError(t, err)
	assert.Equal(t, 5.5, efficiency.AvgAcceptTime)
	assert.Equal(t, 30.0, efficiency.AvgCompleteTime)
	assert.Equal(t, 80.0, efficiency.TodayCompletionRate)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDashboardRepository_GetSalesTrend_Week(t *testing.T) {
	db, mock, sqlDB := setupDashboardTestDB(t)
	defer sqlDB.Close()

	shopID := snowflake.ID(123)

	rows := sqlmock.NewRows([]string{"date", "orders", "revenue"}).
		AddRow("2026-02-20", 10, 1000.0).
		AddRow("2026-02-21", 15, 1500.0)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT DATE_FORMAT(created_at")).
		WillReturnRows(rows)

	repo := NewDashboardRepository(db)
	trend, err := repo.GetSalesTrend(shopID, "week")

	assert.NoError(t, err)
	assert.Equal(t, "week", trend.Period)
	assert.Len(t, trend.Data, 2)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDashboardRepository_GetSalesTrend_Month(t *testing.T) {
	db, mock, sqlDB := setupDashboardTestDB(t)
	defer sqlDB.Close()

	shopID := snowflake.ID(123)

	rows := sqlmock.NewRows([]string{"date", "orders", "revenue"}).
		AddRow("2026-02-01", 20, 2000.0).
		AddRow("2026-02-02", 25, 2500.0)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT DATE_FORMAT(created_at")).
		WillReturnRows(rows)

	repo := NewDashboardRepository(db)
	trend, err := repo.GetSalesTrend(shopID, "month")

	assert.NoError(t, err)
	assert.Equal(t, "month", trend.Period)
	assert.Len(t, trend.Data, 2)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDashboardRepository_GetSalesTrend_Year(t *testing.T) {
	db, mock, sqlDB := setupDashboardTestDB(t)
	defer sqlDB.Close()

	shopID := snowflake.ID(123)

	rows := sqlmock.NewRows([]string{"date", "orders", "revenue"}).
		AddRow("2026-01", 100, 10000.0).
		AddRow("2026-02", 120, 12000.0)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT DATE_FORMAT(created_at")).
		WillReturnRows(rows)

	repo := NewDashboardRepository(db)
	trend, err := repo.GetSalesTrend(shopID, "year")

	assert.NoError(t, err)
	assert.Equal(t, "year", trend.Period)
	assert.Len(t, trend.Data, 2)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDashboardRepository_GetSalesTrend_Default(t *testing.T) {
	db, mock, sqlDB := setupDashboardTestDB(t)
	defer sqlDB.Close()

	shopID := snowflake.ID(123)

	rows := sqlmock.NewRows([]string{"date", "orders", "revenue"}).
		AddRow("2026-02-20", 10, 1000.0)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT DATE_FORMAT(created_at")).
		WillReturnRows(rows)

	repo := NewDashboardRepository(db)
	trend, err := repo.GetSalesTrend(shopID, "invalid")

	assert.NoError(t, err)
	assert.Equal(t, "week", trend.Period)
	assert.NoError(t, mock.ExpectationsWereMet())
}
