package middleware

import (
	"orderease/utils/log2"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

const (
	// 普通接口限流配置
	defaultRate  = 2 // 每秒请求数
	defaultBurst = 5 // 突发请求数

	// 登录接口限流配置
	loginRate  = 1 // 每10秒允许1个请求
	loginBurst = 3 // 最大突发3个请求
)

// 存储每个IP的限流器
var (
	normalLimiters = make(map[string]*rate.Limiter)
	loginLimiters  = make(map[string]*rate.Limiter)
	mu             sync.RWMutex
)

// 获取限流器
func getLimiter(ip string, isLogin bool) *rate.Limiter {
	log2.Debugf("获取限流器, IP: %s, 登录接口: %v", ip, isLogin)
	mu.RLock()
	var limiter *rate.Limiter
	var exists bool

	if isLogin {
		limiter, exists = loginLimiters[ip]
	} else {
		limiter, exists = normalLimiters[ip]
	}
	mu.RUnlock()

	if exists {
		return limiter
	}

	mu.Lock()
	defer mu.Unlock()

	// 双重检查
	if isLogin {
		limiter, exists = loginLimiters[ip]
	} else {
		limiter, exists = normalLimiters[ip]
	}
	if exists {
		return limiter
	}

	// 创建新的限流器
	if isLogin {
		// 登录接口：每10秒1个请求，最多突发3个
		limiter = rate.NewLimiter(rate.Every(10*time.Second), loginBurst)
		loginLimiters[ip] = limiter
	} else {
		// 普通接口：每秒2个请求，最多突发5个
		limiter = rate.NewLimiter(rate.Limit(defaultRate), defaultBurst)
		normalLimiters[ip] = limiter
	}

	return limiter
}

// RateLimitMiddleware 限流中间件
func RateLimitMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 判断是否是登录接口
		// 由于strings.EndsWith是Go 1.20及以上版本才支持的函数，
		// 如果你使用的Go版本低于1.20，可以使用strings.HasSuffix替代
		isLogin := strings.HasSuffix(c.FullPath(), "/login")

		// 获取客户端IP
		ip := c.ClientIP()
		limiter := getLimiter(ip, isLogin)

		// 尝试获取令牌
		if !limiter.Allow() {
			log2.Debugf("IP %s 请求过于频繁 [%s]", ip, c.FullPath())

			var message string
			if isLogin {
				message = "登录尝试过于频繁，请10秒后再试"
			} else {
				message = "请求过于频繁，请稍后再试"
			}

			c.AbortWithStatusJSON(429, gin.H{
				"error": message,
			})
			return
		}

		c.Next()
	}
}

// 定期清理不活跃的限流器
func init() {
	go func() {
		for {
			time.Sleep(time.Hour)
			mu.Lock()
			// 清理限流器
			normalLimiters = make(map[string]*rate.Limiter)
			loginLimiters = make(map[string]*rate.Limiter)
			mu.Unlock()
		}
	}()
}
