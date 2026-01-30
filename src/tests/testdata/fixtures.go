package testdata

import (
	"time"

	"orderease/models"
	"orderease/utils"

	"github.com/bwmarrin/snowflake"
)

// NewMockUser 创建测试用户
func NewMockUser() *models.User {
	return &models.User{
		ID:       utils.GenerateSnowflakeID(),
		Name:     "test_user",
		Password: "HashedPassword123",
		Phone:    "13800138000",
		Address:  "测试地址",
		Role:     models.UserRolePublic,
		Type:     "delivery",
	}
}

// NewMockAdmin 创建测试管理员
func NewMockAdmin() *models.Admin {
	return &models.Admin{
		ID:       utils.GenerateSnowflakeID(),
		Username: "test_admin",
		Password: "HashedPassword123",
	}
}

// NewMockProduct 创建测试商品
func NewMockProduct() *models.Product {
	shopID := utils.GenerateSnowflakeID()
	return &models.Product{
		ID:          utils.GenerateSnowflakeID(),
		ShopID:      shopID,
		Name:        "测试商品",
		Description: "这是一个测试商品",
		Price:       10000, // 100.00
		Image:       "http://example.com/product.jpg",
		Status:      "online",
		Stock:       100,
		CategoryID:  1,
	}
}

// NewMockShop 创建测试店铺
func NewMockShop() *models.Shop {
	return &models.Shop{
		ID:          utils.GenerateSnowflakeID(),
		Name:        "测试店铺",
		Description: "这是一个测试店铺",
		Password:    "hashed_password",
		Status:      "active",
		ValidUntil:  time.Now().AddDate(1, 0, 0), // 1年后过期
	}
}

// NewMockOrder 创建测试订单
func NewMockOrder() *models.Order {
	userID := utils.GenerateSnowflakeID()
	shopID := utils.GenerateSnowflakeID()
	return &models.Order{
		ID:         utils.GenerateSnowflakeID(),
		UserID:     uint64(userID),
		ShopID:     shopID,
		Status:     0, // Pending
		TotalPrice: 10000,
		Address:    "测试地址",
		Remark:     "测试备注",
		CreatedAt:  time.Now(),
	}
}

// NewMockOrderItem 创建测试订单项
func NewMockOrderItem(orderID, productID snowflake.ID) *models.OrderItem {
	return &models.OrderItem{
		ID:        utils.GenerateSnowflakeID(),
		OrderID:   orderID,
		ProductID: productID,
		Quantity:  2,
		Price:     5000, // 50.00
	}
}

// NewMockTag 创建测试标签
func NewMockTag() *models.Tag {
	return &models.Tag{
		ID:        utils.GenerateSnowflakeID(),
		Name:      "测试标签",
		Color:     "#FF0000",
		Icon:      "tag-icon",
	}
}

// NewMockTempToken 创建测试临时令牌
func NewMockTempToken(shopID snowflake.ID) *models.TempToken {
	return &models.TempToken{
		ShopID:    shopID,
		Token:     "123456",
		UserID:    uint64(utils.GenerateSnowflakeID()),
		ExpiresAt: time.Now().Add(1 * time.Hour),
	}
}

// NewMockProductTag 创建测试商品标签关联
func NewMockProductTag(productID, tagID snowflake.ID) *models.ProductTag {
	return &models.ProductTag{
		ProductID: productID,
		TagID:     tagID,
	}
}
