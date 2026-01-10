package shared

// 订单状态类型
type OrderStatus int

// 订单状态常量
const (
	OrderStatusPending  OrderStatus = 1  // 待处理
	OrderStatusAccepted OrderStatus = 2  // 已接单
	OrderStatusRejected OrderStatus = 3  // 已拒绝
	OrderStatusShipped  OrderStatus = 4  // 已发货
	OrderStatusComplete OrderStatus = 10 // 已完成
	OrderStatusCanceled OrderStatus = -1 // 已取消
)

// 商品状态类型
type ProductStatus string

// 商品状态常量
const (
	ProductStatusPending ProductStatus = "pending" // 待上架
	ProductStatusOnline  ProductStatus = "online"  // 已上架
	ProductStatusOffline ProductStatus = "offline" // 已下架
)
