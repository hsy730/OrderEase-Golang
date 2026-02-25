package models

import (
	"testing"
	"time"

	"github.com/bwmarrin/snowflake"
	"github.com/stretchr/testify/assert"
)

func TestProductOptionCategoryStruct(t *testing.T) {
	now := time.Now()
	category := ProductOptionCategory{
		ID:          snowflake.ID(123),
		ProductID:    snowflake.ID(456),
		Name:         "Size",
		IsRequired:   true,
		IsMultiple:   false,
		DisplayOrder: 1,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	assert.Equal(t, snowflake.ID(123), category.ID)
	assert.Equal(t, snowflake.ID(456), category.ProductID)
	assert.Equal(t, "Size", category.Name)
	assert.True(t, category.IsRequired)
	assert.False(t, category.IsMultiple)
	assert.Equal(t, 1, category.DisplayOrder)
}

func TestProductOptionCategoryStructWithOptions(t *testing.T) {
	option := ProductOption{
		ID:         snowflake.ID(789),
		CategoryID: snowflake.ID(123),
		Name:       "Small",
	}

	category := ProductOptionCategory{
		ID:      snowflake.ID(123),
		Options: []ProductOption{option},
	}

	assert.Len(t, category.Options, 1)
	assert.Equal(t, "Small", category.Options[0].Name)
}

func TestProductOptionCategoryStructEmpty(t *testing.T) {
	category := ProductOptionCategory{}

	assert.Zero(t, category.ID)
	assert.Zero(t, category.ProductID)
	assert.Empty(t, category.Name)
	assert.False(t, category.IsRequired)
	assert.False(t, category.IsMultiple)
	assert.Zero(t, category.DisplayOrder)
	assert.Nil(t, category.Options)
}

func TestProductOptionStruct(t *testing.T) {
	now := time.Now()
	option := ProductOption{
		ID:              snowflake.ID(123),
		CategoryID:      snowflake.ID(456),
		Name:            "Small",
		PriceAdjustment: 0.5,
		DisplayOrder:    1,
		IsDefault:       true,
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	assert.Equal(t, snowflake.ID(123), option.ID)
	assert.Equal(t, snowflake.ID(456), option.CategoryID)
	assert.Equal(t, "Small", option.Name)
	assert.Equal(t, 0.5, option.PriceAdjustment)
	assert.Equal(t, 1, option.DisplayOrder)
	assert.True(t, option.IsDefault)
}

func TestProductOptionStructWithCategory(t *testing.T) {
	category := ProductOptionCategory{
		ID:   snowflake.ID(456),
		Name:  "Size",
	}

	option := ProductOption{
		ID:       snowflake.ID(123),
		CategoryID: snowflake.ID(456),
		Name:     "Small",
		Category: &category,
	}

	assert.NotNil(t, option.Category)
	assert.Equal(t, "Size", option.Category.Name)
}

func TestProductOptionStructEmpty(t *testing.T) {
	option := ProductOption{}

	assert.Zero(t, option.ID)
	assert.Zero(t, option.CategoryID)
	assert.Empty(t, option.Name)
	assert.Zero(t, option.PriceAdjustment)
	assert.Zero(t, option.DisplayOrder)
	assert.False(t, option.IsDefault)
	assert.Nil(t, option.Category)
}

func TestProductOption_NegativePriceAdjustment(t *testing.T) {
	option := ProductOption{
		ID:              snowflake.ID(123),
		CategoryID:      snowflake.ID(456),
		Name:            "Small",
		PriceAdjustment: -1.0,
	}

	assert.Equal(t, -1.0, option.PriceAdjustment)
}

func TestProductOption_MultipleSelect(t *testing.T) {
	option := ProductOption{
		ID:         snowflake.ID(123),
		CategoryID: snowflake.ID(456),
		Name:       "Red",
	}

	category := ProductOptionCategory{
		ID:        snowflake.ID(456),
		Name:      "Color",
		IsMultiple: true,
		Options:   []ProductOption{option},
	}

	assert.True(t, category.IsMultiple)
	assert.Len(t, category.Options, 1)
}
