package value_objects

// OrderStatus 订单状态值对象
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

// String 返回状态的标签
func (s OrderStatus) String() string {
	switch s {
	case OrderStatusPending:
		return "待处理"
	case OrderStatusAccepted:
		return "已接单"
	case OrderStatusRejected:
		return "已拒绝"
	case OrderStatusShipped:
		return "已发货"
	case OrderStatusComplete:
		return "已完成"
	case OrderStatusCanceled:
		return "已取消"
	default:
		return "未知状态"
	}
}

// Label 返回状态的英文标签
func (s OrderStatus) Label() string {
	switch s {
	case OrderStatusPending:
		return "Pending"
	case OrderStatusAccepted:
		return "Accepted"
	case OrderStatusRejected:
		return "Rejected"
	case OrderStatusShipped:
		return "Shipped"
	case OrderStatusComplete:
		return "Complete"
	case OrderStatusCanceled:
		return "Canceled"
	default:
		return "Unknown"
	}
}

// Type 返回状态的类型（用于前端UI显示）
func (s OrderStatus) Type() string {
	switch s {
	case OrderStatusPending:
		return "warning"
	case OrderStatusAccepted:
		return "primary"
	case OrderStatusRejected:
		return "danger"
	case OrderStatusShipped:
		return "info"
	case OrderStatusComplete:
		return "success"
	case OrderStatusCanceled:
		return "info"
	default:
		return "default"
	}
}

// IsFinal 判断是否为最终状态（不可再转换）
func (s OrderStatus) IsFinal() bool {
	return s == OrderStatusComplete || s == OrderStatusCanceled || s == OrderStatusRejected
}

// CanTransitionTo 判断是否可以转换到目标状态
// 使用简化的状态转换规则
func (s OrderStatus) CanTransitionTo(to OrderStatus) bool {
	// 如果目标状态与当前状态相同，不允许
	if s == to {
		return false
	}

	// 定义允许的状态转换
	transitions := map[OrderStatus][]OrderStatus{
		OrderStatusPending: {OrderStatusAccepted, OrderStatusCanceled, OrderStatusRejected},
		OrderStatusAccepted: {OrderStatusShipped, OrderStatusCanceled},
		OrderStatusShipped:  {OrderStatusComplete},
	}

	// 最终状态不能再转换
	if s.IsFinal() {
		return false
	}

	// 检查是否允许转换
	allowedStates, exists := transitions[s]
	if !exists {
		return false
	}

	for _, allowed := range allowedStates {
		if allowed == to {
			return true
		}
	}

	return false
}

// Value 返回状态的整数值（用于数据库存储）
func (s OrderStatus) Value() int {
	return int(s)
}

// IsValid 验证状态是否有效
func (s OrderStatus) IsValid() bool {
	switch s {
	case OrderStatusPending, OrderStatusAccepted, OrderStatusRejected,
		OrderStatusShipped, OrderStatusComplete, OrderStatusCanceled:
		return true
	default:
		return false
	}
}

// OrderStatusFromInt 从整数创建 OrderStatus
func OrderStatusFromInt(status int) OrderStatus {
	return OrderStatus(status)
}

// DefaultOrderStatusFlow 默认订单流转状态配置（JSON格式）
// 这个字符串用于初始化新店铺的订单状态流转配置
const DefaultOrderStatusFlow = `{
  "statuses": [
    {
      "value": 1,
      "label": "待处理",
      "type": "warning",
      "isFinal": false,
      "actions": [
        {
          "name": "接单",
          "nextStatus": 2,
          "nextStatusLabel": "已接单"
        },
        {
          "name": "拒绝",
          "nextStatus": 3,
          "nextStatusLabel": "已拒绝"
        }
      ]
    },
    {
      "value": 2,
      "label": "已接单",
      "type": "primary",
      "isFinal": false,
      "actions": [
        {
          "name": "发货",
          "nextStatus": 4,
          "nextStatusLabel": "已发货"
        },
        {
          "name": "取消",
          "nextStatus": -1,
          "nextStatusLabel": "已取消"
        }
      ]
    },
    {
      "value": 4,
      "label": "已发货",
      "type": "info",
      "isFinal": false,
      "actions": [
        {
          "name": "完成",
          "nextStatus": 10,
          "nextStatusLabel": "已完成"
        }
      ]
    },
    {
      "value": 10,
      "label": "已完成",
      "type": "success",
      "isFinal": true,
      "actions": []
    },
    {
      "value": 3,
      "label": "已拒绝",
      "type": "danger",
      "isFinal": true,
      "actions": []
    },
    {
      "value": -1,
      "label": "已取消",
      "type": "info",
      "isFinal": true,
      "actions": []
    }
  ]
}`
