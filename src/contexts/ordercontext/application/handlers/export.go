package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// ExportData 导出所有数据（CSV + ZIP）
func (h *Handler) ExportData(c *gin.Context) {
	if err := h.exportService.ExportAllData(c); err != nil {
		h.logger.Errorf("导出数据失败: %v", err)
		errorResponse(c, http.StatusInternalServerError, "导出失败")
		return
	}
}
