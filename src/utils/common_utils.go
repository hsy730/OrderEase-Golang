package utils

import (
	"bytes"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"os"
	"path/filepath"
)

// 验证图片类型
func IsValidImageType(contentType string) bool {
	validTypes := map[string]bool{
		"image/jpeg": true,
		"image/png":  true,
		"image/gif":  true,
		"image/webp": true,
	}
	return validTypes[contentType]
}

// 压缩图片
// filePath: 图片文件路径
// maxSize: 最大允许大小(字节)
// 返回: 压缩后的文件大小，如果不需要压缩则返回0
func CompressImage(filePath string, maxSize int64) (int64, error) {
	// 获取文件信息
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return 0, err
	}

	// 检查文件大小
	fileSize := fileInfo.Size()
	if fileSize <= maxSize {
		// 不需要压缩
		return 0, nil
	}

	// 打开文件
	file, err := os.Open(filePath)
	if err != nil {
		return 0, err
	}
	defer file.Close()

	// 解码图片
	img, _, err := image.Decode(file)
	if err != nil {
		return 0, err
	}

	// 创建临时缓冲区
	buf := new(bytes.Buffer)

	// 根据文件扩展名选择编码器
	ext := filepath.Ext(filePath)
	var opt interface{}

	switch ext {
	case ".jpg", ".jpeg":
		opt = &jpeg.Options{
			Quality: 70, // 初始质量设置
		}
		if err := jpeg.Encode(buf, img, opt.(*jpeg.Options)); err != nil {
			return 0, err
		}
	case ".png":
		if err := png.Encode(buf, img); err != nil {
			return 0, err
		}
	default:
		// 对于其他格式，不进行压缩
		return 0, nil
	}

	// 检查压缩后的大小
	compressedSize := int64(buf.Len())
	if compressedSize > maxSize {
		// 如果仍然太大，尝试降低JPEG质量
		if ext == ".jpg" || ext == ".jpeg" {
			// 质量逐步降低，直到满足大小要求或达到最低质量
			for quality := 60; quality >= 30; quality -= 10 {
				buf.Reset()
				opt = &jpeg.Options{
					Quality: quality,
				}
				if err := jpeg.Encode(buf, img, opt.(*jpeg.Options)); err != nil {
					return 0, err
				}
				compressedSize = int64(buf.Len())
				if compressedSize <= maxSize {
					break
				}
			}
		}
	}

	// 如果压缩后的大小仍然超过限制，就使用当前压缩结果

	// 保存压缩后的图片
	compressedFile, err := os.Create(filePath)
	if err != nil {
		return 0, err
	}
	defer compressedFile.Close()

	// 写入压缩后的数据
	written, err := io.Copy(compressedFile, buf)
	if err != nil {
		return 0, err
	}

	return written, nil
}
