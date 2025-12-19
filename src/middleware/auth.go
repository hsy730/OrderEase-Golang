package middleware

import (
	"net/http"
	"orderease/utils"
	"orderease/utils/log2"
	"strings"

	"orderease/database"
	"orderease/models"

	"github.com/gin-gonic/gin"
)

func BackendAuthMiddleware(isAdmin bool) gin.HandlerFunc {
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
			log2.Debugf("invalid token: %s, error: %v", token, err)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "无效的token"})
			return
		}
		if claims.Username != "admin" && isAdmin { // 检查是否为管理员
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "没有管理员权限"})
			return
		}

		// 设置shopID

		log2.Debugf("token验证成功, 用户ID: %d, 用户名: %s", claims.UserID, claims.Username)
		// 将用户信息存入上下文
		userInfo := models.UserInfo{
			UserID:   claims.UserID,
			Username: claims.Username,
			IsAdmin:  isAdmin,
		}
		c.Set("userInfo", userInfo)
		c.Next()
	}
}

// FrontendAuthMiddleware 前端用户认证中间件
func FrontendAuthMiddleware() gin.HandlerFunc {
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
			log2.Debugf("invalid token: %s, error: %v", token, err)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "无效的token"})
			return
		}

		// 验证用户是否存在
		var user models.User
		if err := db.Where("id = ?", claims.UserID).First(&user).Error; err != nil {
			log2.Debugf("用户不存在: %d, error: %v", claims.UserID, err)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "用户不存在"})
			return
		}

		// 验证用户名是否匹配
		if user.Name != claims.Username {
			log2.Debugf("用户名不匹配: token中的用户名=%s, 数据库中的用户名=%s", claims.Username, user.Name)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "用户信息不匹配"})
			return
		}

		log2.Debugf("前端用户token验证成功, 用户ID: %d, 用户名: %s", claims.UserID, claims.Username)
		// 将用户信息存入上下文
		userInfo := models.UserInfo{
			UserID:   claims.UserID,
			Username: claims.Username,
			IsAdmin:  false, // 前端用户默认不是管理员
		}
		c.Set("userInfo", userInfo)
		c.Next()
	}
}
