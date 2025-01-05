package routes

import (
	"orderease/handlers"
	"orderease/middleware"

	"github.com/gin-gonic/gin"
)

// SetupRoutes 配置所有路由
func SetupRoutes(r *gin.Engine, h *handlers.Handler) {
	// 公开路由组 - 不需要认证
	public := r.Group("/api/v1/admin")
	{
		public.POST("/login", h.AdminLogin)             // 登录接口不需要认证
		public.POST("/refresh-token", h.RefreshToken)   // 添加刷新token接口
		public.GET("/product/image", h.GetProductImage) // 查看图片不认值，方便前端获取图片
	}

	// 需要认证的路由组
	admin := r.Group("/api/v1/admin")
	admin.Use(middleware.AuthMiddleware())
	{
		// 管理员基础接口
		admin.POST("/change-password", h.ChangeAdminPassword)

		// 商品管理接口
		product := admin.Group("/product")
		{
			product.POST("/create", h.CreateProduct)
			product.GET("/list", h.GetProducts)
			product.GET("/detail", h.GetProduct)
			product.PUT("/update", h.UpdateProduct)
			product.DELETE("/delete", h.DeleteProduct)
			product.POST("/upload-image", h.UploadProductImage)
			product.PUT("/toggle-status", h.ToggleProductStatus)
		}

		// 用户管理接口
		user := admin.Group("/user")
		{
			user.POST("/create", h.CreateUser)
			user.GET("/list", h.GetUsers)
			user.GET("/simple-list", h.GetUserSimpleList)
			user.GET("/detail", h.GetUser)
			user.PUT("/update", h.UpdateUser)
			user.DELETE("/delete", h.DeleteUser)
		}

		// 订单管理接口
		order := admin.Group("/order")
		{
			order.POST("/create", h.CreateOrder)
			order.GET("/list", h.GetOrders)
			order.GET("/detail", h.GetOrder)
			order.PUT("/update", h.UpdateOrder)
			order.DELETE("/delete", h.DeleteOrder)
			order.PUT("/toggle-status", h.ToggleOrderStatus)
		}

		// 数据管理接口
		data := admin.Group("/data")
		{
			data.GET("/export", h.ExportData)
			data.POST("/import", h.ImportData)
		}
	}
}
