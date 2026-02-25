package models

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestBlacklistedTokenStruct(t *testing.T) {
	now := time.Now()
	token := BlacklistedToken{
		ID:        123,
		Token:     "test_token_value",
		ExpiredAt: now.Add(1 * time.Hour),
		CreatedAt: now,
	}

	assert.Equal(t, uint(123), token.ID)
	assert.Equal(t, "test_token_value", token.Token)
	assert.False(t, token.ExpiredAt.IsZero())
}

func TestBlacklistedToken_Empty(t *testing.T) {
	token := BlacklistedToken{}

	assert.Zero(t, token.ID)
	assert.Empty(t, token.Token)
	assert.True(t, token.ExpiredAt.IsZero())
	assert.True(t, token.CreatedAt.IsZero())
}

func TestBlacklistedToken_ExpiredInPast(t *testing.T) {
	now := time.Now()
	token := BlacklistedToken{
		ID:        123,
		Token:     "test_token",
		ExpiredAt: now.Add(-1 * time.Hour),
		CreatedAt: now,
	}

	assert.True(t, token.ExpiredAt.Before(now))
}

func TestBlacklistedToken_LongToken(t *testing.T) {
	now := time.Now()
	longToken := ""
	for i := 0; i < 500; i++ {
		longToken += "a"
	}

	token := BlacklistedToken{
		ID:        123,
		Token:     longToken,
		ExpiredAt: now.Add(1 * time.Hour),
	}

	assert.Len(t, token.Token, 500)
}

