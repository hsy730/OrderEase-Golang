package value_objects

import (
	"errors"
	"regexp"
)

// 预编译正则表达式（性能优化）
var phoneRegex = regexp.MustCompile(`^1[3-9]\d{9}$`)

// Phone 手机号值对象
type Phone string

// NewPhone 创建手机号值对象，带验证
func NewPhone(phone string) (Phone, error) {
	// 空字符串检查（手机号可选）
	if phone == "" {
		return Phone(""), nil
	}

	// 验证手机号格式：1开头，11位数字
	if !phoneRegex.MatchString(phone) {
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
	return phoneRegex.MatchString(string(p))
}

// IsEmpty 检查是否为空
func (p Phone) IsEmpty() bool {
	return p == ""
}

// Masked 返回脱敏的手机号（中间4位显示为*）
// 例如: 13812345678 -> 138****5678
func (p Phone) Masked() string {
	s := string(p)
	if len(s) != 11 {
		return s
	}
	return s[:3] + "****" + s[7:]
}

// Carrier 返回运营商标识（基于号段）
// 返回: "移动", "联通", "电信", "未知"
func (p Phone) Carrier() string {
	if p.IsEmpty() || len(p) != 11 {
		return "未知"
	}
	prefix := string(p)[:3]

	// 移动号段
	mobilePrefixes := []string{"134", "135", "136", "137", "138", "139",
		"147", "150", "151", "152", "157", "158", "159",
		"172", "178", "182", "183", "184", "187", "188", "198"}
	for _, v := range mobilePrefixes {
		if prefix == v {
			return "移动"
		}
	}

	// 联通号段
	unicomPrefixes := []string{"130", "131", "132", "145", "155", "156",
		"166", "171", "175", "176", "185", "186"}
	for _, v := range unicomPrefixes {
		if prefix == v {
			return "联通"
		}
	}

	// 电信号段
	telecomPrefixes := []string{"133", "149", "153", "173", "177",
		"180", "181", "189", "191", "199"}
	for _, v := range telecomPrefixes {
		if prefix == v {
			return "电信"
		}
	}

	return "未知"
}
