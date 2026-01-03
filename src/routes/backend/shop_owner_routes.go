package backend

import (
	"orderease/config"
	"orderease/handlers"
	"orderease/middleware"

	"github.com/gin-gonic/gin"
)

// SetupShopRoutes 配置商家相关路由
func SetupShopRoutes(r *gin.Engine, h *handlers.Handler) {
	// 获取基础路径
	basePath := config.AppConfig.Server.BasePath

	// 需要认证的路由组
	shopOwner := r.Group(basePath + "/shopOwner")
	shopOwner.Use(middleware.RateLimitMiddleware())
	shopOwner.Use(middleware.BackendAuthMiddleware(false))

	// 商家基础路由
	setupShopOwnerBaseRoutes(shopOwner, h)
	
	// 商品管理路由
	setupShopOwnerProductRoutes(shopOwner, h)
	
	// 订单管理路由
	setupShopOwnerOrderRoutes(shopOwner, h)
	
	// 标签管理路由
	setupShopOwnerTagRoutes(shopOwner, h)
	
	// 店铺管理路由
	setupShopOwnerShopRoutes(shopOwner, h)
	
	// 用户管理路由
	setupShopOwnerUserRoutes(shopOwner, h)
}

// setupShopOwnerBaseRoutes 配置商家基础路由
func setupShopOwnerBaseRoutes(group *gin.RouterGroup, h *handlers.Handler) {
	// @Summary 商家登出
	// @Description 商家登出接口
	// @Tags 商家基础
	// @Accept json
	// @Produce json
	// @Success 200 {object} Response
	// @Router /api/shopOwner/logout [post]
	group.POST("/logout", h.Logout)
	
	// @Summary 修改商家密码
	// @Description 修改商家登录密码
	// @Tags 商家基础
	// @Accept json
	// @Produce json
	// @Param passwordRequest body ChangePasswordRequest true "密码信息"
	// @Success 200 {object} Response
	// @Router /api/shopOwner/change-password [post]
	group.POST("/change-password", h.ChangeShopPassword)
}

// setupShopOwnerProductRoutes 配置商家商品管理路由
func setupShopOwnerProductRoutes(group *gin.RouterGroup, h *handlers.Handler) {
	product := group.Group("/product")
	{
		// @Summary 创建商品
		// @Description 创建新商品
		// @Tags 商品管理
		// @Accept json
		// @Produce json
		// @Param product body CreateProductRequest true "商品信息"
		// @Success 200 {object} Response
		// @Router /api/shopOwner/product/create [post]
		product.POST("/create", h.CreateProduct)
		
		// @Summary 获取商品列表
		// @Description 获取商品列表，支持分页和筛选
		// @Tags 商品管理
		// @Accept json
		// @Produce json
		// @Param page query int false "页码"
		// @Param pageSize query int false "每页数量"
		// @Param status query string false "商品状态"
		// @Success 200 {object} Response
		// @Router /api/shopOwner/product/list [get]
		product.GET("/list", h.GetProducts)
		
		// @Summary 获取商品详情
		// @Description 获取指定商品的详细信息
		// @Tags 商品管理
		// @Accept json
		// @Produce json
		// @Param productId query string true "商品ID"
		// @Success 200 {object} Response
		// @Router /api/shopOwner/product/detail [get]
		product.GET("/detail", h.GetProduct)
		
		// @Summary 更新商品信息
		// @Description 更新商品基本信息
		// @Tags 商品管理
		// @Accept json
		// @Produce json
		// @Param product body UpdateProductRequest true "商品信息"
		// @Success 200 {object} Response
		// @Router /api/shopOwner/product/update [put]
		product.PUT("/update", h.UpdateProduct)
		
		// @Summary 删除商品
		// @Description 删除指定商品
		// @Tags 商品管理
		// @Accept json
		// @Produce json
		// @Param productId query string true "商品ID"
		// @Success 200 {object} Response
		// @Router /api/shopOwner/product/delete [delete]
		product.DELETE("/delete", h.DeleteProduct)
		
		// @Summary 上传商品图片
		// @Description 上传商品图片
		// @Tags 商品管理
		// @Accept multipart/form-data
		// @Produce json
		// @Param image formData file true "商品图片"
		// @Success 200 {object} Response
		// @Router /api/shopOwner/product/upload-image [post]
		product.POST("/upload-image", h.UploadProductImage)
		
		// @Summary 切换商品状态
		// @Description 切换商品上架/下架状态
		// @Tags 商品管理
		// @Accept json
		// @Produce json
		// @Param productId query string true "商品ID"
		// @Success 200 {object} Response
		// @Router /api/shopOwner/product/toggle-status [put]
		product.PUT("/toggle-status", h.ToggleProductStatus)
		
		// @Summary 获取商品图片
		// @Description 获取指定商品的图片
		// @Tags 商品管理
		// @Accept json
		// @Produce json
		// @Param productId query string true "商品ID"
		// @Success 200 {object} Response
		// @Router /api/shopOwner/product/image [get]
		product.GET("/image", h.GetProductImage)
	}
}

// setupShopOwnerOrderRoutes 配置商家订单管理路由
func setupShopOwnerOrderRoutes(group *gin.RouterGroup, h *handlers.Handler) {
	order := group.Group("/order")
	{
		// @Summary 创建订单
		// @Description 创建新订单
		// @Tags 订单管理
		// @Accept json
		// @Produce json
		// @Param order body CreateOrderRequest true "订单信息"
		// @Success 200 {object} Response
		// @Router /api/shopOwner/order/create [post]
		order.POST("/create", h.CreateOrder)
		
		// @Summary 获取订单列表
		// @Description 获取订单列表，支持分页和筛选
		// @Tags 订单管理
		// @Accept json
		// @Produce json
		// @Param page query int false "页码"
		// @Param pageSize query int false "每页数量"
		// @Param status query string false "订单状态"
		// @Success 200 {object} Response
		// @Router /api/shopOwner/order/list [get]
		order.GET("/list", h.GetOrders)
		
		// @Summary 获取订单详情
		// @Description 获取指定订单的详细信息
		// @Tags 订单管理
		// @Accept json
		// @Produce json
		// @Param orderId query string true "订单ID"
		// @Success 200 {object} Response
		// @Router /api/shopOwner/order/detail [get]
		order.GET("/detail", h.GetOrder)
		
		// @Summary 更新订单信息
		// @Description 更新订单基本信息
		// @Tags 订单管理
		// @Accept json
		// @Produce json
		// @Param order body UpdateOrderRequest true "订单信息"
		// @Success 200 {object} Response
		// @Router /api/shopOwner/order/update [put]
		order.PUT("/update", h.UpdateOrder)
		
		// @Summary 删除订单
		// @Description 删除指定订单
		// @Tags 订单管理
		// @Accept json
		// @Produce json
		// @Param orderId query string true "订单ID"
		// @Success 200 {object} Response
		// @Router /api/shopOwner/order/delete [delete]
		order.DELETE("/delete", h.DeleteOrder)
		
		// @Summary 切换订单状态
		// @Description 切换订单状态
		// @Tags 订单管理
		// @Accept json
		// @Produce json
		// @Param orderId query string true "订单ID"
		// @Param status query string true "目标状态"
		// @Success 200 {object} Response
		// @Router /api/shopOwner/order/toggle-status [put]
		order.PUT("/toggle-status", h.ToggleOrderStatus)
		
		// @Summary 获取订单状态流转
		// @Description 获取订单状态流转信息
		// @Tags 订单管理
		// @Accept json
		// @Produce json
		// @Param orderId query string true "订单ID"
		// @Success 200 {object} Response
		// @Router /api/shopOwner/order/status-flow [get]
		order.GET("/status-flow", h.GetOrderStatusFlow)
		
		// @Summary 获取未完成订单列表
		// @Description 获取未完成的订单列表
		// @Tags 订单管理
		// @Accept json
		// @Produce json
		// @Success 200 {object} Response
		// @Router /api/shopOwner/order/unfinished-list [get]
		order.GET("/unfinished-list", h.GetUnfinishedOrders)
		
		// @Summary SSE连接
		// @Description 建立SSE连接，用于实时推送订单状态更新
		// @Tags 订单管理
		// @Accept json
		// @Produce text/event-stream
		// @Success 200 {string} string "SSE流"
		// @Router /api/shopOwner/order/sse [get]
		order.GET("/sse", h.SSEConnection)
		
		// @Summary 高级搜索订单
		// @Description 使用多种条件高级搜索订单
		// @Tags 订单管理
		// @Accept json
		// @Produce json
		// @Param searchRequest body AdvanceSearchOrderRequest true "搜索条件"
		// @Success 200 {object} Response
		// @Router /api/shopOwner/order/advance-search [post]
		order.POST("/advance-search", h.GetAdvanceSearchOrders)
	}
}

// setupShopOwnerTagRoutes 配置商家标签管理路由
func setupShopOwnerTagRoutes(group *gin.RouterGroup, h *handlers.Handler) {
	tag := group.Group("/tag")
	{
		// @Summary 创建标签
		// @Description 创建新标签
		// @Tags 标签管理
		// @Accept json
		// @Produce json
		// @Param tag body CreateTagRequest true "标签信息"
		// @Success 200 {object} Response
		// @Router /api/shopOwner/tag/create [post]
		tag.POST("/create", h.CreateTag)
		
		// @Summary 获取标签列表
		// @Description 获取标签列表，支持分页和筛选
		// @Tags 标签管理
		// @Accept json
		// @Produce json
		// @Param page query int false "页码"
		// @Param pageSize query int false "每页数量"
		// @Success 200 {object} Response
		// @Router /api/shopOwner/tag/list [get]
		tag.GET("/list", h.GetTagsForBackend)
		
		// @Summary 获取标签详情
		// @Description 获取指定标签的详细信息
		// @Tags 标签管理
		// @Accept json
		// @Produce json
		// @Param tagId query string true "标签ID"
		// @Success 200 {object} Response
		// @Router /api/shopOwner/tag/detail [get]
		tag.GET("/detail", h.GetTag)
		
		// @Summary 更新标签信息
		// @Description 更新标签基本信息
		// @Tags 标签管理
		// @Accept json
		// @Produce json
		// @Param tag body UpdateTagRequest true "标签信息"
		// @Success 200 {object} Response
		// @Router /api/shopOwner/tag/update [put]
		tag.PUT("/update", h.UpdateTag)
		
		// @Summary 删除标签
		// @Description 删除指定标签
		// @Tags 标签管理
		// @Accept json
		// @Produce json
		// @Param tagId query string true "标签ID"
		// @Success 200 {object} Response
		// @Router /api/shopOwner/tag/delete [delete]
		tag.DELETE("/delete", h.DeleteTag)
		
		// @Summary 批量打标签
		// @Description 批量给商品打标签
		// @Tags 标签管理
		// @Accept json
		// @Produce json
		// @Param batchTagRequest body BatchTagRequest true "批量打标签信息"
		// @Success 200 {object} Response
		// @Router /api/shopOwner/tag/batch-tag [post]
		tag.POST("/batch-tag", h.BatchTagProducts)
		
		// @Summary 获取标签关联的已上架商品
		// @Description 获取指定标签关联的已上架商品列表
		// @Tags 标签管理
		// @Accept json
		// @Produce json
		// @Param tagId query string true "标签ID"
		// @Success 200 {object} Response
		// @Router /api/shopOwner/tag/online-products [get]
		tag.GET("/online-products", h.GetTagOnlineProducts)
		
		// @Summary 获取商品已绑定的标签
		// @Description 获取指定商品已绑定的标签列表
		// @Tags 标签管理
		// @Accept json
		// @Produce json
		// @Param productId query string true "商品ID"
		// @Success 200 {object} Response
		// @Router /api/shopOwner/tag/bound-tags [get]
		tag.GET("/bound-tags", h.GetBoundTags)
		
		// @Summary 获取商品未绑定的标签
		// @Description 获取指定商品未绑定的标签列表
		// @Tags 标签管理
		// @Accept json
		// @Produce json
		// @Param productId query string true "商品ID"
		// @Success 200 {object} Response
		// @Router /api/shopOwner/tag/unbound-tags [get]
		tag.GET("/unbound-tags", h.GetUnboundTags)
		
		// @Summary 批量设置商品标签
		// @Description 批量设置商品的标签
		// @Tags 标签管理
		// @Accept json
		// @Produce json
		// @Param batchTagProductRequest body BatchTagProductRequest true "批量设置标签信息"
		// @Success 200 {object} Response
		// @Router /api/shopOwner/tag/batch-tag-product [post]
		tag.POST("/batch-tag-product", h.BatchTagProduct)
		
		// @Summary 批量解绑商品标签
		// @Description 批量解绑商品的标签
		// @Tags 标签管理
		// @Accept json
		// @Produce json
		// @Param batchUntagRequest body BatchUntagRequest true "批量解绑标签信息"
		// @Success 200 {object} Response
		// @Router /api/shopOwner/tag/batch-untag [delete]
		tag.DELETE("/batch-untag", h.BatchUntagProducts)
		
		// @Summary 获取没有绑定商品的标签列表
		// @Description 获取没有绑定任何商品的标签列表
		// @Tags 标签管理
		// @Accept json
		// @Produce json
		// @Success 200 {object} Response
		// @Router /api/shopOwner/tag/unbound-list [get]
		tag.GET("/unbound-list", h.GetUnboundTagsList)
		
		// @Summary 获取标签未绑定的商品列表
		// @Description 获取指定标签未绑定的商品列表
		// @Tags 标签管理
		// @Accept json
		// @Produce json
		// @Param tagId query string true "标签ID"
		// @Success 200 {object} Response
		// @Router /api/shopOwner/tag/unbound-products [get]
		tag.GET("/unbound-products", h.GetUnboundProductsForTag)
		
		// @Summary 获取标签已绑定的商品列表
		// @Description 获取指定标签已绑定的商品列表
		// @Tags 标签管理
		// @Accept json
		// @Produce json
		// @Param tagId query string true "标签ID"
		// @Success 200 {object} Response
		// @Router /api/shopOwner/tag/bound-products [get]
		tag.GET("/bound-products", h.GetTagBoundProducts)
	}
}

// setupShopOwnerShopRoutes 配置商家店铺管理路由
func setupShopOwnerShopRoutes(group *gin.RouterGroup, h *handlers.Handler) {
	shop := group.Group("/shop")
	{
		// @Summary 获取店铺详情
		// @Description 获取指定店铺的详细信息
		// @Tags 店铺管理
		// @Accept json
		// @Produce json
		// @Success 200 {object} Response
		// @Router /api/shopOwner/shop/detail [get]
		shop.GET("/detail", h.GetShopInfo)
		
		// @Summary 获取店铺图片
		// @Description 获取指定店铺的图片
		// @Tags 店铺管理
		// @Accept json
		// @Produce json
		// @Success 200 {object} Response
		// @Router /api/shopOwner/shop/image [get]
		shop.GET("/image", h.GetShopImage)
		
		// @Summary 更新店铺信息
		// @Description 更新店铺基本信息
		// @Tags 店铺管理
		// @Accept json
		// @Produce json
		// @Param shop body UpdateShopRequest true "店铺信息"
		// @Success 200 {object} Response
		// @Router /api/shopOwner/shop/update [put]
		shop.PUT("/update", h.UpdateShop)
		
		// @Summary 更新订单状态流转
		// @Description 更新店铺的订单状态流转配置
		// @Tags 店铺管理
		// @Accept json
		// @Produce json
		// @Param flow body OrderStatusFlowRequest true "状态流转信息"
		// @Success 200 {object} Response
		// @Router /api/shopOwner/shop/update-order-status-flow [put]
		shop.PUT("/update-order-status-flow", h.UpdateOrderStatusFlow)

		// @Summary 获取临时令牌
		// @Description 获取店铺临时令牌
		// @Tags 店铺管理
		// @Accept json
		// @Produce json
		// @Success 200 {object} Response
		// @Router /api/shopOwner/shop/temp-token [get]
		shop.GET("/temp-token", h.GetShopTempToken)
	}
}

// setupShopOwnerUserRoutes 配置商家用户管理路由
func setupShopOwnerUserRoutes(group *gin.RouterGroup, h *handlers.Handler) {
	user := group.Group("/user")
	{
		// @Summary 创建用户
		// @Description 创建新用户
		// @Tags 用户管理
		// @Accept json
		// @Produce json
		// @Param user body CreateUserRequest true "用户信息"
		// @Success 200 {object} Response
		// @Router /api/shopOwner/user/create [post]
		user.POST("/create", h.CreateUser)
		
		// @Summary 获取用户列表
		// @Description 获取用户列表，支持分页和筛选
		// @Tags 用户管理
		// @Accept json
		// @Produce json
		// @Param page query int false "页码"
		// @Param pageSize query int false "每页数量"
		// @Param status query string false "用户状态"
		// @Success 200 {object} Response
		// @Router /api/shopOwner/user/list [get]
		user.GET("/list", h.GetUsers)
		
		// @Summary 获取用户简单列表
		// @Description 获取用户简单信息列表，用于下拉选择等场景
		// @Tags 用户管理
		// @Accept json
		// @Produce json
		// @Success 200 {object} Response
		// @Router /api/shopOwner/user/simple-list [get]
		user.GET("/simple-list", h.GetUserSimpleList)
		
		// @Summary 获取用户详情
		// @Description 获取指定用户的详细信息
		// @Tags 用户管理
		// @Accept json
		// @Produce json
		// @Param userId query string true "用户ID"
		// @Success 200 {object} Response
		// @Router /api/shopOwner/user/detail [get]
		user.GET("/detail", h.GetUser)
		
		// @Summary 更新用户信息
		// @Description 更新用户基本信息
		// @Tags 用户管理
		// @Accept json
		// @Produce json
		// @Param user body UpdateUserRequest true "用户信息"
		// @Success 200 {object} Response
		// @Router /api/shopOwner/user/update [put]
		user.PUT("/update", h.UpdateUser)
		
		// @Summary 删除用户
		// @Description 删除指定用户
		// @Tags 用户管理
		// @Accept json
		// @Produce json
		// @Param userId query string true "用户ID"
		// @Success 200 {object} Response
		// @Router /api/shopOwner/user/delete [delete]
		user.DELETE("/delete", h.DeleteUser)
	}
}
