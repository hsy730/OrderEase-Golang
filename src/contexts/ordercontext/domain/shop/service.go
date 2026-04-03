// Package shop (service) 提供店铺领域服务。
//
// 职责：
//   - 处理店铺删除的业务编排
//   - 有效期处理和默认值设置
//   - 订单状态流转配置解析
//
// 特点：
//   - 部分方法管理事务（如 DeleteShop）
//   - 协调领域实体和基础设施层
//
// 事务边界：
//   - DeleteShop: 内部管理事务
//   - ProcessValidUntil/ParseOrderStatusFlow: 无事务
package shop

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/bwmarrin/snowflake"
	"gorm.io/datatypes"
	"gorm.io/gorm"
	// TODO(DDD-P3): 移除 models 依赖，改用领域内部值对象 + Infrastructure Mapper
	// TODO(DDD-P3): 移除 *gorm.DB 直接注入，改为通过 Repository 接口操作数据
	"orderease/models"
)

// Service 店铺领域服务
//
// 职责边界：
//   - 需要跨实体查询的业务逻辑
//   - 需要事务保证的操作
//   - 复杂数据转换
//
// 依赖：
//   - *gorm.DB: 数据库连接
type Service struct {
	db *gorm.DB
}

// NewService 创建 Shop 领域服务
func NewService(db *gorm.DB) *Service {
	return &Service{
		db: db,
	}
}

// DeleteShop 删除店铺（带事务）
//
// 执行流程：
//   1. 查询店铺信息
//   2. 检查关联商品数量
//   3. 检查关联订单数量
//   4. 使用领域实体验证可删除性
//   5. 开启事务
//   6. 删除店铺（及相关数据）
//   7. 提交事务
//
// 参数：
//   - shopID: 店铺ID
//
// 返回：
//   - nil:   删除成功
//   - error: 验证失败或删除失败
//
// 注意：
//   - 此方法内部管理事务
//   - 有关联数据时无法删除
//   - 删除后无法恢复
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
//
// 参数：
//   - validUntilStr: 有效期字符串（RFC3339 格式），可为空
//
// 返回：
//   - time.Time: 处理后的有效期
//   - error:     解析错误
//
// 默认值：
//   - 如果 validUntilStr 为空，返回当前时间 + 1年
//
// 格式示例：
//   - "2024-12-31T23:59:59Z"
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
func (s *Service) ParseOrderStatusFlow(orderStatusFlow *models.OrderStatusFlow) (models.OrderStatusFlow, error) {
	var flow models.OrderStatusFlow

	if err := json.Unmarshal([]byte(models.DefaultOrderStatusFlow), &flow); err != nil {
		return models.OrderStatusFlow{}, errors.New("解析默认订单流转配置失败")
	}

	if orderStatusFlow != nil {
		flow = *orderStatusFlow
	}

	return flow, nil
}

// ShopUpdates 店铺更新数据（DTO）
type ShopUpdates struct {
	Name            string
	ContactPhone    string
	ContactEmail    string
	Description     string
	Address         string
	Settings        datatypes.JSON
	OrderStatusFlow *models.OrderStatusFlow
	ValidUntil      string
	OwnerUsername   string
	OwnerPassword   *string
}

// UpdateInfo 更新店铺基本信息
func (s *Service) UpdateInfo(shopID snowflake.ID, updates ShopUpdates) (*models.Shop, error) {
	var shopModel models.Shop
	if err := s.db.First(&shopModel, shopID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("店铺不存在")
		}
		return nil, fmt.Errorf("查询店铺失败: %w", err)
	}

	shopEntity := ShopFromModel(&shopModel)

	if updates.Name != "" {
		shopEntity.SetName(updates.Name)
	}
	if updates.ContactPhone != "" {
		shopEntity.SetContactPhone(updates.ContactPhone)
	}
	if updates.ContactEmail != "" {
		shopEntity.SetContactEmail(updates.ContactEmail)
	}
	if updates.Description != "" {
		shopEntity.SetDescription(updates.Description)
	}
	if updates.Address != "" {
		shopEntity.SetAddress(updates.Address)
	}
	if updates.Settings != nil {
		shopEntity.SetSettings(updates.Settings)
	}
	if updates.OrderStatusFlow != nil {
		shopEntity.SetOrderStatusFlow(*updates.OrderStatusFlow)
	} else if len(shopModel.OrderStatusFlow.Statuses) == 0 {
		var defaultFlow models.OrderStatusFlow
		if err := json.Unmarshal([]byte(models.DefaultOrderStatusFlow), &defaultFlow); err == nil {
			shopEntity.SetOrderStatusFlow(defaultFlow)
		}
	}
	if updates.ValidUntil != "" {
		validUntil, err := time.Parse(time.RFC3339, updates.ValidUntil)
		if err != nil {
			return nil, fmt.Errorf("无效的有效期格式")
		}
		shopEntity.SetValidUntil(validUntil)
	}
	if updates.OwnerUsername != "" {
		shopEntity.SetOwnerUsername(updates.OwnerUsername)
	}
	if updates.OwnerPassword != nil && *updates.OwnerPassword != "" {
		shopEntity.SetOwnerPassword(*updates.OwnerPassword)
	}

	updatedModel := shopEntity.ToModel()

	if err := s.db.Save(updatedModel).Error; err != nil {
		return nil, fmt.Errorf("更新店铺失败: %w", err)
	}

	return updatedModel, nil
}

// UpdateOrderStatusFlow 更新店铺订单状态流转配置
func (s *Service) UpdateOrderStatusFlow(shopID snowflake.ID, flow models.OrderStatusFlow) (*models.Shop, error) {
	var shopModel models.Shop
	if err := s.db.First(&shopModel, shopID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("店铺不存在")
		}
		return nil, fmt.Errorf("查询店铺失败: %w", err)
	}

	shopModel.OrderStatusFlow = flow

	if err := s.db.Save(&shopModel).Error; err != nil {
		return nil, fmt.Errorf("更新店铺订单流转状态配置失败: %w", err)
	}

	return &shopModel, nil
}

// UpdatePassword 更新店铺密码
func (s *Service) UpdatePassword(shopID snowflake.ID, newPassword string) error {
	var shopModel models.Shop
	if err := s.db.First(&shopModel, shopID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("店铺账户不存在")
		}
		return fmt.Errorf("查询店铺失败: %w", err)
	}

	shopEntity := ShopFromModel(&shopModel)
	shopEntity.SetOwnerPassword(newPassword)
	updatedModel := shopEntity.ToModel()

	if err := s.db.Save(updatedModel).Error; err != nil {
		return fmt.Errorf("保存新密码失败: %w", err)
	}

	return nil
}
