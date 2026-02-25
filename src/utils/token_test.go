package utils

import (
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGenerateTempToken(t *testing.T) {
	token := GenerateTempToken()

	assert.Len(t, token, 6)

	for _, char := range token {
		assert.True(t, char >= '0' && char <= '9')
	}
}

func TestGenerateTempToken_Range(t *testing.T) {
	for i := 0; i < 100; i++ {
		token := GenerateTempToken()

		value, _ := strconv.Atoi(token)

		assert.GreaterOrEqual(t, value, 100000)
		assert.LessOrEqual(t, value, 999999)
	}
}

func TestIsTokenExpired(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name      string
		expiresAt time.Time
		expected  bool
	}{
		{
			name:      "expired 1 second ago",
			expiresAt: now.Add(-1 * time.Second),
			expected:  true,
		},
		{
			name:      "expired 1 hour ago",
			expiresAt: now.Add(-1 * time.Hour),
			expected:  true,
		},
		{
			name:      "expires in 1 hour",
			expiresAt: now.Add(1 * time.Hour),
			expected:  false,
		},
		{
			name:      "expires in 1 second",
			expiresAt: now.Add(1 * time.Second),
			expected:  false,
		},
		{
			name:      "zero time",
			expiresAt: time.Time{},
			expected:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsTokenExpired(tt.expiresAt)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsTokenExpired_EdgeCases(t *testing.T) {
	justExpired := time.Now().Add(-1 * time.Millisecond)
	justValid := time.Now().Add(1 * time.Millisecond)

	assert.True(t, IsTokenExpired(justExpired))
	assert.False(t, IsTokenExpired(justValid))
}
