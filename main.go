package main

import (
	"OrderEase/config"
	"OrderEase/handlers"
	"OrderEase/models"
	"OrderEase/utils"
	"fmt"
	"net/http"
	"os"
	"runtime/debug"
	"time"
)

// 添加日志中间件
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// 记录请求信息
		utils.Logger.Printf("开始处理请求: %s %s", r.Method, r.URL.Path)
		utils.Logger.Printf("请求头: %v", r.Header)

		// 捕获panic
		defer func() {
			if err := recover(); err != nil {
				utils.Logger.Printf("请求处理panic: %v\n堆栈: %s", err, debug.Stack())
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
		}()

		next.ServeHTTP(w, r)

		// 记录响应时间
		utils.Logger.Printf("请求处理完成: %s %s 耗时: %v",
			r.Method, r.URL.Path, time.Since(start))
	})
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

	h := &handlers.Handler{DB: db}

	// 创建路由基础路径
	basePath := config.AppConfig.Server.BasePath

	// 添加CORS中间件
	corsMiddleware := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// 设置CORS头
			w.Header().Set("Access-Control-Allow-Origin", "*") // 在生产环境中应该设置为具体的域名
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

			// 处理预检请求
			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}

			next.ServeHTTP(w, r)
		})
	}

	// 创建路由器
	mux := http.NewServeMux()

	// 注册路由
	// 商品相关路由
	mux.HandleFunc(fmt.Sprintf("%s/product/create", basePath), h.CreateProduct)
	mux.HandleFunc(fmt.Sprintf("%s/product/list", basePath), h.GetProducts)
	mux.HandleFunc(fmt.Sprintf("%s/product/detail", basePath), h.GetProduct)
	mux.HandleFunc(fmt.Sprintf("%s/product/update", basePath), h.UpdateProduct)
	mux.HandleFunc(fmt.Sprintf("%s/product/delete", basePath), h.DeleteProduct)
	mux.HandleFunc(fmt.Sprintf("%s/product/upload-image", basePath), h.UploadProductImage)
	mux.HandleFunc(fmt.Sprintf("%s/product/image", basePath), h.GetProductImage)

	// 订单相关路由
	mux.HandleFunc(fmt.Sprintf("%s/order/create", basePath), h.CreateOrder)
	mux.HandleFunc(fmt.Sprintf("%s/order/update", basePath), h.UpdateOrder)
	mux.HandleFunc(fmt.Sprintf("%s/order/list", basePath), h.GetOrders)
	mux.HandleFunc(fmt.Sprintf("%s/order/detail", basePath), h.GetOrder)
	mux.HandleFunc(fmt.Sprintf("%s/order/delete", basePath), h.DeleteOrder)
	mux.HandleFunc(fmt.Sprintf("%s/order/toggle-status", basePath), h.ToggleOrderStatus)

	// 添加静态文件服务
	mux.Handle("/uploads/", http.StripPrefix("/uploads/", http.FileServer(http.Dir("./uploads"))))

	// 确保上传目录存在
	if err := os.MkdirAll("./uploads/products", 0755); err != nil {
		utils.Logger.Fatal("创建上传目录失败:", err)
	}
	utils.Logger.Println("上传目录创建成功")

	// 添加中间件链
	handler := loggingMiddleware(corsMiddleware(mux))

	// 启动服务器
	serverAddr := fmt.Sprintf("%s:%d", config.AppConfig.Server.Host, config.AppConfig.Server.Port)
	utils.Logger.Printf("服务器启动在 %s", serverAddr)
	utils.Logger.Fatal(http.ListenAndServe(serverAddr, handler))
}
