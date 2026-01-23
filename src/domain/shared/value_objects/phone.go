package value_objects

import (
	"errors"
	"regexp"
)

// Phone 手机号值对象
type Phone string

// NewPhone 创建手机号值对象，带验证
func NewPhone(phone string) (Phone, error) {
	// 空字符串检查（手机号可选）
	if phone == "" {
		return Phone(""), nil
	}

	// 验证手机号格式：1开头，11位数字
	matched, _ := regexp.MatchString(`^1[3-9]\d{9}$`, phone)
	if !matched {
		return "", errors.New("手机号必须为11位数字且以1开头")
	}

	return Phone(phone), nil
}

// String 返回手机号字符串
func (p Phone) String() string {
	return string(p)
}

// IsValid 验证手机号是否有效
func (p Phone) IsValid() bool {
	if p == "" {
		return true // 空值视为有效（可选字段）
	}
	matched, _ := regexp.MatchString(`^1[3-9]\d{9}$`, string(p))
	return matched
}

// IsEmpty 检查是否为空
func (p Phone) IsEmpty() bool {
	return p == ""
}
