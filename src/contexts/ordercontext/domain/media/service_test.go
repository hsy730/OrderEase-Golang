package media

import (
	"mime/multipart"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	log2 "orderease/utils/log2"
)

func TestNewImageUploadService(t *testing.T) {
	logger := log2.GetLogger()
	service := NewImageUploadService(logger)

	assert.NotNil(t, service)
	assert.NotNil(t, service.logger)
}

func TestValidateImageType(t *testing.T) {
	logger := log2.GetLogger()
	service := NewImageUploadService(logger)

	tests := []struct {
		name           string
		contentType    string
		expectedResult string
		expectError    bool
	}{
		{
			name:           "valid jpeg",
			contentType:    "image/jpeg",
			expectedResult: "image/jpeg",
			expectError:    false,
		},
		{
			name:           "valid jpg (should normalize to jpeg)",
			contentType:    "image/jpg",
			expectedResult: "image/jpeg",
			expectError:    false,
		},
		{
			name:           "valid png",
			contentType:    "image/png",
			expectedResult: "image/png",
			expectError:    false,
		},
		{
			name:           "valid gif",
			contentType:    "image/gif",
			expectedResult: "image/gif",
			expectError:    false,
		},
		{
			name:           "valid webp",
			contentType:    "image/webp",
			expectedResult: "image/webp",
			expectError:    false,
		},
		{
			name:           "invalid pdf",
			contentType:    "application/pdf",
			expectedResult: "",
			expectError:    true,
		},
		{
			name:           "invalid bmp",
			contentType:    "image/bmp",
			expectedResult: "",
			expectError:    true,
		},
		{
			name:           "empty string",
			contentType:    "",
			expectedResult: "",
			expectError:    true,
		},
		{
			name:           "uppercase PNG",
			contentType:    "image/PNG",
			expectedResult: "image/png",
			expectError:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := service.ValidateImageType(tt.contentType)

			if tt.expectError {
				assert.Error(t, err)
				assert.Empty(t, result)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedResult, result)
			}
		})
	}
}

func TestValidateImageSize(t *testing.T) {
	logger := log2.GetLogger()
	service := NewImageUploadService(logger)

	tests := []struct {
		name        string
		fileSize    int64
		maxSize     int64
		expectError bool
	}{
		{
			name:        "file within limit",
			fileSize:    1024,
			maxSize:     2048,
			expectError: false,
		},
		{
			name:        "file exactly at limit",
			fileSize:    2048,
			maxSize:     2048,
			expectError: false,
		},
		{
			name:        "file exceeds limit",
			fileSize:    4096,
			maxSize:     2048,
			expectError: true,
		},
		{
			name:        "empty file",
			fileSize:    0,
			maxSize:     2048,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fileHeader := &multipart.FileHeader{
				Size: tt.fileSize,
			}

			err := service.ValidateImageSize(fileHeader, tt.maxSize)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "文件大小超过限制")
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestGenerateUniqueFileName(t *testing.T) {
	logger := log2.GetLogger()
	service := NewImageUploadService(logger)

	tests := []struct {
		name             string
		entityType       string
		entityID         uint64
		originalFilename string
		prefix           string
		ext              string
	}{
		{
			name:             "shop image",
			entityType:       "shop",
			entityID:         123,
			originalFilename: "image.jpg",
			prefix:           "shop_123_",
			ext:              ".jpg",
		},
		{
			name:             "product image",
			entityType:       "product",
			entityID:         456,
			originalFilename: "photo.png",
			prefix:           "product_456_",
			ext:              ".png",
		},
		{
			name:             "image without extension",
			entityType:       "shop",
			entityID:         789,
			originalFilename: "noext",
			prefix:           "shop_789_",
			ext:              "",
		},
		{
			name:             "image with multiple dots",
			entityType:       "product",
			entityID:         999,
			originalFilename: "my.image.gif",
			prefix:           "product_999_",
			ext:              ".gif",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.GenerateUniqueFileName(tt.entityType, tt.entityID, tt.originalFilename)

			assert.Contains(t, result, tt.prefix)
			assert.True(t, len(result) > len(tt.prefix))
			assert.Contains(t, result, tt.ext)
		})
	}
}

func TestGenerateUniqueFileName_Uniqueness(t *testing.T) {
	logger := log2.GetLogger()
	service := NewImageUploadService(logger)

	filename1 := service.GenerateUniqueFileName("shop", 123, "image.jpg")
	time.Sleep(2 * time.Second)
	filename2 := service.GenerateUniqueFileName("shop", 123, "image.jpg")

	assert.NotEqual(t, filename1, filename2)
}

func TestBuildFilePath(t *testing.T) {
	logger := log2.GetLogger()
	service := NewImageUploadService(logger)

	tests := []struct {
		name      string
		uploadDir string
		filename  string
		expected  string
	}{
		{
			name:      "standard path",
			uploadDir: "uploads",
			filename:  "image.jpg",
			expected:  "uploads/image.jpg",
		},
		{
			name:      "nested directory",
			uploadDir: "static/uploads",
			filename:  "photo.png",
			expected:  "static/uploads/photo.png",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.BuildFilePath(tt.uploadDir, tt.filename)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetUploadMessage(t *testing.T) {
	logger := log2.GetLogger()
	service := NewImageUploadService(logger)

	t.Run("had old image", func(t *testing.T) {
		message := service.GetUploadMessage(true)
		assert.Equal(t, "图片更新成功", message)
	})

	t.Run("no old image", func(t *testing.T) {
		message := service.GetUploadMessage(false)
		assert.Equal(t, "图片上传成功", message)
	})
}

func TestGetOperationType(t *testing.T) {
	logger := log2.GetLogger()
	service := NewImageUploadService(logger)

	t.Run("create operation", func(t *testing.T) {
		opType := service.GetOperationType("图片上传成功")
		assert.Equal(t, "create", opType)
	})

	t.Run("update operation", func(t *testing.T) {
		opType := service.GetOperationType("图片更新成功")
		assert.Equal(t, "update", opType)
	})

	t.Run("unknown operation", func(t *testing.T) {
		opType := service.GetOperationType("其他消息")
		assert.Equal(t, "update", opType)
	})
}

func TestCompressImage(t *testing.T) {
	logger := log2.GetLogger()
	service := NewImageUploadService(logger)

	result, err := service.CompressImage("dummy/path", 1024)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.False(t, result.Success)
}
