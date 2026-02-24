package utils

import (
	"strings"
	"testing"
	"time"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

// setupJWTConfig sets up viper configuration for JWT tests
func setupJWTConfig() {
	viper.Set("jwt.secret", "e6jf493kdhbms9ew6mv2v1a4dx2")
	viper.Set("jwt.expiration", 7200)
}

func TestGenerateToken(t *testing.T) {
	setupJWTConfig()
	userID := uint64(123456789)
	username := "testuser"

	token, _, err := GenerateToken(userID, username)
	assert.NoError(t, err)
	assert.NotEmpty(t, token)

	// Token should be valid JWT format (3 parts separated by dots)
	parts := strings.Split(token, ".")
	assert.Len(t, parts, 3)
}

func TestGenerateToken_ExpirationTime(t *testing.T) {
	setupJWTConfig()
	userID := uint64(1)
	username := "test"

	token1, expiredAt1, err1 := GenerateToken(userID, username)
	assert.NoError(t, err1)
	assert.NotEmpty(t, token1)
	assert.True(t, expiredAt1.After(time.Now()))

	time.Sleep(10 * time.Millisecond)

	token2, expiredAt2, err2 := GenerateToken(userID, username)
	assert.NoError(t, err2)
	assert.NotEmpty(t, token2)
	assert.True(t, expiredAt2.After(time.Now()))

	// Expiration times should be different due to clock
	assert.NotEqual(t, expiredAt1, expiredAt2)

	// Both should be roughly 2 hours from now (from config)
	// Allow for some timing variation (around 7200 seconds)
	duration1 := expiredAt1.Sub(time.Now())
	duration2 := expiredAt2.Sub(time.Now())

	assert.True(t, duration1 > 1*time.Hour && duration1 < 3*time.Hour)
	assert.True(t, duration2 > 1*time.Hour && duration2 < 3*time.Hour)
}

func TestParseToken(t *testing.T) {
	setupJWTConfig()
	// Generate a token first
	userID := uint64(123456789)
	username := "testuser"
	token, _, err := GenerateToken(userID, username)
	assert.NoError(t, err)

	// Parse and validate token
	claims, err := ParseToken(token)
	assert.NoError(t, err)
	assert.NotNil(t, claims)
	assert.Equal(t, userID, claims.UserID)
	assert.Equal(t, username, claims.Username)
}

func TestParseToken_Invalid(t *testing.T) {
	invalidTokens := []struct {
		name  string
		token string
	}{
		{"empty token", ""},
		{"invalid format", "invalid.token.format"},
		{"malformed base64", "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.aGVudGVyInRiZXZjcnNvbS5qMjM0"},
		{"invalid signature", "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.aGVudGVyInRiZXZjcnNvbS5qMjM0Kjg3MTQzOTZ"},
	}

	for _, tt := range invalidTokens {
		t.Run("invalid: "+tt.name, func(t *testing.T) {
			claims, err := ParseToken(tt.token)
			assert.Error(t, err)
			assert.Nil(t, claims)
		})
	}
}

func TestParseToken_Wrong(t *testing.T) {
	setupJWTConfig()
	// Create a valid token
	userID := uint64(123456789)
	username := "testuser"
	token, _, _ := GenerateToken(userID, username)

	// Modify signature part (last character after last dot)
	lastDot := strings.LastIndex(token, ".")
	if lastDot != -1 {
		invalidToken := token[:lastDot] + "A" + token[lastDot+1:]
		claims, err := ParseToken(invalidToken)
		assert.Error(t, err)
		assert.Nil(t, claims)
	}
}
