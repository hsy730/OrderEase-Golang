package utils

import (
	"fmt"
	"unicode"
)

// ValidatePassword 验证密码强度
func ValidatePassword(password string) error {
	if len(password) < 8 {
		return fmt.Errorf("密码长度至少为8位")
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
		return fmt.Errorf("密码必须包含数字")
	}
	if !hasLower {
		return fmt.Errorf("密码必须包含小写字母")
	}
	if !hasUpper {
		return fmt.Errorf("密码必须包含大写字母")
	}
	if !hasSpecial {
		return fmt.Errorf("密码必须包含特殊字符")
	}

	return nil
}
