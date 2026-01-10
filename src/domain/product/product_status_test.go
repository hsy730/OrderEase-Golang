package product

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestProductStatus_IsValid(t *testing.T) {
	tests := []struct {
		name string
		s    ProductStatus
		want bool
	}{
		{"pending status", ProductStatusPending, true},
		{"online status", ProductStatusOnline, true},
		{"offline status", ProductStatusOffline, true},
		{"invalid status", ProductStatus("unknown"), false},
		{"empty status", ProductStatus(""), false},
		{"random status", ProductStatus("active"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.s.IsValid()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestProductStatus_CanTransitionTo(t *testing.T) {
	tests := []struct {
		name      string
		from      ProductStatus
		to        ProductStatus
		allowed   bool
	}{
		// Valid transitions
		{"pending to online", ProductStatusPending, ProductStatusOnline, true},
		{"online to offline", ProductStatusOnline, ProductStatusOffline, true},
		{"offline to online", ProductStatusOffline, ProductStatusOnline, true},

		// Invalid transitions
		{"pending to offline", ProductStatusPending, ProductStatusOffline, false},
		{"pending to pending", ProductStatusPending, ProductStatusPending, false},
		{"online to pending", ProductStatusOnline, ProductStatusPending, false},
		{"online to online", ProductStatusOnline, ProductStatusOnline, false},
		{"offline to pending", ProductStatusOffline, ProductStatusPending, false},
		{"offline to offline", ProductStatusOffline, ProductStatusOffline, false},

		// Invalid statuses
		{"unknown to online", ProductStatus("unknown"), ProductStatusOnline, false},
		{"online to unknown", ProductStatusOnline, ProductStatus("unknown"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.from.CanTransitionTo(tt.to)
			assert.Equal(t, tt.allowed, got)
		})
	}
}

func TestProductStatus_String(t *testing.T) {
	tests := []struct {
		name string
		s    ProductStatus
		want string
	}{
		{"pending", ProductStatusPending, "pending"},
		{"online", ProductStatusOnline, "online"},
		{"offline", ProductStatusOffline, "offline"},
		{"custom", ProductStatus("custom"), "custom"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.s.String()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestProductStatus_AllStatuses(t *testing.T) {
	// Verify all defined statuses are valid
	statuses := []ProductStatus{
		ProductStatusPending,
		ProductStatusOnline,
		ProductStatusOffline,
	}

	for _, status := range statuses {
		t.Run(status.String(), func(t *testing.T) {
			assert.True(t, status.IsValid(), "%s should be valid", status)
		})
	}
}

func TestProductStatus_TransitionMatrix(t *testing.T) {
	// Test the complete transition matrix
	transitions := map[string]map[string]bool{
		"pending": {
			"online":  true,
			"offline": false,
		},
		"online": {
			"offline": true,
			"pending": false,
		},
		"offline": {
			"online":  true,
			"pending": false,
		},
	}

	for from, toMap := range transitions {
		for to, expected := range toMap {
			t.Run(from+" to "+to, func(t *testing.T) {
				fromStatus := ProductStatus(from)
				toStatus := ProductStatus(to)
				got := fromStatus.CanTransitionTo(toStatus)
				assert.Equal(t, expected, got)
			})
		}
	}
}
