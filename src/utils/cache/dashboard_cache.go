package cache

import (
	"fmt"
	"orderease/utils/log2"
	"time"

	"github.com/patrickmn/go-cache"
)

var (
	globalCache *cache.Cache
)

func GetCache() *cache.Cache {
	if globalCache == nil {
		globalCache = cache.New(5*time.Minute, 10*time.Minute)
	}
	return globalCache
}

func ResetCache() {
	globalCache = nil
}

const (
	CacheKeyOrderStats      = "dashboard:order_stats"
	CacheKeyProductStats    = "dashboard:product_stats"
	CacheKeyUserStats       = "dashboard:user_stats"
	CacheKeyOrderEfficiency = "dashboard:efficiency"
	CacheKeySalesTrend      = "dashboard:sales_trend"
	CacheKeyHotProducts     = "dashboard:hot_products"
	CacheKeyRecentOrders    = "dashboard:recent_orders"
)

func BuildCacheKey(prefix string, shopID int64, suffix ...string) string {
	key := fmt.Sprintf("%s:%d", prefix, shopID)
	for _, s := range suffix {
		key += ":" + s
	}
	return key
}

func InvalidateDashboardCache(shopID int64) {
	c := GetCache()

	keysToDelete := []string{
		BuildCacheKey(CacheKeyOrderStats, shopID),
		BuildCacheKey(CacheKeyProductStats, shopID),
		BuildCacheKey(CacheKeyUserStats, shopID),
		BuildCacheKey(CacheKeyOrderEfficiency, shopID),
		BuildCacheKey(CacheKeySalesTrend, shopID),
		BuildCacheKey(CacheKeyHotProducts, shopID),
		BuildCacheKey(CacheKeyRecentOrders, shopID),
		BuildCacheKey(CacheKeyOrderStats, shopID, time.Now().Format("2006-01-02")),
		BuildCacheKey(CacheKeyUserStats, shopID, time.Now().Format("2006-01-02")),
		BuildCacheKey(CacheKeyOrderEfficiency, shopID, time.Now().Format("2006-01-02")),
		BuildCacheKey(CacheKeySalesTrend, shopID, "week"),
		BuildCacheKey(CacheKeySalesTrend, shopID, "month"),
		BuildCacheKey(CacheKeySalesTrend, shopID, "year"),
	}

	for _, key := range keysToDelete {
		c.Delete(key)
	}

	keys := make([]string, 0, len(c.Items()))
	for key := range c.Items() {
		if containsShopID(key, shopID) {
			keys = append(keys, key)
		}
	}
	for _, key := range keys {
		c.Delete(key)
	}

	log2.Infof("Dashboard cache invalidated for shop: %d", shopID)
}

func containsShopID(key string, shopID int64) bool {
	// 检查键是否以 :shopID 结尾
	endPattern := fmt.Sprintf(":%d", shopID)
	if len(key) >= len(endPattern) && key[len(key)-len(endPattern):] == endPattern {
		return true
	}
	
	// 检查键是否包含 :shopID:
	middlePattern := fmt.Sprintf(":%d:", shopID)
	for i := 0; i <= len(key)-len(middlePattern); i++ {
		if key[i:i+len(middlePattern)] == middlePattern {
			return true
		}
	}
	return false
}

func InvalidateDashboardCacheByPrefix(shopID int64, keyPrefix string) {
	c := GetCache()
	keys := make([]string, 0, len(c.Items()))
	for key := range c.Items() {
		if containsShopID(key, shopID) && len(key) >= len(keyPrefix) && key[:len(keyPrefix)] == keyPrefix {
			keys = append(keys, key)
		}
	}
	for _, key := range keys {
		c.Delete(key)
	}
	log2.Debugf("Dashboard cache invalidated for shop: %d, prefix: %s", shopID, keyPrefix)
}
