package backend

import (
	"orderease/config"
	"orderease/handlers"
	"orderease/middleware"

	"github.com/gin-gonic/gin"
)

// SetupRoutes 配置所有路由
func SetupShopRoutes(r *gin.Engine, h *handlers.Handler) {
	// 获取基础路径
	basePath := config.AppConfig.Server.BasePath

	// 应用限流中间件到所有管理员接口

	// 需要认证的路由组
	shopOwner := r.Group(basePath + "/shopOwner")
	shopOwner.Use(middleware.RateLimitMiddleware())

	shopOwner.Use(middleware.BackendAuthMiddleware(false))
	{
		shopOwner.POST("/logout", h.Logout) // 添加登出接口
		// 商户基础接口
		shopOwner.POST("/change-password", h.ChangeShopPassword)

		// 商品管理接口
		product := shopOwner.Group("/product")
		{
			product.POST("/create", h.CreateProduct)
			product.GET("/list", h.GetProducts)
			product.GET("/detail", h.GetProduct)
			product.PUT("/update", h.UpdateProduct)
			product.DELETE("/delete", h.DeleteProduct)
			product.POST("/upload-image", h.UploadProductImage)
			product.PUT("/toggle-status", h.ToggleProductStatus)
		}

		// 订单管理接口
		order := shopOwner.Group("/order")
		{
			order.POST("/create", h.CreateOrder)
			order.GET("/list", h.GetOrders)
			order.GET("/detail", h.GetOrder)
			order.PUT("/update", h.UpdateOrder)
			order.DELETE("/delete", h.DeleteOrder)
			order.PUT("/toggle-status", h.ToggleOrderStatus)
			order.GET("/sse", h.SSEConnection)
		}

		// 标签管理接口
		tag := shopOwner.Group("/tag")
		{
			tag.POST("/create", h.CreateTag)
			tag.GET("/list", h.GetTagsForBackend)
			tag.GET("/detail", h.GetTag)
			tag.PUT("/update", h.UpdateTag)
			tag.DELETE("/delete", h.DeleteTag)
			tag.POST("/batch-tag", h.BatchTagProducts)               // 批量打标签接口
			tag.GET("/online-products", h.GetTagOnlineProducts)      // 获取标签关联的已上架商品
			tag.GET("/bound-tags", h.GetBoundTags)                   // 获取商品已绑定的标签
			tag.GET("/unbound-tags", h.GetUnboundTags)               // 获取商品未绑定的标签
			tag.POST("/batch-tag-product", h.BatchTagProduct)        // 批量设置商品标签
			tag.DELETE("/batch-untag", h.BatchUntagProducts)         // 批量解绑商品标签
			tag.GET("/unbound-list", h.GetUnboundTagsList)           // 获取没有绑定商品的标签列表
			tag.GET("/unbound-products", h.GetUnboundProductsForTag) // 获取标签未绑定的商品列表
			tag.GET("/bound-products", h.GetTagBoundProducts)        // 获取标签已绑定的商品列表
		}

		shop := shopOwner.Group("/shop")
		{
			shop.GET("/detail", h.GetShopInfo) // 新增店铺信息查询
		}

		// 用户管理接口
		user := shopOwner.Group("/user")
		{
			user.POST("/create", h.CreateUser)
			user.GET("/list", h.GetUsers)
			user.GET("/simple-list", h.GetUserSimpleList)
			user.GET("/detail", h.GetUser)
			user.PUT("/update", h.UpdateUser)
			user.DELETE("/delete", h.DeleteUser)
		}
	}
}
