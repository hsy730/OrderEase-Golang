package services

import (
	"errors"
	"fmt"
	"orderease/application/dto"
	"orderease/domain/order"
	"orderease/domain/product"
	"orderease/domain/shared"
	"orderease/domain/shop"
	"orderease/utils"
	"orderease/utils/log2"
	"os"
	"path/filepath"
	"time"

	"gorm.io/gorm"
)

type ShopService struct {
	shopRepo       shop.ShopRepository
	tagRepo        shop.TagRepository
	productRepo    product.ProductRepository
	db             *gorm.DB
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

func (s *ShopService) UploadShopImage(id shared.ID, file *os.File, filename string) (string, error) {
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

func (s *ShopService) toShopResponse(shopEntity *shop.Shop) *dto.ShopResponse {
	return &dto.ShopResponse{
		ID:            shopEntity.ID,
		Name:          shopEntity.Name,
		OwnerUsername: shopEntity.OwnerUsername,
		ContactPhone:  shopEntity.ContactPhone,
		ContactEmail:  shopEntity.ContactEmail,
		Address:       shopEntity.Address,
		Description:   shopEntity.Description,
		CreatedAt:     shopEntity.CreatedAt,
		UpdatedAt:     shopEntity.UpdatedAt,
		ValidUntil:    shopEntity.ValidUntil,
		Settings:      shopEntity.Settings,
		ImageURL:      shopEntity.ImageURL,
		OrderStatusFlow: shopEntity.OrderStatusFlow,
	}
}
