package product

import (
	"fmt"

	"gorm.io/gorm"
	"orderease/models"
)

// Service 商品领域服务
// 负责跨实体的商品编排逻辑
type Service struct {
	db *gorm.DB
}

// NewService 创建商品领域服务
func NewService(db *gorm.DB) *Service {
	return &Service{db: db}
}

// ValidateForDeletion 验证商品是否可以删除
// 检查是否存在关联的订单项
func (s *Service) ValidateForDeletion(productID uint64) error {
	var orderCount int64
	if err := s.db.Model(&models.OrderItem{}).
		Where("product_id = ?", productID).
		Count(&orderCount).Error; err != nil {
		return fmt.Errorf("检查商品订单关联失败: %v", err)
	}

	if orderCount > 0 {
		return fmt.Errorf("该商品有 %d 个关联订单，不能删除。建议将商品下架而不是删除", orderCount)
	}

	return nil
}

// CanTransitionTo 验证商品状态流转是否合法
func (s *Service) CanTransitionTo(currentStatus, newStatus string) bool {
	// 定义状态流转规则
	transitions := map[string][]string{
		models.ProductStatusPending: {models.ProductStatusOnline},
		models.ProductStatusOnline:  {models.ProductStatusOffline},
		models.ProductStatusOffline: {models.ProductStatusOnline}, // 允许下架后重新上架
	}

	allowedStates, exists := transitions[currentStatus]
	if !exists {
		return false
	}

	for _, allowed := range allowedStates {
		if allowed == newStatus {
			return true
		}
	}

	return false
}

// GetDomainStatusFromModel 从模型状态字符串转换为领域状态
func GetDomainStatusFromModel(status string) ProductStatus {
	switch status {
	case models.ProductStatusOnline:
		return ProductStatusOnline
	case models.ProductStatusOffline:
		return ProductStatusOffline
	case models.ProductStatusPending, "":
		return ProductStatusPending
	default:
		return ProductStatusPending
	}
}

// GetModelStatusFromDomain 从领域状态转换为模型状态字符串
func GetModelStatusFromDomain(status ProductStatus) string {
	switch status {
	case ProductStatusOnline:
		return models.ProductStatusOnline
	case ProductStatusOffline:
		return models.ProductStatusOffline
	case ProductStatusPending:
		return models.ProductStatusPending
	default:
		return models.ProductStatusPending
	}
}
