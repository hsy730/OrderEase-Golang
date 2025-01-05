package middleware

import (
	"net/http"
	"orderease/utils"
	"strings"

	"orderease/database"
	"orderease/models"

	"github.com/gin-gonic/gin"
)

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		db := database.GetDB()
		token := c.GetHeader("Authorization")
		if token == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "未认证"})
			return
		}

		// 去掉Bearer前缀
		token = strings.TrimPrefix(token, "Bearer ")

		// 检查token是否在黑名单中
		var blacklistedToken models.BlacklistedToken
		if err := db.Where("token = ?", token).First(&blacklistedToken).Error; err == nil {
			// token在黑名单中
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "token已失效"})
			return
		}

		// 验证token
		claims, err := utils.ParseToken(token)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "无效的token"})
			return
		}

		// 将用户信息存入上下文
		c.Set("userID", claims.UserID)
		c.Set("username", claims.Username)

		c.Next()
	}
}
