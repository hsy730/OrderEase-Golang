package routes

import (
	"orderease/handlers"

	"github.com/gin-gonic/gin"
)

// SetupRoutes 配置所有路由
func SetupRoutes(r *gin.Engine, h *handlers.Handler) {
	// 创建路由组
	api := r.Group("/api/v1")
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

		// 数据管理相关路由
		data := api.Group("/data")
		{
			data.GET("/export", h.ExportData)
			data.POST("/import", h.ImportData)
		}

		// 用户相关路由
		user := api.Group("/user")
		{
			user.POST("/create", h.CreateUser)
			user.GET("/list", h.GetUsers)
			user.GET("/simple-list", h.GetUserSimpleList)
			user.GET("/detail", h.GetUser)
			user.PUT("/update", h.UpdateUser)
			user.DELETE("/delete", h.DeleteUser)
		}

		// 管理员相关路由
		admin := api.Group("/admin")
		{
			admin.POST("/login", h.AdminLogin)
			admin.PUT("/change-password", h.ChangeAdminPassword)
		}
	}
}
