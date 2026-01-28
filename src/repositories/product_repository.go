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
	err := r.DB.Where("shop_id = ?", shopID).
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
	if err := r.DB.Select("id").Where("id IN (?) AND shop_id = ?", ids, shopID).Find(&products).Error; err != nil {
		return nil, err
	}
	return products, nil
}

// CheckShopExists 校验店铺ID合法性
func (r *ProductRepository) CheckShopExists(shopID uint64) (bool, error) {
    var count int64
    if err := r.DB.Model(&models.Shop{}).Where("id = ?", shopID).Count(&count).Error; err != nil {
        log2.Errorf("CheckShopExists failed: %v", err)
        return false, errors.New("店铺校验失败")
    }
    return count > 0, nil
}

// GetShopProducts 获取指定店铺的商品（用于批量操作前的店铺校验）
func (r *ProductRepository) GetShopProducts(shopID uint64, productIDs []snowflake.ID) ([]models.Product, error) {
    var products []models.Product
    if err := r.DB.Select("id").
        Where("shop_id = ? AND id IN (?)", shopID, productIDs).
        Find(&products).Error; err != nil {
        return nil, err
    }
    return products, nil
}

// UpdateStatus 更新商品状态
func (r *ProductRepository) UpdateStatus(productID uint64, shopID uint64, status string) error {
	result := r.DB.Model(&models.Product{}).
		Where("id = ? AND shop_id = ?", productID, shopID).
		Update("status", status)
	if result.Error != nil {
		log2.Errorf("UpdateStatus failed: %v", result.Error)
		return errors.New("更新商品状态失败")
	}
	if result.RowsAffected == 0 {
		return errors.New("商品不存在")
	}
	return nil
}

// UpdateImageURL 更新商品图片URL
func (r *ProductRepository) UpdateImageURL(productID uint64, shopID uint64, imageURL string) error {
	result := r.DB.Model(&models.Product{}).
		Where("id = ? AND shop_id = ?", productID, shopID).
		Update("image_url", imageURL)
	if result.Error != nil {
		log2.Errorf("UpdateImageURL failed: %v", result.Error)
		return errors.New("更新商品图片失败")
	}
	if result.RowsAffected == 0 {
		return errors.New("商品不存在")
	}
	return nil
}

// ProductListResult 商品列表查询结果
type ProductListResult struct {
	Products []models.Product
	Total    int64
}

// GetProductsByShop 获取店铺商品列表（分页，预加载选项类别）
func (r *ProductRepository) GetProductsByShop(shopID uint64, page int, pageSize int, search string) (*ProductListResult, error) {
	var products []models.Product
	var total int64

	offset := (page - 1) * pageSize

	// 只查询未下架的商品（待上架和已上架）
	query := r.DB.Where("status != ? and shop_id = ?", models.ProductStatusOffline, shopID)

	// 如果有搜索关键词，添加模糊搜索条件
	if search != "" {
		query = query.Where("name LIKE ?", "%"+search+"%")
	}

	// 获取总数
	if err := query.Model(&models.Product{}).Count(&total).Error; err != nil {
		log2.Errorf("GetProductsByShop count failed: %v", err)
		return nil, errors.New("获取商品总数失败")
	}

	// 获取分页数据，并预加载参数类别和选项信息
	if err := query.Offset(offset).Limit(pageSize).
		Order("created_at DESC").
		Preload("OptionCategories").
		Preload("OptionCategories.Options").
		Find(&products).Error; err != nil {
		log2.Errorf("GetProductsByShop find failed: %v", err)
		return nil, errors.New("获取商品列表失败")
	}

	return &ProductListResult{
		Products: products,
		Total:    total,
	}, nil
}
