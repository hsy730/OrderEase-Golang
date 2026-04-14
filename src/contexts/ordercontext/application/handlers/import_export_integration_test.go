package handlers

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	vsql "github.com/dolthub/vitess/go/mysql"
	"github.com/bwmarrin/snowflake"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	gormmysql "gorm.io/driver/mysql"
	"gorm.io/gorm"
	"orderease/models"
	"orderease/utils/log2"

	sqle "github.com/dolthub/go-mysql-server"
	"github.com/dolthub/go-mysql-server/memory"
	"github.com/dolthub/go-mysql-server/server"
	gsql "github.com/dolthub/go-mysql-server/sql"

	services "orderease/contexts/ordercontext/application/services"
)

// TestImportExportIntegration 使用 go-mysql-server 内存 MySQL 服务器进行真实的导入导出集成测试
func TestImportExportIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过集成测试")
	}

	gin.SetMode(gin.TestMode)

	db, srv, cleanup := setupTestMySQLServer(t)
	defer cleanup()

	err := autoMigrateAllTables(db)
	require.NoError(t, err, "自动迁移表结构失败")

	testData := createComprehensiveTestData(t)
	insertTestData(t, db, testData)

	t.Run("完整导出流程测试", func(t *testing.T) {
		testExportFlow(t, db)
	})

	t.Run("完整导入流程测试", func(t *testing.T) {
		testImportFlow(t, db)
	})

	t.Run("特殊类型数据处理测试", func(t *testing.T) {
		testSpecialDataTypes(t, db)
	})

	t.Run("循环导入导出一致性测试", func(t *testing.T) {
		testRoundTripConsistency(t, db)
	})

	t.Cleanup(func() {
		if srv != nil {
			srv.Close()
		}
	})
}

// setupTestMySQLServer 创建基于 go-mysql-server 的内存 MySQL 服务器
func setupTestMySQLServer(t *testing.T) (*gorm.DB, *server.Server, func()) {
	t.Helper()

	dbName := "test_db"

	memDB := memory.NewDatabase(dbName)
	pro := memory.NewDBProvider(memDB)
	engine := sqle.NewDefault(pro)

	cfg := server.Config{
		Protocol: "tcp",
		Address:  "127.0.0.1:0",
	}

	sessionBuilder := func(ctx context.Context, c *vsql.Conn, addr string) (gsql.Session, error) {
		host := ""
		user := ""
		mysqlConnectionUser, ok := c.UserData.(gsql.MysqlConnectionUser)
		if ok {
			host = mysqlConnectionUser.Host
			user = mysqlConnectionUser.User
		}
		client := gsql.Client{Address: host, User: user, Capabilities: c.Capabilities}
		return memory.NewSession(gsql.NewBaseSessionWithClientServer(addr, client, c.ConnectionID), pro), nil
	}

	srv, err := server.NewServer(cfg, engine, gsql.NewContext, sessionBuilder, nil)
	require.NoError(t, err, "创建 MySQL 服务器失败")

	go func() {
		if err := srv.Start(); err != nil && !strings.Contains(err.Error(), "use of closed network connection") {
			log2.Errorf("MySQL 服务器错误: %v", err)
		}
	}()

	time.Sleep(200 * time.Millisecond)

	addr := srv.Listener.Addr()
	require.NotEmpty(t, addr, "服务器地址不应该为空")
	t.Logf("✓ go-mysql-server 成功启动在: %s", addr)

	dsn := fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		"root",
		"",
		addr.String(),
		dbName,
	)

	var gormDB *gorm.DB
	var connErr error
	for i := 0; i < 10; i++ {
		gormDB, connErr = gorm.Open(gormmysql.Open(dsn), &gorm.Config{
			SkipDefaultTransaction: true,
			PrepareStmt:            true,
			DisableForeignKeyConstraintWhenMigrating: true,
		})
		if connErr == nil {
			break
		}
		time.Sleep(50 * time.Millisecond)
	}
	require.NoError(t, connErr, "连接到测试数据库失败")

	sqlDB, err := gormDB.DB()
	require.NoError(t, err)

	cleanup := func() {
		sqlDB.Close()
	}

	return gormDB, srv, cleanup
}

// autoMigrateAllTables 自动迁移所有需要的表
func autoMigrateAllTables(db *gorm.DB) error {
	tables := []interface{}{
		&models.User{},
		&models.Product{},
		&models.Order{},
		&models.OrderItem{},
		&models.OrderItemOption{},
		&models.Tag{},
		&models.ProductTag{},
		&models.ProductOption{},
		&models.ProductOptionCategory{},
		&models.Shop{},
		&models.Admin{},
		&models.UserThirdpartyBinding{},
	}

	for _, table := range tables {
		if err := db.AutoMigrate(table); err != nil {
			errMsg := err.Error()
			if strings.Contains(errMsg, "already exists") || strings.Contains(errMsg, "duplicate") {
				continue
			}
			return fmt.Errorf("迁移表 %T 失败: %w", table, err)
		}
	}

	m := db.Migrator()

	tableNames := []string{
		"users", "products", "orders", "order_items", "order_item_options",
		"tags", "product_tags", "product_options", "product_option_categories",
		"shops", "admins", "user_thirdparty_bindings",
	}

	for _, tableName := range tableNames {
		if !m.HasTable(tableName) {
			return fmt.Errorf("表 %s 未创建成功", tableName)
		}
	}

	return nil
}

type TestData struct {
	Users                 []models.User
	Products              []models.Product
	Orders                []models.Order
	OrderItems            []models.OrderItem
	OrderItemOptions      []models.OrderItemOption
	Tags                  []models.Tag
	ProductTags           []models.ProductTag
	ProductOptions        []models.ProductOption
	ProductOptionCategories []models.ProductOptionCategory
	Shops                 []models.Shop
	Admins                []models.Admin
	UserBindings          []models.UserThirdpartyBinding
}

// createComprehensiveTestData 创建包含各种特殊类型的综合测试数据
func createComprehensiveTestData(t *testing.T) TestData {
	t.Helper()

	now := time.Now().UTC().Truncate(time.Second)

	shopID := snowflake.ID(1001)
	userID1 := snowflake.ID(2001)
	userID2 := snowflake.ID(2002)
	productID1 := snowflake.ID(3001)
	productID2 := snowflake.ID(3002)
	orderID := snowflake.ID(4001)
	orderItemID := snowflake.ID(5001)
	tagID1 := 1
	tagID2 := 2
	categoryID := snowflake.ID(6001)
	optionID1 := snowflake.ID(7001)
	optionID2 := snowflake.ID(7002)
	bindingID := uint(1)
	shopIDForTag := models.FromSnowflakeID(shopID)

	defaultOrderStatusFlow := models.OrderStatusFlow{
		Statuses: []models.OrderStatus{
			{
				Value:   0,
				Label:   "待处理",
				Type:    "warning",
				IsFinal: false,
				Actions: []models.OrderStatusAction{
					{Name: "接单", NextStatus: 1, NextStatusLabel: "已接单"},
					{Name: "取消", NextStatus: 10, NextStatusLabel: "已取消"},
				},
			},
			{
				Value:   1,
				Label:   "已接单",
				Type:    "primary",
				IsFinal: false,
				Actions: []models.OrderStatusAction{
					{Name: "完成", NextStatus: 9, NextStatusLabel: "已完成"},
					{Name: "取消", NextStatus: 10, NextStatusLabel: "已取消"},
				},
			},
			{
				Value:   9,
				Label:   "已完成",
				Type:    "success",
				IsFinal: true,
				Actions: []models.OrderStatusAction{},
			},
			{
				Value:   10,
				Label:   "已取消",
				Type:    "info",
				IsFinal: true,
				Actions: []models.OrderStatusAction{},
			},
		},
	}

	metadata := models.Metadata{
		"access_token":  "test_access_token_12345",
		"refresh_token": "test_refresh_token_67890",
		"openid":        "test_openid",
		"unionid":       "test_unionid",
	}

	lastLoginAt := now.Add(-24 * time.Hour)

	return TestData{
		Shops: []models.Shop{
			{
				ID:              shopID,
				Name:            "测试店铺",
				OwnerUsername:   "shop_owner",
				OwnerPassword:   "hashed_password_123",
				ContactPhone:    "13800138000",
				ContactEmail:    "shop@test.com",
				Address:         "北京市海淀区测试街道123号",
				ImageURL:        "https://example.com/shop.jpg",
				Description:     "这是一个用于测试的店铺描述，包含各种字符：中文、English、数字123、特殊符号@#$%",
				CreatedAt:       now.Add(-30 * 24 * time.Hour),
				UpdatedAt:       now,
				ValidUntil:      now.Add(365 * 24 * time.Hour),
				Settings:        json.RawMessage(`{"theme":"dark","language":"zh-CN","notifications":true}`),
				OrderStatusFlow: defaultOrderStatusFlow,
			},
		},

		Users: []models.User{
			{
				ID:        userID1,
				Name:      "张三",
				Role:      "private_user",
				Password:  "hashed_password_user1",
				Phone:     "13900139001",
				Address:   "上海市浦东新区测试路456号",
				Type:      "delivery",
				Nickname:  "小明同学",
				Avatar:    "https://example.com/avatar1.jpg",
				CreatedAt: now.Add(-20 * 24 * time.Hour),
				UpdatedAt: now,
			},
			{
				ID:        userID2,
				Name:      "李四",
				Role:      "public_user",
				Password:  "hashed_password_user2",
				Phone:     "13800138002",
				Address:   "",
				Type:      "pickup",
				Nickname:  "",
				Avatar:    "",
				CreatedAt: now.Add(-15 * 24 * time.Hour),
				UpdatedAt: now,
			},
		},

		Admins: []models.Admin{
			{
				ID:        1,
				Username:  "admin",
				Password:  "$2a$10$hashed_admin_password",
				CreatedAt: now.Add(-60 * 24 * time.Hour),
				UpdatedAt: now,
			},
		},

		Tags: []models.Tag{
			{
				ID:          tagID1,
				ShopID:      shopIDForTag,
				Name:        "热销",
				Description: "热门销售商品标签",
				CreatedAt:   now.Add(-25 * 24 * time.Hour),
				UpdatedAt:   now,
			},
			{
				ID:          tagID2,
				ShopID:      shopIDForTag,
				Name:        "新品",
				Description: "新上架商品标签",
				CreatedAt:   now.Add(-10 * 24 * time.Hour),
				UpdatedAt:   now,
			},
		},

		Products: []models.Product{
			{
				ID:          productID1,
				ShopID:      shopID,
				Name:        "经典拿铁咖啡 ☕",
				Description: "精选阿拉比卡豆，口感醇厚，奶泡绵密。支持多种糖度选择。\n包含特殊字符：<>&\"'和换行",
				Price:       28.50,
				Stock:       100,
				ImageURL:    "https://example.com/coffee.jpg",
				Status:      "online",
				CreatedAt:   now.Add(-15 * 24 * time.Hour),
				UpdatedAt:   now,
			},
			{
				ID:          productID2,
				ShopID:      shopID,
				Name:        "草莓蛋糕 🍰",
				Description: "新鲜草莓制作，甜而不腻",
				Price:       38.00,
				Stock:       50,
				ImageURL:    "https://example.com/cake.jpg",
				Status:      "online",
				CreatedAt:   now.Add(-10 * 24 * time.Hour),
				UpdatedAt:   now,
			},
		},

		ProductOptionCategories: []models.ProductOptionCategory{
			{
				ID:           categoryID,
				ProductID:    productID1,
				Name:         "杯型大小",
				IsRequired:   true,
				IsMultiple:   false,
				DisplayOrder: 1,
				CreatedAt:    now.Add(-14 * 24 * time.Hour),
				UpdatedAt:    now,
			},
		},

		ProductOptions: []models.ProductOption{
			{
				ID:              optionID1,
				CategoryID:      categoryID,
				Name:            "小杯 (12oz)",
				PriceAdjustment: 0.0,
				DisplayOrder:    1,
				IsDefault:       true,
				CreatedAt:       now.Add(-14 * 24 * time.Hour),
				UpdatedAt:       now,
			},
			{
				ID:              optionID2,
				CategoryID:      categoryID,
				Name:            "大杯 (16oz)",
				PriceAdjustment: 5.00,
				DisplayOrder:    2,
				IsDefault:       false,
				CreatedAt:       now.Add(-14 * 24 * time.Hour),
				UpdatedAt:       now,
			},
		},

		ProductTags: []models.ProductTag{
			{ProductID: productID1, TagID: tagID1, ShopID: shopID, CreatedAt: now, UpdatedAt: now},
			{ProductID: productID2, TagID: tagID2, ShopID: shopID, CreatedAt: now, UpdatedAt: now},
		},

		Orders: []models.Order{
			{
				ID:         orderID,
				UserID:     userID1,
				ShopID:     shopID,
				TotalPrice: models.Price(33.50),
				Status:     1,
				Remark:     "请少加糖，谢谢！🙏\n多行备注\n特殊字符：<>\"'&",
				CreatedAt:  now.Add(-5 * 24 * time.Hour),
				UpdatedAt:  now,
			},
		},

		OrderItems: []models.OrderItem{
			{
				ID:                 orderItemID,
				OrderID:            orderID,
				ProductID:          productID1,
				Quantity:           1,
				Price:              models.Price(28.50),
				TotalPrice:         models.Price(33.50),
				ProductName:        "经典拿铁咖啡 ☕",
				ProductDescription: "精选阿拉比卡豆，口感醇厚",
				ProductImageURL:    "https://example.com/coffee.jpg",
			},
		},

		OrderItemOptions: []models.OrderItemOption{
			{
				ID:              snowflake.ID(8001),
				OrderItemID:     orderItemID,
				CategoryID:      categoryID,
				OptionID:        optionID2,
				OptionName:      "大杯 (16oz)",
				CategoryName:    "杯型大小",
				PriceAdjustment: 5.00,
				CreatedAt:       now.Add(-5 * 24 * time.Hour),
				UpdatedAt:       now,
			},
		},

		UserBindings: []models.UserThirdpartyBinding{
			{
				ID:             bindingID,
				UserID:         userID1,
				Provider:       "wechat",
				ProviderUserID: "wx_openid_test_12345",
				UnionID:        "wx_unionid_test_67890",
				Nickname:       "微信用户昵称 🎉",
				AvatarURL:      "https://thirdparty.example.com/avatar.jpg",
				Gender:         1,
				Country:        "中国",
				Province:       "上海",
				City:           "浦东新区",
				Metadata:       metadata,
				IsActive:       true,
				LastLoginAt:    &lastLoginAt,
				CreatedAt:      now.Add(-7 * 24 * time.Hour),
				UpdatedAt:      now,
			},
		},
	}
}

// insertTestData 将测试数据插入数据库
func insertTestData(t *testing.T, db *gorm.DB, data TestData) {
	t.Helper()

	assert.NoError(t, db.Create(&data.Shops).Error, "插入 Shops 失败")
	assert.NoError(t, db.Create(&data.Users).Error, "插入 Users 失败")
	assert.NoError(t, db.Create(&data.Admins).Error, "插入 Admins 失败")
	assert.NoError(t, db.Create(&data.Tags).Error, "插入 Tags 失败")
	assert.NoError(t, db.Create(&data.Products).Error, "插入 Products 失败")
	assert.NoError(t, db.Create(&data.ProductOptionCategories).Error, "插入 ProductOptionCategories 失败")
	assert.NoError(t, db.Create(&data.ProductOptions).Error, "插入 ProductOptions 失败")
	assert.NoError(t, db.Create(&data.ProductTags).Error, "插入 ProductTags 失败")
	assert.NoError(t, db.Create(&data.Orders).Error, "插入 Orders 失败")
	assert.NoError(t, db.Create(&data.OrderItems).Error, "插入 OrderItems 失败")
	assert.NoError(t, db.Create(&data.OrderItemOptions).Error, "插入 OrderItemOptions 失败")
	assert.NoError(t, db.Create(&data.UserBindings).Error, "插入 UserBindings 失败")
}

// testExportFlow 测试完整的导出流程
func testExportFlow(t *testing.T, db *gorm.DB) {
	t.Helper()

	exportService := services.NewExportService(db)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/export", nil)

	err := exportService.ExportAllData(c)
	assert.NoError(t, err, "导出数据失败")

	assert.Equal(t, http.StatusOK, w.Code, "HTTP 状态码应该是 200")
	assert.Contains(t, w.Header().Get("Content-Type"), "application/zip", "Content-Type 应该是 application/zip")

	contentDisposition := w.Header().Get("Content-Disposition")
	assert.True(t, strings.HasPrefix(contentDisposition, "attachment; filename="), "应该有 Content-Disposition header")
	assert.True(t, strings.HasSuffix(contentDisposition, ".zip"), "文件名应该以 .zip 结尾")

	body := w.Body.Bytes()
	assert.Greater(t, len(body), 0, "导出的 ZIP 文件不应该为空")

	reader, err := zip.NewReader(bytes.NewReader(body), int64(len(body)))
	assert.NoError(t, err, "读取 ZIP 文件失败")

	expectedFiles := map[string]bool{
		"admins.csv":                   false,
		"product_option_categories.csv": false,
		"product_options.csv":          false,
		"shops.csv":                    false,
		"users.csv":                    false,
		"products.csv":                 false,
		"tags.csv":                     false,
		"product_tags.csv":             false,
		"orders.csv":                   false,
		"order_items.csv":              false,
		"order_item_options.csv":       false,
		"user_thirdparty_bindings.csv": false,
	}

	for _, file := range reader.File {
		if _, ok := expectedFiles[file.Name]; ok {
			expectedFiles[file.Name] = true
			t.Logf("✓ 找到文件: %s (大小: %d bytes)", file.Name, file.UncompressedSize64)
		}
	}

	for fileName, found := range expectedFiles {
		assert.True(t, found, "ZIP 中缺少文件: %s", fileName)
	}

	t.Logf("导出完成，共包含 %d 个 CSV 文件", countTrue(expectedFiles))
}

// testImportFlow 测试完整的导入流程
func testImportFlow(t *testing.T, db *gorm.DB) {
	t.Helper()

	exportService := services.NewExportService(db)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/export", nil)

	err := exportService.ExportAllData(c)
	require.NoError(t, err, "导出数据失败")

	zipContent := w.Body.Bytes()

	tempDir := t.TempDir()
	zipPath := filepath.Join(tempDir, "export.zip")

	err = os.WriteFile(zipPath, zipContent, 0644)
	require.NoError(t, err, "写入临时 ZIP 文件失败")

	clearAllTables(t, db)

	handler := createTestHandler(db, t)

	router := gin.New()
	router.POST("/api/import", handler.ImportData)

	reqBody, contentType := createMultipartRequest(t, zipPath)
	req := httptest.NewRequest(http.MethodPost, "/api/import", reqBody)
	req.Header.Set("Content-Type", contentType)

	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req)

	assert.Equal(t, http.StatusOK, w2.Code, "导入应该成功，HTTP 状态码: %d, 响应: %s", w2.Code, w2.Body.String())

	verifyImportedData(t, db)
}

// testSpecialDataTypes 测试特殊数据类型的处理
func testSpecialDataTypes(t *testing.T, db *gorm.DB) {
	t.Helper()

	t.Run("OrderStatusFlow 类型", func(t *testing.T) {
		var shops []models.Shop
		err := db.Find(&shops).Error
		assert.NoError(t, err)
		assert.Len(t, shops, 1, "应该有 1 个店铺")

		shop := shops[0]
		assert.NotEmpty(t, shop.OrderStatusFlow.Statuses, "OrderStatusFlow 不应该为空")
		assert.Equal(t, 4, len(shop.OrderStatusFlow.Statuses), "应该有 4 个状态配置")
		assert.Equal(t, "待处理", shop.OrderStatusFlow.Statuses[0].Label, "第一个状态应该是 '待处理'")
		assert.Equal(t, 2, len(shop.OrderStatusFlow.Statuses[0].Actions), "第一个状态应该有 2 个动作")
	})

	t.Run("Metadata 类型", func(t *testing.T) {
		var bindings []models.UserThirdpartyBinding
		err := db.Find(&bindings).Error
		assert.NoError(t, err)
		assert.Len(t, bindings, 1, "应该有 1 条绑定记录")

		binding := bindings[0]
		assert.NotNil(t, binding.Metadata, "Metadata 不应该为 nil")
		assert.Equal(t, "test_access_token_12345", binding.Metadata.GetAccessToken(), "Access Token 应该正确")
		assert.Equal(t, "test_refresh_token_67890", binding.Metadata.GetRefreshToken(), "Refresh Token 应该正确")
	})

	t.Run("*time.Time 指针类型", func(t *testing.T) {
		var bindings []models.UserThirdpartyBinding
		err := db.Find(&bindings).Error
		assert.NoError(t, err)

		binding := bindings[0]
		assert.NotNil(t, binding.LastLoginAt, "LastLoginAt 不应该为 nil")
		assert.False(t, binding.LastLoginAt.IsZero(), "LastLoginAt 不应该是零值")
	})

	t.Run("Price 自定义类型", func(t *testing.T) {
		var orders []models.Order
		err := db.Find(&orders).Error
		assert.NoError(t, err)
		assert.Len(t, orders, 1, "应该有 1 个订单")

		order := orders[0]
		assert.Equal(t, models.Price(33.50), order.TotalPrice, "TotalPrice 应该是 33.50")
	})

	t.Run("JSON RawMessage 类型", func(t *testing.T) {
		var shops []models.Shop
		err := db.Find(&shops).Error
		assert.NoError(t, err)

		shop := shops[0]
		assert.NotEmpty(t, shop.Settings, "Settings 不应该为空")

		var settings map[string]interface{}
		err = json.Unmarshal(shop.Settings, &settings)
		assert.NoError(t, err, "解析 Settings JSON 失败")
		assert.Equal(t, "dark", settings["theme"], "主题应该是 dark")
	})

	t.Run("特殊字符和多行文本", func(t *testing.T) {
		var orders []models.Order
		err := db.Find(&orders).Error
		assert.NoError(t, err)

		order := orders[0]
		assert.Contains(t, order.Remark, "请少加糖", "备注应该包含指定文本")
		assert.Contains(t, order.Remark, "\n", "备注应该包含换行符")
		assert.Contains(t, order.Remark, "🙏", "备注应该包含 emoji")
	})

	t.Run("雪花 ID 类型", func(t *testing.T) {
		var users []models.User
		err := db.Find(&users).Error
		assert.NoError(t, err)
		assert.Len(t, users, 2, "应该有 2 个用户")

		for _, user := range users {
			assert.NotEqual(t, snowflake.ID(0), user.ID, "用户 ID 不应该是 0")
			assert.Greater(t, int64(user.ID), int64(0), "用户 ID 应该大于 0")
		}
	})
}

// testRoundTripConsistency 测试导入导出的数据一致性（往返测试）
func testRoundTripConsistency(t *testing.T, db *gorm.DB) {
	t.Helper()

	exportService := services.NewExportService(db)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/export", nil)

	err := exportService.ExportAllData(c)
	require.NoError(t, err, "第一次导出失败")

	firstExport := w.Body.Bytes()

	tempDir := t.TempDir()
	zipPath := filepath.Join(tempDir, "export.zip")
	os.WriteFile(zipPath, firstExport, 0644)

	clearAllTables(t, db)

	handler := createTestHandler(db, t)

	router := gin.New()
	router.POST("/api/import", handler.ImportData)

	reqBody, contentType := createMultipartRequest(t, zipPath)
	req := httptest.NewRequest(http.MethodPost, "/api/import", reqBody)
	req.Header.Set("Content-Type", contentType)

	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req)
	assert.Equal(t, http.StatusOK, w2.Code, "第一次导入失败")

	w3 := httptest.NewRecorder()
	c2, _ := gin.CreateTestContext(w3)
	c2.Request = httptest.NewRequest(http.MethodGet, "/api/export", nil)

	err = exportService.ExportAllData(c2)
	require.NoError(t, err, "第二次导出失败")

	secondExport := w3.Body.Bytes()

	reader1, err := zip.NewReader(bytes.NewReader(firstExport), int64(len(firstExport)))
	require.NoError(t, err)

	reader2, err := zip.NewReader(bytes.NewReader(secondExport), int64(len(secondExport)))
	require.NoError(t, err)

	fileMap1 := make(map[string]*zip.File)
	for _, f := range reader1.File {
		if !strings.HasPrefix(f.Name, "uploads/") {
			fileMap1[f.Name] = f
		}
	}

	fileMap2 := make(map[string]*zip.File)
	for _, f := range reader2.File {
		if !strings.HasPrefix(f.Name, "uploads/") {
			fileMap2[f.Name] = f
		}
	}

	assert.Equal(t, len(fileMap1), len(fileMap2), "两次导出的文件数量应该相同")

	for name, file1 := range fileMap1 {
		file2, ok := fileMap2[name]
		assert.True(t, ok, "第二次导出缺少文件: %s", name)

		if ok && !strings.HasSuffix(name, ".csv") {
			continue
		}

		rc1, err := file1.Open()
		require.NoError(t, err)
		content1, _ := io.ReadAll(rc1)
		rc1.Close()

		rc2, err := file2.Open()
		require.NoError(t, err)
		content2, _ := io.ReadAll(rc2)
		rc2.Close()

		assert.Equal(t, string(content1), string(content2), "文件 %s 内容不一致", name)
		t.Logf("✓ 文件 %s 一致性验证通过 (%d bytes)", name, len(content1))
	}

	t.Logf("✓ 往返测试通过：导入导出数据完全一致")
}

// clearAllTables 清空所有表的数据
func clearAllTables(t *testing.T, db *gorm.DB) {
	t.Helper()

	tablesToClean := []string{
		"order_item_options",
		"order_items",
		"orders",
		"product_tags",
		"tags",
		"products",
		"user_thirdparty_bindings",
		"users",
		"admins",
		"product_options",
		"product_option_categories",
		"shops",
	}

	for _, tableName := range tablesToClean {
		err := db.Exec(fmt.Sprintf("DELETE FROM %s", tableName)).Error
		if err != nil {
			t.Logf("⚠️  清空表 %s 失败: %v (尝试继续)", tableName, err)
		} else {
			t.Logf("✓ 清空表: %s", tableName)
		}
	}

	for _, tableName := range tablesToClean {
		var count int64
		db.Table(tableName).Count(&count)
		if count > 0 {
			t.Logf("⚠️  表 %s 仍有 %d 条记录", tableName, count)
		}
	}
}

// verifyImportedData 验证导入后的数据完整性
func verifyImportedData(t *testing.T, db *gorm.DB) {
	t.Helper()

	var userCount int64
	db.Model(&models.User{}).Count(&userCount)
	assert.Equal(t, int64(2), userCount, "用户数量不匹配")

	var shopCount int64
	db.Model(&models.Shop{}).Count(&shopCount)
	assert.Equal(t, int64(1), shopCount, "店铺数量不匹配")

	var productCount int64
	db.Model(&models.Product{}).Count(&productCount)
	assert.Equal(t, int64(2), productCount, "商品数量不匹配")

	var orderCount int64
	db.Model(&models.Order{}).Count(&orderCount)
	assert.Equal(t, int64(1), orderCount, "订单数量不匹配")

	var tagCount int64
	db.Model(&models.Tag{}).Count(&tagCount)
	assert.Equal(t, int64(2), tagCount, "标签数量不匹配")

	var bindingCount int64
	db.Model(&models.UserThirdpartyBinding{}).Count(&bindingCount)
	assert.Equal(t, int64(1), bindingCount, "第三方绑定数量不匹配")

	t.Logf("✅ 数据验证通过 - Users: %d, Shops: %d, Products: %d, Orders: %d, Tags: %d, Bindings: %d",
		userCount, shopCount, productCount, orderCount, tagCount, bindingCount)
}

// createTestHandler 创建测试用的 Handler 实例
func createTestHandler(db *gorm.DB, t *testing.T) *Handler {
	t.Helper()

	logger := log2.GetLogger()

	handler := &Handler{
		DB:            db,
		logger:        logger,
		exportService: services.NewExportService(db),
	}

	return handler
}

// createMultipartRequest 创建 multipart 表单请求
func createMultipartRequest(t *testing.T, filePath string) (*bytes.Buffer, string) {
	t.Helper()

	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	file, err := os.Open(filePath)
	require.NoError(t, err, "打开文件失败")
	defer file.Close()

	part, err := writer.CreateFormFile("file", filepath.Base(filePath))
	require.NoError(t, err, "创建表单字段失败")

	_, err = io.Copy(part, file)
	require.NoError(t, err, "复制文件内容失败")

	err = writer.Close()
	require.NoError(t, err, "关闭 multipart writer 失败")

	return &buf, writer.FormDataContentType()
}

// 辅助函数
func countTrue(m map[string]bool) int {
	count := 0
	for _, v := range m {
		if v {
			count++
		}
	}
	return count
}
