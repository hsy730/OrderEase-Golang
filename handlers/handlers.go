package handlers

import (
	"errors"
	"log"
	"orderease/models"
	"orderease/utils/log2"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

const (
	maxFileSize = 32 << 20 // 32MB
)

type Handler struct {
	DB     *gorm.DB
	logger *log.Logger
}

// 创建处理器实例
func NewHandler(db *gorm.DB) *Handler {
	return &Handler{
		DB:     db,
		logger: log2.Logger,
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

func (h *Handler) applyShopIdPolicy(c *gin.Context, setShopFunc func(models.UserInfo) error) error {
	requestUser, err := h.getRequestUserInfo(c)
	if err != nil {
		return errors.New("获取用户信息失败")
	}

	return setShopFunc(*requestUser) // 非管理员，设置shopID为用户ID
}
