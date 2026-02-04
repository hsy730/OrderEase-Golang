package value_objects

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOrderStatus_String(t *testing.T) {
	tests := []struct {
		name     string
		status   OrderStatus
		expected string
	}{
		{
			name:     "Pending",
			status:   OrderStatusPending,
			expected: "待处理",
		},
		{
			name:     "Accepted",
			status:   OrderStatusAccepted,
			expected: "已接单",
		},
		{
			name:     "Rejected",
			status:   OrderStatusRejected,
			expected: "已拒绝",
		},
		{
			name:     "Shipped",
			status:   OrderStatusShipped,
			expected: "已发货",
		},
		{
			name:     "Complete",
			status:   OrderStatusComplete,
			expected: "已完成",
		},
		{
			name:     "Canceled",
			status:   OrderStatusCanceled,
			expected: "已取消",
		},
		{
			name:     "Invalid status",
			status:   OrderStatus(999),
			expected: "未知状态",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.status.String()
			assert.Equal(t, tt.expected, got)
		})
	}
}

func TestOrderStatus_Label(t *testing.T) {
	tests := []struct {
		name     string
		status   OrderStatus
		expected string
	}{
		{
			name:     "Pending",
			status:   OrderStatusPending,
			expected: "Pending",
		},
		{
			name:     "Accepted",
			status:   OrderStatusAccepted,
			expected: "Accepted",
		},
		{
			name:     "Rejected",
			status:   OrderStatusRejected,
			expected: "Rejected",
		},
		{
			name:     "Shipped",
			status:   OrderStatusShipped,
			expected: "Shipped",
		},
		{
			name:     "Complete",
			status:   OrderStatusComplete,
			expected: "Complete",
		},
		{
			name:     "Canceled",
			status:   OrderStatusCanceled,
			expected: "Canceled",
		},
		{
			name:     "Invalid status",
			status:   OrderStatus(999),
			expected: "Unknown",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.status.Label()
			assert.Equal(t, tt.expected, got)
		})
	}
}

func TestOrderStatus_Type(t *testing.T) {
	tests := []struct {
		name     string
		status   OrderStatus
		expected string
	}{
		{
			name:     "Pending",
			status:   OrderStatusPending,
			expected: "warning",
		},
		{
			name:     "Accepted",
			status:   OrderStatusAccepted,
			expected: "primary",
		},
		{
			name:     "Rejected",
			status:   OrderStatusRejected,
			expected: "danger",
		},
		{
			name:     "Shipped",
			status:   OrderStatusShipped,
			expected: "info",
		},
		{
			name:     "Complete",
			status:   OrderStatusComplete,
			expected: "success",
		},
		{
			name:     "Canceled",
			status:   OrderStatusCanceled,
			expected: "info",
		},
		{
			name:     "Invalid status",
			status:   OrderStatus(999),
			expected: "default",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.status.Type()
			assert.Equal(t, tt.expected, got)
		})
	}
}

func TestOrderStatus_IsFinal(t *testing.T) {
	tests := []struct {
		name       string
		status     OrderStatus
		wantFinal  bool
	}{
		{
			name:      "Pending - not final",
			status:    OrderStatusPending,
			wantFinal: false,
		},
		{
			name:      "Accepted - not final",
			status:    OrderStatusAccepted,
			wantFinal: false,
		},
		{
			name:      "Shipped - not final",
			status:    OrderStatusShipped,
			wantFinal: false,
		},
		{
			name:      "Complete - final",
			status:    OrderStatusComplete,
			wantFinal: true,
		},
		{
			name:      "Canceled - final",
			status:    OrderStatusCanceled,
			wantFinal: true,
		},
		{
			name:      "Rejected - final",
			status:    OrderStatusRejected,
			wantFinal: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.status.IsFinal()
			assert.Equal(t, tt.wantFinal, got)
		})
	}
}

func TestOrderStatus_CanTransitionTo(t *testing.T) {
	tests := []struct {
		name     string
		from     OrderStatus
		to       OrderStatus
		allowed  bool
	}{
		// Pending 状态的转换
		{
			name:    "Pending -> Accepted (allowed)",
			from:    OrderStatusPending,
			to:      OrderStatusAccepted,
			allowed: true,
		},
		{
			name:    "Pending -> Canceled (allowed)",
			from:    OrderStatusPending,
			to:      OrderStatusCanceled,
			allowed: true,
		},
		{
			name:    "Pending -> Rejected (allowed)",
			from:    OrderStatusPending,
			to:      OrderStatusRejected,
			allowed: true,
		},
		{
			name:    "Pending -> Shipped (not allowed)",
			from:    OrderStatusPending,
			to:      OrderStatusShipped,
			allowed: false,
		},
		{
			name:    "Pending -> Complete (not allowed)",
			from:    OrderStatusPending,
			to:      OrderStatusComplete,
			allowed: false,
		},
		// Accepted 状态的转换
		{
			name:    "Accepted -> Shipped (allowed)",
			from:    OrderStatusAccepted,
			to:      OrderStatusShipped,
			allowed: true,
		},
		{
			name:    "Accepted -> Canceled (allowed)",
			from:    OrderStatusAccepted,
			to:      OrderStatusCanceled,
			allowed: true,
		},
		{
			name:    "Accepted -> Pending (not allowed)",
			from:    OrderStatusAccepted,
			to:      OrderStatusPending,
			allowed: false,
		},
		{
			name:    "Accepted -> Complete (not allowed)",
			from:    OrderStatusAccepted,
			to:      OrderStatusComplete,
			allowed: false,
		},
		// Shipped 状态的转换
		{
			name:    "Shipped -> Complete (allowed)",
			from:    OrderStatusShipped,
			to:      OrderStatusComplete,
			allowed: true,
		},
		{
			name:    "Shipped -> Canceled (not allowed)",
			from:    OrderStatusShipped,
			to:      OrderStatusCanceled,
			allowed: false,
		},
		{
			name:    "Shipped -> Pending (not allowed)",
			from:    OrderStatusShipped,
			to:      OrderStatusPending,
			allowed: false,
		},
		// 最终状态的转换
		{
			name:    "Complete -> Accepted (not allowed)",
			from:    OrderStatusComplete,
			to:      OrderStatusAccepted,
			allowed: false,
		},
		{
			name:    "Complete -> Shipped (not allowed)",
			from:    OrderStatusComplete,
			to:      OrderStatusShipped,
			allowed: false,
		},
		{
			name:    "Canceled -> Pending (not allowed)",
			from:    OrderStatusCanceled,
			to:      OrderStatusPending,
			allowed: false,
		},
		{
			name:    "Rejected -> Pending (not allowed)",
			from:    OrderStatusRejected,
			to:      OrderStatusPending,
			allowed: false,
		},
		// 相同状态
		{
			name:    "Pending -> Pending (not allowed)",
			from:    OrderStatusPending,
			to:      OrderStatusPending,
			allowed: false,
		},
		{
			name:    "Accepted -> Accepted (not allowed)",
			from:    OrderStatusAccepted,
			to:      OrderStatusAccepted,
			allowed: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.from.CanTransitionTo(tt.to)
			assert.Equal(t, tt.allowed, got)
		})
	}
}

func TestOrderStatus_Value(t *testing.T) {
	tests := []struct {
		name      string
		status    OrderStatus
		expected  int
	}{
		{
			name:     "Pending",
			status:   OrderStatusPending,
			expected: 1,
		},
		{
			name:     "Accepted",
			status:   OrderStatusAccepted,
			expected: 2,
		},
		{
			name:     "Rejected",
			status:   OrderStatusRejected,
			expected: 3,
		},
		{
			name:     "Shipped",
			status:   OrderStatusShipped,
			expected: 4,
		},
		{
			name:     "Complete",
			status:   OrderStatusComplete,
			expected: 10,
		},
		{
			name:     "Canceled",
			status:   OrderStatusCanceled,
			expected: -1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.status.Value()
			assert.Equal(t, tt.expected, got)
		})
	}
}

func TestOrderStatus_IsValid(t *testing.T) {
	tests := []struct {
		name      string
		status    OrderStatus
		wantValid bool
	}{
		{
			name:      "Pending - valid",
			status:    OrderStatusPending,
			wantValid: true,
		},
		{
			name:      "Accepted - valid",
			status:    OrderStatusAccepted,
			wantValid: true,
		},
		{
			name:      "Rejected - valid",
			status:    OrderStatusRejected,
			wantValid: true,
		},
		{
			name:      "Shipped - valid",
			status:    OrderStatusShipped,
			wantValid: true,
		},
		{
			name:      "Complete - valid",
			status:    OrderStatusComplete,
			wantValid: true,
		},
		{
			name:      "Canceled - valid",
			status:    OrderStatusCanceled,
			wantValid: true,
		},
		{
			name:      "Invalid status - 999",
			status:    OrderStatus(999),
			wantValid: false,
		},
		{
			name:      "Invalid status - 0",
			status:    OrderStatus(0),
			wantValid: false,
		},
		{
			name:      "Invalid status - negative",
			status:    OrderStatus(-999),
			wantValid: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.status.IsValid()
			assert.Equal(t, tt.wantValid, got)
		})
	}
}

func TestOrderStatusFromInt(t *testing.T) {
	tests := []struct {
		name      string
		statusInt int
		expected  OrderStatus
	}{
		{
			name:      "Pending",
			statusInt: 1,
			expected:  OrderStatusPending,
		},
		{
			name:      "Accepted",
			statusInt: 2,
			expected:  OrderStatusAccepted,
		},
		{
			name:      "Rejected",
			statusInt: 3,
			expected:  OrderStatusRejected,
		},
		{
			name:      "Shipped",
			statusInt: 4,
			expected:  OrderStatusShipped,
		},
		{
			name:      "Complete",
			statusInt: 10,
			expected:  OrderStatusComplete,
		},
		{
			name:      "Canceled",
			statusInt: -1,
			expected:  OrderStatusCanceled,
		},
		{
			name:      "Custom status",
			statusInt: 999,
			expected:  OrderStatus(999),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := OrderStatusFromInt(tt.statusInt)
			assert.Equal(t, tt.expected, got)
		})
	}
}

func TestDefaultOrderStatusFlow(t *testing.T) {
	// 测试默认状态流转配置不为空
	assert.NotEmpty(t, DefaultOrderStatusFlow)
	// 验证包含所有必要的状态
	assert.Contains(t, DefaultOrderStatusFlow, "待处理")
	assert.Contains(t, DefaultOrderStatusFlow, "已接单")
	assert.Contains(t, DefaultOrderStatusFlow, "已完成")
	assert.Contains(t, DefaultOrderStatusFlow, "已取消")
	assert.Contains(t, DefaultOrderStatusFlow, "已拒绝")
	assert.Contains(t, DefaultOrderStatusFlow, "已发货")
}
