package main

import (
	"OrderEase/config"
	"OrderEase/handlers"
	"OrderEase/models"
	"OrderEase/utils"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
)

// 日志中间件
func LoggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		utils.Logger.Printf("开始处理请求: %s %s", c.Request.Method, path)
		utils.Logger.Printf("请求头: %v", c.Request.Header)

		// 处理请求
		c.Next()

		// 计算延迟
		latency := time.Since(start)

		utils.Logger.Printf("请求处理完成: %s %s?%s 耗时: %v",
			c.Request.Method, path, raw, latency)
	}
}

func main() {
	// 初始化日志
	utils.InitLogger()
	utils.Logger.Println("服务启动...")

	// 加载配置文件
	if err := config.LoadConfig("config/config.yaml"); err != nil {
		utils.Logger.Fatal("加载配置文件失败:", err)
	}
	utils.Logger.Println("配置加载成功")

	// 设置 Gin 模式
	gin.SetMode(gin.ReleaseMode)

	// 创建 Gin 引擎
	r := gin.New()

	// 使用中间件
	r.Use(gin.Recovery())
	r.Use(LoggerMiddleware())

	// CORS 中间件
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusOK)
			return
		}

		c.Next()
	})

	// 连接数据库
	db, err := config.InitDB()
	if err != nil {
		utils.Logger.Fatal("数据库连接失败:", err)
	}
	utils.Logger.Println("数据库连接成功")

	// 数据库迁移
	tables := []interface{}{
		&models.Product{},
		&models.Order{},
		&models.OrderItem{},
		&models.OrderStatusLog{},
	}

	for _, table := range tables {
		if err := db.AutoMigrate(table); err != nil {
			utils.Logger.Fatalf("迁移表 %T 失败: %v", table, err)
		}
		utils.Logger.Printf("表 %T 迁移成功", table)
	}

	utils.Logger.Println("所有数据库表迁移完成")

	// 创建处理器
	h := handlers.NewHandler(db)

	// 创建路由组
	api := r.Group(config.AppConfig.Server.BasePath)
	{
		// 商品相关路由
		product := api.Group("/product")
		{
			product.POST("/create", h.CreateProduct)
			product.GET("/list", h.GetProducts)
			product.GET("/detail", h.GetProduct)
			product.PUT("/update", h.UpdateProduct)
			product.DELETE("/delete", h.DeleteProduct)
			product.POST("/upload-image", h.UploadProductImage)
			product.GET("/image", h.GetProductImage)
		}

		// 订单相关路由
		order := api.Group("/order")
		{
			order.POST("/create", h.CreateOrder)
			order.PUT("/update", h.UpdateOrder)
			order.GET("/list", h.GetOrders)
			order.GET("/detail", h.GetOrder)
			order.DELETE("/delete", h.DeleteOrder)
			order.PUT("/toggle-status", h.ToggleOrderStatus)
		}
	}

	// 静态文件服务
	r.Static("/uploads", "./uploads")

	// 确保上传目录存在
	if err := os.MkdirAll("./uploads/products", 0755); err != nil {
		utils.Logger.Fatal("创建上传目录失败:", err)
	}
	utils.Logger.Println("上传目录创建成功")

	// 启动服务器
	serverAddr := fmt.Sprintf("%s:%d", config.AppConfig.Server.Host, config.AppConfig.Server.Port)
	utils.Logger.Printf("服务器启动在 %s", serverAddr)
	utils.Logger.Fatal(r.Run(serverAddr))
}
