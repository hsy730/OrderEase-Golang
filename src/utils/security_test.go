package utils

import (
	"testing"
)

func TestValidateImageURL(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		folder  string
		wantErr bool
	}{
		{"valid product", "product_123456789_987654321.jpg", "product", false},
		{"valid shop", "shop_123456789_987654321.png", "shop", false},
		{"invalid extension", "product_123_456.bmp", "product", true},
		{"missing timestamp", "product_987654321.jpg", "product", true},
		{"extra characters", "product_123_456_extra.jpg", "product", true},
		{"path traversal", "../product_123_456.jpg", "product", true},
		{"invalid folder type", "category_123_456.jpg", "category", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateImageURL(tt.url, tt.folder)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateImageURL(%q, %q) error = %v, wantErr %v",
					tt.url, tt.folder, err, tt.wantErr)
			}
		})
	}
}