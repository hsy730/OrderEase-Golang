package main

import (
	"fmt"
	"log"
	"net/http"
	"orderease/config"
	"orderease/handlers"
	"orderease/routes"
	"orderease/utils/log2"
	"os"
	"time"

	"orderease/database"
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

		log2.Debugf("开始处理请求: %s %s", c.Request.Method, path)
		log2.Debugf("请求头: %v", c.Request.Header)

		// 处理请求
		c.Next()

		// 计算延迟
		latency := time.Since(start)

		log2.Debugf("请求处理完成: %s %s?%s 耗时: %v",
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
	log2.InitLogger()
	log2.Info("服务启动...")

	// 加载配置文件
	if err := config.LoadConfig("config/config.yaml"); err != nil {
		log2.Fatal("加载配置文件失败:", err)
	}
	log2.Info("配置加载成功")

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
	log2.Info("数据库连接成功")

	// 创建处理器
	h := handlers.NewHandler(db)

	// 设置路由
	routes.SetupRoutes(r, h)

	// 静态文件服务
	r.Static("/uploads", "./uploads")

	// 确保上传目录存在
	if err := os.MkdirAll("./uploads/products", 0755); err != nil {
		log2.Fatal("创建上传目录失败:", err)
	}
	log2.Info("上传目录创建成功")

	// 初始化清理任务
	cleanupTask := tasks.NewCleanupTask(db)
	cleanupTask.StartCleanupTask()

	// 启动服务器
	serverAddr := fmt.Sprintf("%s:%d", config.AppConfig.Server.Host, config.AppConfig.Server.Port)
	log2.Info("服务器启动在 %s", serverAddr)
	log2.Fatal(r.Run(serverAddr))
}
