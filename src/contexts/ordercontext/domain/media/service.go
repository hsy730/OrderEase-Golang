package media

import (
	"fmt"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"

	"orderease/utils/log2"
)

// ImageUploadService 图片上传领域服务
type ImageUploadService struct {
	logger *log2.Logger
}

// NewImageUploadService 创建图片上传服务
func NewImageUploadService(logger *log2.Logger) *ImageUploadService {
	return &ImageUploadService{
		logger: logger,
	}
}

// AllowedImageTypes 允许的图片类型
var AllowedImageTypes = map[string]bool{
	"image/jpeg": true,
	"image/jpg":  true,
	"image/png":  true,
	"image/gif":  true,
	"image/webp": true,
}

// ValidateImageType 验证图片类型
// 返回 (contentType, error) - 标准化后的内容类型或错误
func (s *ImageUploadService) ValidateImageType(contentType string) (string, error) {
	// 标准化 content-type (处理 image/jpg -> image/jpeg)
	normalizedType := strings.ToLower(contentType)
	if normalizedType == "image/jpg" {
		normalizedType = "image/jpeg"
	}

	if !AllowedImageTypes[normalizedType] {
		return "", fmt.Errorf("不支持的文件类型: %s", contentType)
	}
	return normalizedType, nil
}

// ValidateImageSize 验证图片大小（通过限制读取大小）
func (s *ImageUploadService) ValidateImageSize(fileHeader *multipart.FileHeader, maxSizeBytes int64) error {
	if fileHeader.Size > maxSizeBytes {
		return fmt.Errorf("文件大小超过限制: %d 字节 (最大 %d 字节)", fileHeader.Size, maxSizeBytes)
	}
	return nil
}

// GenerateUniqueFileName 生成唯一的文件名
// entityType: "shop" 或 "product"
// entityID: 实体 ID
// originalFilename: 原始文件名
func (s *ImageUploadService) GenerateUniqueFileName(entityType string, entityID uint64, originalFilename string) string {
	ext := filepath.Ext(originalFilename)
	timestamp := time.Now().Unix()
	return fmt.Sprintf("%s_%d_%d%s", entityType, entityID, timestamp, ext)
}

// CreateUploadDir 创建上传目录
func (s *ImageUploadService) CreateUploadDir(uploadDir string) error {
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		return fmt.Errorf("创建上传目录失败: %w", err)
	}
	return nil
}

// RemoveOldImage 删除旧图片
func (s *ImageUploadService) RemoveOldImage(imageURL string) error {
	if imageURL == "" {
		return nil
	}

	oldImagePath := strings.TrimPrefix(imageURL, "/")
	if err := os.Remove(oldImagePath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("删除旧图片失败: %w", err)
	}
	return nil
}

// SaveUploadedFile 保存上传的文件
func (s *ImageUploadService) SaveUploadedFile(c interface {
	SaveUploadedFile(*multipart.FileHeader, string) error
}, file *multipart.FileHeader, filePath string) error {
	if err := c.SaveUploadedFile(file, filePath); err != nil {
		return fmt.Errorf("保存文件失败: %w", err)
	}
	return nil
}

// BuildFilePath 构建文件路径
func (s *ImageUploadService) BuildFilePath(uploadDir, filename string) string {
	return fmt.Sprintf("%s/%s", uploadDir, filename)
}

// CompressImageResult 压缩结果
type CompressImageResult struct {
	OriginalSize int64
	CompressedSize int64
	Success bool
}

// CompressImage 压缩图片（委托给 utils 函数，但封装在服务中）
// 注意：实际压缩实现在 utils.CompressImage，这里提供领域服务接口
func (s *ImageUploadService) CompressImage(filePath string, maxSize int64) (*CompressImageResult, error) {
	// 这里需要导入 utils.CompressImage
	// 为了避免循环依赖，暂时返回未实现的结果
	// 实际实现需要在 handlers 中继续使用 utils.CompressImage
	// 或者将 utils.CompressImage 移动到这里
	return &CompressImageResult{
		Success: false,
	}, nil
}

// ValidateImageURL 验证图片 URL 是否有效
func (s *ImageUploadService) ValidateImageURL(imageURL string, folder string) error {
	if imageURL == "" {
		return fmt.Errorf("图片URL不能为空")
	}

	// 检查文件名格式
	if !strings.HasPrefix(imageURL, folder) {
		return fmt.Errorf("图片路径不正确")
	}

	// 检查文件扩展名
	ext := filepath.Ext(imageURL)
	validExts := map[string]bool{
		".jpg":  true,
		".jpeg": true,
		".png":  true,
		".gif":  true,
		".webp": true,
	}
	if !validExts[strings.ToLower(ext)] {
		return fmt.Errorf("不支持的图片格式: %s", ext)
	}

	return nil
}

// GetUploadMessage 根据是否有旧图片生成消息
func (s *ImageUploadService) GetUploadMessage(hadOldImage bool) string {
	if hadOldImage {
		return "图片更新成功"
	}
	return "图片上传成功"
}

// GetOperationType 获取操作类型
func (s *ImageUploadService) GetOperationType(message string) string {
	if message == "图片上传成功" {
		return "create"
	}
	return "update"
}
