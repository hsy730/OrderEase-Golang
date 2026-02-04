package value_objects

import (
	"errors"
	"regexp"
	"strings"
	"unicode"
)

// Password 密码值对象
type Password string

// NewPassword 创建密码值对象（宽松规则：6-20位，字母+数字，支持特殊字符）
// 用于前端用户注册和登录
func NewPassword(password string) (Password, error) {
	// 长度验证：6-20位
	if len(password) < 6 || len(password) > 20 {
		return "", errors.New("密码长度必须在6-20位")
	}

	// 必须包含字母和数字
	hasLetter := regexp.MustCompile(`[a-zA-Z]`).MatchString(password)
	hasDigit := regexp.MustCompile(`\d`).MatchString(password)

	if !hasLetter || !hasDigit {
		return "", errors.New("密码必须包含字母和数字")
	}

	// 特殊字符可选，不做硬性要求

	return Password(password), nil
}

// NewStrictPassword 创建强密码值对象（管理员/店主：8+位，大小写+数字+特殊字符）
// 用于管理员和店主密码
func NewStrictPassword(password string) (Password, error) {
	// 长度验证：至少8位
	if len(password) < 8 {
		return "", errors.New("密码长度至少为8位")
	}

	var (
		hasNumber  bool
		hasLower   bool
		hasUpper   bool
		hasSpecial bool
	)

	for _, char := range password {
		switch {
		case unicode.IsNumber(char):
			hasNumber = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSpecial = true
		}
	}

	if !hasNumber {
		return "", errors.New("密码必须包含数字")
	}
	if !hasLower {
		return "", errors.New("密码必须包含小写字母")
	}
	if !hasUpper {
		return "", errors.New("密码必须包含大写字母")
	}
	if !hasSpecial {
		return "", errors.New("密码必须包含特殊字符")
	}

	return Password(password), nil
}

// NewSimplePassword 创建简单密码（用于前端用户6位字母或数字）
func NewSimplePassword(password string) (Password, error) {
	// 长度验证：6位
	if len(password) != 6 {
		return "", errors.New("密码必须为6位")
	}

	// 必须全是字母或数字
	matched, _ := regexp.MatchString(`^[a-zA-Z0-9]{6}$`, password)
	if !matched {
		return "", errors.New("密码必须为6位字母或数字")
	}

	return Password(password), nil
}

// String 返回密码字符串
func (p Password) String() string {
	return string(p)
}

// Hash 返回密码的哈希值（用于bcrypt）
func (p Password) Hash() (string, error) {
	// 如果已经哈希过，直接返回
	if strings.HasPrefix(string(p), "$2a$") {
		return string(p), nil
	}
	// 否则返回空，由调用方处理哈希
	return "", nil
}

// IsValid 验证密码是否符合格式
func (p Password) IsValid() bool {
	_, err := NewPassword(string(p))
	return err == nil
}

// IsStrictValid 验证密码是否符合强密码格式
func (p Password) IsStrictValid() bool {
	_, err := NewStrictPassword(string(p))
	return err == nil
}
