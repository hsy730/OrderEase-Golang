package handlers

import (
	"fmt"
	"time"

	"orderease/models"
	"orderease/utils"

	"github.com/bwmarrin/snowflake"
	"github.com/golang-jwt/jwt/v4"
	"github.com/spf13/viper"
)

// JWTService JWT 服务
type JWTService struct{}

// NewJWTService 创建 JWT 服务
func NewJWTService() *JWTService {
	return &JWTService{}
}

// GenerateToken 生成 JWT token
func (s *JWTService) GenerateToken(user *models.User) (string, int64, error) {
	// 获取过期时间配置
	expirationSeconds := viper.GetInt("jwt.expiration")

	// 转换 ID 为 uint64
	var userID uint64
	userID = uint64(user.ID)

	// 使用现有的 GenerateToken 函数
	tokenString, _, err := utils.GenerateToken(userID, user.Name)
	if err != nil {
		return "", 0, fmt.Errorf("generate token failed: %w", err)
	}

	return tokenString, int64(expirationSeconds), nil
}

// ValidateToken 验证 JWT token
func (s *JWTService) ValidateToken(tokenString string) (*utils.Claims, error) {
	return utils.ParseToken(tokenString)
}

// GetUserIDFromToken 从 token 获取用户 ID
func (s *JWTService) GetUserIDFromToken(tokenString string) (snowflake.ID, error) {
	claims, err := s.ValidateToken(tokenString)
	if err != nil {
		return 0, fmt.Errorf("validate token failed: %w", err)
	}

	return snowflake.ID(claims.UserID), nil
}

// GenerateTokenWithClaims 使用自定义 claims 生成 token
func (s *JWTService) GenerateTokenWithClaims(userID uint64, username string, isAdmin bool) (string, time.Time, error) {
	expirationSeconds := viper.GetInt("jwt.expiration")
	expirationTime := time.Now().Add(time.Duration(expirationSeconds) * time.Second)

	claims := &utils.Claims{
		UserID:   userID,
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(viper.GetString("jwt.secret")))
	if err != nil {
		return "", time.Time{}, fmt.Errorf("sign token failed: %w", err)
	}

	return tokenString, expirationTime, nil
}
