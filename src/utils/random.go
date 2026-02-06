package utils

import (
	"crypto/rand"
	"encoding/base64"
	"math/big"
)

// GenerateRandomString 生成指定长度的随机字符串
func GenerateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	result := make([]byte, length)
	charsetLen := big.NewInt(int64(len(charset)))

	for i := range result {
		n, err := rand.Int(rand.Reader, charsetLen)
		if err != nil {
			// 如果生成随机数失败，使用简单的备用方案
			result[i] = charset[i%len(charset)]
			continue
		}
		result[i] = charset[n.Int64()]
	}

	return string(result)
}

// GenerateRandomBase64 生成随机 base64 字符串
func GenerateRandomBase64(length int) string {
	b := make([]byte, length)
	_, err := rand.Read(b)
	if err != nil {
		// 备用方案
		return GenerateRandomString(length)
	}
	return base64.URLEncoding.EncodeToString(b)[:length]
}
