package handlers

import (
	"log"
	"orderease/utils"

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
		logger: utils.Logger,
	}
}
