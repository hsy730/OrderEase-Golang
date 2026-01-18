package handlers

import (
	"errors"
	"orderease/models"
	"orderease/services"
	"orderease/utils/log2"

	"orderease/repositories"

	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

const (
	maxFileSize = 32 << 20 // 32MB
)

type Handler struct {
	DB               *gorm.DB
	logger           *log2.Logger
	productRepo      *repositories.ProductRepository
	userRepo         *repositories.UserRepository
	adminRepo        *repositories.AdminRepository
	tempTokenService *services.TempTokenService
}

// 创建处理器实例
func NewHandler(db *gorm.DB) *Handler {
	return &Handler{
		DB:               db,
		productRepo:      repositories.NewProductRepository(db),
		userRepo:         repositories.NewUserRepository(db),
		adminRepo:        repositories.NewAdminRepository(db),
		logger:           log2.GetLogger(),
		tempTokenService: services.NewTempTokenService(),
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

func (h *Handler) validAndReturnShopID(c *gin.Context, shopID uint64) (uint64, error) {
	// 如果是管理端接口，普通用户（店主）需要使用绑定的shopId
	if strings.Contains(c.Request.URL.Path, "/shopOwner/") {
		requestUser, err := h.getRequestUserInfo(c)
		if err != nil {
			return 0, errors.New("获取用户信息失败")
		}
		if !requestUser.IsAdmin {
			shopID = requestUser.UserID // 非管理员，设置shopID为用户ID
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
