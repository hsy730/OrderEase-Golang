package repositories

import (
	"errors"
	"orderease/models"
	"orderease/utils/log2"

	"github.com/bwmarrin/snowflake"
	"gorm.io/gorm"
)

// TagRepository 标签数据访问层
type TagRepository struct {
	DB *gorm.DB
}

// NewTagRepository 创建TagRepository实例
func NewTagRepository(db *gorm.DB) *TagRepository {
	return &TagRepository{DB: db}
}

// Create 创建标签
func (r *TagRepository) Create(tag *models.Tag) error {
	if err := r.DB.Create(tag).Error; err != nil {
		log2.Errorf("Create tag failed: %v", err)
		return errors.New("创建标签失败")
	}
	return nil
}

// Update 更新标签
func (r *TagRepository) Update(tag *models.Tag) error {
	if err := r.DB.Save(tag).Error; err != nil {
		log2.Errorf("Update tag failed: %v", err)
		return errors.New("更新标签失败")
	}
	return nil
}

// Delete 删除标签
func (r *TagRepository) Delete(tag *models.Tag) error {
	if err := r.DB.Delete(tag).Error; err != nil {
		log2.Errorf("Delete tag failed: %v", err)
		return errors.New("删除标签失败")
	}
	return nil
}

// GetByIDAndShopID 根据ID和店铺ID获取标签
func (r *TagRepository) GetByIDAndShopID(id int, shopID uint64) (*models.Tag, error) {
	var tag models.Tag
	err := r.DB.Where("shop_id = ?", shopID).First(&tag, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.New("标签不存在")
	}
	if err != nil {
		log2.Errorf("GetByIDAndShopID failed: %v", err)
		return nil, errors.New("查询标签失败")
	}
	return &tag, nil
}

// GetListByShopID 获取店铺的标签列表
func (r *TagRepository) GetListByShopID(shopID uint64) ([]models.Tag, error) {
	var tags []models.Tag
	err := r.DB.Where("shop_id = ?", shopID).Order("created_at DESC").Find(&tags).Error
	if err != nil {
		log2.Errorf("GetListByShopID failed: %v", err)
		return nil, errors.New("查询标签列表失败")
	}
	return tags, nil
}

// 在 ProductRepository 结构体新增方法
func (r *ProductRepository) GetCurrentProductTags(productID snowflake.ID, shopID uint64) ([]models.Tag, error) {
	var tags []models.Tag
	err := r.DB.Joins("JOIN product_tags ON product_tags.tag_id = tags.id").
		Where("product_tags.product_id = ? AND product_tags.shop_id = ?", productID, shopID).
		Find(&tags).Error
	if err != nil {
		log2.Errorf("GetCurrentProductTags failed: %v", err)
		return nil, err
	}
	return tags, err
}

func (r *ProductRepository) CheckProductsBelongToShop(productIDs []uint, shopID uint64) ([]uint, error) {
	var validIDs []uint
	err := r.DB.Model(&models.Product{}).
		Where("id IN (?) AND shop_id = ?", productIDs, shopID).
		Pluck("id", &validIDs).Error
	return validIDs, err
}

func (r *ProductRepository) GetShopTagsByID(shopID uint64) ([]models.Tag, error) {
	tags := make([]models.Tag, 0)
	err := r.DB.Where("shop_id = ?", shopID).Find(&tags).Error
	if err != nil {
		return nil, err
	}
	return tags, nil
}

// GetUnboundTags 获取商品未绑定的标签（该店铺下未被该商品绑定的标签）
func (r *ProductRepository) GetUnboundTags(productID snowflake.ID, shopID uint64) ([]models.Tag, error) {
	var tags []models.Tag
	err := r.DB.Where("shop_id = ? AND id NOT IN (SELECT tag_id FROM product_tags WHERE product_id = ?)", shopID, productID).
		Find(&tags).Error
	if err != nil {
		log2.Errorf("GetUnboundTags failed: %v", err)
		return nil, err
	}
	return tags, nil
}
