// Package product (service) 提供商品领域服务。
//
// 职责：
//   - 处理商品删除前的关联检查
//   - 验证商品状态流转的合法性
//   - 领域状态与模型状态的转换
//
// 与应用层的区别：
//   - 领域服务关注业务规则验证
//   - 不处理 HTTP、事务、权限等
package product

import (
	"fmt"

	"gorm.io/gorm"
	"orderease/models"
)

// Service 商品领域服务
//
// 职责边界：
//   - 跨实体查询（检查商品订单关联）
//   - 状态流转规则定义和验证
//   - 状态值转换
//
// 依赖：
//   - *gorm.DB: 用于查询关联数据
type Service struct {
	db *gorm.DB
}

// NewService 创建商品领域服务
func NewService(db *gorm.DB) *Service {
	return &Service{db: db}
}

// ValidateForDeletion 验证商品是否可删除
//
// 检查逻辑：
//   - 查询是否存在关联的订单项
//   - 如果有历史订单，建议下架而非删除
//
// 参数：
//   - productID: 商品ID
//
// 返回：
//   - nil:   可以删除
//   - error: 存在关联订单，不可删除
//
// 业务建议：
//   - 有历史订单的商品建议下架（offline）而非删除
//   - 保留历史订单数据完整性
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
//
// 状态流转规则：
//   - pending -> online:  上架
//   - online  -> offline: 下架
//   - offline -> online:  重新上架
//
// 禁止流转：
//   - online  -> pending: 不能直接回到待上架
//   - offline -> pending: 不能直接回到待上架
//
// 参数：
//   - currentStatus: 当前状态
//   - newStatus:     目标状态
//
// 返回：
//   - true:  流转合法
//   - false: 流转不合法
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
