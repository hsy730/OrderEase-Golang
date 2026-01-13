package repositories

import (
	"errors"
	"orderease/domain/shop"
	"orderease/domain/shared"
	"orderease/infrastructure/persistence"
	"orderease/models"
	"orderease/utils/log2"

	"gorm.io/gorm"
)

type ShopRepositoryImpl struct {
	db *gorm.DB
}

func NewShopRepository(db *gorm.DB) shop.ShopRepository {
	return &ShopRepositoryImpl{db: db}
}

func (r *ShopRepositoryImpl) Save(s *shop.Shop) error {
	model := persistence.ShopToModel(s)
	if err := r.db.Create(model).Error; err != nil {
		log2.Errorf("保存店铺失败: %v", err)
		return errors.New("保存店铺失败")
	}
	s.ID = shared.ID(model.ID)
	return nil
}

func (r *ShopRepositoryImpl) FindByID(id shared.ID) (*shop.Shop, error) {
	var model models.Shop
	if err := r.db.Preload("Tags").First(&model, id.Value()).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("店铺不存在")
		}
		log2.Errorf("查询店铺失败: %v", err)
		return nil, errors.New("查询店铺失败")
	}
	return persistence.ShopToDomain(model), nil
}

func (r *ShopRepositoryImpl) FindByName(name string) (*shop.Shop, error) {
	var model models.Shop
	if err := r.db.Where("name = ?", name).First(&model).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("店铺不存在")
		}
		log2.Errorf("查询店铺失败: %v", err)
		return nil, errors.New("查询店铺失败")
	}
	return persistence.ShopToDomain(model), nil
}

func (r *ShopRepositoryImpl) FindByOwnerUsername(username string) (*shop.Shop, error) {
	var model models.Shop
	if err := r.db.Where("owner_username = ?", username).First(&model).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("店铺不存在")
		}
		log2.Errorf("查询店铺失败: %v", err)
		return nil, errors.New("查询店铺失败")
	}
	return persistence.ShopToDomain(model), nil
}

func (r *ShopRepositoryImpl) FindAll(page, pageSize int, search string) ([]shop.Shop, int64, error) {
	query := r.db.Model(&models.Shop{})

	if search != "" {
		search = "%" + search + "%"
		query = query.Where("name LIKE ? OR owner_username LIKE ?", search, search)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		log2.Errorf("查询店铺总数失败: %v", err)
		return nil, 0, errors.New("查询店铺总数失败")
	}

	var modelsList []models.Shop
	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Order("created_at DESC").Preload("Tags").Find(&modelsList).Error; err != nil {
		log2.Errorf("查询店铺列表失败: %v", err)
		return nil, 0, errors.New("查询店铺列表失败")
	}

	shops := make([]shop.Shop, len(modelsList))
	for i, m := range modelsList {
		shops[i] = *persistence.ShopToDomain(m)
	}
	return shops, total, nil
}

func (r *ShopRepositoryImpl) Delete(id shared.ID) error {
	if err := r.db.Delete(&models.Shop{}, id.Value()).Error; err != nil {
		log2.Errorf("删除店铺失败: %v", err)
		return errors.New("删除店铺失败")
	}
	return nil
}

func (r *ShopRepositoryImpl) Update(s *shop.Shop) error {
	model := persistence.ShopToModel(s)
	if err := r.db.Save(model).Error; err != nil {
		log2.Errorf("更新店铺失败: %v", err)
		return errors.New("更新店铺失败")
	}
	return nil
}

func (r *ShopRepositoryImpl) Exists(id shared.ID) (bool, error) {
	var count int64
	if err := r.db.Model(&models.Shop{}).Where("id = ?", id.Value()).Count(&count).Error; err != nil {
		log2.Errorf("检查店铺是否存在失败: %v", err)
		return false, errors.New("检查店铺是否存在失败")
	}
	return count > 0, nil
}

type TagRepositoryImpl struct {
	db *gorm.DB
}

func NewTagRepository(db *gorm.DB) shop.TagRepository {
	return &TagRepositoryImpl{db: db}
}

func (r *TagRepositoryImpl) Save(tag *shop.Tag) error {
	model := persistence.TagToModel(tag)
	if err := r.db.Create(model).Error; err != nil {
		log2.Errorf("保存标签失败: %v", err)
		return errors.New("保存标签失败")
	}
	tag.ID = model.ID
	return nil
}

func (r *TagRepositoryImpl) FindByID(id int) (*shop.Tag, error) {
	var model models.Tag
	if err := r.db.First(&model, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("标签不存在")
		}
		log2.Errorf("查询标签失败: %v", err)
		return nil, errors.New("查询标签失败")
	}
	return persistence.TagToDomain(model), nil
}

func (r *TagRepositoryImpl) FindByShopID(shopID shared.ID) ([]shop.Tag, error) {
	var modelsList []models.Tag
	if err := r.db.Where("shop_id = ?", shopID.Value()).Find(&modelsList).Error; err != nil {
		log2.Errorf("查询标签失败: %v", err)
		return nil, errors.New("查询标签失败")
	}

	tags := make([]shop.Tag, len(modelsList))
	for i, m := range modelsList {
		tags[i] = *persistence.TagToDomain(m)
	}
	return tags, nil
}

func (r *TagRepositoryImpl) Delete(id int) error {
	if err := r.db.Delete(&models.Tag{}, id).Error; err != nil {
		log2.Errorf("删除标签失败: %v", err)
		return errors.New("删除标签失败")
	}
	return nil
}

func (r *TagRepositoryImpl) Update(tag *shop.Tag) error {
	model := persistence.TagToModel(tag)
	if err := r.db.Save(model).Error; err != nil {
		log2.Errorf("更新标签失败: %v", err)
		return errors.New("更新标签失败")
	}
	return nil
}
