package utils

import (
	"fmt"
	"html"
	"orderease/models"
	"orderease/utils/log2"
	"path/filepath"
	"regexp"
	"strings"
)

// 防止XSS的字符串清理函数
func SanitizeString(input string) string {
	// HTML转义
	escaped := html.EscapeString(input)
	// 移除可能的脚本标签
	escaped = regexp.MustCompile(`<script[^>]*>.*?</script>`).ReplaceAllString(escaped, "")
	// 移除可能的事件处理器
	escaped = regexp.MustCompile(`\bon\w+\s*=`).ReplaceAllString(escaped, "")
	return escaped
}

// 验证图片URL
func ValidateImageURL(imageURL string, folder string) error {
	// 检查是否为空
	if imageURL == "" {
		return nil
	}

	// 确保路径以 /uploads/products/ 开头
	if !strings.HasPrefix(imageURL, fmt.Sprintf("/uploads/%ss/", folder)) {
		return fmt.Errorf("invalid image path prefix:" + imageURL)
	}

	// 验证文件扩展名
	ext := strings.ToLower(filepath.Ext(imageURL))
	validExts := map[string]bool{
		".jpg":  true,
		".jpeg": true,
		".png":  true,
		".gif":  true,
	}
	if !validExts[ext] {
		return fmt.Errorf("invalid image extension: %s", ext)
	}

	// 检查路径中是否包含危险字符
	if strings.Contains(imageURL, "..") {
		return fmt.Errorf("path traversal attempt detected")
	}

	// 验证文件名格式
	validName := regexp.MustCompile(fmt.Sprintf(`^/uploads/%ss/%s_\d+_\d+\.(jpg|jpeg|png|gif)$`, folder, folder))
	if !validName.MatchString(imageURL) {
		return fmt.Errorf("invalid image filename format")
	}

	return nil
}

// 清理订单数据
func SanitizeOrder(order *models.Order) {
	order.Status = SanitizeString(order.Status)
	order.Remark = SanitizeString(order.Remark)
}

// 清理商品数据
func SanitizeProduct(product *models.Product) {
	product.Name = SanitizeString(product.Name)
	product.Description = SanitizeString(product.Description)
	if err := ValidateImageURL(product.ImageURL, "products"); err != nil {
		product.ImageURL = "" // 如果图片URL无效，清空它
		log2.Errorf("Invalid image URL detected: %v", err)
	}
}
