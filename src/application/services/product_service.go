package services

import (
	"errors"
	"fmt"
	"orderease/application/dto"
	"orderease/domain/product"
	"orderease/domain/shared"
	"orderease/utils"
	"orderease/utils/log2"
	"os"
	"path/filepath"
	"time"

	"gorm.io/gorm"
)

type ProductService struct {
	productRepo         product.ProductRepository
	productCategoryRepo product.ProductOptionCategoryRepository
	productOptionRepo   product.ProductOptionRepository
	productTagRepo      product.ProductTagRepository
	orderItemRepo       product.ProductRepository
	db                  *gorm.DB
}

func NewProductService(
	productRepo product.ProductRepository,
	productCategoryRepo product.ProductOptionCategoryRepository,
	productOptionRepo product.ProductOptionRepository,
	productTagRepo product.ProductTagRepository,
	orderItemRepo product.ProductRepository,
	db *gorm.DB,
) *ProductService {
	return &ProductService{
		productRepo:         productRepo,
		productCategoryRepo: productCategoryRepo,
		productOptionRepo:   productOptionRepo,
		productTagRepo:      productTagRepo,
		orderItemRepo:       orderItemRepo,
		db:                  db,
	}
}

func (s *ProductService) CreateProduct(req *dto.CreateProductRequest) (*dto.ProductResponse, error) {
	prod, err := product.NewProduct(req.ShopID, req.Name, req.Description, shared.Price(req.Price), req.Stock)
	if err != nil {
		return nil, err
	}

	prod.ID = shared.ID(utils.GenerateSnowflakeID())
	prod.Status = product.ProductStatusPending
	prod.ImageURL = req.ImageURL

	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if err := s.productRepo.Save(prod); err != nil {
		tx.Rollback()
		return nil, errors.New("创建商品失败")
	}

	for _, catReq := range req.OptionCategories {
		cat, err := product.NewProductOptionCategory(prod.ID, catReq.Name, catReq.IsRequired, catReq.IsMultiple, catReq.DisplayOrder)
		if err != nil {
			tx.Rollback()
			return nil, err
		}

		cat.ID = shared.ID(utils.GenerateSnowflakeID())

		if err := s.productCategoryRepo.Save(cat); err != nil {
			tx.Rollback()
			return nil, errors.New("创建商品参数类别失败")
		}

		for _, optReq := range catReq.Options {
			opt, err := product.NewProductOption(cat.ID, optReq.Name, optReq.PriceAdjustment, optReq.IsDefault, optReq.DisplayOrder)
			if err != nil {
				tx.Rollback()
				return nil, err
			}

			opt.ID = shared.ID(utils.GenerateSnowflakeID())

			if err := s.productOptionRepo.Save(opt); err != nil {
				tx.Rollback()
				return nil, errors.New("创建商品参数选项失败")
			}
		}
	}

	tx.Commit()

	return s.getProductResponse(prod.ID)
}

func (s *ProductService) GetProduct(id shared.ID, shopID uint64) (*dto.ProductResponse, error) {
	prod, err := s.productRepo.FindByIDAndShopID(id, shopID)
	if err != nil {
		return nil, err
	}

	return s.toProductResponse(prod), nil
}

func (s *ProductService) GetProducts(shopID uint64, page, pageSize int, search string) (*dto.ProductListResponse, error) {
	products, total, err := s.productRepo.FindByShopID(shopID, page, pageSize, search, true)
	if err != nil {
		return nil, err
	}

	data := make([]dto.ProductResponse, len(products))
	for i, prod := range products {
		data[i] = *s.toProductResponse(&prod)
	}

	return &dto.ProductListResponse{
		Total:    total,
		Page:     page,
		PageSize: pageSize,
		Data:     data,
	}, nil
}

func (s *ProductService) UpdateProduct(id shared.ID, shopID uint64, req *dto.CreateProductRequest) (*dto.ProductResponse, error) {
	prod, err := s.productRepo.FindByIDAndShopID(id, shopID)
	if err != nil {
		return nil, err
	}

	prod.Name = req.Name
	prod.Description = req.Description
	prod.Price = shared.Price(req.Price)
	prod.Stock = req.Stock
	prod.ImageURL = req.ImageURL

	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if err := s.productRepo.Update(prod); err != nil {
		tx.Rollback()
		return nil, errors.New("更新商品失败")
	}

	if err := s.productCategoryRepo.DeleteByProductID(prod.ID); err != nil {
		tx.Rollback()
		return nil, errors.New("删除商品参数类别失败")
	}

	for _, catReq := range req.OptionCategories {
		cat, err := product.NewProductOptionCategory(prod.ID, catReq.Name, catReq.IsRequired, catReq.IsMultiple, catReq.DisplayOrder)
		if err != nil {
			tx.Rollback()
			return nil, err
		}

		cat.ID = shared.ID(utils.GenerateSnowflakeID())

		if err := s.productCategoryRepo.Save(cat); err != nil {
			tx.Rollback()
			return nil, errors.New("创建商品参数类别失败")
		}

		for _, optReq := range catReq.Options {
			opt, err := product.NewProductOption(cat.ID, optReq.Name, optReq.PriceAdjustment, optReq.IsDefault, optReq.DisplayOrder)
			if err != nil {
				tx.Rollback()
				return nil, err
			}

			opt.ID = shared.ID(utils.GenerateSnowflakeID())

			if err := s.productOptionRepo.Save(opt); err != nil {
				tx.Rollback()
				return nil, errors.New("创建商品参数选项失败")
			}
		}
	}

	tx.Commit()

	return s.getProductResponse(prod.ID)
}

func (s *ProductService) DeleteProduct(id shared.ID, shopID uint64) error {
	prod, err := s.productRepo.FindByIDAndShopID(id, shopID)
	if err != nil {
		return err
	}

	count, err := s.productRepo.CountByProductID(prod.ID)
	if err != nil {
		return err
	}

	if count > 0 {
		return fmt.Errorf("该商品有 %d 个关联订单，不能删除。建议将商品下架而不是删除", count)
	}

	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if prod.ImageURL != "" {
		imagePath := filepath.Join("./uploads/products", prod.ImageURL)
		if err := os.Remove(imagePath); err != nil && !os.IsNotExist(err) {
			log2.Errorf("删除商品图片失败: %v", err)
		}
	}

	if err := s.productOptionRepo.DeleteByCategoryID(prod.ID); err != nil {
		tx.Rollback()
		return errors.New("删除商品参数选项失败")
	}

	if err := s.productCategoryRepo.DeleteByProductID(prod.ID); err != nil {
		tx.Rollback()
		return errors.New("删除商品参数类别失败")
	}

	if err := s.productTagRepo.DeleteByProductID(prod.ID); err != nil {
		tx.Rollback()
		return errors.New("删除商品标签关联失败")
	}

	if err := s.productRepo.Delete(prod.ID); err != nil {
		tx.Rollback()
		return errors.New("删除商品失败: " + err.Error())
	}

	tx.Commit()

	return nil
}

func (s *ProductService) UpdateProductStatus(req *dto.UpdateProductStatusRequest, shopID uint64) error {
	prod, err := s.productRepo.FindByIDAndShopID(req.ID, shopID)
	if err != nil {
		return err
	}

	if err := prod.ChangeStatus(req.Status); err != nil {
		return err
	}

	if err := s.productRepo.Update(prod); err != nil {
		return errors.New("更新商品状态失败")
	}

	return nil
}

func (s *ProductService) UploadProductImage(id shared.ID, shopID uint64, file *os.File, filename string) (string, error) {
	prod, err := s.productRepo.FindByIDAndShopID(id, shopID)
	if err != nil {
		return "", err
	}

	uploadDir := "./uploads/products"
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		return "", errors.New("创建上传目录失败")
	}

	if prod.ImageURL != "" {
		oldImagePath := filepath.Join(uploadDir, prod.ImageURL)
		if err := os.Remove(oldImagePath); err != nil && !os.IsNotExist(err) {
			log2.Errorf("删除旧图片失败: %v", err)
		}
	}

	newFilename := fmt.Sprintf("product_%d_%d%s", id, time.Now().Unix(), filepath.Ext(filename))
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

	prod.ImageURL = newFilename
	if err := s.productRepo.Update(prod); err != nil {
		return "", errors.New("更新商品图片失败")
	}

	return newFilename, nil
}

func (s *ProductService) getProductResponse(id shared.ID) (*dto.ProductResponse, error) {
	prod, err := s.productRepo.FindByID(id)
	if err != nil {
		return nil, err
	}
	return s.toProductResponse(prod), nil
}

func (s *ProductService) toProductResponse(prod *product.Product) *dto.ProductResponse {
	categories := make([]dto.ProductOptionCategoryResponse, len(prod.OptionCategories))
	for i, cat := range prod.OptionCategories {
		options := make([]dto.ProductOptionResponse, len(cat.Options))
		for j, opt := range cat.Options {
			options[j] = dto.ProductOptionResponse{
				ID:              opt.ID,
				CategoryID:      opt.CategoryID,
				Name:            opt.Name,
				PriceAdjustment: opt.PriceAdjustment,
				DisplayOrder:    opt.DisplayOrder,
				IsDefault:       opt.IsDefault,
			}
		}

		categories[i] = dto.ProductOptionCategoryResponse{
			ID:           cat.ID,
			ProductID:    cat.ProductID,
			Name:         cat.Name,
			IsRequired:   cat.IsRequired,
			IsMultiple:   cat.IsMultiple,
			DisplayOrder: cat.DisplayOrder,
			Options:      options,
		}
	}

	return &dto.ProductResponse{
		ID:               prod.ID,
		ShopID:           prod.ShopID,
		Name:             prod.Name,
		Description:      prod.Description,
		Price:            prod.Price,
		Stock:            prod.Stock,
		ImageURL:         prod.ImageURL,
		Status:           prod.Status,
		CreatedAt:        prod.CreatedAt,
		UpdatedAt:        prod.UpdatedAt,
		OptionCategories: categories,
	}
}
