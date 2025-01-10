package main

import (
	"fmt"
	"log"
	"net/http"
	"orderease/config"
	"orderease/handlers"
	"orderease/utils"
	"os"
	"time"

	"orderease/database"
	"orderease/routes"
	"orderease/tasks"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
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

func init() {
	// 设置配置文件路径
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./config")

	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("读取配置文件失败: %v", err)
	}

	// 验证必要的配置项
	if viper.GetString("jwt.secret") == "" {
		log.Fatal("JWT secret 未配置")
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
	db, err := database.Init()
	if err != nil {
		log.Fatalf("数据库初始化失败: %v", err)
	}
	utils.Logger.Println("数据库连接成功")

	// 创建处理器
	h := handlers.NewHandler(db)

	// 设置路由
	routes.SetupBackedRoutes(r, h)
	routes.SetupFrontRoutes(r, h)

	// 静态文件服务
	r.Static("/uploads", "./uploads")

	// 确保上传目录存在
	if err := os.MkdirAll("./uploads/products", 0755); err != nil {
		utils.Logger.Fatal("创建上传目录失败:", err)
	}
	utils.Logger.Println("上传目录创建成功")

	// 初始化清理任务
	cleanupTask := tasks.NewCleanupTask(db)
	cleanupTask.StartCleanupTask()

	// 启动服务器
	serverAddr := fmt.Sprintf("%s:%d", config.AppConfig.Server.Host, config.AppConfig.Server.Port)
	utils.Logger.Printf("服务器启动在 %s", serverAddr)
	utils.Logger.Fatal(r.Run(serverAddr))
}
