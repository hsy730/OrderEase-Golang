package utils

import (
	"fmt"
	"html"
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
	validFolders := map[string]bool{"product": true, "shop": true, "avatar": true}
	if !validFolders[folder] {
		return fmt.Errorf("invalid folder type: %s", folder)
	}

	var pattern string
	if folder == "avatar" {
		// 头像文件名格式: {随机字符串}_{时间戳}.{扩展名}
		pattern = `^[a-zA-Z0-9]+_\d+\.(jpg|jpeg|png|gif)$`
	} else {
		// 商品/店铺文件名格式: {folder}_{id}_{时间戳}.{扩展名}
		pattern = fmt.Sprintf(`^%s_\d+_\d+\.(jpg|jpeg|png|gif)$`, folder)
	}
	re := regexp.MustCompile(pattern)

	if !re.MatchString(imageURL) {
		return fmt.Errorf("invalid image url format: %s", imageURL)
	}
	return nil
}

// 注意：SanitizeOrder 已被删除（未被使用）
// 订单备注清理应该在 Domain 层处理（如需要，可在 Order 实体中添加 Sanitize 方法）
