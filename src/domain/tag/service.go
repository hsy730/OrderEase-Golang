package tag

import (
	"github.com/bwmarrin/snowflake"
	"gorm.io/gorm"
	"orderease/models"
)

// Service 标签领域服务
type Service struct {
	db *gorm.DB
}

// NewService 创建标签领域服务
func NewService(db *gorm.DB) *Service {
	return &Service{
		db: db,
	}
}

// UpdateProductTagsDTO 更新商品标签 DTO
type UpdateProductTagsDTO struct {
	CurrentTags []models.Tag
	NewTagIDs   []int
	ProductID   snowflake.ID
	ShopID      snowflake.ID
}

// UpdateProductTagsResult 更新结果
type UpdateProductTagsResult struct {
	AddedCount   int
	DeletedCount int
}

// UpdateProductTags 更新商品标签关联（领域服务方法）
// 计算当前标签和新标签的差异，执行添加和删除操作
func (s *Service) UpdateProductTags(dto UpdateProductTagsDTO) (*UpdateProductTagsResult, error) {
	// 计算差异
	currentTagMap := make(map[int]bool)
	for _, tag := range dto.CurrentTags {
		currentTagMap[tag.ID] = true
	}

	newTagMap := make(map[int]bool)
	for _, tagID := range dto.NewTagIDs {
		newTagMap[tagID] = true
	}

	// 准备操作数据
	var tagsToAdd []models.ProductTag
	var tagsToDelete []int

	// 计算需要添加的标签
	for _, tagID := range dto.NewTagIDs {
		if !currentTagMap[tagID] {
			tagsToAdd = append(tagsToAdd, models.ProductTag{
				ProductID: dto.ProductID,
				TagID:     tagID,
				ShopID:    dto.ShopID,
			})
		}
	}

	// 计算需要删除的标签
	for _, tag := range dto.CurrentTags {
		if !newTagMap[tag.ID] {
			tagsToDelete = append(tagsToDelete, tag.ID)
		}
	}

	// 执行事务操作
	err := s.db.Transaction(func(tx *gorm.DB) error {
		if len(tagsToAdd) > 0 {
			if err := tx.Create(&tagsToAdd).Error; err != nil {
				return err
			}
		}

		if len(tagsToDelete) > 0 {
			if err := tx.Where("product_id = ? AND tag_id IN (?)", dto.ProductID, tagsToDelete).
				Delete(&models.ProductTag{}).Error; err != nil {
				return err
			}
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return &UpdateProductTagsResult{
		AddedCount:   len(tagsToAdd),
		DeletedCount: len(tagsToDelete),
	}, nil
}
