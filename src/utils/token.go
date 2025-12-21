package utils

import (
	"math/rand"
	"strconv"
	"time"
)

// GenerateTempToken 生成6位数字临时令牌
func GenerateTempToken() string {
	// 确保每次生成不同的随机数
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	// 生成6位数字，范围是100000到999999
	token := r.Intn(900000) + 100000

	return strconv.Itoa(token)
}

// IsTokenExpired 检查令牌是否过期
func IsTokenExpired(expiresAt time.Time) bool {
	return time.Now().After(expiresAt)
}
