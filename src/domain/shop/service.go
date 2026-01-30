package shop

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/bwmarrin/snowflake"
	"gorm.io/gorm"
	"orderease/models"
)

// Service Shop 领域服务
// 处理需要多个实体协作或需要基础设施的业务逻辑
type Service struct {
	db *gorm.DB
}

// NewService 创建 Shop 领域服务
func NewService(db *gorm.DB) *Service {
	return &Service{
		db: db,
	}
}

// DeleteShop 删除店铺（业务编排）
// 1. 检查店铺是否可删除（使用 Shop.CanDelete）
// 2. 开启事务
// 3. 删除关联数据（如果有）
// 4. 删除店铺
func (s *Service) DeleteShop(shopID snowflake.ID) error {
	// 查询店铺
	var shopModel models.Shop
	if err := s.db.First(&shopModel, shopID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("店铺不存在")
		}
		return fmt.Errorf("查询店铺失败: %w", err)
	}

	// 转换为领域实体
	shopEntity := ShopFromModel(&shopModel)

	// 检查关联商品数量
	var productCount int64
	if err := s.db.Model(&models.Product{}).Where("shop_id = ?", shopID).Count(&productCount).Error; err != nil {
		return fmt.Errorf("查询关联商品失败: %w", err)
	}

	// 检查关联订单数量
	var orderCount int64
	if err := s.db.Model(&models.Order{}).Where("shop_id = ?", shopID).Count(&orderCount).Error; err != nil {
		return fmt.Errorf("查询关联订单失败: %w", err)
	}

	// 使用领域实体验证是否可删除
	if err := shopEntity.CanDelete(int(productCount), int(orderCount)); err != nil {
		return err
	}

	// 开启事务
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// TODO: 如果需要级联删除关联数据，在这里添加
	// 例如：标签关联、订单状态日志等

	// 删除店铺记录
	if err := tx.Delete(&shopModel).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("删除店铺失败: %w", err)
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("提交事务失败: %w", err)
	}

	return nil
}

// ProcessValidUntil 处理店铺有效期
// 解析用户提供的有效期字符串，如果为空则使用默认值（1年）
func (s *Service) ProcessValidUntil(validUntilStr string) (time.Time, error) {
	// 默认有效期1年
	validUntil := time.Now().AddDate(1, 0, 0)

	// 如果提供了有效期，则解析
	if validUntilStr != "" {
		parsedValidUntil, err := time.Parse(time.RFC3339, validUntilStr)
		if err != nil {
			return time.Time{}, errors.New("无效的有效期格式，请使用 RFC3339 格式（如：2024-01-01T00:00:00Z）")
		}
		validUntil = parsedValidUntil
	}

	return validUntil, nil
}

// ParseOrderStatusFlow 解析订单状态流转配置
// 如果提供了配置则使用提供的配置，否则使用默认配置
func (s *Service) ParseOrderStatusFlow(orderStatusFlow *models.OrderStatusFlow) (models.OrderStatusFlow, error) {
	var flow models.OrderStatusFlow

	// 解析默认订单流转配置
	if err := json.Unmarshal([]byte(models.DefaultOrderStatusFlow), &flow); err != nil {
		return models.OrderStatusFlow{}, errors.New("解析默认订单流转配置失败")
	}

	// 如果提供了订单流转配置，则使用提供的配置
	if orderStatusFlow != nil {
		flow = *orderStatusFlow
	}

	return flow, nil
}
