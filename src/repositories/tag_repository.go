package repositories

import (
	"errors"
	"orderease/models"
	"orderease/utils/log2"

	"github.com/bwmarrin/snowflake"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
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

// ==================== Tag 复杂查询方法 ====================

// GetUnboundProductsCount 获取店铺中未绑定任何标签的商品数量
func (r *TagRepository) GetUnboundProductsCount(shopID uint64) (int64, error) {
	var count int64
	err := r.DB.Raw(`SELECT COUNT(*) FROM products
		WHERE shop_id = ? AND id NOT IN (SELECT product_id FROM product_tags)`, shopID).Scan(&count).Error
	if err != nil {
		log2.Errorf("GetUnboundProductsCount failed: %v", err)
		return 0, errors.New("查询未绑定商品数量失败")
	}
	return count, nil
}

// GetUnboundProductsForTag 获取可绑定到指定标签的商品列表（未绑定该标签的商品）
func (r *TagRepository) GetUnboundProductsForTag(tagID int, shopID uint64, page, pageSize int) ([]models.Product, int64, error) {
	offset := (page - 1) * pageSize

	var products []models.Product
	err := r.DB.Raw(`
		SELECT * FROM products
		WHERE id NOT IN (
			SELECT product_id FROM product_tags
			WHERE tag_id = ? AND shop_id = ?
		) ORDER BY created_at DESC LIMIT ? OFFSET ?`, tagID, shopID, pageSize, offset).Scan(&products).Error

	if err != nil {
		log2.Errorf("GetUnboundProductsForTag failed: %v", err)
		return nil, 0, errors.New("查询未绑定商品失败")
	}

	var total int64
	r.DB.Raw(`
		SELECT COUNT(*) FROM products
		WHERE id NOT IN (
			SELECT product_id FROM product_tags
			WHERE tag_id = ? AND shop_id = ?
		)`, tagID, shopID).Scan(&total)

	return products, total, nil
}

// GetUnboundTagsList 获取店铺中未绑定任何商品的标签列表（分页）
func (r *TagRepository) GetUnboundTagsList(shopID uint64, page, pageSize int) ([]models.Tag, int64, error) {
	offset := (page - 1) * pageSize

	var tags []models.Tag
	err := r.DB.Raw(`
		SELECT * FROM tags
		WHERE shop_id = ? AND id NOT IN (
			SELECT DISTINCT tag_id FROM product_tags
		) ORDER BY created_at DESC LIMIT ? OFFSET ?`, shopID, pageSize, offset).Scan(&tags).Error

	if err != nil {
		log2.Errorf("GetUnboundTagsList failed: %v", err)
		return nil, 0, errors.New("查询未绑定标签失败")
	}

	var total int64
	r.DB.Raw(`
		SELECT COUNT(*) FROM tags
		WHERE shop_id = ? AND id NOT IN (
			SELECT DISTINCT tag_id FROM product_tags
		)`, shopID).Scan(&total)

	return tags, total, nil
}

// GetTagBoundProductIDs 获取已绑定到指定标签的商品ID列表
func (r *TagRepository) GetTagBoundProductIDs(tagID int, shopID uint64) ([]uint, error) {
	var productIDs []uint
	err := r.DB.Raw(`
		SELECT product_id FROM product_tags
		WHERE tag_id = ? AND shop_id = ?`, tagID, shopID).Scan(&productIDs).Error

	if err != nil {
		log2.Errorf("GetTagBoundProductIDs failed: %v", err)
		return nil, errors.New("获取绑定商品ID列表失败")
	}

	return productIDs, nil
}

// GetOnlineProductsByTag 获取标签关联的在线商品列表
func (r *TagRepository) GetOnlineProductsByTag(tagID int, shopID uint64) ([]models.Product, error) {
	var products []models.Product
	err := r.DB.Joins("JOIN product_tags ON product_tags.product_id = products.id").
		Where("product_tags.tag_id = ? AND products.status = ? AND products.shop_id = ?",
			tagID, models.ProductStatusOnline, shopID).
		Find(&products).Error

	if err != nil {
		log2.Errorf("GetOnlineProductsByTag failed: %v", err)
		return nil, errors.New("查询标签关联商品失败")
	}

	return products, nil
}

// BatchTagProductsResult 批量打标结果
type BatchTagProductsResult struct {
	Total      int
	Successful int
}

// BatchTagProducts 批量打标签（事务）
func (r *TagRepository) BatchTagProducts(productIDs []snowflake.ID, tagID int, shopID uint64) (*BatchTagProductsResult, error) {
	tx := r.DB.Begin()

	// 批量查询商品的店铺信息
	var validProducts []models.Product
	if err := tx.Select("id").Where("id IN (?) AND shop_id = ?", productIDs, shopID).Find(&validProducts).Error; err != nil {
		tx.Rollback()
		log2.Errorf("BatchTagProducts query products failed: %v", err)
		return nil, errors.New("批量查询商品失败")
	}

	// 构建有效商品ID集合
	validProductMap := make(map[snowflake.ID]bool)
	for _, p := range validProducts {
		validProductMap[p.ID] = true
	}

	// 过滤无效商品并生成关联记录
	var productTags []models.ProductTag
	successCount := 0
	for _, productID := range productIDs {
		if validProductMap[productID] {
			productTags = append(productTags, models.ProductTag{
				ProductID: productID,
				TagID:     tagID,
				ShopID:    shopID,
			})
			successCount++
		}
	}

	// 使用 INSERT IGNORE 避免重复插入错误
	if len(productTags) > 0 {
		if err := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&productTags).Error; err != nil {
			tx.Rollback()
			log2.Errorf("BatchTagProducts create failed: %v", err)
			return nil, errors.New("批量打标签失败")
		}
	}

	if err := tx.Commit().Error; err != nil {
		log2.Errorf("BatchTagProducts commit failed: %v", err)
		return nil, errors.New("批量打标签失败")
	}

	return &BatchTagProductsResult{
		Total:      len(productIDs),
		Successful: successCount,
	}, nil
}

// BoundProductsResult 绑定商品查询结果
type BoundProductsResult struct {
	Products []models.Product
	Total    int64
}

// GetBoundProductsWithPagination 获取标签绑定的商品（分页）
// onlyOnline: true 表示只查询已上架商品（客户端），false 表示查询所有商品（管理端）
func (r *TagRepository) GetBoundProductsWithPagination(tagID int, shopID uint64, page, pageSize int, onlyOnline bool) (*BoundProductsResult, error) {
	// 获取已绑定商品的ID列表
	productIDs, err := r.GetTagBoundProductIDs(tagID, shopID)
	if err != nil {
		return nil, err
	}

	var total int64
	var products []models.Product

	offset := (page - 1) * pageSize

	// 构建查询条件
	query := r.DB.Model(&models.Product{}).Where("id IN (?)", productIDs)
	if onlyOnline {
		query = query.Where("status = ?", models.ProductStatusOnline)
	}

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		log2.Errorf("GetBoundProductsWithPagination count failed: %v", err)
		return nil, errors.New("获取商品总数失败")
	}

	// 查询完整商品数据并预加载选项
	if err := query.Preload("OptionCategories.Options").
		Order("created_at DESC").
		Limit(pageSize).Offset(offset).
		Find(&products).Error; err != nil {
		log2.Errorf("GetBoundProductsWithPagination find failed: %v", err)
		return nil, errors.New("查询商品详情失败")
	}

	return &BoundProductsResult{
		Products: products,
		Total:    total,
	}, nil
}

// GetUnboundProductsWithPagination 获取未绑定任何标签的商品（分页）
// onlyOnline: true 表示只查询已上架商品（客户端），false 表示查询所有商品（管理端）
func (r *TagRepository) GetUnboundProductsWithPagination(shopID uint64, page, pageSize int, onlyOnline bool) (*BoundProductsResult, error) {
	offset := (page - 1) * pageSize
	var products []models.Product
	var total int64

	query := r.DB.Model(&models.Product{}).
		Where("shop_id = ? AND id NOT IN (SELECT product_id FROM product_tags)", shopID)

	// 如果是客户端请求，只查询已上架商品
	if onlyOnline {
		query = query.Where("status = ?", models.ProductStatusOnline)
	}

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		log2.Errorf("GetUnboundProductsWithPagination count failed: %v", err)
		return nil, errors.New("获取未绑定商品总数失败")
	}

	// 查询商品数据并预加载选项
	if err := query.Offset(offset).
		Limit(pageSize).Order("created_at DESC").
		Preload("OptionCategories.Options").
		Find(&products).Error; err != nil {
		log2.Errorf("GetUnboundProductsWithPagination find failed: %v", err)
		return nil, errors.New("查询未绑定商品失败")
	}

	return &BoundProductsResult{
		Products: products,
		Total:    total,
	}, nil
}

// BatchUntagProductsResult 批量解绑结果
type BatchUntagProductsResult struct {
	Total      int
	Successful int64
}

// BatchUntagProducts 批量解绑商品标签
func (r *TagRepository) BatchUntagProducts(productIDs []snowflake.ID, tagID uint, shopID uint64) (*BatchUntagProductsResult, error) {
	result := r.DB.Where("shop_id = ? AND tag_id = ? AND product_id IN (?)", shopID, tagID, productIDs).
		Delete(&models.ProductTag{})

	if result.Error != nil {
		log2.Errorf("BatchUntagProducts delete failed: %v", result.Error)
		return nil, errors.New("批量解绑标签失败")
	}

	return &BatchUntagProductsResult{
		Total:      len(productIDs),
		Successful: result.RowsAffected,
	}, nil
}
