package models

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/bwmarrin/snowflake"
	"github.com/stretchr/testify/assert"
)

func TestOrderStatusActionStruct(t *testing.T) {
	action := OrderStatusAction{
		Name:            "接单",
		NextStatus:      2,
		NextStatusLabel: "已接单",
	}

	assert.Equal(t, "接单", action.Name)
	assert.Equal(t, 2, action.NextStatus)
	assert.Equal(t, "已接单", action.NextStatusLabel)
}

func TestOrderStatusStruct(t *testing.T) {
	actions := []OrderStatusAction{
		{
			Name:            "接单",
			NextStatus:      1,
			NextStatusLabel: "已接单",
		},
	}

	status := OrderStatus{
		Value:   0,
		Label:   "待处理",
		Type:    "warning",
		IsFinal: false,
		Actions: actions,
	}

	assert.Equal(t, 0, status.Value)
	assert.Equal(t, "待处理", status.Label)
	assert.Equal(t, "warning", status.Type)
	assert.False(t, status.IsFinal)
	assert.Len(t, status.Actions, 1)
}

func TestOrderStatusFlow_Value(t *testing.T) {
	flow := OrderStatusFlow{
		Statuses: []OrderStatus{
			{
				Value:   0,
				Label:   "待处理",
				Type:    "warning",
				IsFinal: false,
				Actions: []OrderStatusAction{},
			},
		},
	}

	value, err := flow.Value()
	assert.NoError(t, err)
	assert.NotNil(t, value)

	// Should be valid JSON
	var decoded map[string]interface{}
	err = json.Unmarshal(value.([]byte), &decoded)
	assert.NoError(t, err)
	assert.Contains(t, decoded, "statuses")
}

func TestOrderStatusFlow_Scan(t *testing.T) {
	jsonData := `{
		"statuses": [
			{
				"value": 0,
				"label": "待处理",
				"type": "warning",
				"isFinal": false,
				"actions": []
			}
		]
	}`

	var flow OrderStatusFlow
	err := flow.Scan([]byte(jsonData))
	assert.NoError(t, err)
	assert.Len(t, flow.Statuses, 1)
	assert.Equal(t, 0, flow.Statuses[0].Value)
	assert.Equal(t, "待处理", flow.Statuses[0].Label)
}

func TestOrderStatusFlow_ScanInvalidType(t *testing.T) {
	var flow OrderStatusFlow
	err := flow.Scan("not a byte slice")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "type assertion")
}

func TestShopStruct(t *testing.T) {
	now := time.Now()
	shop := Shop{
		ID:            snowflake.ID(123),
		Name:          "Test Shop",
		OwnerUsername:  "shopowner",
		OwnerPassword:  "hashedpassword",
		ContactPhone:   "13800138000",
		ContactEmail:   "shop@example.com",
		Address:        "123 Shop Street",
		ImageURL:       "https://example.com/shop.jpg",
		Description:    "A test shop",
		CreatedAt:      now,
		UpdatedAt:      now,
		ValidUntil:     now.AddDate(1, 0, 0),
	}

	assert.Equal(t, snowflake.ID(123), shop.ID)
	assert.Equal(t, "Test Shop", shop.Name)
	assert.Equal(t, "shopowner", shop.OwnerUsername)
	assert.Equal(t, "hashedpassword", shop.OwnerPassword)
	assert.Equal(t, "13800138000", shop.ContactPhone)
	assert.Equal(t, "shop@example.com", shop.ContactEmail)
	assert.Equal(t, "123 Shop Street", shop.Address)
	assert.Equal(t, "https://example.com/shop.jpg", shop.ImageURL)
	assert.Equal(t, "A test shop", shop.Description)
}

func TestShopStructWithProducts(t *testing.T) {
	product := Product{
		ID:    snowflake.ID(456),
		ShopID: snowflake.ID(123),
		Name:   "Test Product",
		Price:  99.99,
	}

	shop := Shop{
		ID:       snowflake.ID(123),
		Name:     "Test Shop",
		Products: []Product{product},
	}

	assert.Len(t, shop.Products, 1)
	assert.Equal(t, "Test Product", shop.Products[0].Name)
}

func TestShopStructWithTags(t *testing.T) {
	tag := Tag{
		ID:     1,
		ShopID: snowflake.ID(123),
		Name:   "Electronics",
	}

	shop := Shop{
		ID:    snowflake.ID(123),
		Name:  "Test Shop",
		Tags:   []Tag{tag},
	}

	assert.Len(t, shop.Tags, 1)
	assert.Equal(t, "Electronics", shop.Tags[0].Name)
}

func TestShopStructWithOrderStatusFlow(t *testing.T) {
	flow := OrderStatusFlow{
		Statuses: []OrderStatus{
			{
				Value:   0,
				Label:   "待处理",
				Type:    "warning",
				IsFinal: false,
				Actions: []OrderStatusAction{},
			},
		},
	}

	shop := Shop{
		ID:             snowflake.ID(123),
		Name:           "Test Shop",
		OrderStatusFlow: flow,
	}

	assert.Len(t, shop.OrderStatusFlow.Statuses, 1)
	assert.Equal(t, 0, shop.OrderStatusFlow.Statuses[0].Value)
}

func TestShopStructEmpty(t *testing.T) {
	shop := Shop{}

	assert.Zero(t, shop.ID)
	assert.Empty(t, shop.Name)
	assert.Empty(t, shop.OwnerUsername)
	assert.Empty(t, shop.OwnerPassword)
	assert.Empty(t, shop.ContactPhone)
	assert.Empty(t, shop.ContactEmail)
	assert.Empty(t, shop.Address)
	assert.Empty(t, shop.ImageURL)
	assert.Empty(t, shop.Description)
	assert.True(t, shop.ValidUntil.IsZero())
	assert.Nil(t, shop.Products)
	assert.Nil(t, shop.Tags)
}

func TestOrderStatusFlow_ValueScanRoundtrip(t *testing.T) {
	originalFlow := OrderStatusFlow{
		Statuses: []OrderStatus{
			{
				Value:   0,
				Label:   "待处理",
				Type:    "warning",
				IsFinal: false,
				Actions: []OrderStatusAction{
					{
						Name:            "接单",
						NextStatus:      1,
						NextStatusLabel: "已接单",
					},
				},
			},
		},
	}

	value, err := originalFlow.Value()
	assert.NoError(t, err)

	var scannedFlow OrderStatusFlow
	err = scannedFlow.Scan(value)
	assert.NoError(t, err)

	assert.Equal(t, len(originalFlow.Statuses), len(scannedFlow.Statuses))
	assert.Equal(t, originalFlow.Statuses[0].Value, scannedFlow.Statuses[0].Value)
	assert.Equal(t, originalFlow.Statuses[0].Label, scannedFlow.Statuses[0].Label)
}
