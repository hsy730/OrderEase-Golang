package main

import (
	"fmt"
	"log"
	"net/http"
	"orderease/providers"
	"orderease/routes"
	"orderease/utils/log2"
	"os"
	"strings"
	"time"

	"github.com/gin-contrib/gzip"
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
	// 使用依赖注入容器初始化应用
	container, err := providers.InitializeApp("config/config.yaml")
	if err != nil {
		log.Fatal("初始化应用失败:", err)
	}
	log2.Info("应用初始化成功")

	// 设置 Gin 模式
	gin.SetMode(gin.ReleaseMode)

	// 创建 Gin 引擎
	r := gin.New()

	// 使用中间件
	r.Use(gin.Recovery())
	r.Use(LoggerMiddleware())
	// Gzip 压缩中间件 - 减少传输体积
	r.Use(gzip.Gzip(gzip.DefaultCompression))

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

	// 设置路由
	routes.SetupRoutes(r, container.Handler)

	// 静态文件服务
	r.Static("/uploads", "./uploads")

	// 前后台UI静态文件服务
	r.Static("/order-ease-iui", "./static/order-ease-iui")
	r.Static("/order-ease-adminiui", "./static/order-ease-adminiui")

	// 根路径重定向到前台UI
	r.GET("/", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "/order-ease-iui/")
	})

	// SPA fallback - 处理前端路由
	r.NoRoute(func(c *gin.Context) {
		path := c.Request.URL.Path
		if strings.HasPrefix(path, "/order-ease-iui/") && !strings.Contains(path, ".") {
			c.File("./static/order-ease-iui/index.html")
			return
		}
		if strings.HasPrefix(path, "/order-ease-adminiui/") && !strings.Contains(path, ".") {
			c.File("./static/order-ease-adminiui/index.html")
			return
		}
		c.JSON(http.StatusNotFound, gin.H{"error": "Not Found"})
	})

	// 确保上传目录存在
	if err := os.MkdirAll("./uploads/products", 0755); err != nil {
		log2.Fatal("创建上传目录失败:", err)
	}
	log2.Info("上传目录创建成功")

	// 启动后台任务
	container.CleanupTask.StartCleanupTask()
	container.TempTokenService.SetupCronJob()
	log2.Info("后台任务已启动")

	// 启动服务器
	serverAddr := fmt.Sprintf("%s:%d", container.Config.Server.Host, container.Config.Server.Port)
	log2.Info("服务器启动在 %s", serverAddr)
	log2.Fatal(r.Run(serverAddr))
}
