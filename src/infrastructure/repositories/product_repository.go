package repositories

import (
	"errors"
	"orderease/domain/product"
	"orderease/domain/shared"
	"orderease/infrastructure/persistence"
	"orderease/models"
	"orderease/utils/log2"

	"github.com/bwmarrin/snowflake"
	"gorm.io/gorm"
)

type ProductRepositoryImpl struct {
	db *gorm.DB
}

func NewProductRepository(db *gorm.DB) product.ProductRepository {
	return &ProductRepositoryImpl{db: db}
}

func (r *ProductRepositoryImpl) Save(prod *product.Product) error {
	model := persistence.ProductToModel(prod)
	if err := r.db.Create(model).Error; err != nil {
		log2.Errorf("保存商品失败: %v", err)
		return errors.New("保存商品失败")
	}
	prod.ID = shared.ID(model.ID)
	return nil
}

func (r *ProductRepositoryImpl) FindByID(id shared.ID) (*product.Product, error) {
	var model models.Product
	if err := r.db.Preload("OptionCategories").Preload("OptionCategories.Options").First(&model, id.Value()).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("商品不存在")
		}
		log2.Errorf("查询商品失败: %v", err)
		return nil, errors.New("查询商品失败")
	}
	return persistence.ProductToDomain(model), nil
}

func (r *ProductRepositoryImpl) FindByIDAndShopID(id shared.ID, shopID uint64) (*product.Product, error) {
	var model models.Product
	if err := r.db.Where("shop_id = ?", shopID).Preload("OptionCategories").Preload("OptionCategories.Options").First(&model, id.Value()).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("商品不存在")
		}
		log2.Errorf("查询商品失败: %v", err)
		return nil, errors.New("查询商品失败")
	}
	return persistence.ProductToDomain(model), nil
}

func (r *ProductRepositoryImpl) FindByShopID(shopID uint64, page, pageSize int, search string, excludeOffline bool) ([]product.Product, int64, error) {
	query := r.db.Model(&models.Product{}).Where("shop_id = ?", shopID)

	if excludeOffline {
		query = query.Where("status != ?", product.ProductStatusOffline)
	}

	if search != "" {
		query = query.Where("name LIKE ?", "%"+search+"%")
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		log2.Errorf("查询商品总数失败: %v", err)
		return nil, 0, errors.New("查询商品总数失败")
	}

	var modelsList []models.Product
	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Order("created_at DESC").Preload("OptionCategories").Preload("OptionCategories.Options").Find(&modelsList).Error; err != nil {
		log2.Errorf("查询商品列表失败: %v", err)
		return nil, 0, errors.New("查询商品列表失败")
	}

	products := make([]product.Product, len(modelsList))
	for i, m := range modelsList {
		products[i] = *persistence.ProductToDomain(m)
	}
	return products, total, nil
}

func (r *ProductRepositoryImpl) FindByIDs(ids []shared.ID) ([]product.Product, error) {
	if len(ids) == 0 {
		return []product.Product{}, nil
	}

	idValues := make([]uint64, len(ids))
	for i, id := range ids {
		idValues[i] = id.ToUint64()
	}

	var modelsList []models.Product
	if err := r.db.Where("id IN (?)", idValues).Find(&modelsList).Error; err != nil {
		log2.Errorf("查询商品失败: %v", err)
		return nil, errors.New("查询商品失败")
	}

	products := make([]product.Product, len(modelsList))
	for i, m := range modelsList {
		products[i] = *persistence.ProductToDomain(m)
	}
	return products, nil
}

func (r *ProductRepositoryImpl) Delete(id shared.ID) error {
	if err := r.db.Delete(&models.Product{}, id.Value()).Error; err != nil {
		log2.Errorf("删除商品失败: %v", err)
		return errors.New("删除商品失败")
	}
	return nil
}

func (r *ProductRepositoryImpl) Update(prod *product.Product) error {
	model := persistence.ProductToModel(prod)
	if err := r.db.Save(model).Error; err != nil {
		log2.Errorf("更新商品失败: %v", err)
		return errors.New("更新商品失败")
	}
	return nil
}

func (r *ProductRepositoryImpl) CountByProductID(productID shared.ID) (int64, error) {
	var count int64
	if err := r.db.Model(&models.OrderItem{}).Where("product_id = ?", productID.Value()).Count(&count).Error; err != nil {
		log2.Errorf("查询商品订单关联数失败: %v", err)
		return 0, errors.New("查询商品订单关联数失败")
	}
	return count, nil
}

func (r *ProductRepositoryImpl) FindOptionByID(id shared.ID) (*product.ProductOption, error) {
	var model models.ProductOption
	if err := r.db.First(&model, id.Value()).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("商品参数选项不存在")
		}
		log2.Errorf("查询商品参数选项失败: %v", err)
		return nil, errors.New("查询商品参数选项失败")
	}
	return persistence.ProductOptionToDomain(model), nil
}

func (r *ProductRepositoryImpl) FindOptionCategoryByID(id shared.ID) (*product.ProductOptionCategory, error) {
	var model models.ProductOptionCategory
	if err := r.db.Preload("Options").First(&model, id.Value()).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("商品参数类别不存在")
		}
		log2.Errorf("查询商品参数类别失败: %v", err)
		return nil, errors.New("查询商品参数类别失败")
	}
	return persistence.ProductOptionCategoryToDomain(model), nil
}

type ProductOptionCategoryRepositoryImpl struct {
	db *gorm.DB
}

func NewProductOptionCategoryRepository(db *gorm.DB) product.ProductOptionCategoryRepository {
	return &ProductOptionCategoryRepositoryImpl{db: db}
}

func (r *ProductOptionCategoryRepositoryImpl) Save(category *product.ProductOptionCategory) error {
	model := persistence.ProductOptionCategoryToModel(*category)
	if err := r.db.Create(&model).Error; err != nil {
		log2.Errorf("保存商品参数类别失败: %v", err)
		return errors.New("保存商品参数类别失败")
	}
	category.ID = shared.ID(model.ID)
	return nil
}

func (r *ProductOptionCategoryRepositoryImpl) FindByID(id shared.ID) (*product.ProductOptionCategory, error) {
	var model models.ProductOptionCategory
	if err := r.db.Preload("Options").First(&model, id.Value()).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("商品参数类别不存在")
		}
		log2.Errorf("查询商品参数类别失败: %v", err)
		return nil, errors.New("查询商品参数类别失败")
	}
	return persistence.ProductOptionCategoryToDomain(model), nil
}

func (r *ProductOptionCategoryRepositoryImpl) FindByProductID(productID shared.ID) ([]product.ProductOptionCategory, error) {
	var modelsList []models.ProductOptionCategory
	if err := r.db.Preload("Options").Where("product_id = ?", productID.Value()).Find(&modelsList).Error; err != nil {
		log2.Errorf("查询商品参数类别失败: %v", err)
		return nil, errors.New("查询商品参数类别失败")
	}

	categories := make([]product.ProductOptionCategory, len(modelsList))
	for i, m := range modelsList {
		categories[i] = *persistence.ProductOptionCategoryToDomain(m)
	}
	return categories, nil
}

func (r *ProductOptionCategoryRepositoryImpl) DeleteByProductID(productID shared.ID) error {
	if err := r.db.Where("product_id = ?", productID.Value()).Delete(&models.ProductOptionCategory{}).Error; err != nil {
		log2.Errorf("删除商品参数类别失败: %v", err)
		return errors.New("删除商品参数类别失败")
	}
	return nil
}

type ProductOptionRepositoryImpl struct {
	db *gorm.DB
}

func NewProductOptionRepository(db *gorm.DB) product.ProductOptionRepository {
	return &ProductOptionRepositoryImpl{db: db}
}

func (r *ProductOptionRepositoryImpl) Save(option *product.ProductOption) error {
	model := persistence.ProductOptionToModel(*option)
	if err := r.db.Create(&model).Error; err != nil {
		log2.Errorf("保存商品参数选项失败: %v", err)
		return errors.New("保存商品参数选项失败")
	}
	option.ID = shared.ID(model.ID)
	return nil
}

func (r *ProductOptionRepositoryImpl) FindByID(id shared.ID) (*product.ProductOption, error) {
	var model models.ProductOption
	if err := r.db.First(&model, id.Value()).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("商品参数选项不存在")
		}
		log2.Errorf("查询商品参数选项失败: %v", err)
		return nil, errors.New("查询商品参数选项失败")
	}
	return persistence.ProductOptionToDomain(model), nil
}

func (r *ProductOptionRepositoryImpl) FindByCategoryID(categoryID shared.ID) ([]product.ProductOption, error) {
	var modelsList []models.ProductOption
	if err := r.db.Where("category_id = ?", categoryID.Value()).Find(&modelsList).Error; err != nil {
		log2.Errorf("查询商品参数选项失败: %v", err)
		return nil, errors.New("查询商品参数选项失败")
	}

	options := make([]product.ProductOption, len(modelsList))
	for i, m := range modelsList {
		options[i] = *persistence.ProductOptionToDomain(m)
	}
	return options, nil
}

func (r *ProductOptionRepositoryImpl) DeleteByCategoryID(categoryID shared.ID) error {
	if err := r.db.Where("category_id = ?", categoryID.Value()).Delete(&models.ProductOption{}).Error; err != nil {
		log2.Errorf("删除商品参数选项失败: %v", err)
		return errors.New("删除商品参数选项失败")
	}
	return nil
}

type ProductTagRepositoryImpl struct {
	db *gorm.DB
}

func NewProductTagRepository(db *gorm.DB) product.ProductTagRepository {
	return &ProductTagRepositoryImpl{db: db}
}

func (r *ProductTagRepositoryImpl) Save(productID shared.ID, tagID int, shopID uint64) error {
	productTag := models.ProductTag{
		ProductID: productID.Value(),
		TagID:     tagID,
		ShopID:    snowflake.ID(shopID),
	}
	if err := r.db.Create(&productTag).Error; err != nil {
		log2.Errorf("保存商品标签关联失败: %v", err)
		return errors.New("保存商品标签关联失败")
	}
	return nil
}

func (r *ProductTagRepositoryImpl) FindByProductID(productID shared.ID) ([]int, error) {
	var productTags []models.ProductTag
	if err := r.db.Where("product_id = ?", productID.Value()).Find(&productTags).Error; err != nil {
		log2.Errorf("查询商品标签失败: %v", err)
		return nil, errors.New("查询商品标签失败")
	}

	tagIDs := make([]int, len(productTags))
	for i, pt := range productTags {
		tagIDs[i] = pt.TagID
	}
	return tagIDs, nil
}

func (r *ProductTagRepositoryImpl) FindByTagID(tagID int) ([]shared.ID, error) {
	var productTags []models.ProductTag
	if err := r.db.Where("tag_id = ?", tagID).Find(&productTags).Error; err != nil {
		log2.Errorf("查询标签商品失败: %v", err)
		return nil, errors.New("查询标签商品失败")
	}

	productIDs := make([]shared.ID, len(productTags))
	for i, pt := range productTags {
		productIDs[i] = shared.ID(pt.ProductID)
	}
	return productIDs, nil
}

func (r *ProductTagRepositoryImpl) DeleteByProductID(productID shared.ID) error {
	if err := r.db.Where("product_id = ?", productID.Value()).Delete(&models.ProductTag{}).Error; err != nil {
		log2.Errorf("删除商品标签关联失败: %v", err)
		return errors.New("删除商品标签关联失败")
	}
	return nil
}
