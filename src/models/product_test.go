package models

import (
	"testing"

	"github.com/bwmarrin/snowflake"
	"github.com/stretchr/testify/assert"
)

func TestProductStatusConstants(t *testing.T) {
	t.Run("pending status", func(t *testing.T) {
		assert.Equal(t, "pending", ProductStatusPending)
	})

	t.Run("online status", func(t *testing.T) {
		assert.Equal(t, "online", ProductStatusOnline)
	})

	t.Run("offline status", func(t *testing.T) {
		assert.Equal(t, "offline", ProductStatusOffline)
	})
}

func TestProductStruct(t *testing.T) {
	product := Product{
		ID:          snowflake.ID(123),
		ShopID:      snowflake.ID(456),
		Name:        "Test Product",
		Description: "Test Description",
		Price:       99.99,
		Stock:       10,
		ImageURL:    "https://example.com/image.jpg",
		Status:      ProductStatusOnline,
	}

	assert.Equal(t, snowflake.ID(123), product.ID)
	assert.Equal(t, snowflake.ID(456), product.ShopID)
	assert.Equal(t, "Test Product", product.Name)
	assert.Equal(t, "Test Description", product.Description)
	assert.Equal(t, 99.99, product.Price)
	assert.Equal(t, 10, product.Stock)
	assert.Equal(t, "https://example.com/image.jpg", product.ImageURL)
	assert.Equal(t, ProductStatusOnline, product.Status)
}

func TestProductTagStruct(t *testing.T) {
	productTag := ProductTag{
		ProductID: snowflake.ID(123),
		TagID:     1,
		ShopID:    snowflake.ID(456),
	}

	assert.Equal(t, snowflake.ID(123), productTag.ProductID)
	assert.Equal(t, 1, productTag.TagID)
	assert.Equal(t, snowflake.ID(456), productTag.ShopID)
}
