package backend

import (
	"orderease/config"
	"orderease/handlers"
	"orderease/middleware"

	"github.com/gin-gonic/gin"
)

// SetupAdminRoutes 配置管理员相关路由
func SetupAdminRoutes(r *gin.Engine, h *handlers.Handler) {
	// 获取基础路径
	basePath := config.AppConfig.Server.BasePath

	// 需要认证的路由组
	admin := r.Group(basePath + "/admin")
	admin.Use(middleware.RateLimitMiddleware())
	admin.Use(middleware.BackendAuthMiddleware(true))

	// 管理员基础路由
	setupAdminBaseRoutes(admin, h)
	
	// 店铺管理路由
	setupAdminShopRoutes(admin, h)
	
	// 商品管理路由
	setupAdminProductRoutes(admin, h)
	
	// 用户管理路由
	setupAdminUserRoutes(admin, h)
	
	// 订单管理路由
	setupAdminOrderRoutes(admin, h)
	
	// 标签管理路由
	setupAdminTagRoutes(admin, h)
	
	// 数据管理路由
	setupAdminDataRoutes(admin, h)
}

// setupAdminBaseRoutes 配置管理员基础路由
func setupAdminBaseRoutes(group *gin.RouterGroup, h *handlers.Handler) {
	// @Summary 管理员登出
	// @Description 管理员登出接口
	// @Tags 管理员基础
	// @Accept json
	// @Produce json
	// @Success 200 {object} Response
	// @Router /api/admin/logout [post]
	group.POST("/logout", h.Logout)
	
	// @Summary 修改管理员密码
	// @Description 修改管理员登录密码
	// @Tags 管理员基础
	// @Accept json
	// @Produce json
	// @Param passwordRequest body ChangePasswordRequest true "密码信息"
	// @Success 200 {object} Response
	// @Router /api/admin/change-password [post]
	group.POST("/change-password", h.ChangeAdminPassword)
}

// setupAdminShopRoutes 配置管理员店铺管理路由
func setupAdminShopRoutes(group *gin.RouterGroup, h *handlers.Handler) {
	shop := group.Group("/shop")
	{
		// @Summary 创建店铺
		// @Description 创建新店铺
		// @Tags 店铺管理
		// @Accept json
		// @Produce json
		// @Param shop body CreateShopRequest true "店铺信息"
		// @Success 200 {object} Response
		// @Router /api/admin/shop/create [post]
		shop.POST("/create", h.CreateShop)
		
		// @Summary 更新店铺信息
		// @Description 更新店铺基本信息
		// @Tags 店铺管理
		// @Accept json
		// @Produce json
		// @Param shop body UpdateShopRequest true "店铺信息"
		// @Success 200 {object} Response
		// @Router /api/admin/shop/update [put]
		shop.PUT("/update", h.UpdateShop)
		
		// @Summary 获取店铺详情
		// @Description 获取指定店铺的详细信息
		// @Tags 店铺管理
		// @Accept json
		// @Produce json
		// @Param shopId query string true "店铺ID"
		// @Success 200 {object} Response
		// @Router /api/admin/shop/detail [get]
		shop.GET("/detail", h.GetShopInfo)
		
		// @Summary 获取店铺列表
		// @Description 获取店铺列表，支持分页和筛选
		// @Tags 店铺管理
		// @Accept json
		// @Produce json
		// @Param page query int false "页码"
		// @Param pageSize query int false "每页数量"
		// @Param status query string false "店铺状态"
		// @Success 200 {object} Response
		// @Router /api/admin/shop/list [get]
		shop.GET("/list", h.GetShopList)
		
		// @Summary 删除店铺
		// @Description 删除指定店铺
		// @Tags 店铺管理
		// @Accept json
		// @Produce json
		// @Param shopId query string true "店铺ID"
		// @Success 200 {object} Response
		// @Router /api/admin/shop/delete [delete]
		shop.DELETE("/delete", h.DeleteShop)
		
		// @Summary 上传店铺图片
		// @Description 上传店铺图片
		// @Tags 店铺管理
		// @Accept multipart/form-data
		// @Produce json
		// @Param image formData file true "店铺图片"
		// @Success 200 {object} Response
		// @Router /api/admin/shop/upload-image [post]
		shop.POST("/upload-image", h.UploadShopImage)
		
		// @Summary 检查店铺名称是否存在
		// @Description 检查店铺名称是否已被使用
		// @Tags 店铺管理
		// @Accept json
		// @Produce json
		// @Param name query string true "店铺名称"
		// @Success 200 {object} Response
		// @Router /api/admin/shop/check-name [get]
		shop.GET("/check-name", h.CheckShopNameExists)
		
		// @Summary 获取店铺图片
		// @Description 获取指定店铺的图片
		// @Tags 店铺管理
		// @Accept json
		// @Produce json
		// @Param shopId query string true "店铺ID"
		// @Success 200 {object} Response
		// @Router /api/admin/shop/image [get]
		shop.GET("/image", h.GetShopImage)
		
		// @Summary 获取店铺临时令牌
		// @Description 为店铺生成临时访问令牌
		// @Tags 店铺管理
		// @Accept json
		// @Produce json
		// @Param shopId query string true "店铺ID"
		// @Success 200 {object} Response
		// @Router /api/admin/shop/temp-token [get]
		shop.GET("/temp-token", h.GetShopTempToken)
		
		// @Summary 更新订单状态流转
		// @Description 更新店铺的订单状态流转配置
		// @Tags 店铺管理
		// @Accept json
		// @Produce json
		// @Param flow body OrderStatusFlowRequest true "状态流转信息"
		// @Success 200 {object} Response
		// @Router /api/admin/shop/update-order-status-flow [put]
		shop.PUT("/update-order-status-flow", h.UpdateOrderStatusFlow)
	}
}

// setupAdminProductRoutes 配置管理员商品管理路由
func setupAdminProductRoutes(group *gin.RouterGroup, h *handlers.Handler) {
	product := group.Group("/product")
	{
		// @Summary 创建商品
		// @Description 创建新商品
		// @Tags 商品管理
		// @Accept json
		// @Produce json
		// @Param product body CreateProductRequest true "商品信息"
		// @Success 200 {object} Response
		// @Router /api/admin/product/create [post]
		product.POST("/create", h.CreateProduct)
		
		// @Summary 获取商品列表
		// @Description 获取商品列表，支持分页和筛选
		// @Tags 商品管理
		// @Accept json
		// @Produce json
		// @Param page query int false "页码"
		// @Param pageSize query int false "每页数量"
		// @Param shopId query string false "店铺ID"
		// @Param status query string false "商品状态"
		// @Success 200 {object} Response
		// @Router /api/admin/product/list [get]
		product.GET("/list", h.GetProducts)
		
		// @Summary 获取商品详情
		// @Description 获取指定商品的详细信息
		// @Tags 商品管理
		// @Accept json
		// @Produce json
		// @Param productId query string true "商品ID"
		// @Success 200 {object} Response
		// @Router /api/admin/product/detail [get]
		product.GET("/detail", h.GetProduct)
		
		// @Summary 更新商品信息
		// @Description 更新商品基本信息
		// @Tags 商品管理
		// @Accept json
		// @Produce json
		// @Param product body UpdateProductRequest true "商品信息"
		// @Success 200 {object} Response
		// @Router /api/admin/product/update [put]
		product.PUT("/update", h.UpdateProduct)
		
		// @Summary 删除商品
		// @Description 删除指定商品
		// @Tags 商品管理
		// @Accept json
		// @Produce json
		// @Param productId query string true "商品ID"
		// @Success 200 {object} Response
		// @Router /api/admin/product/delete [delete]
		product.DELETE("/delete", h.DeleteProduct)
		
		// @Summary 上传商品图片
		// @Description 上传商品图片
		// @Tags 商品管理
		// @Accept multipart/form-data
		// @Produce json
		// @Param image formData file true "商品图片"
		// @Success 200 {object} Response
		// @Router /api/admin/product/upload-image [post]
		product.POST("/upload-image", h.UploadProductImage)
		
		// @Summary 切换商品状态
		// @Description 切换商品上架/下架状态
		// @Tags 商品管理
		// @Accept json
		// @Produce json
		// @Param productId query string true "商品ID"
		// @Success 200 {object} Response
		// @Router /api/admin/product/toggle-status [put]
		product.PUT("/toggle-status", h.ToggleProductStatus)
		
		// @Summary 获取商品图片
		// @Description 获取指定商品的图片
		// @Tags 商品管理
		// @Accept json
		// @Produce json
		// @Param productId query string true "商品ID"
		// @Success 200 {object} Response
		// @Router /api/admin/product/image [get]
		product.GET("/image", h.GetProductImage)
	}
}

// setupAdminUserRoutes 配置管理员用户管理路由
func setupAdminUserRoutes(group *gin.RouterGroup, h *handlers.Handler) {
	user := group.Group("/user")
	{
		// @Summary 创建用户
		// @Description 创建新用户
		// @Tags 用户管理
		// @Accept json
		// @Produce json
		// @Param user body CreateUserRequest true "用户信息"
		// @Success 200 {object} Response
		// @Router /api/admin/user/create [post]
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
		// @Router /api/admin/user/list [get]
		user.GET("/list", h.GetUsers)
		
		// @Summary 获取用户简单列表
		// @Description 获取用户简单信息列表，用于下拉选择等场景
		// @Tags 用户管理
		// @Accept json
		// @Produce json
		// @Success 200 {object} Response
		// @Router /api/admin/user/simple-list [get]
		user.GET("/simple-list", h.GetUserSimpleList)
		
		// @Summary 获取用户详情
		// @Description 获取指定用户的详细信息
		// @Tags 用户管理
		// @Accept json
		// @Produce json
		// @Param userId query string true "用户ID"
		// @Success 200 {object} Response
		// @Router /api/admin/user/detail [get]
		user.GET("/detail", h.GetUser)
		
		// @Summary 更新用户信息
		// @Description 更新用户基本信息
		// @Tags 用户管理
		// @Accept json
		// @Produce json
		// @Param user body UpdateUserRequest true "用户信息"
		// @Success 200 {object} Response
		// @Router /api/admin/user/update [put]
		user.PUT("/update", h.UpdateUser)
		
		// @Summary 删除用户
		// @Description 删除指定用户
		// @Tags 用户管理
		// @Accept json
		// @Produce json
		// @Param userId query string true "用户ID"
		// @Success 200 {object} Response
		// @Router /api/admin/user/delete [delete]
		user.DELETE("/delete", h.DeleteUser)
	}
}

// setupAdminOrderRoutes 配置管理员订单管理路由
func setupAdminOrderRoutes(group *gin.RouterGroup, h *handlers.Handler) {
	order := group.Group("/order")
	{
		// @Summary 创建订单
		// @Description 创建新订单
		// @Tags 订单管理
		// @Accept json
		// @Produce json
		// @Param order body CreateOrderRequest true "订单信息"
		// @Success 200 {object} Response
		// @Router /api/admin/order/create [post]
		order.POST("/create", h.CreateOrder)
		
		// @Summary 获取订单列表
		// @Description 获取订单列表，支持分页和筛选
		// @Tags 订单管理
		// @Accept json
		// @Produce json
		// @Param page query int false "页码"
		// @Param pageSize query int false "每页数量"
		// @Param status query string false "订单状态"
		// @Param shopId query string false "店铺ID"
		// @Success 200 {object} Response
		// @Router /api/admin/order/list [get]
		order.GET("/list", h.GetOrders)
		
		// @Summary 获取订单详情
		// @Description 获取指定订单的详细信息
		// @Tags 订单管理
		// @Accept json
		// @Produce json
		// @Param orderId query string true "订单ID"
		// @Success 200 {object} Response
		// @Router /api/admin/order/detail [get]
		order.GET("/detail", h.GetOrder)
		
		// @Summary 更新订单信息
		// @Description 更新订单基本信息
		// @Tags 订单管理
		// @Accept json
		// @Produce json
		// @Param order body UpdateOrderRequest true "订单信息"
		// @Success 200 {object} Response
		// @Router /api/admin/order/update [put]
		order.PUT("/update", h.UpdateOrder)
		
		// @Summary 删除订单
		// @Description 删除指定订单
		// @Tags 订单管理
		// @Accept json
		// @Produce json
		// @Param orderId query string true "订单ID"
		// @Success 200 {object} Response
		// @Router /api/admin/order/delete [delete]
		order.DELETE("/delete", h.DeleteOrder)
		
		// @Summary 切换订单状态
		// @Description 切换订单状态
		// @Tags 订单管理
		// @Accept json
		// @Produce json
		// @Param orderId query string true "订单ID"
		// @Param status query string true "目标状态"
		// @Success 200 {object} Response
		// @Router /api/admin/order/toggle-status [put]
		order.PUT("/toggle-status", h.ToggleOrderStatus)
		
		// @Summary 获取订单状态流转
		// @Description 获取订单状态流转信息
		// @Tags 订单管理
		// @Accept json
		// @Produce json
		// @Param orderId query string true "订单ID"
		// @Success 200 {object} Response
		// @Router /api/admin/order/status-flow [get]
		order.GET("/status-flow", h.GetOrderStatusFlow)
		
		// @Summary SSE连接
		// @Description 建立SSE连接，用于实时推送订单状态更新
		// @Tags 订单管理
		// @Accept json
		// @Produce text/event-stream
		// @Success 200 {string} string "SSE流"
		// @Router /api/admin/order/sse [get]
		order.GET("/sse", h.SSEConnection)
		
		// @Summary 高级搜索订单
		// @Description 使用多种条件高级搜索订单
		// @Tags 订单管理
		// @Accept json
		// @Produce json
		// @Param searchRequest body AdvanceSearchOrderRequest true "搜索条件"
		// @Success 200 {object} Response
		// @Router /api/admin/order/advance-search [post]
		order.POST("/advance-search", h.GetAdvanceSearchOrders)
	}
}

// setupAdminTagRoutes 配置管理员标签管理路由
func setupAdminTagRoutes(group *gin.RouterGroup, h *handlers.Handler) {
	tag := group.Group("/tag")
	{
		// @Summary 创建标签
		// @Description 创建新标签
		// @Tags 标签管理
		// @Accept json
		// @Produce json
		// @Param tag body CreateTagRequest true "标签信息"
		// @Success 200 {object} Response
		// @Router /api/admin/tag/create [post]
		tag.POST("/create", h.CreateTag)
		
		// @Summary 获取标签列表
		// @Description 获取标签列表，支持分页和筛选
		// @Tags 标签管理
		// @Accept json
		// @Produce json
		// @Param page query int false "页码"
		// @Param pageSize query int false "每页数量"
		// @Success 200 {object} Response
		// @Router /api/admin/tag/list [get]
		tag.GET("/list", h.GetTagsForBackend)
		
		// @Summary 获取标签详情
		// @Description 获取指定标签的详细信息
		// @Tags 标签管理
		// @Accept json
		// @Produce json
		// @Param tagId query string true "标签ID"
		// @Success 200 {object} Response
		// @Router /api/admin/tag/detail [get]
		tag.GET("/detail", h.GetTag)
		
		// @Summary 更新标签信息
		// @Description 更新标签基本信息
		// @Tags 标签管理
		// @Accept json
		// @Produce json
		// @Param tag body UpdateTagRequest true "标签信息"
		// @Success 200 {object} Response
		// @Router /api/admin/tag/update [put]
		tag.PUT("/update", h.UpdateTag)
		
		// @Summary 删除标签
		// @Description 删除指定标签
		// @Tags 标签管理
		// @Accept json
		// @Produce json
		// @Param tagId query string true "标签ID"
		// @Success 200 {object} Response
		// @Router /api/admin/tag/delete [delete]
		tag.DELETE("/delete", h.DeleteTag)
		
		// @Summary 批量打标签
		// @Description 批量给商品打标签
		// @Tags 标签管理
		// @Accept json
		// @Produce json
		// @Param batchTagRequest body BatchTagRequest true "批量打标签信息"
		// @Success 200 {object} Response
		// @Router /api/admin/tag/batch-tag [post]
		tag.POST("/batch-tag", h.BatchTagProducts)
		
		// @Summary 获取标签关联的已上架商品
		// @Description 获取指定标签关联的已上架商品列表
		// @Tags 标签管理
		// @Accept json
		// @Produce json
		// @Param tagId query string true "标签ID"
		// @Success 200 {object} Response
		// @Router /api/admin/tag/online-products [get]
		tag.GET("/online-products", h.GetTagOnlineProducts)
		
		// @Summary 获取商品已绑定的标签
		// @Description 获取指定商品已绑定的标签列表
		// @Tags 标签管理
		// @Accept json
		// @Produce json
		// @Param productId query string true "商品ID"
		// @Success 200 {object} Response
		// @Router /api/admin/tag/bound-tags [get]
		tag.GET("/bound-tags", h.GetBoundTags)
		
		// @Summary 获取商品未绑定的标签
		// @Description 获取指定商品未绑定的标签列表
		// @Tags 标签管理
		// @Accept json
		// @Produce json
		// @Param productId query string true "商品ID"
		// @Success 200 {object} Response
		// @Router /api/admin/tag/unbound-tags [get]
		tag.GET("/unbound-tags", h.GetUnboundTags)
		
		// @Summary 批量设置商品标签
		// @Description 批量设置商品的标签
		// @Tags 标签管理
		// @Accept json
		// @Produce json
		// @Param batchTagProductRequest body BatchTagProductRequest true "批量设置标签信息"
		// @Success 200 {object} Response
		// @Router /api/admin/tag/batch-tag-product [post]
		tag.POST("/batch-tag-product", h.BatchTagProduct)
		
		// @Summary 批量解绑商品标签
		// @Description 批量解绑商品的标签
		// @Tags 标签管理
		// @Accept json
		// @Produce json
		// @Param batchUntagRequest body BatchUntagRequest true "批量解绑标签信息"
		// @Success 200 {object} Response
		// @Router /api/admin/tag/batch-untag [delete]
		tag.DELETE("/batch-untag", h.BatchUntagProducts)
		
		// @Summary 获取没有绑定商品的标签列表
		// @Description 获取没有绑定任何商品的标签列表
		// @Tags 标签管理
		// @Accept json
		// @Produce json
		// @Success 200 {object} Response
		// @Router /api/admin/tag/unbound-list [get]
		tag.GET("/unbound-list", h.GetUnboundTagsList)
		
		// @Summary 获取标签未绑定的商品列表
		// @Description 获取指定标签未绑定的商品列表
		// @Tags 标签管理
		// @Accept json
		// @Produce json
		// @Param tagId query string true "标签ID"
		// @Success 200 {object} Response
		// @Router /api/admin/tag/unbound-products [get]
		tag.GET("/unbound-products", h.GetUnboundProductsForTag)
		
		// @Summary 获取标签已绑定的商品列表
		// @Description 获取指定标签已绑定的商品列表
		// @Tags 标签管理
		// @Accept json
		// @Produce json
		// @Param tagId query string true "标签ID"
		// @Success 200 {object} Response
		// @Router /api/admin/tag/bound-products [get]
		tag.GET("/bound-products", h.GetTagBoundProducts)
	}
}

// setupAdminDataRoutes 配置管理员数据管理路由
func setupAdminDataRoutes(group *gin.RouterGroup, h *handlers.Handler) {
	data := group.Group("/data")
	{
		// @Summary 导出数据
		// @Description 导出系统数据
		// @Tags 数据管理
		// @Accept json
		// @Produce application/octet-stream
		// @Param type query string true "数据类型"
		// @Success 200 {file} file "导出文件"
		// @Router /api/admin/data/export [get]
		data.GET("/export", h.ExportData)
		
		// @Summary 导入数据
		// @Description 导入系统数据
		// @Tags 数据管理
		// @Accept multipart/form-data
		// @Produce json
		// @Param file formData file true "导入文件"
		// @Success 200 {object} Response
		// @Router /api/admin/data/import [post]
		data.POST("/import", h.ImportData)
	}
}
