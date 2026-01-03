package order

import (
	"orderease/domain/shared"
	"time"
)

type OrderRepository interface {
	Save(order *Order) error
	FindByID(id shared.ID) (*Order, error)
	FindByIDAndShopID(id shared.ID, shopID uint64) (*Order, error)
	FindByShopID(shopID uint64, page, pageSize int) ([]Order, int64, error)
	FindByUserID(userID shared.ID, shopID uint64, page, pageSize int) ([]Order, int64, error)
	FindUnfinishedByShopID(shopID uint64, flow OrderStatusFlow, page, pageSize int) ([]Order, int64, error)
	Search(shopID uint64, userID string, statuses []OrderStatus, startTime, endTime time.Time, page, pageSize int) ([]Order, int64, error)
	Delete(id shared.ID) error
	Update(order *Order) error
}

type OrderItemRepository interface {
	Save(item *OrderItem) error
	FindByOrderID(orderID shared.ID) ([]OrderItem, error)
	DeleteByOrderID(orderID shared.ID) error
}

type OrderItemOptionRepository interface {
	Save(option *OrderItemOption) error
	FindByOrderItemID(orderItemID shared.ID) ([]OrderItemOption, error)
	DeleteByOrderItemID(orderItemID shared.ID) error
}

type OrderStatusLogRepository interface {
	Save(log *OrderStatusLog) error
	FindByOrderID(orderID shared.ID) ([]OrderStatusLog, error)
	DeleteByOrderID(orderID shared.ID) error
}
