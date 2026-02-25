package models

import (
	"testing"
	"time"

	"github.com/bwmarrin/snowflake"
	"github.com/stretchr/testify/assert"
)

func TestTagStruct(t *testing.T) {
	now := time.Now()
	tag := Tag{
		ID:          1,
		ShopID:      snowflake.ID(123),
		Name:        "Electronics",
		Description: "Electronic products",
		CreatedAt:   now,
		UpdatedAt:   now,
		Products:    []Product{},
	}

	assert.Equal(t, 1, tag.ID)
	assert.Equal(t, snowflake.ID(123), tag.ShopID)
	assert.Equal(t, "Electronics", tag.Name)
	assert.Equal(t, "Electronic products", tag.Description)
	assert.NotNil(t, tag.Products)
}

func TestTagStructWithProducts(t *testing.T) {
	product := Product{
		ID:     snowflake.ID(456),
		ShopID: snowflake.ID(123),
		Name:   "Laptop",
		Price:  999.99,
		Stock:  5,
		Status: ProductStatusOnline,
	}

	tag := Tag{
		ID:       1,
		ShopID:   snowflake.ID(123),
		Name:     "Electronics",
		Products: []Product{product},
	}

	assert.Len(t, tag.Products, 1)
	assert.Equal(t, "Laptop", tag.Products[0].Name)
	assert.Equal(t, 999.99, tag.Products[0].Price)
}

func TestTagStructEmpty(t *testing.T) {
	tag := Tag{}

	assert.Zero(t, tag.ID)
	assert.Zero(t, tag.ShopID)
	assert.Empty(t, tag.Name)
	assert.Empty(t, tag.Description)
	assert.Nil(t, tag.Products)
}
