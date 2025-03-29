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
	err := r.DB.Where("shop_id = ?", shopID).First(&product, id).Error
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
	if err := r.DB.Select("id").Where("id IN (?) AND shop_id = ?", ids, shopID).Find(&products).Error; err != nil {
		return nil, err
	}
	return products, nil
}
