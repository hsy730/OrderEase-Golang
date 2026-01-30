package repositories

import (
	"errors"
	"orderease/models"
	"orderease/utils/log2"

	"github.com/bwmarrin/snowflake"
	"gorm.io/gorm"
)

// ShopRepository 店铺数据访问层
type ShopRepository struct {
	DB *gorm.DB
}

// NewShopRepository 创建ShopRepository实例
func NewShopRepository(db *gorm.DB) *ShopRepository {
	return &ShopRepository{DB: db}
}

// GetShopByID 根据ID获取店铺
func (r *ShopRepository) GetShopByID(shopID snowflake.ID) (*models.Shop, error) {
	var shop models.Shop
	err := r.DB.First(&shop, shopID).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.New("店铺不存在")
	}
	if err != nil {
		log2.Errorf("GetShopByID failed: %v", err)
		return nil, errors.New("查询店铺失败")
	}
	return &shop, nil
}

// GetShopList 获取店铺列表（分页+搜索）
func (r *ShopRepository) GetShopList(page, pageSize int, search string) ([]models.Shop, int64, error) {
	var shops []models.Shop
	var total int64

	baseQuery := r.DB.Model(&models.Shop{}).Preload("Tags")

	// 如果提供了搜索参数，则添加模糊匹配条件（搜索名称和店主用户名）
	if search != "" {
		searchPattern := "%" + search + "%"
		baseQuery = baseQuery.Where("name LIKE ? OR owner_username LIKE ?", searchPattern, searchPattern)
	}

	// 获取总数
	if err := baseQuery.Count(&total).Error; err != nil {
		log2.Errorf("GetShopList count failed: %v", err)
		return nil, 0, errors.New("获取店铺总数失败")
	}

	// 分页查询
	offset := (page - 1) * pageSize
	if err := baseQuery.Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&shops).Error; err != nil {
		log2.Errorf("GetShopList query failed: %v", err)
		return nil, 0, errors.New("查询店铺列表失败")
	}

	return shops, total, nil
}

// CheckShopNameExists 检查店铺名称是否存在
func (r *ShopRepository) CheckShopNameExists(name string) (bool, error) {
	var count int64
	err := r.DB.Model(&models.Shop{}).Where("name = ?", name).Count(&count).Error
	if err != nil {
		log2.Errorf("CheckShopNameExists failed: %v", err)
		return false, errors.New("检查店铺名称失败")
	}
	return count > 0, nil
}

// GetOrderStatusFlow 获取店铺的订单状态流转配置
func (r *ShopRepository) GetOrderStatusFlow(shopID snowflake.ID) (models.OrderStatusFlow, error) {
	var shop models.Shop
	err := r.DB.Select("order_status_flow").First(&shop, shopID).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return models.OrderStatusFlow{}, errors.New("店铺不存在")
	}
	if err != nil {
		log2.Errorf("GetOrderStatusFlow failed: %v", err)
		return models.OrderStatusFlow{}, errors.New("获取订单状态流转配置失败")
	}
	return shop.OrderStatusFlow, nil
}

// CheckUsernameExists 检查店主用户名是否已存在
func (r *ShopRepository) CheckUsernameExists(username string) (bool, error) {
	var count int64
	err := r.DB.Model(&models.Shop{}).Where("owner_username = ?", username).Count(&count).Error
	if err != nil {
		log2.Errorf("CheckUsernameExists failed: %v", err)
		return false, errors.New("检查用户名失败")
	}
	return count > 0, nil
}

// Create 创建店铺
func (r *ShopRepository) Create(shop *models.Shop) error {
	err := r.DB.Create(shop).Error
	if err != nil {
		log2.Errorf("Create shop failed: %v", err)
		return errors.New("创建店铺失败")
	}
	return nil
}

// Update 更新店铺
func (r *ShopRepository) Update(shop *models.Shop) error {
	err := r.DB.Save(shop).Error
	if err != nil {
		log2.Errorf("Update shop failed: %v", err)
		return errors.New("更新店铺失败")
	}
	return nil
}

// UpdatePassword 更新店铺密码
func (r *ShopRepository) UpdatePassword(shopID snowflake.ID, hashedPassword string) error {
	err := r.DB.Model(&models.Shop{}).
		Where("id = ?", shopID).
		Update("owner_password", hashedPassword).Error
	if err != nil {
		log2.Errorf("UpdatePassword failed: %v", err)
		return errors.New("更新密码失败")
	}
	return nil
}

// GetByUsername 根据用户名获取店铺
func (r *ShopRepository) GetByUsername(username string) (*models.Shop, error) {
	var shop models.Shop
	err := r.DB.Where("owner_username = ?", username).First(&shop).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.New("店铺不存在")
	}
	if err != nil {
		log2.Errorf("GetByUsername failed: %v", err)
		return nil, errors.New("查询店铺失败")
	}
	return &shop, nil
}

// GetWithTags 获取店铺及其标签
func (r *ShopRepository) GetWithTags(shopID snowflake.ID) (*models.Shop, error) {
	var shop models.Shop
	err := r.DB.Preload("Tags").First(&shop, shopID).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.New("店铺不存在")
	}
	if err != nil {
		log2.Errorf("GetWithTags failed: %v", err)
		return nil, errors.New("查询店铺失败")
	}
	return &shop, nil
}

// UpdateImageURL 更新店铺图片URL
func (r *ShopRepository) UpdateImageURL(shopID snowflake.ID, imageURL string) error {
	err := r.DB.Model(&models.Shop{}).
		Where("id = ?", shopID).
		Update("image_url", imageURL).Error
	if err != nil {
		log2.Errorf("UpdateImageURL failed: %v", err)
		return errors.New("更新图片URL失败")
	}
	return nil
}
