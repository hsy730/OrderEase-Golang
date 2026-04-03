package cache

import (
	"testing"
	"time"
)

func TestGetCache(t *testing.T) {
	// 测试获取缓存实例
	cache1 := GetCache()
	cache2 := GetCache()
	
	// 验证返回的是同一个实例
	if cache1 != cache2 {
		t.Errorf("Expected same cache instance, got different ones")
	}
	
	// 测试重置缓存后获取新实例
	ResetCache()
	cache3 := GetCache()
	if cache1 == cache3 {
		t.Errorf("Expected different cache instance after reset, got same one")
	}
}

func TestResetCache(t *testing.T) {
	// 先获取缓存实例
	cache1 := GetCache()
	
	// 重置缓存
	ResetCache()
	
	// 验证全局缓存被重置为nil
	if globalCache != nil {
		t.Errorf("Expected globalCache to be nil after reset, got %v", globalCache)
	}
	
	// 再次获取缓存，应该创建新实例
	cache2 := GetCache()
	if cache1 == cache2 {
		t.Errorf("Expected different cache instance after reset, got same one")
	}
}

func TestBuildCacheKey(t *testing.T) {
	// 测试基本构建
	key := BuildCacheKey("prefix", 123)
	expected := "prefix:123"
	if key != expected {
		t.Errorf("Expected key %s, got %s", expected, key)
	}
	
	// 测试带后缀的构建
	key = BuildCacheKey("prefix", 123, "suffix1", "suffix2")
	expected = "prefix:123:suffix1:suffix2"
	if key != expected {
		t.Errorf("Expected key %s, got %s", expected, key)
	}
}

func TestContainsShopID(t *testing.T) {
	// 测试包含店铺ID的情况
	key := "dashboard:order_stats:123:2024-01-01"
	if !containsShopID(key, 123) {
		t.Errorf("Expected key to contain shop ID 123, but it didn't")
	}
	
	// 测试包含店铺ID在末尾的情况
	key = "dashboard:order_stats:123"
	if !containsShopID(key, 123) {
		t.Errorf("Expected key '%s' to contain shop ID 123, but it didn't", key)
	}
	
	// 测试不包含店铺ID的情况
	if containsShopID(key, 456) {
		t.Errorf("Expected key to not contain shop ID 456, but it did")
	}
	
	// 测试边界情况
	key = "123"
	if containsShopID(key, 123) {
		t.Errorf("Expected key '%s' to not contain shop ID 123, but it did", key)
	}
}

func TestInvalidateDashboardCache(t *testing.T) {
	// 重置缓存
	ResetCache()
	c := GetCache()
	
	// 添加一些测试数据
	shopID := int64(123)
	testKeys := []string{
		BuildCacheKey(CacheKeyOrderStats, shopID),
		BuildCacheKey(CacheKeyProductStats, shopID),
		BuildCacheKey(CacheKeyUserStats, shopID),
		BuildCacheKey(CacheKeyOrderStats, shopID, "2024-01-01"),
		BuildCacheKey(CacheKeySalesTrend, shopID, "week"),
	}
	
	for _, key := range testKeys {
		c.Set(key, "test value", 5*time.Minute)
	}
	
	// 添加其他店铺的数据，应该不受影响
	otherShopID := int64(456)
	otherKey := BuildCacheKey(CacheKeyOrderStats, otherShopID)
	c.Set(otherKey, "other value", 5*time.Minute)
	
	// 使缓存失效
	InvalidateDashboardCache(shopID)
	
	// 验证目标店铺的缓存被清除
	for _, key := range testKeys {
		_, found := c.Get(key)
		if found {
			t.Errorf("Expected key %s to be deleted, but it was found", key)
		}
	}
	
	// 验证其他店铺的缓存未被清除
	_, found := c.Get(otherKey)
	if !found {
		t.Errorf("Expected key %s to not be deleted, but it was not found", otherKey)
	}
}

func TestInvalidateDashboardCacheByPrefix(t *testing.T) {
	// 重置缓存
	ResetCache()
	c := GetCache()
	
	// 添加一些测试数据
	shopID := int64(123)
	testKeys := []string{
		BuildCacheKey(CacheKeyOrderStats, shopID),
		BuildCacheKey(CacheKeyOrderStats, shopID, "2024-01-01"),
		BuildCacheKey(CacheKeyProductStats, shopID), // 不同前缀
	}
	
	for _, key := range testKeys {
		c.Set(key, "test value", 5*time.Minute)
	}
	
	// 按前缀使缓存失效
	InvalidateDashboardCacheByPrefix(shopID, CacheKeyOrderStats)
	
	// 验证匹配前缀的缓存被清除
	orderStatsKey := BuildCacheKey(CacheKeyOrderStats, shopID)
	_, found := c.Get(orderStatsKey)
	if found {
		t.Errorf("Expected key %s to be deleted, but it was found", orderStatsKey)
	}
	
	orderStatsDateKey := BuildCacheKey(CacheKeyOrderStats, shopID, "2024-01-01")
	_, found = c.Get(orderStatsDateKey)
	if found {
		t.Errorf("Expected key %s to be deleted, but it was found", orderStatsDateKey)
	}
	
	// 验证不匹配前缀的缓存未被清除
	productStatsKey := BuildCacheKey(CacheKeyProductStats, shopID)
	_, found = c.Get(productStatsKey)
	if !found {
		t.Errorf("Expected key %s to not be deleted, but it was not found", productStatsKey)
	}
}
