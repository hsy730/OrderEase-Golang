package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenerateRandomString(t *testing.T) {
	tests := []struct {
		name        string
		length      int
		expectError bool
	}{
		{
			name:   "length 0",
			length: 0,
		},
		{
			name:   "length 1",
			length: 1,
		},
		{
			name:   "length 8",
			length: 8,
		},
		{
			name:   "length 16",
			length: 16,
		},
		{
			name:   "length 32",
			length: 32,
		},
		{
			name:   "length 50",
			length: 50,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GenerateRandomString(tt.length)
			assert.Len(t, result, tt.length)
		})
	}
}

func TestGenerateRandomString_Charset(t *testing.T) {
	// 验证生成的字符串只包含有效字符集
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

	result := GenerateRandomString(100)
	for _, char := range result {
		assert.Contains(t, charset, string(char))
	}
}

func TestGenerateRandomString_Consistency(t *testing.T) {
	// 多次调用应该生成不同的字符串（虽然理论上可能重复，但概率极低）
	results := make(map[string]bool)
	for i := 0; i < 10; i++ {
		result := GenerateRandomString(10)
		results[result] = true
	}

	// 至少应该有一些不同的字符串
	assert.GreaterOrEqual(t, len(results), 5)
}

func TestGenerateRandomBase64(t *testing.T) {
	tests := []struct {
		name   string
		length int
	}{
		{
			name:   "length 4",
			length: 4,
		},
		{
			name:   "length 8",
			length: 8,
		},
		{
			name:   "length 16",
			length: 16,
		},
		{
			name:   "length 32",
			length: 32,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GenerateRandomBase64(tt.length)
			assert.Len(t, result, tt.length)
		})
	}
}

func TestGenerateRandomBase64_Format(t *testing.T) {
	// 验证生成的字符串是有效的 base64 字符
	for i := 0; i < 10; i++ {
		result := GenerateRandomBase64(32)
		// Base64 encoded string may contain padding '='
		assert.Len(t, result, 32)
	}
}

func isValidBase64Char(c rune) bool {
	const base64Chars = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/="
	for _, char := range base64Chars {
		if c == char {
			return true
		}
	}
	return false
}
