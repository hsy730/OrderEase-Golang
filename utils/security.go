package utils

import (
	"fmt"
	"html"
	"orderease/models"
	"orderease/utils/log2"
	"regexp"
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
	if imageURL == "" {
		return nil
	}

	// 添加文件夹类型白名单校验
	validFolders := map[string]bool{"product": true, "shop": true}
	if !validFolders[folder] {
		return fmt.Errorf("invalid folder type: %s", folder)
	}

	pattern := fmt.Sprintf(`^%s_\d+_\d+\.(jpg|jpeg|png|gif)$`, folder)
	re := regexp.MustCompile(pattern)

	if !re.MatchString(imageURL) {
		return fmt.Errorf("invalid image url format: %s", imageURL)
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
