package handlers

import (
	"errors"
	"orderease/contexts/ordercontext/domain/media"
	"orderease/contexts/ordercontext/domain/order"
	"orderease/contexts/ordercontext/domain/product"
	"orderease/contexts/ordercontext/domain/shop"
	"orderease/contexts/ordercontext/domain/tag"
	"orderease/contexts/ordercontext/domain/user"
	"orderease/contexts/ordercontext/infrastructure/repositories"
	"orderease/models"
	"orderease/services"
	"orderease/utils/log2"

	"strings"

	"github.com/bwmarrin/snowflake"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

const (
	maxFileSize = 32 << 20 // 32MB
)

type Handler struct {
	DB                    *gorm.DB
	logger                *log2.Logger
	productRepo           *repositories.ProductRepository
	userRepo              *repositories.UserRepository
	adminRepo             *repositories.AdminRepository
	orderRepo             *repositories.OrderRepository
	shopRepo              *repositories.ShopRepository
	tagRepo               *repositories.TagRepository
	tokenRepo             *repositories.TokenRepository
	dashboardRepo         *repositories.DashboardRepository
	tempTokenService       *services.TempTokenService
	userDomain            *user.Service
	orderService          *order.Service
	productService        *product.Service
	mediaService          *media.ImageUploadService
	shopService           *shop.Service
	tagService            *tag.Service
	miniProgramAuthHandler *MiniProgramAuthHandler
}

// 创建处理器实例
func NewHandler(db *gorm.DB) *Handler {
	userRepo := repositories.NewUserRepository(db)

	// 创建 Repository 适配器，将 repositories.UserRepository 适配到 domain.Repository
	userRepoAdapter := user.NewRepositoryAdapter(
		// createFunc
		func(u *user.User) error {
			model := u.ToModel()
			return userRepo.Create(model)
		},
		// getByIDFunc
		func(id user.UserID) (*user.User, error) {
			// 从 models.User 转换为 domain.User
			model, err := userRepo.GetUserByID(string(id))
			if err != nil {
				return nil, err
			}
			return user.UserFromModel(model), nil
		},
		// getByUsernameFunc
		func(username string) (*user.User, error) {
			// 需要先在 UserRepository 中添加 GetByUsername 方法
			// 暂时返回错误，表示未实现
			return nil, errors.New("GetByUsername not implemented")
		},
		// phoneExistsFunc
		func(phone string) (bool, error) {
			return userRepo.CheckPhoneExists(phone)
		},
		// usernameExistsFunc
		func(username string) (bool, error) {
			return userRepo.CheckUsernameExists(username)
		},
		// updateFunc
		func(u *user.User) error {
			model := u.ToModel()
			return userRepo.Update(model)
		},
		// deleteFunc
		func(u *user.User) error {
			model := u.ToModel()
			return userRepo.Delete(model)
		},
	)

	userDomain := user.NewService(userRepoAdapter)
	orderService := order.NewService(db)
	productService := product.NewService(db)
	mediaService := media.NewImageUploadService(log2.GetLogger())
	shopService := shop.NewService(db)
	tagService := tag.NewService(db)

	// 初始化小程序认证处理器
	miniProgramHandler, err := NewMiniProgramAuthHandler(db)
	if err != nil {
		log2.GetLogger().Warnf("小程序认证处理器初始化失败: %v", err)
		miniProgramHandler = nil
	}

	return &Handler{
		DB:                    db,
		productRepo:           repositories.NewProductRepository(db),
		userRepo:              userRepo,
		adminRepo:             repositories.NewAdminRepository(db),
		orderRepo:             repositories.NewOrderRepository(db),
		shopRepo:              repositories.NewShopRepository(db),
		tagRepo:               repositories.NewTagRepository(db),
		tokenRepo:             repositories.NewTokenRepository(db),
		dashboardRepo:         repositories.NewDashboardRepository(db),
		logger:                log2.GetLogger(),
		tempTokenService:       services.NewTempTokenService(db),
		userDomain:            userDomain,
		orderService:          orderService,
		productService:        productService,
		mediaService:          mediaService,
		shopService:           shopService,
		tagService:            tagService,
		miniProgramAuthHandler: miniProgramHandler,
	}
}

func (h *Handler) getRequestUserInfo(c *gin.Context) (*models.UserInfo, error) {
	user_info, ok := c.Get("userInfo")
	if !ok {
		return nil, errors.New("未找到用户信息")
	}

	userInfo, ok := user_info.(models.UserInfo)
	if !ok {
		return nil, errors.New("用户信息格式错误")
	}

	return &userInfo, nil
}

func (h *Handler) validAndReturnShopID(c *gin.Context, shopID snowflake.ID) (snowflake.ID, error) {
	// 如果是管理端接口，普通用户（店主）需要使用绑定的shopId
	if strings.Contains(c.Request.URL.Path, "/shopOwner/") {
		requestUser, err := h.getRequestUserInfo(c)
		if err != nil {
			return 0, errors.New("获取用户信息失败")
		}
		if !requestUser.IsAdmin {
			shopID = snowflake.ID(requestUser.UserID) // 非管理员，设置shopID为用户ID
		}
	} else {
		log2.Debugf("非管理端接口，shopID: %d", shopID)
	}

	exist, err := h.productRepo.CheckShopExists(shopID)
	if err != nil {
		return 0, err
	}
	if !exist {
		return 0, errors.New("店铺不存在")
	}
	return shopID, nil
}

// checkShopExpiration 检查店铺是否过期
func (h *Handler) checkShopExpiration(shopModel *models.Shop) error {
	shopDomain := shop.ShopFromModel(shopModel)
	if shopDomain.IsExpired() {
		return errors.New("店铺服务已到期")
	}
	return nil
}

// GetMiniProgramAuthHandler 获取小程序认证处理器
func (h *Handler) GetMiniProgramAuthHandler() *MiniProgramAuthHandler {
	return h.miniProgramAuthHandler
}
