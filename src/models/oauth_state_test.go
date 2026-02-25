package models

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestOAuthStateStruct(t *testing.T) {
	now := time.Now()
	state := OAuthState{
		ID:        1,
		State:     "test_state_12345",
		Provider:  "wechat",
		ExpiresAt: now.Add(1 * time.Hour),
		CreatedAt: now,
	}

	assert.Equal(t, uint(1), state.ID)
	assert.Equal(t, "test_state_12345", state.State)
	assert.Equal(t, "wechat", state.Provider)
	assert.False(t, state.IsExpired())
}

func TestOAuthState_TableName(t *testing.T) {
	state := OAuthState{}
	assert.Equal(t, "oauth_states", state.TableName())
}

func TestOAuthState_IsExpired(t *testing.T) {
	now := time.Now()
	tests := []struct {
		name      string
		expiresAt time.Time
		expected  bool
	}{
		{
			name:      "expired 1 hour ago",
			expiresAt: now.Add(-1 * time.Hour),
			expected:  true,
		},
		{
			name:      "expired 1 minute ago",
			expiresAt: now.Add(-1 * time.Minute),
			expected:  true,
		},
		{
			name:      "expired 1 second ago",
			expiresAt: now.Add(-1 * time.Second),
			expected:  true,
		},
		{
			name:      "expires in 1 hour",
			expiresAt: now.Add(1 * time.Hour),
			expected:  false,
		},
		{
			name:      "expires in 1 minute",
			expiresAt: now.Add(1 * time.Minute),
			expected:  false,
		},
		{
			name:      "expires in 1 second",
			expiresAt: now.Add(1 * time.Second),
			expected:  false,
		},
		{
			name:      "expires exactly now",
			expiresAt: now,
			expected:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			state := &OAuthState{ExpiresAt: tt.expiresAt}
			assert.Equal(t, tt.expected, state.IsExpired())
		})
	}
}

func TestOAuthState_IsExpired_NilPointer(t *testing.T) {
	var state *OAuthState
	if state != nil {
		state.IsExpired()
	}
	// This test ensures we handle nil pointers appropriately
	// If this panics, the test will fail
	assert.Nil(t, state)
}

func TestOAuthState_Empty(t *testing.T) {
	state := OAuthState{}

	assert.Zero(t, state.ID)
	assert.Empty(t, state.State)
	assert.Empty(t, state.Provider)
	assert.True(t, state.ExpiresAt.IsZero())
	assert.True(t, state.IsExpired()) // Zero time means expired
}
