package order

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOrderStatus_IsFinal(t *testing.T) {
	tests := []struct {
		name string
		s    OrderStatus
		want bool
	}{
		{"pending is not final", OrderStatusPending, false},
		{"accepted is not final", OrderStatusAccepted, false},
		{"shipped is not final", OrderStatusShipped, false},
		{"complete is final", OrderStatusComplete, true},
		{"rejected is final", OrderStatusRejected, true},
		{"canceled is final", OrderStatusCanceled, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.s.IsFinal()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestOrderStatusFlow_CanTransition(t *testing.T) {
	// Create a simple flow for testing
	tests := []struct {
		name         string
		flow         OrderStatusFlow
		from         OrderStatus
		to           OrderStatus
		expected     bool
	}{
		{
			name: "pending to accepted - allowed",
			flow: OrderStatusFlow{
				Statuses: []OrderStatusConfig{
					{
						Value:   OrderStatusPending,
						IsFinal: false,
						Actions: []OrderStatusTransition{
							{NextStatus: OrderStatusAccepted},
						},
					},
				},
			},
			from:     OrderStatusPending,
			to:       OrderStatusAccepted,
			expected: true,
		},
		{
			name: "pending to rejected - allowed",
			flow: OrderStatusFlow{
				Statuses: []OrderStatusConfig{
					{
						Value:   OrderStatusPending,
						IsFinal: false,
						Actions: []OrderStatusTransition{
							{NextStatus: OrderStatusAccepted},
							{NextStatus: OrderStatusRejected},
						},
					},
				},
			},
			from:     OrderStatusPending,
			to:       OrderStatusRejected,
			expected: true,
		},
		{
			name: "pending to shipped - not allowed",
			flow: OrderStatusFlow{
				Statuses: []OrderStatusConfig{
					{
						Value:   OrderStatusPending,
						IsFinal: false,
						Actions: []OrderStatusTransition{
							{NextStatus: OrderStatusAccepted},
						},
					},
				},
			},
			from:     OrderStatusPending,
			to:       OrderStatusShipped,
			expected: false,
		},
		{
			name: "final status cannot transition",
			flow: OrderStatusFlow{
				Statuses: []OrderStatusConfig{
					{
						Value:   OrderStatusComplete,
						IsFinal: true,
						Actions: []OrderStatusTransition{},
					},
				},
			},
			from:     OrderStatusComplete,
			to:       OrderStatusPending,
			expected: false,
		},
		{
			name: "unknown source status",
			flow: OrderStatusFlow{
				Statuses: []OrderStatusConfig{
					{
						Value:   OrderStatusPending,
						IsFinal: false,
						Actions: []OrderStatusTransition{},
					},
				},
			},
			from:     OrderStatusShipped,
			to:       OrderStatusComplete,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.flow.CanTransition(tt.from, tt.to)
			assert.Equal(t, tt.expected, got)
		})
	}
}

func TestOrderStatusFlow_GetUnfinishedStatuses(t *testing.T) {
	tests := []struct {
		name              string
		flow              OrderStatusFlow
		expectedCount     int
		expectedStatuses  []OrderStatus
	}{
		{
			name: "mixed statuses",
			flow: OrderStatusFlow{
				Statuses: []OrderStatusConfig{
					{Value: OrderStatusPending, IsFinal: false},
					{Value: OrderStatusAccepted, IsFinal: false},
					{Value: OrderStatusComplete, IsFinal: true},
					{Value: OrderStatusCanceled, IsFinal: true},
				},
			},
			expectedCount: 2,
			expectedStatuses: []OrderStatus{
				OrderStatusPending,
				OrderStatusAccepted,
			},
		},
		{
			name: "all unfinished",
			flow: OrderStatusFlow{
				Statuses: []OrderStatusConfig{
					{Value: OrderStatusPending, IsFinal: false},
					{Value: OrderStatusAccepted, IsFinal: false},
					{Value: OrderStatusShipped, IsFinal: false},
				},
			},
			expectedCount: 3,
			expectedStatuses: []OrderStatus{
				OrderStatusPending,
				OrderStatusAccepted,
				OrderStatusShipped,
			},
		},
		{
			name: "all final",
			flow: OrderStatusFlow{
				Statuses: []OrderStatusConfig{
					{Value: OrderStatusComplete, IsFinal: true},
					{Value: OrderStatusRejected, IsFinal: true},
				},
			},
			expectedCount:     0,
			expectedStatuses:  []OrderStatus{},
		},
		{
			name:              "empty flow",
			flow:              OrderStatusFlow{Statuses: []OrderStatusConfig{}},
			expectedCount:     0,
			expectedStatuses:  []OrderStatus{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.flow.GetUnfinishedStatuses()
			assert.Equal(t, tt.expectedCount, len(got))

			for _, expected := range tt.expectedStatuses {
				assert.Contains(t, got, expected)
			}
		})
	}
}

func TestOrderStatus_String(t *testing.T) {
	tests := []struct {
		name string
		s    OrderStatus
		want string
	}{
		{"pending", OrderStatusPending, "待处理"},
		{"accepted", OrderStatusAccepted, "已接单"},
		{"rejected", OrderStatusRejected, "已拒绝"},
		{"shipped", OrderStatusShipped, "已发货"},
		{"complete", OrderStatusComplete, "已完成"},
		{"canceled", OrderStatusCanceled, "已取消"},
		{"unknown", OrderStatus(999), "未知状态"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.s.String()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestOrderStatus_AllStatuses(t *testing.T) {
	// Verify all defined statuses
	statuses := []OrderStatus{
		OrderStatusPending,
		OrderStatusAccepted,
		OrderStatusRejected,
		OrderStatusShipped,
		OrderStatusComplete,
		OrderStatusCanceled,
	}

	for _, status := range statuses {
		t.Run(status.String(), func(t *testing.T) {
			assert.NotEmpty(t, status.String())
		})
	}
}

func TestOrderStatus_TransitionWorkflow(t *testing.T) {
	// Test a typical order workflow
	t.Run("typical workflow: pending -> accepted -> shipped -> complete", func(t *testing.T) {
		flow := OrderStatusFlow{
			Statuses: []OrderStatusConfig{
				{
					Value:   OrderStatusPending,
					IsFinal: false,
					Actions: []OrderStatusTransition{
						{NextStatus: OrderStatusAccepted},
						{NextStatus: OrderStatusRejected},
						{NextStatus: OrderStatusCanceled},
					},
				},
				{
					Value:   OrderStatusAccepted,
					IsFinal: false,
					Actions: []OrderStatusTransition{
						{NextStatus: OrderStatusShipped},
						{NextStatus: OrderStatusCanceled},
					},
				},
				{
					Value:   OrderStatusShipped,
					IsFinal: false,
					Actions: []OrderStatusTransition{
						{NextStatus: OrderStatusComplete},
					},
				},
				{
					Value:   OrderStatusComplete,
					IsFinal: true,
					Actions: []OrderStatusTransition{},
				},
				{
					Value:   OrderStatusRejected,
					IsFinal: true,
					Actions: []OrderStatusTransition{},
				},
				{
					Value:   OrderStatusCanceled,
					IsFinal: true,
					Actions: []OrderStatusTransition{},
				},
			},
		}

		// Verify all valid transitions
		assert.True(t, flow.CanTransition(OrderStatusPending, OrderStatusAccepted))
		assert.True(t, flow.CanTransition(OrderStatusAccepted, OrderStatusShipped))
		assert.True(t, flow.CanTransition(OrderStatusShipped, OrderStatusComplete))

		// Verify invalid transitions
		assert.False(t, flow.CanTransition(OrderStatusPending, OrderStatusComplete))
		assert.False(t, flow.CanTransition(OrderStatusComplete, OrderStatusPending))
		assert.False(t, flow.CanTransition(OrderStatusAccepted, OrderStatusPending))
	})
}
