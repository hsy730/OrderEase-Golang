package repositories

import (
	"errors"
	"orderease/models"
	"orderease/utils/log2"

	"gorm.io/gorm"

	"github.com/bwmarrin/snowflake"
)

type ProductRepository struct {
	DB *gorm.DB
}

func NewProductRepository(db *gorm.DB) *ProductRepository {
	return &ProductRepository{DB: db}
}

// 根据ID和店铺ID查询商品
func (r *ProductRepository) GetProductByID(id uint64, shopID uint64) (*models.Product, error) {
	var product models.Product
	// 使用嵌套Preload预加载商品的选项类别及其选项
	err := r.DB.Where("shop_id = ?", snowflake.ID(shopID)).
		Preload("OptionCategories.Options"). // 嵌套预加载选项类别及其选项
		First(&product, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.New("商品不存在")
	}
	if err != nil {
		log2.Errorf("GetProductByID failed: %v", err)
		return nil, errors.New("服务器内部错误")
	}
	return &product, nil
}

// 通用商品查询方法
func (r *ProductRepository) GetProductsByIDs(ids []snowflake.ID, shopID uint64) ([]models.Product, error) {
	var products []models.Product
	if err := r.DB.Select("id").Where("id IN (?) AND shop_id = ?", ids, snowflake.ID(shopID)).Find(&products).Error; err != nil {
		return nil, err
	}
	return products, nil
}

// CheckShopExists 校验店铺ID合法性
func (r *ProductRepository) CheckShopExists(shopID uint64) (bool, error) {
    var count int64
    if err := r.DB.Model(&models.Shop{}).Where("id = ?", snowflake.ID(shopID)).Count(&count).Error; err != nil {
        log2.Errorf("CheckShopExists failed: %v", err)
        return false, errors.New("店铺校验失败")
    }
    return count > 0, nil
}

// GetShopProducts 获取指定店铺的商品（用于批量操作前的店铺校验）
func (r *ProductRepository) GetShopProducts(shopID uint64, productIDs []snowflake.ID) ([]models.Product, error) {
    var products []models.Product
    if err := r.DB.Select("id").
        Where("shop_id = ? AND id IN (?)", snowflake.ID(shopID), productIDs).
        Find(&products).Error; err != nil {
        return nil, err
    }
    return products, nil
}
