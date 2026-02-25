package models

import (
	"testing"
	"time"

	"github.com/bwmarrin/snowflake"
	"github.com/stretchr/testify/assert"
)

func TestTempTokenStruct(t *testing.T) {
	now := time.Now()
	token := TempToken{
		ID:        snowflake.ID(123),
		ShopID:    snowflake.ID(456),
		UserID:    789,
		Token:     "123456",
		ExpiresAt: now.Add(5 * time.Minute),
		CreatedAt: now,
		UpdatedAt: now,
	}

	assert.Equal(t, snowflake.ID(123), token.ID)
	assert.Equal(t, snowflake.ID(456), token.ShopID)
	assert.Equal(t, uint64(789), token.UserID)
	assert.Equal(t, "123456", token.Token)
}
