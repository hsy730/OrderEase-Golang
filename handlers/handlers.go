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

func (h *Handler) getUserInfo(c *gin.Context) (uint, error) {

	user_info, ok := c.Get("userInfo")
	if !ok {
		return 0, errors.New("未找到用户信息")
	}

	userInfo, ok := user_info.(models.UserInfo)
	if !ok {
		return 0, errors.New("用户信息格式错误")
	}

	// 根据商家用户名获取商户ishopid
	var shop models.Shop
	if err := h.DB.Where("owner_username = ?", userInfo.Username).First(&shop).Error; err != nil {
		return 0, errors.New("未找到商家用户名")

	}
	return shop.ID, nil
}
