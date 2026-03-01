package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func TestRateLimitMiddleware_NormalRequest(t *testing.T) {
	router := gin.New()
	router.Use(RateLimitMiddleware())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-Forwarded-For", "192.168.1.1")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestRateLimitMiddleware_MultipleRequests(t *testing.T) {
	router := gin.New()
	router.Use(RateLimitMiddleware())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	for i := 0; i < defaultBurst; i++ {
		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("X-Forwarded-For", "192.168.1.2")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	}
}

func TestRateLimitMiddleware_RateLimited(t *testing.T) {
	router := gin.New()
	router.Use(RateLimitMiddleware())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	for i := 0; i < defaultBurst+10; i++ {
		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("X-Forwarded-For", "192.168.1.3")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if i < defaultBurst {
			assert.Equal(t, http.StatusOK, w.Code)
		} else {
			assert.Equal(t, 429, w.Code)
		}
	}
}

func TestRateLimitMiddleware_DifferentIPs(t *testing.T) {
	router := gin.New()
	router.Use(RateLimitMiddleware())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	for i := 0; i < 5; i++ {
		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("X-Forwarded-For", "192.168.1.10")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	}

	for i := 0; i < 5; i++ {
		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("X-Forwarded-For", "192.168.1.11")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	}
}

func TestRateLimitMiddleware_LoginEndpoint(t *testing.T) {
	router := gin.New()
	router.Use(RateLimitMiddleware())
	router.POST("/api/login", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	for i := 0; i < loginBurst; i++ {
		req := httptest.NewRequest("POST", "/api/login", nil)
		req.Header.Set("X-Forwarded-For", "192.168.1.20")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	}

	req := httptest.NewRequest("POST", "/api/login", nil)
	req.Header.Set("X-Forwarded-For", "192.168.1.20")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, 429, w.Code)
}

func TestRateLimitMiddleware_LoginErrorMessage(t *testing.T) {
	router := gin.New()
	router.Use(RateLimitMiddleware())
	router.POST("/api/login", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	for i := 0; i < loginBurst+1; i++ {
		req := httptest.NewRequest("POST", "/api/login", nil)
		req.Header.Set("X-Forwarded-For", "192.168.1.21")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
	}

	req := httptest.NewRequest("POST", "/api/login", nil)
	req.Header.Set("X-Forwarded-For", "192.168.1.21")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, 429, w.Code)
	assert.Contains(t, w.Body.String(), "登录尝试过于频繁")
}

func TestRateLimitMiddleware_NormalErrorMessage(t *testing.T) {
	router := gin.New()
	router.Use(RateLimitMiddleware())
	router.GET("/api/data", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	for i := 0; i < defaultBurst+1; i++ {
		req := httptest.NewRequest("GET", "/api/data", nil)
		req.Header.Set("X-Forwarded-For", "192.168.1.22")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
	}

	req := httptest.NewRequest("GET", "/api/data", nil)
	req.Header.Set("X-Forwarded-For", "192.168.1.22")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, 429, w.Code)
	assert.Contains(t, w.Body.String(), "请求过于频繁")
}

func TestGetLimiter_SameIPNormal(t *testing.T) {
	ip := "192.168.1.100"

	limiter1 := getLimiter(ip, false)
	limiter2 := getLimiter(ip, false)

	assert.NotNil(t, limiter1)
	assert.NotNil(t, limiter2)
	assert.Equal(t, limiter1, limiter2)
}

func TestGetLimiter_SameIPLogin(t *testing.T) {
	ip := "192.168.1.101"

	limiter1 := getLimiter(ip, true)
	limiter2 := getLimiter(ip, true)

	assert.NotNil(t, limiter1)
	assert.NotNil(t, limiter2)
	assert.Equal(t, limiter1, limiter2)
}

func TestGetLimiter_DifferentTypes(t *testing.T) {
	ip := "192.168.1.102"

	limiterNormal := getLimiter(ip, false)
	limiterLogin := getLimiter(ip, true)

	assert.NotNil(t, limiterNormal)
	assert.NotNil(t, limiterLogin)
	assert.NotEqual(t, limiterNormal, limiterLogin)
}

func TestGetLimiter_ConcurrentAccess(t *testing.T) {
	done := make(chan bool)

	for i := 0; i < 10; i++ {
		go func(index int) {
			ip := "192.168.1.200"
			limiter := getLimiter(ip, false)
			assert.NotNil(t, limiter)
			done <- true
		}(i)
	}

	for i := 0; i < 10; i++ {
		<-done
	}
}

func TestRateLimitMiddleware_NoIP(t *testing.T) {
	router := gin.New()
	router.Use(RateLimitMiddleware())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestRateLimitConstants(t *testing.T) {
	assert.Equal(t, 5, defaultRate)
	assert.Equal(t, 20, defaultBurst)
	assert.Equal(t, 1, loginRate)
	assert.Equal(t, 3, loginBurst)
}
