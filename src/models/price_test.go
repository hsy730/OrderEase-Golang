package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPrice_String(t *testing.T) {
	tests := []struct {
		name     string
		price    Price
		expected string
	}{
		{
			name:     "zero price",
			price:    0,
			expected: "0.00",
		},
		{
			name:     "positive price",
			price:    123.45,
			expected: "123.45",
		},
		{
			name:     "negative price",
			price:    -10.50,
			expected: "-10.50",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.price.String()
			assert.Equal(t, tt.expected, result)
		})
	}
}
