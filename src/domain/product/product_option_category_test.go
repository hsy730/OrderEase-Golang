package product

import (
	"testing"
	"time"

	"orderease/domain/shared"

	"github.com/stretchr/testify/assert"
)

func TestNewProductOptionCategory(t *testing.T) {
	tests := []struct {
		name         string
		productID    shared.ID
		categoryName string
		isRequired   bool
		isMultiple   bool
		displayOrder int
		wantErr      bool
		errMsg       string
		validate     func(*testing.T, *ProductOptionCategory)
	}{
		{
			name:         "valid required single option category",
			productID:    shared.ID(123),
			categoryName: "尺寸",
			isRequired:   true,
			isMultiple:   false,
			displayOrder: 1,
			wantErr:      false,
			validate: func(t *testing.T, c *ProductOptionCategory) {
				assert.Equal(t, shared.ID(0), c.ID)
				assert.Equal(t, shared.ID(123), c.ProductID)
				assert.Equal(t, "尺寸", c.Name)
				assert.True(t, c.IsRequired)
				assert.False(t, c.IsMultiple)
				assert.Equal(t, 1, c.DisplayOrder)
				assert.False(t, c.CreatedAt.IsZero())
				assert.False(t, c.UpdatedAt.IsZero())
			},
		},
		{
			name:         "valid optional multiple option category",
			productID:    shared.ID(456),
			categoryName: "配料",
			isRequired:   false,
			isMultiple:   true,
			displayOrder: 2,
			wantErr:      false,
		},
		{
			name:         "zero display order valid",
			productID:    shared.ID(123),
			categoryName: "颜色",
			isRequired:   true,
			isMultiple:   false,
			displayOrder: 0,
			wantErr:      false,
		},
		{
			name:         "empty productID",
			productID:    shared.ID(0),
			categoryName: "尺寸",
			isRequired:   true,
			isMultiple:   false,
			displayOrder: 1,
			wantErr:      true,
			errMsg:       "商品ID不能为空",
		},
		{
			name:         "empty name",
			productID:    shared.ID(123),
			categoryName: "",
			isRequired:   true,
			isMultiple:   false,
			displayOrder: 1,
			wantErr:      true,
			errMsg:       "类别名称不能为空",
		},
		{
			name:         "negative display order valid",
			productID:    shared.ID(123),
			categoryName: "测试",
			isRequired:   false,
			isMultiple:   false,
			displayOrder: -1,
			wantErr:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewProductOptionCategory(tt.productID, tt.categoryName, tt.isRequired, tt.isMultiple, tt.displayOrder)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
				assert.Nil(t, got)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, got)
				if tt.validate != nil {
					tt.validate(t, got)
				}
			}
		})
	}
}

func TestNewProductOption(t *testing.T) {
	tests := []struct {
		name            string
		categoryID      shared.ID
		optionName      string
		priceAdjustment float64
		isDefault       bool
		displayOrder    int
		wantErr         bool
		errMsg          string
		validate        func(*testing.T, *ProductOption)
	}{
		{
			name:            "valid default option",
			categoryID:      shared.ID(123),
			optionName:      "大",
			priceAdjustment: 5.0,
			isDefault:       true,
			displayOrder:    1,
			wantErr:         false,
			validate: func(t *testing.T, o *ProductOption) {
				assert.Equal(t, shared.ID(0), o.ID)
				assert.Equal(t, shared.ID(123), o.CategoryID)
				assert.Equal(t, "大", o.Name)
				assert.Equal(t, 5.0, o.PriceAdjustment)
				assert.True(t, o.IsDefault)
				assert.Equal(t, 1, o.DisplayOrder)
				assert.False(t, o.CreatedAt.IsZero())
				assert.False(t, o.UpdatedAt.IsZero())
			},
		},
		{
			name:            "valid non-default option with negative adjustment",
			categoryID:      shared.ID(456),
			optionName:      "小",
			priceAdjustment: -2.5,
			isDefault:       false,
			displayOrder:    0,
			wantErr:         false,
		},
		{
			name:            "zero price adjustment valid",
			categoryID:      shared.ID(123),
			optionName:      "中",
			priceAdjustment: 0,
			isDefault:       false,
			displayOrder:    2,
			wantErr:         false,
		},
		{
			name:            "empty categoryID",
			categoryID:      shared.ID(0),
			optionName:      "大",
			priceAdjustment: 5.0,
			isDefault:       true,
			displayOrder:    1,
			wantErr:         true,
			errMsg:          "类别ID不能为空",
		},
		{
			name:            "empty name",
			categoryID:      shared.ID(123),
			optionName:      "",
			priceAdjustment: 5.0,
			isDefault:       true,
			displayOrder:    1,
			wantErr:         true,
			errMsg:          "选项名称不能为空",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewProductOption(tt.categoryID, tt.optionName, tt.priceAdjustment, tt.isDefault, tt.displayOrder)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
				assert.Nil(t, got)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, got)
				if tt.validate != nil {
					tt.validate(t, got)
				}
			}
		})
	}
}

func TestProductOptionCategory_Timestamps(t *testing.T) {
	before := time.Now()
	category, err := NewProductOptionCategory(shared.ID(123), "测试", true, false, 1)
	after := time.Now()

	assert.NoError(t, err)
	assert.True(t, category.CreatedAt.After(before) || category.CreatedAt.Equal(before))
	assert.True(t, category.CreatedAt.Before(after) || category.CreatedAt.Equal(after))
	assert.True(t, category.UpdatedAt.After(before) || category.UpdatedAt.Equal(before))
	assert.True(t, category.UpdatedAt.Before(after) || category.UpdatedAt.Equal(after))
}

func TestProductOption_Timestamps(t *testing.T) {
	before := time.Now()
	option, err := NewProductOption(shared.ID(123), "测试", 1.5, false, 1)
	after := time.Now()

	assert.NoError(t, err)
	assert.True(t, option.CreatedAt.After(before) || option.CreatedAt.Equal(before))
	assert.True(t, option.CreatedAt.Before(after) || option.CreatedAt.Equal(after))
	assert.True(t, option.UpdatedAt.After(before) || option.UpdatedAt.Equal(before))
	assert.True(t, option.UpdatedAt.Before(after) || option.UpdatedAt.Equal(after))
}
