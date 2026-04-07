package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// ==================== GetFileExtension Tests ====================

func TestGetFileExtension(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		expected string
	}{
		{
			name:     "jpg file",
			filename: "photo.jpg",
			expected: ".jpg",
		},
		{
			name:     "jpeg file",
			filename: "image.jpeg",
			expected: ".jpeg",
		},
		{
			name:     "png file",
			filename: "screenshot.png",
			expected: ".png",
		},
		{
			name:     "gif file",
			filename: "animation.gif",
			expected: ".gif",
		},
		{
			name:     "uppercase extension",
			filename: "IMAGE.PNG",
			expected: ".png",
		},
		{
			name:     "mixed case extension",
			filename: "Photo.JPG",
			expected: ".jpg",
		},
		{
			name:     "file with multiple dots",
			filename: "my.photo.2024.jpg",
			expected: ".jpg",
		},
		{
			name:     "file without extension",
			filename: "README",
			expected: "",
		},
		{
			name:     "empty filename",
			filename: "",
			expected: "",
		},
		{
			name:     "hidden file (Unix)",
			filename: ".bashrc",
			expected: ".bashrc",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetFileExtension(tt.filename)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// ==================== IsAllowedImageExt Tests ====================

func TestIsAllowedImageExt(t *testing.T) {
	tests := []struct {
		name     string
		ext      string
		expected bool
	}{
		{
			name:     "jpg - allowed",
			ext:      ".jpg",
			expected: true,
		},
		{
			name:     "jpeg - allowed",
			ext:      ".jpeg",
			expected: true,
		},
		{
			name:     "png - allowed",
			ext:      ".png",
			expected: true,
		},
		{
			name:     "gif - allowed",
			ext:      ".gif",
			expected: true,
		},
		{
			name:     "JPG uppercase - not allowed (case sensitive)",
			ext:      ".JPG",
			expected: false,
		},
		{
			name:     "PNG uppercase - not allowed",
			ext:      ".PNG",
			expected: false,
		},
		{
			name:     "webp - not allowed",
			ext:      ".webp",
			expected: false,
		},
		{
			name:     "svg - not allowed",
			ext:      ".svg",
			expected: false,
		},
		{
			name:     "bmp - not allowed",
			ext:      ".bmp",
			expected: false,
		},
		{
			name:     "empty string - not allowed",
			ext:      "",
			expected: false,
		},
		{
			name:     "no dot prefix - not allowed",
			ext:      "jpg",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsAllowedImageExt(tt.ext)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// ==================== GenerateUniqueFilename Tests ====================

func TestGenerateUniqueFilename(t *testing.T) {
	tests := []struct {
		name           string
		originalName   string
		checkExtension bool
	}{
		{
			name:           "jpg file",
			originalName:   "photo.jpg",
			checkExtension: true,
		},
		{
			name:           "png file",
			originalName:   "image.png",
			checkExtension: true,
		},
		{
			name:           "jpeg file with spaces",
			originalName:   "my photo.jpeg",
			checkExtension: true,
		},
		{
			name:           "gif file",
			originalName:   "animation.gif",
			checkExtension: true,
		},
		{
			name:           "file without extension",
			originalName:   "README",
			checkExtension: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GenerateUniqueFilename(tt.originalName)

			assert.NotEmpty(t, result)
			assert.True(t, len(result) > 10)

			if tt.checkExtension {
				ext := GetFileExtension(result)
				originalExt := GetFileExtension(tt.originalName)
				assert.Equal(t, originalExt, ext)
			}

			assert.Contains(t, result, "_")
		})
	}
}

func TestGenerateUniqueFilename_Uniqueness(t *testing.T) {
	results := make(map[string]bool)

	for i := 0; i < 100; i++ {
		result := GenerateUniqueFilename("test.jpg")

		if results[result] {
			t.Errorf("Duplicate filename generated: %s", result)
		}
		results[result] = true
	}
}

// ==================== Integration Tests ====================

func TestFileUtils_Integration(t *testing.T) {
	t.Run("complete file processing workflow", func(t *testing.T) {
		filename := "user_upload_photo.jpg"
		
		ext := GetFileExtension(filename)
		assert.Equal(t, ".jpg", ext)

		isAllowed := IsAllowedImageExt(ext)
		assert.True(t, isAllowed)

		newName := GenerateUniqueFilename(filename)
		assert.NotEqual(t, filename, newName)
	})

	t.Run("reject disallowed file type", func(t *testing.T) {
		badFilename := "malicious_script.exe"
		
		ext := GetFileExtension(badFilename)
		assert.Equal(t, ".exe", ext)

		isAllowed := IsAllowedImageExt(ext)
		assert.False(t, isAllowed)
	})
}
