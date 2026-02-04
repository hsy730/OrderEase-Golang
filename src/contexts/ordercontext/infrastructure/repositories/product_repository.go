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
func (r *ProductRepository) GetProductByID(id uint64, shopID snowflake.ID) (*models.Product, error) {
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
func (r *ProductRepository) GetProductsByIDs(ids []snowflake.ID, shopID snowflake.ID) ([]models.Product, error) {
	var products []models.Product
	if err := r.DB.Select("id").Where("id IN (?) AND shop_id = ?", ids, shopID).Find(&products).Error; err != nil {
		return nil, err
	}
	return products, nil
}

// CheckShopExists 校验店铺ID合法性
func (r *ProductRepository) CheckShopExists(shopID snowflake.ID) (bool, error) {
    var count int64
    if err := r.DB.Model(&models.Shop{}).Where("id = ?", shopID).Count(&count).Error; err != nil {
        log2.Errorf("CheckShopExists failed: %v", err)
        return false, errors.New("店铺校验失败")
    }
    return count > 0, nil
}

// GetShopProducts 获取指定店铺的商品（用于批量操作前的店铺校验）
func (r *ProductRepository) GetShopProducts(shopID snowflake.ID, productIDs []snowflake.ID) ([]models.Product, error) {
    var products []models.Product
    if err := r.DB.Select("id").
        Where("shop_id = ? AND id IN (?)", shopID, productIDs).
        Find(&products).Error; err != nil {
        return nil, err
    }
    return products, nil
}

// UpdateStatus 更新商品状态
func (r *ProductRepository) UpdateStatus(productID uint64, shopID snowflake.ID, status string) error {
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
func (r *ProductRepository) UpdateImageURL(productID uint64, shopID snowflake.ID, imageURL string) error {
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

// CreateWithCategories 创建商品及其参数类别（事务）
func (r *ProductRepository) CreateWithCategories(product *models.Product, categories []models.ProductOptionCategory) error {
	tx := r.DB.Begin()

	// 创建商品
	if err := tx.Create(product).Error; err != nil {
		tx.Rollback()
		log2.Errorf("CreateWithCategories create product failed: %v", err)
		return errors.New("创建商品失败")
	}

	// 创建商品参数类别
	for i := range categories {
		category := categories[i]
		category.ProductID = product.ID

		if err := tx.Create(&category).Error; err != nil {
			tx.Rollback()
			log2.Errorf("CreateWithCategories create category failed: %v", err)
			return errors.New("创建商品参数失败")
		}
	}

	if err := tx.Commit().Error; err != nil {
		log2.Errorf("CreateWithCategories commit failed: %v", err)
		return errors.New("创建商品失败")
	}

	return nil
}

// UpdateWithCategories 更新商品及其参数类别（事务）
func (r *ProductRepository) UpdateWithCategories(product *models.Product, categories []models.ProductOptionCategory) error {
	tx := r.DB.Begin()

	// 保存更新后的商品信息
	if err := tx.Save(product).Error; err != nil {
		tx.Rollback()
		log2.Errorf("UpdateWithCategories save product failed: %v", err)
		return errors.New("更新商品失败")
	}

	// 删除旧的参数类别
	if err := tx.Where("product_id = ?", product.ID).Delete(&models.ProductOptionCategory{}).Error; err != nil {
		tx.Rollback()
		log2.Errorf("UpdateWithCategories delete old categories failed: %v", err)
		return errors.New("更新商品参数失败")
	}

	// 创建新的参数类别
	for i := range categories {
		category := categories[i]
		category.ProductID = product.ID

		if err := tx.Create(&category).Error; err != nil {
			tx.Rollback()
			log2.Errorf("UpdateWithCategories create category failed: %v", err)
			return errors.New("更新商品参数失败")
		}
	}

	if err := tx.Commit().Error; err != nil {
		log2.Errorf("UpdateWithCategories commit failed: %v", err)
		return errors.New("更新商品失败")
	}

	return nil
}

// DeleteWithDependencies 删除商品及其关联数据（事务）
func (r *ProductRepository) DeleteWithDependencies(productID uint64, shopID snowflake.ID) error {
	tx := r.DB.Begin()

	// 删除商品参数选项（先删除选项）
	if err := tx.Where(`category_id IN (
		SELECT id FROM product_option_categories WHERE product_id = ?
	)`, productID).Delete(&models.ProductOption{}).Error; err != nil {
		tx.Rollback()
		log2.Errorf("DeleteWithDependencies delete options failed: %v", err)
		return errors.New("删除商品参数选项失败")
	}

	// 删除商品参数类别
	if err := tx.Where("product_id = ?", productID).Delete(&models.ProductOptionCategory{}).Error; err != nil {
		tx.Rollback()
		log2.Errorf("DeleteWithDependencies delete categories failed: %v", err)
		return errors.New("删除商品参数类别失败")
	}

	// 删除商品记录
	result := tx.Where("id = ? AND shop_id = ?", productID, shopID).Delete(&models.Product{})
	if result.Error != nil {
		tx.Rollback()
		log2.Errorf("DeleteWithDependencies delete product failed: %v", result.Error)
		return errors.New("删除商品记录失败")
	}
	if result.RowsAffected == 0 {
		tx.Rollback()
		return errors.New("商品不存在")
	}

	if err := tx.Commit().Error; err != nil {
		log2.Errorf("DeleteWithDependencies commit failed: %v", err)
		return errors.New("删除商品失败")
	}

	return nil
}

// GetProductsByShop 获取店铺商品列表（分页，预加载选项类别）
// onlyOnline: true 表示只查询已上架商品（用户查询），false 表示查询所有商品（管理员查询）
func (r *ProductRepository) GetProductsByShop(shopID snowflake.ID, page int, pageSize int, search string, onlyOnline bool) (*ProductListResult, error) {
	var products []models.Product
	var total int64

	offset := (page - 1) * pageSize

	// 构建查询条件
	query := r.DB.Where("shop_id = ?", shopID)

	// 如果是用户查询，只查询已上架商品
	if onlyOnline {
		query = query.Where("status = ?", models.ProductStatusOnline)
	}

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
