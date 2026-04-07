package utils

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// ==================== SanitizeString Tests ====================

func TestSanitizeString(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "plain text",
			input:    "Hello World",
			expected: "Hello World",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "HTML tags - escaped",
			input:    "<script>alert('xss')</script>",
			expected: "&lt;script&gt;alert(&#39;xss&#39;)&lt;/script&gt;",
		},
		{
			name:     "HTML entities - escaped",
			input:    "<div>Content</div>",
			expected: "&lt;div&gt;Content&lt;/div&gt;",
		},
		{
			name:     "event handlers - removed after escape",
			input:    "<img src=x onerror=alert(1)>",
			expected: "&lt;img src=x alert(1)&gt;",
		},
		{
			name:     "mixed content - escaped with script preserved as entities",
			input:    "Hello <b>World</b> <script>evil()</script>",
			expected: "Hello &lt;b&gt;World&lt;/b&gt; &lt;script&gt;evil()&lt;/script&gt;",
		},
		{
			name:     "special characters - escaped to numeric entities",
			input:    "Test & < > \" ' / \\",
			expected: "Test &amp; &lt; &gt; &#34; &#39; / \\",
		},
		{
			name:     "Chinese characters - preserved",
			input:    "测试中文内容",
			expected: "测试中文内容",
		},
		{
			name:     "newline and tabs - preserved",
			input:    "Line1\nLine2\tTabbed",
			expected: "Line1\nLine2\tTabbed",
		},
		{
			name:     "SQL injection attempt - special chars escaped",
			input:    "'; DROP TABLE users; --",
			expected: "&#39;; DROP TABLE users; --",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SanitizeString(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSanitizeString_Security(t *testing.T) {
	t.Run("should escape HTML tags (not remove content)", func(t *testing.T) {
		input := "<script>alert('XSS')</script>"
		result := SanitizeString(input)

		assert.NotContains(t, result, "<")
		assert.NotContains(t, result, ">")
		assert.Contains(t, result, "&lt;")
		assert.Contains(t, result, "&gt;")
	})

	t.Run("should remove event handler attributes", func(t *testing.T) {
		inputs := []string{
			"onclick=evil()",
			"onload=hack()",
			"onerror=bad()",
			"onmouseover=steal()",
		}

		for _, input := range inputs {
			result := SanitizeString(input)
			assert.NotContains(t, strings.ToLower(result), "on")
		}
	})

	t.Run("should handle nested HTML tags by escaping them", func(t *testing.T) {
		input := "<script><script>alert(1)</script></script>"
		result := SanitizeString(input)

		assert.NotContains(t, result, "<")
		assert.NotContains(t, result, ">")
	})
}

// ==================== ValidateImageURL Tests ====================

func TestValidateImageURL(t *testing.T) {
	tests := []struct {
		name      string
		imageURL  string
		folder    string
		wantErr   bool
		errMsg    string
	}{
		{
			name:     "valid product image URL",
			imageURL: "product_1234567890_9876543210.jpg",
			folder:   "product",
			wantErr:  false,
		},
		{
			name:     "valid shop image URL (png)",
			imageURL: "shop_1111111111_2222222222.png",
			folder:   "shop",
			wantErr:  false,
		},
		{
			name:     "valid shop image URL (jpeg)",
			imageURL: "shop_9999999999_8888888888.jpeg",
			folder:   "shop",
			wantErr:  false,
		},
		{
			name:     "valid product image URL (gif)",
			imageURL: "product_123_456.gif",
			folder:   "product",
			wantErr:  false,
		},
		{
			name:     "empty URL should pass validation",
			imageURL: "",
			folder:   "product",
			wantErr:  false,
		},
		{
			name:     "invalid folder type",
			imageURL: "product_123_456.jpg",
			folder:   "invalid_folder",
			wantErr:  true,
			errMsg:   "invalid folder type",
		},
		{
			name:     "missing timestamp in URL",
			imageURL: "product_123.jpg",
			folder:   "product",
			wantErr:  true,
			errMsg:   "invalid image url format",
		},
		{
			name:     "wrong extension",
			imageURL: "product_123_456.exe",
			folder:   "product",
			wantErr:  true,
			errMsg:   "invalid image url format",
		},
		{
			name:     "missing underscore separators",
			imageURL: "product12345678909876543210.jpg",
			folder:   "product",
			wantErr:  true,
			errMsg:   "invalid image url format",
		},
		{
			name:     "URL with path traversal attempt",
			imageURL: "../product_123_456.jpg",
			folder:   "product",
			wantErr:  true,
			errMsg:   "invalid image url format",
		},
		{
			name:     "URL with special characters",
			imageURL: "product_12$%&_456.jpg",
			folder:   "product",
			wantErr:  true,
			errMsg:   "invalid image url format",
		},
		{
			name:     "folder mismatch - shop URL with product folder",
			imageURL: "shop_123_456.jpg",
			folder:   "product",
			wantErr:  true,
			errMsg:   "invalid image url format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateImageURL(tt.imageURL, tt.folder)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateImageURL_EdgeCases(t *testing.T) {
	t.Run("very long numeric IDs", func(t *testing.T) {
		longID := strings.Repeat("1", 20)
		url := "product_" + longID + "_" + longID + ".jpg"
		
		err := ValidateImageURL(url, "product")
		assert.NoError(t, err)
	})

	t.Run("single digit IDs", func(t *testing.T) {
		err := ValidateImageURL("product_1_2.jpg", "product")
		assert.NoError(t, err)
	})

	t.Run("uppercase extension rejected", func(t *testing.T) {
		err := ValidateImageURL("product_123_456.JPG", "product")
		assert.Error(t, err)
	})
}

// ==================== Integration Tests ====================

func TestSecurityUtils_Integration(t *testing.T) {
	t.Run("sanitize then validate workflow", func(t *testing.T) {
		userInput := "<script>alert('xss')</script>product_123_456.jpg"
		
		sanitized := SanitizeString(userInput)

		assert.NotContains(t, sanitized, "<")
		assert.NotContains(t, sanitized, ">")

		err := ValidateImageURL(sanitized, "product")
		assert.Error(t, err)
	})

	t.Run("file upload security check simulation", func(t *testing.T) {
		filename := "image.jpg"
		
		ext := GetFileExtension(filename)
		assert.True(t, IsAllowedImageExt(ext))

		sanitizedFilename := SanitizeString(filename)
		assert.Equal(t, filename, sanitizedFilename)
	})
}
