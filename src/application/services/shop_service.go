package services

import (
	"errors"
	"fmt"
	"io"
	"orderease/application/dto"
	"orderease/domain/order"
	"orderease/domain/product"
	"orderease/domain/shared"
	"orderease/domain/shop"
	"orderease/models"
	"orderease/utils"
	"orderease/utils/log2"
	"os"
	"path/filepath"
	"time"

	"github.com/bwmarrin/snowflake"
	"gorm.io/gorm"
)

type ShopService struct {
	shopRepo    shop.ShopRepository
	tagRepo     shop.TagRepository
	productRepo product.ProductRepository
	db          *gorm.DB
}

func NewShopService(
	shopRepo shop.ShopRepository,
	tagRepo shop.TagRepository,
	productRepo product.ProductRepository,
	db *gorm.DB,
) *ShopService {
	return &ShopService{
		shopRepo:    shopRepo,
		tagRepo:     tagRepo,
		productRepo: productRepo,
		db:          db,
	}
}

func (s *ShopService) CreateShop(req *dto.CreateShopRequest) (*dto.ShopResponse, error) {
	_, err := s.shopRepo.FindByOwnerUsername(req.OwnerUsername)
	if err == nil {
		return nil, errors.New("店主用户名已存在")
	}

	shopEntity, err := shop.NewShop(req.Name, req.OwnerUsername, req.OwnerPassword, req.ValidUntil)
	if err != nil {
		return nil, err
	}

	shopEntity.ContactPhone = req.ContactPhone
	shopEntity.ContactEmail = req.ContactEmail
	shopEntity.Description = req.Description
	shopEntity.Address = req.Address
	shopEntity.Settings = req.Settings

	if req.OrderStatusFlow != nil {
		shopEntity.OrderStatusFlow = *req.OrderStatusFlow
	}

	if err := s.shopRepo.Save(shopEntity); err != nil {
		return nil, errors.New("创建店铺失败")
	}

	return s.toShopResponse(shopEntity), nil
}

func (s *ShopService) GetShop(id shared.ID) (*dto.ShopResponse, error) {
	shopEntity, err := s.shopRepo.FindByID(id)
	if err != nil {
		return nil, err
	}

	return s.toShopResponse(shopEntity), nil
}

func (s *ShopService) GetShops(page, pageSize int, search string) (*dto.ShopListResponse, error) {
	shops, total, err := s.shopRepo.FindAll(page, pageSize, search)
	if err != nil {
		return nil, err
	}

	data := make([]dto.ShopResponse, len(shops))
	for i, shopEntity := range shops {
		data[i] = *s.toShopResponse(&shopEntity)
	}

	return &dto.ShopListResponse{
		Total:    total,
		Page:     page,
		PageSize: pageSize,
		Data:     data,
	}, nil
}

func (s *ShopService) UpdateShop(req *dto.UpdateShopRequest) (*dto.ShopResponse, error) {
	shopEntity, err := s.shopRepo.FindByID(req.ID)
	if err != nil {
		return nil, err
	}

	if req.Name != "" {
		shopEntity.Name = req.Name
	}
	if req.ContactPhone != "" {
		shopEntity.ContactPhone = req.ContactPhone
	}
	if req.ContactEmail != "" {
		shopEntity.ContactEmail = req.ContactEmail
	}
	if req.Description != "" {
		shopEntity.Description = req.Description
	}
	if req.Address != "" {
		shopEntity.Address = req.Address
	}
	if req.Settings != "" {
		shopEntity.Settings = req.Settings
	}
	if req.OwnerUsername != "" {
		shopEntity.OwnerUsername = req.OwnerUsername
	}
	if req.OwnerPassword != nil {
		shopEntity.OwnerPassword = *req.OwnerPassword
	}

	if !req.ValidUntil.IsZero() {
		if err := shopEntity.UpdateValidUntil(req.ValidUntil); err != nil {
			return nil, err
		}
	}

	if req.OrderStatusFlow != nil {
		if err := shopEntity.UpdateOrderStatusFlow(*req.OrderStatusFlow); err != nil {
			return nil, err
		}
	}

	if err := s.shopRepo.Update(shopEntity); err != nil {
		return nil, errors.New("更新店铺失败")
	}

	return s.toShopResponse(shopEntity), nil
}

func (s *ShopService) DeleteShop(id shared.ID) error {
	products, _, err := s.productRepo.FindByShopID(id.ToUint64(), 1, 1, "", false)
	if err == nil && len(products) > 0 {
		return errors.New("存在关联商品，无法删除店铺")
	}

	if err := s.shopRepo.Delete(id); err != nil {
		return errors.New("删除店铺失败")
	}

	return nil
}

func (s *ShopService) CheckShopNameExists(name string) (bool, error) {
	_, err := s.shopRepo.FindByName(name)
	if err != nil {
		if err.Error() == "店铺不存在" {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (s *ShopService) UploadShopImage(id shared.ID, file io.Reader, filename string) (string, error) {
	shopEntity, err := s.shopRepo.FindByID(id)
	if err != nil {
		return "", err
	}

	uploadDir := "./uploads/shops"
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		return "", errors.New("创建上传目录失败")
	}

	if shopEntity.ImageURL != "" {
		oldImagePath := filepath.Join(uploadDir, shopEntity.ImageURL)
		if err := os.Remove(oldImagePath); err != nil && !os.IsNotExist(err) {
			log2.Errorf("删除旧图片失败: %v", err)
		}
	}

	newFilename := fmt.Sprintf("shop_%d_%d%s", id.ToUint64(), time.Now().Unix(), filepath.Ext(filename))
	imagePath := filepath.Join(uploadDir, newFilename)

	dst, err := os.Create(imagePath)
	if err != nil {
		return "", errors.New("创建文件失败")
	}
	defer dst.Close()

	if _, err := dst.ReadFrom(file); err != nil {
		return "", errors.New("保存文件失败")
	}

	if _, err := utils.CompressImage(imagePath, 512*1024); err != nil {
		log2.Errorf("压缩图片失败: %v", err)
	}

	shopEntity.ImageURL = newFilename
	if err := s.shopRepo.Update(shopEntity); err != nil {
		return "", errors.New("更新店铺图片失败")
	}

	return newFilename, nil
}

func (s *ShopService) UpdateOrderStatusFlow(shopID shared.ID, flow order.OrderStatusFlow) error {
	shopEntity, err := s.shopRepo.FindByID(shopID)
	if err != nil {
		return err
	}

	if err := shopEntity.UpdateOrderStatusFlow(flow); err != nil {
		return err
	}

	if err := s.shopRepo.Update(shopEntity); err != nil {
		return errors.New("更新店铺订单流转状态配置失败")
	}

	return nil
}

func (s *ShopService) GetShopTags(shopID shared.ID) (*dto.TagListResponse, error) {
	tags, err := s.tagRepo.FindByShopID(shopID)
	if err != nil {
		return nil, err
	}

	tagResponses := make([]dto.TagResponse, len(tags))
	for i, tag := range tags {
		tagResponses[i] = dto.TagResponse{
			ID:          tag.ID,
			ShopID:      shared.ParseIDFromUint64(uint64(tag.ShopID)),
			Name:        tag.Name,
			Description: tag.Description,
			CreatedAt:   tag.CreatedAt,
			UpdatedAt:   tag.UpdatedAt,
		}
	}

	return &dto.TagListResponse{
		Total: int64(len(tags)),
		Tags:  tagResponses,
	}, nil
}

func (s *ShopService) CreateTag(req *dto.CreateTagRequest) (*dto.TagResponse, error) {
	tagEntity, err := shop.NewTag(req.ShopID, req.Name, req.Description)
	if err != nil {
		return nil, err
	}

	if err := s.tagRepo.Save(tagEntity); err != nil {
		return nil, errors.New("创建标签失败")
	}

	return &dto.TagResponse{
		ID:          tagEntity.ID,
		ShopID:      tagEntity.ShopID,
		Name:        tagEntity.Name,
		Description: tagEntity.Description,
		CreatedAt:   tagEntity.CreatedAt,
		UpdatedAt:   tagEntity.UpdatedAt,
	}, nil
}

func (s *ShopService) UpdateTag(id int, req *dto.CreateTagRequest) (*dto.TagResponse, error) {
	tagEntity, err := s.tagRepo.FindByID(id)
	if err != nil {
		return nil, err
	}

	if err := tagEntity.Update(req.Name, req.Description); err != nil {
		return nil, err
	}

	if err := s.tagRepo.Update(tagEntity); err != nil {
		return nil, errors.New("更新标签失败")
	}

	return &dto.TagResponse{
		ID:          tagEntity.ID,
		ShopID:      tagEntity.ShopID,
		Name:        tagEntity.Name,
		Description: tagEntity.Description,
		CreatedAt:   tagEntity.CreatedAt,
		UpdatedAt:   tagEntity.UpdatedAt,
	}, nil
}

func (s *ShopService) DeleteTag(id int) error {
	// TODO: 检查是否有关联商品
	return s.tagRepo.Delete(id)
}

func (s *ShopService) GetTag(id int) (*dto.TagResponse, error) {
	tagEntity, err := s.tagRepo.FindByID(id)
	if err != nil {
		return nil, err
	}

	return &dto.TagResponse{
		ID:          tagEntity.ID,
		ShopID:      tagEntity.ShopID,
		Name:        tagEntity.Name,
		Description: tagEntity.Description,
		CreatedAt:   tagEntity.CreatedAt,
		UpdatedAt:   tagEntity.UpdatedAt,
	}, nil
}

func (s *ShopService) toShopResponse(shopEntity *shop.Shop) *dto.ShopResponse {
	return &dto.ShopResponse{
		ID:              shopEntity.ID,
		Name:            shopEntity.Name,
		OwnerUsername:   shopEntity.OwnerUsername,
		ContactPhone:    shopEntity.ContactPhone,
		ContactEmail:    shopEntity.ContactEmail,
		Address:         shopEntity.Address,
		Description:     shopEntity.Description,
		CreatedAt:       shopEntity.CreatedAt,
		UpdatedAt:       shopEntity.UpdatedAt,
		ValidUntil:      shopEntity.ValidUntil,
		Settings:        shopEntity.Settings,
		ImageURL:        shopEntity.ImageURL,
		OrderStatusFlow: shopEntity.OrderStatusFlow,
	}
}

// GetBoundTags 获取商品已绑定的标签
func (s *ShopService) GetBoundTags(productID string, shopID uint64) ([]interface{}, error) {
	type TagResult struct {
		ID          uint64 `json:"id"`
		ShopID      uint64 `json:"shop_id"`
		Name        string `json:"name"`
		Description string `json:"description"`
	}

	var tags []TagResult
	err := s.db.Raw(`
		SELECT tags.id, tags.shop_id, tags.name, tags.description FROM tags
		JOIN product_tags ON product_tags.tag_id = tags.id
		WHERE product_tags.product_id = ?
		AND tags.shop_id = ?`, productID, shopID).Scan(&tags).Error

	if err != nil {
		return nil, err
	}

	result := make([]interface{}, len(tags))
	for i, tag := range tags {
		result[i] = tag
	}
	return result, nil
}

// GetUnboundTags 获取商品未绑定的标签
func (s *ShopService) GetUnboundTags(productID string, shopID uint64) ([]interface{}, error) {
	type TagResult struct {
		ID          uint64 `json:"id"`
		ShopID      uint64 `json:"shop_id"`
		Name        string `json:"name"`
		Description string `json:"description"`
	}

	var tags []TagResult
	err := s.db.Raw(`
		SELECT * FROM tags
		WHERE id NOT IN (
			SELECT tag_id FROM product_tags
			WHERE product_id = ?
		)
		AND shop_id = ?`, productID, shopID).Scan(&tags).Error

	if err != nil {
		return nil, err
	}

	result := make([]interface{}, len(tags))
	for i, tag := range tags {
		result[i] = tag
	}
	return result, nil
}

// BatchTagProducts 批量打标签
func (s *ShopService) BatchTagProducts(tagID int, productIDs []string, shopID uint64) error {
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	for _, productID := range productIDs {
		var existingCount int64
		if err := tx.Raw("SELECT COUNT(*) FROM product_tags WHERE product_id = ? AND tag_id = ? AND shop_id = ?",
			productID, tagID, shopID).Count(&existingCount).Error; err != nil {
			tx.Rollback()
			return err
		}

		if existingCount == 0 {
			if err := tx.Exec("INSERT INTO product_tags (product_id, tag_id, shop_id) VALUES (?, ?, ?)",
				productID, tagID, shopID).Error; err != nil {
				tx.Rollback()
				return err
			}
		}
	}

	return tx.Commit().Error
}

// BatchUntagProducts 批量解绑商品标签
func (s *ShopService) BatchUntagProducts(tagID int, productIDs []string, shopID uint64) error {
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	for _, productID := range productIDs {
		if err := tx.Exec("DELETE FROM product_tags WHERE product_id = ? AND tag_id = ?",
			productID, tagID).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit().Error
}

// BatchTagProduct 批量设置商品标签
func (s *ShopService) BatchTagProduct(productID string, tagIDs []string, shopID uint64) error {
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 先删除该商品的所有标签关联
	if err := tx.Exec("DELETE FROM product_tags WHERE product_id = ? AND shop_id = ?", productID, shopID).Error; err != nil {
		tx.Rollback()
		return err
	}

	// 批量插入新的标签关联
	for _, tagID := range tagIDs {
		if err := tx.Exec("INSERT INTO product_tags (product_id, tag_id, shop_id) VALUES (?, ?, ?)",
			productID, tagID, shopID).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit().Error
}

// GetTagBoundProducts 获取标签已绑定的商品列表
func (s *ShopService) GetTagBoundProducts(tagID string, shopID uint64, page, pageSize int) (map[string]interface{}, error) {
	type ProductResult struct {
		ID          uint64  `json:"id"`
		ShopID      uint64  `json:"shop_id"`
		Name        string  `json:"name"`
		Description string  `json:"description"`
		Price       float64 `json:"price"`
		Stock       int     `json:"stock"`
		Status      string  `json:"status"`
		ImageURL    string  `json:"image_url"`
	}

	offset := (page - 1) * pageSize

	var total int64
	var products []ProductResult

	// tag_id=-1 表示查询未绑定任何标签的商品
	if tagID == "-1" {
		// 查询未绑定任何标签的商品
		if err := s.db.Raw(`
			SELECT COUNT(*) FROM products
			WHERE shop_id = ? AND id NOT IN (
				SELECT product_id FROM product_tags WHERE shop_id = ?
			)`, shopID, shopID).Scan(&total).Error; err != nil {
			return nil, err
		}

		if err := s.db.Raw(`
			SELECT id, shop_id, name, description, price, stock, status, image_url
			FROM products
			WHERE shop_id = ? AND id NOT IN (
				SELECT product_id FROM product_tags WHERE shop_id = ?
			)
			ORDER BY created_at DESC
			LIMIT ? OFFSET ?`, shopID, shopID, pageSize, offset).Scan(&products).Error; err != nil {
			return nil, err
		}
	} else {
		// 查询指定标签绑定的商品
		if err := s.db.Raw(`
			SELECT COUNT(*) FROM products
			JOIN product_tags ON product_tags.product_id = products.id
			WHERE product_tags.tag_id = ? AND products.shop_id = ?`, tagID, shopID).Scan(&total).Error; err != nil {
			return nil, err
		}

		if err := s.db.Raw(`
			SELECT products.id, products.shop_id, products.name, products.description,
			       products.price, products.stock, products.status, products.image_url
			FROM products
			JOIN product_tags ON product_tags.product_id = products.id
			WHERE product_tags.tag_id = ? AND products.shop_id = ?
			ORDER BY products.created_at DESC
			LIMIT ? OFFSET ?`, tagID, shopID, pageSize, offset).Scan(&products).Error; err != nil {
			return nil, err
		}
	}

	result := make([]interface{}, len(products))
	for i, p := range products {
		result[i] = p
	}

	return map[string]interface{}{
		"total":     total,
		"page":      page,
		"page_size": pageSize,
		"data":      result,
	}, nil
}

func (h *Handler) getUnboundProducts(shopID snowflake.ID, page int, pageSize int) ([]models.Product, int64, error) {
	offset := (page - 1) * pageSize
	var products []models.Product
	var total int64

	query := h.DB.Model(&models.Product{}).
		Where("shop_id = ? AND id NOT IN (SELECT product_id FROM product_tags WHERE shop_id = ?)", shopID, shopID)

	// 获取总数
	if err := query.
		Model(&models.Product{}).
		Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Offset(offset).
		Limit(pageSize).Order("created_at DESC").
		Preload("OptionCategories.Options").
		Find(&products).Error; err != nil {
		return nil, 0, err
	}

	return products, total, nil
}

// GetUnboundProductsForTag 获取标签未绑定的商品列表
func (s *ShopService) GetUnboundProductsForTag(tagID string, shopID uint64, page, pageSize int) (map[string]interface{}, error) {
	type ProductResult struct {
		ID          uint64  `json:"id"`
		ShopID      uint64  `json:"shop_id"`
		Name        string  `json:"name"`
		Description string  `json:"description"`
		Price       float64 `json:"price"`
		Stock       int     `json:"stock"`
		Status      string  `json:"status"`
		ImageURL    string  `json:"image_url"`
	}

	offset := (page - 1) * pageSize

	var total int64
	if err := s.db.Raw(`
		SELECT COUNT(*) FROM products
		WHERE id NOT IN (
			SELECT product_id FROM product_tags WHERE tag_id = ?
		)
		AND shop_id = ?`, tagID, shopID).Scan(&total).Error; err != nil {
		return nil, err
	}

	var products []ProductResult
	if err := s.db.Raw(`
		SELECT id, shop_id, name, description, price, stock, status, image_url
		FROM products
		WHERE id NOT IN (
			SELECT product_id FROM product_tags WHERE tag_id = ?
		)
		AND shop_id = ?
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?`, tagID, shopID, pageSize, offset).Scan(&products).Error; err != nil {
		return nil, err
	}

	result := make([]interface{}, len(products))
	for i, p := range products {
		result[i] = p
	}

	return map[string]interface{}{
		"total":     total,
		"page":      page,
		"page_size": pageSize,
		"data":      result,
	}, nil
}

// GetUnboundTagsList 获取没有绑定商品的标签列表
func (s *ShopService) GetUnboundTagsList(shopID uint64) ([]interface{}, error) {
	type TagResult struct {
		ID          uint64 `json:"id"`
		ShopID      uint64 `json:"shop_id"`
		Name        string `json:"name"`
		Description string `json:"description"`
	}

	var tags []TagResult
	err := s.db.Raw(`
		SELECT tags.id, tags.shop_id, tags.name, tags.description FROM tags
		WHERE tags.shop_id = ?
		AND tags.id NOT IN (
			SELECT DISTINCT tag_id FROM product_tags
		)`, shopID).Scan(&tags).Error

	if err != nil {
		return nil, err
	}

	result := make([]interface{}, len(tags))
	for i, tag := range tags {
		result[i] = tag
	}
	return result, nil
}

// GetTagOnlineProducts 获取标签关联的已上架商品
func (s *ShopService) GetTagOnlineProducts(tagID string, shopID uint64) ([]interface{}, error) {
	type ProductResult struct {
		ID          uint64  `json:"id"`
		ShopID      uint64  `json:"shop_id"`
		Name        string  `json:"name"`
		Description string  `json:"description"`
		Price       float64 `json:"price"`
		Stock       int     `json:"stock"`
		Status      string  `json:"status"`
		ImageURL    string  `json:"image_url"`
	}

	var products []ProductResult
	if err := s.db.Raw(`
		SELECT products.id, products.shop_id, products.name, products.description,
		       products.price, products.stock, products.status, products.image_url
		FROM products
		JOIN product_tags ON product_tags.product_id = products.id
		WHERE product_tags.tag_id = ? AND products.shop_id = ? AND products.status = 'online'
		ORDER BY products.created_at DESC`, tagID, shopID).Scan(&products).Error; err != nil {
		return nil, err
	}

	result := make([]interface{}, len(products))
	for i, p := range products {
		result[i] = p
	}
	return result, nil
}
