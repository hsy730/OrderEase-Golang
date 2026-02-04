package product

import (
	"testing"
	"time"

	"github.com/bwmarrin/snowflake"
	"github.com/stretchr/testify/assert"
	"orderease/models"
)

// ==================== Constructor Tests ====================

func TestNewProduct(t *testing.T) {
	shopID := snowflake.ID(123)
	name := "测试商品"
	price := 99.99
	stock := 100

	product := NewProduct(shopID, name, price, stock)

	assert.NotNil(t, product)
	assert.Equal(t, shopID, product.ShopID())
	assert.Equal(t, name, product.Name())
	assert.Equal(t, price, product.Price())
	assert.Equal(t, stock, product.Stock())
	assert.Equal(t, ProductStatusPending, product.Status())
	assert.Equal(t, "", product.Description())
	assert.Equal(t, "", product.ImageURL())
	assert.Empty(t, product.OptionCategories())
	assert.False(t, product.createdAt.IsZero())
	assert.False(t, product.updatedAt.IsZero())
}

func TestNewProductWithDefaults(t *testing.T) {
	shopID := snowflake.ID(123)
	name := "测试商品"
	price := 99.99
	stock := 100
	description := "这是测试商品描述"
	imageURL := "http://example.com/image.jpg"
	optionCategories := []models.ProductOptionCategory{
		{ID: 1, Name: "杯型", Options: []models.ProductOption{
			{ID: 1, Name: "大杯", PriceAdjustment: 500},
			{ID: 2, Name: "中杯", PriceAdjustment: 0},
		}},
	}

	product := NewProductWithDefaults(shopID, name, price, stock, description, imageURL, optionCategories)

	assert.NotNil(t, product)
	assert.Equal(t, shopID, product.ShopID())
	assert.Equal(t, name, product.Name())
	assert.Equal(t, price, product.Price())
	assert.Equal(t, stock, product.Stock())
	assert.Equal(t, ProductStatusPending, product.Status())
	assert.Equal(t, description, product.Description())
	assert.Equal(t, imageURL, product.ImageURL())
	assert.Len(t, product.OptionCategories(), 1)
	assert.False(t, product.createdAt.IsZero())
	assert.False(t, product.updatedAt.IsZero())
}

// ==================== Getter Tests ====================

func TestProduct_Getters(t *testing.T) {
	product := NewProduct(123, "测试商品", 99.99, 100)
	testID := snowflake.ID(456)
	product.SetID(testID)

	assert.Equal(t, testID, product.ID())
	assert.Equal(t, snowflake.ID(123), product.ShopID())
	assert.Equal(t, "测试商品", product.Name())
	assert.Equal(t, 99.99, product.Price())
	assert.Equal(t, 100, product.Stock())
	assert.Equal(t, "", product.Description())
	assert.Equal(t, "", product.ImageURL())
	assert.Equal(t, ProductStatusPending, product.Status())
	assert.Empty(t, product.OptionCategories())
	assert.False(t, product.CreatedAt().IsZero())
	assert.False(t, product.UpdatedAt().IsZero())
}

// ==================== Setter Tests ====================

func TestProduct_Setters(t *testing.T) {
	product := NewProduct(123, "测试商品", 99.99, 100)

	// Test SetID
	testID := snowflake.ID(456)
	product.SetID(testID)
	assert.Equal(t, testID, product.ID())

	// Test SetName
	product.SetName("新商品名")
	assert.Equal(t, "新商品名", product.Name())

	// Test SetDescription
	product.SetDescription("新描述")
	assert.Equal(t, "新描述", product.Description())

	// Test SetPrice
	product.SetPrice(199.99)
	assert.Equal(t, 199.99, product.Price())

	// Test SetStock
	product.SetStock(200)
	assert.Equal(t, 200, product.Stock())

	// Test SetImageURL
	product.SetImageURL("http://example.com/new-image.jpg")
	assert.Equal(t, "http://example.com/new-image.jpg", product.ImageURL())

	// Test SetStatus
	product.SetStatus(ProductStatusOnline)
	assert.Equal(t, ProductStatusOnline, product.Status())

	// Test SetOptionCategories
	categories := []models.ProductOptionCategory{{ID: 1, Name: "杯型"}}
	product.SetOptionCategories(categories)
	assert.Len(t, product.OptionCategories(), 1)

	// Test SetCreatedAt
	now := time.Now()
	product.SetCreatedAt(now)
	assert.Equal(t, now, product.CreatedAt())

	// Test SetUpdatedAt
	updated := time.Now()
	product.SetUpdatedAt(updated)
	assert.Equal(t, updated, product.UpdatedAt())
}

// ==================== Business Logic Tests ====================

func TestProduct_IsOnline(t *testing.T) {
	tests := []struct {
		name     string
		status   ProductStatus
		expected bool
	}{
		{
			name:     "online status - true",
			status:   ProductStatusOnline,
			expected: true,
		},
		{
			name:     "offline status - false",
			status:   ProductStatusOffline,
			expected: false,
		},
		{
			name:     "pending status - false",
			status:   ProductStatusPending,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			product := NewProduct(123, "测试商品", 99.99, 100)
			product.SetStatus(tt.status)
			got := product.IsOnline()
			assert.Equal(t, tt.expected, got)
		})
	}
}

func TestProduct_IsOffline(t *testing.T) {
	tests := []struct {
		name     string
		status   ProductStatus
		expected bool
	}{
		{
			name:     "offline status - true",
			status:   ProductStatusOffline,
			expected: true,
		},
		{
			name:     "online status - false",
			status:   ProductStatusOnline,
			expected: false,
		},
		{
			name:     "pending status - false",
			status:   ProductStatusPending,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			product := NewProduct(123, "测试商品", 99.99, 100)
			product.SetStatus(tt.status)
			got := product.IsOffline()
			assert.Equal(t, tt.expected, got)
		})
	}
}

func TestProduct_IsPending(t *testing.T) {
	tests := []struct {
		name     string
		status   ProductStatus
		expected bool
	}{
		{
			name:     "pending status - true",
			status:   ProductStatusPending,
			expected: true,
		},
		{
			name:     "online status - false",
			status:   ProductStatusOnline,
			expected: false,
		},
		{
			name:     "offline status - false",
			status:   ProductStatusOffline,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			product := NewProduct(123, "测试商品", 99.99, 100)
			product.SetStatus(tt.status)
			got := product.IsPending()
			assert.Equal(t, tt.expected, got)
		})
	}
}

func TestProduct_InStock(t *testing.T) {
	tests := []struct {
		name     string
		stock    int
		expected bool
	}{
		{
			name:     "has stock - true",
			stock:    100,
			expected: true,
		},
		{
			name:     "no stock - false",
			stock:    0,
			expected: false,
		},
		{
			name:     "negative stock - false",
			stock:    -10,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			product := NewProduct(123, "测试商品", 99.99, tt.stock)
			got := product.InStock()
			assert.Equal(t, tt.expected, got)
		})
	}
}

func TestProduct_HasEnoughStock(t *testing.T) {
	tests := []struct {
		name     string
		stock    int
		quantity int
		expected bool
	}{
		{
			name:     "enough stock - true",
			stock:    100,
			quantity: 50,
			expected: true,
		},
		{
			name:     "exact stock - true",
			stock:    50,
			quantity: 50,
			expected: true,
		},
		{
			name:     "not enough stock - false",
			stock:    30,
			quantity: 50,
			expected: false,
		},
		{
			name:     "zero stock - false",
			stock:    0,
			quantity: 1,
			expected: false,
		},
		{
			name:     "zero quantity - true",
			stock:    10,
			quantity: 0,
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			product := NewProduct(123, "测试商品", 99.99, tt.stock)
			got := product.HasEnoughStock(tt.quantity)
			assert.Equal(t, tt.expected, got)
		})
	}
}

func TestProduct_DecreaseStock(t *testing.T) {
	tests := []struct {
		name         string
		initialStock int
		quantity     int
		expectedStock int
	}{
		{
			name:         "normal decrease",
			initialStock: 100,
			quantity:     30,
			expectedStock: 70,
		},
		{
			name:         "decrease all stock",
			initialStock: 50,
			quantity:     50,
			expectedStock: 0,
		},
		{
			name:         "decrease more than stock - no change",
			initialStock: 20,
			quantity:     50,
			expectedStock: 20,
		},
		{
			name:         "decrease zero",
			initialStock: 100,
			quantity:     0,
			expectedStock: 100,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			product := NewProduct(123, "测试商品", 99.99, tt.initialStock)
			product.DecreaseStock(tt.quantity)
			assert.Equal(t, tt.expectedStock, product.Stock())
		})
	}
}

func TestProduct_IncreaseStock(t *testing.T) {
	tests := []struct {
		name         string
		initialStock int
		quantity     int
		expectedStock int
	}{
		{
			name:         "normal increase",
			initialStock: 100,
			quantity:     30,
			expectedStock: 130,
		},
		{
			name:         "increase zero",
			initialStock: 100,
			quantity:     0,
			expectedStock: 100,
		},
		{
			name:         "increase from zero",
			initialStock: 0,
			quantity:     50,
			expectedStock: 50,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			product := NewProduct(123, "测试商品", 99.99, tt.initialStock)
			product.IncreaseStock(tt.quantity)
			assert.Equal(t, tt.expectedStock, product.Stock())
		})
	}
}

func TestProduct_Sanitize(t *testing.T) {
	tests := []struct {
		name             string
		inputName        string
		inputDescription string
		expectedName     string
		expectedDesc     string
	}{
		{
			name:             "normal text",
			inputName:        "测试商品",
			inputDescription: "这是描述",
			expectedName:     "测试商品",
			expectedDesc:     "这是描述",
		},
		{
			name:             "text with script tag - HTML escaped",
			inputName:        "<script>alert('xss')</script>商品",
			inputDescription: "描述<script>alert('xss')</script>",
			expectedName:     "&lt;script&gt;alert(&#39;xss&#39;)&lt;/script&gt;商品",
			expectedDesc:     "描述&lt;script&gt;alert(&#39;xss&#39;)&lt;/script&gt;",
		},
		{
			name:             "empty strings",
			inputName:        "",
			inputDescription: "",
			expectedName:     "",
			expectedDesc:     "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			product := NewProduct(123, tt.inputName, 99.99, 100)
			product.SetDescription(tt.inputDescription)
			product.Sanitize()
			assert.Equal(t, tt.expectedName, product.Name())
			assert.Equal(t, tt.expectedDesc, product.Description())
		})
	}
}

// ==================== Model Conversion Tests ====================

func TestProduct_ToModel(t *testing.T) {
	productID := snowflake.ID(123)
	shopID := snowflake.ID(456)
	optionCategories := []models.ProductOptionCategory{
		{ID: 1, Name: "杯型", Options: []models.ProductOption{
			{ID: 1, Name: "大杯", PriceAdjustment: 500},
		}},
	}

	product := NewProductWithDefaults(
		shopID,
		"测试商品",
		99.99,
		100,
		"这是描述",
		"http://example.com/image.jpg",
		optionCategories,
	)
	product.SetID(productID)
	product.SetStatus(ProductStatusOnline)

	model := product.ToModel()

	assert.Equal(t, productID, model.ID)
	assert.Equal(t, shopID, model.ShopID)
	assert.Equal(t, "测试商品", model.Name)
	assert.Equal(t, "这是描述", model.Description)
	assert.Equal(t, 99.99, model.Price)
	assert.Equal(t, 100, model.Stock)
	assert.Equal(t, "http://example.com/image.jpg", model.ImageURL)
	assert.Equal(t, string(ProductStatusOnline), model.Status)
	assert.Len(t, model.OptionCategories, 1)
}

func TestProductFromModel(t *testing.T) {
	productID := snowflake.ID(123)
	shopID := snowflake.ID(456)
	optionCategories := []models.ProductOptionCategory{
		{ID: 1, Name: "杯型", Options: []models.ProductOption{
			{ID: 1, Name: "大杯", PriceAdjustment: 500},
		}},
	}

	model := &models.Product{
		ID:               productID,
		ShopID:           shopID,
		Name:             "测试商品",
		Description:      "这是描述",
		Price:            99.99,
		Stock:            100,
		ImageURL:         "http://example.com/image.jpg",
		Status:           string(ProductStatusOnline),
		OptionCategories: optionCategories,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	product := ProductFromModel(model)

	assert.Equal(t, productID, product.ID())
	assert.Equal(t, shopID, product.ShopID())
	assert.Equal(t, "测试商品", product.Name())
	assert.Equal(t, "这是描述", product.Description())
	assert.Equal(t, 99.99, product.Price())
	assert.Equal(t, 100, product.Stock())
	assert.Equal(t, "http://example.com/image.jpg", product.ImageURL())
	assert.Equal(t, ProductStatusOnline, product.Status())
	assert.Len(t, product.OptionCategories(), 1)
}

// ==================== Service Tests ====================

func TestGetDomainStatusFromModel(t *testing.T) {
	tests := []struct {
		name     string
		status   string
		expected ProductStatus
	}{
		{
			name:     "online status",
			status:   models.ProductStatusOnline,
			expected: ProductStatusOnline,
		},
		{
			name:     "offline status",
			status:   models.ProductStatusOffline,
			expected: ProductStatusOffline,
		},
		{
			name:     "pending status",
			status:   models.ProductStatusPending,
			expected: ProductStatusPending,
		},
		{
			name:     "empty status - default to pending",
			status:   "",
			expected: ProductStatusPending,
		},
		{
			name:     "invalid status - default to pending",
			status:   "invalid",
			expected: ProductStatusPending,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetDomainStatusFromModel(tt.status)
			assert.Equal(t, tt.expected, got)
		})
	}
}

func TestGetModelStatusFromDomain(t *testing.T) {
	tests := []struct {
		name     string
		status   ProductStatus
		expected string
	}{
		{
			name:     "online status",
			status:   ProductStatusOnline,
			expected: models.ProductStatusOnline,
		},
		{
			name:     "offline status",
			status:   ProductStatusOffline,
			expected: models.ProductStatusOffline,
		},
		{
			name:     "pending status",
			status:   ProductStatusPending,
			expected: models.ProductStatusPending,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetModelStatusFromDomain(tt.status)
			assert.Equal(t, tt.expected, got)
		})
	}
}

func TestService_CanTransitionTo(t *testing.T) {
	service := &Service{}

	tests := []struct {
		name         string
		currentStatus string
		newStatus    string
		allowed      bool
	}{
		{
			name:         "pending to online - allowed",
			currentStatus: models.ProductStatusPending,
			newStatus:    models.ProductStatusOnline,
			allowed:      true,
		},
		{
			name:         "online to offline - allowed",
			currentStatus: models.ProductStatusOnline,
			newStatus:    models.ProductStatusOffline,
			allowed:      true,
		},
		{
			name:         "offline to online - allowed",
			currentStatus: models.ProductStatusOffline,
			newStatus:    models.ProductStatusOnline,
			allowed:      true,
		},
		{
			name:         "pending to offline - not allowed",
			currentStatus: models.ProductStatusPending,
			newStatus:    models.ProductStatusOffline,
			allowed:      false,
		},
		{
			name:         "online to pending - not allowed",
			currentStatus: models.ProductStatusOnline,
			newStatus:    models.ProductStatusPending,
			allowed:      false,
		},
		{
			name:         "invalid status - not allowed",
			currentStatus: "invalid",
			newStatus:    models.ProductStatusOnline,
			allowed:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := service.CanTransitionTo(tt.currentStatus, tt.newStatus)
			assert.Equal(t, tt.allowed, got)
		})
	}
}
