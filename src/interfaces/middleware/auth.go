package middleware

import (
	"net/http"
	"orderease/utils"
	"orderease/utils/log2"
	"strings"

	"github.com/gin-gonic/gin"
)

type UserInfo struct {
	UserID   uint64
	IsAdmin  bool
	UserName string
}

func (u UserInfo) IsAdminUser() bool {
	return u.IsAdmin
}

func (u UserInfo) GetUserID() uint64 {
	return u.UserID
}

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			errorResponse(c, http.StatusUnauthorized, "未提供认证令牌")
			c.Abort()
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			errorResponse(c, http.StatusUnauthorized, "无效的认证令牌格式")
			c.Abort()
			return
		}

		claims, err := utils.ParseToken(tokenString)
		if err != nil {
			log2.Errorf("解析令牌失败: %v", err)
			errorResponse(c, http.StatusUnauthorized, "无效的认证令牌")
			c.Abort()
			return
		}

		userInfo := UserInfo{
			UserID:   claims.UserID,
			IsAdmin:  claims.IsAdmin,
			UserName: claims.Username,
		}

		c.Set("userInfo", userInfo)
		c.Next()
	}
}

func AdminMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		userInfo, exists := c.Get("userInfo")
		if !exists {
			errorResponse(c, http.StatusUnauthorized, "未找到用户信息")
			c.Abort()
			return
		}

		user, ok := userInfo.(UserInfo)
		if !ok {
			errorResponse(c, http.StatusInternalServerError, "用户信息格式错误")
			c.Abort()
			return
		}

		if !user.IsAdminUser() {
			errorResponse(c, http.StatusForbidden, "需要管理员权限")
			c.Abort()
			return
		}

		c.Next()
	}
}

func errorResponse(c *gin.Context, code int, message string) {
	log2.Errorf("错误响应: %d - %s", code, message)
	c.JSON(code, gin.H{"error": message})
}
