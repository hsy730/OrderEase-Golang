package utils

import (
	"regexp"
	"orderease/domain/shared/value_objects"
)

// ValidatePassword 验证密码强度（管理员/店主强密码规则）
// 内部调用值对象的 NewStrictPassword 进行验证
func ValidatePassword(password string) error {
	_, err := value_objects.NewStrictPassword(password)
	if err != nil {
		return err
	}
	return nil
}

// ValidatePhoneWithRegex 验证中国大陆手机号格式
func ValidatePhoneWithRegex(phone string) bool {
	matched, _ := regexp.MatchString(`^1[3-9]\d{9}$`, phone)
	return matched
}
