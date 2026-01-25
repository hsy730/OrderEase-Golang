package handlers

import "errors"

// ValidatePaginationParams 验证分页参数
func ValidatePaginationParams(page, pageSize int) error {
	if page < 1 {
		return errors.New("页码必须大于0")
	}
	if pageSize < 1 || pageSize > 100 {
		return errors.New("每页数量必须在1-100之间")
	}
	return nil
}
