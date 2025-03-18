package handlers

import (
	"archive/zip"
	"encoding/csv"
	"fmt"
	"io"
	"net/http"
	"orderease/models"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"orderease/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// 导入数据
func (h *Handler) ImportData(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		errorResponse(c, http.StatusBadRequest, "请上传ZIP文件")
		return
	}

	if !strings.HasSuffix(file.Filename, ".zip") {
		errorResponse(c, http.StatusBadRequest, "只支持ZIP文件")
		return
	}

	f, err := file.Open()
	if err != nil {
		h.logger.Printf("打开上传文件失败: %v", err)
		errorResponse(c, http.StatusInternalServerError, "文件处理失败")
		return
	}
	defer f.Close()

	// 读取 ZIP 文件
	zipReader, err := zip.NewReader(f, file.Size)
	if err != nil {
		h.logger.Printf("读取ZIP文件失败: %v", err)
		errorResponse(c, http.StatusInternalServerError, "文件处理失败")
		return
	}

	tx := h.DB.Begin()

	// 按照依赖关系的反序清空表（先清除依赖表，再清除主表）
	tablesToClean := []interface{}{
		&models.OrderItem{},  // 先清除订单项
		&models.Order{},      // 再清除订单
		&models.ProductTag{}, // 清除产品标签关联
		&models.Tag{},        // 清除标签
		&models.Product{},    // 清除商品
		&models.User{},       // 最后清除用户
	}

	// 清空所有表
	for _, table := range tablesToClean {
		if err := tx.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(table).Error; err != nil {
			tx.Rollback()
			h.logger.Printf("清空表失败: %v", err)
			errorResponse(c, http.StatusInternalServerError, "清空数据失败")
			return
		}
	}

	// 按照依赖关系顺序处理文件
	fileOrder := []string{
		"users.csv",        // 先导入用户
		"products.csv",     // 再导入商品
		"tags.csv",         // 然后是标签
		"product_tags.csv", // 产品标签关联
		"orders.csv",       // 然后是订单
		"order_items.csv",  // 最后是订单项
	}

	// 创建一个文件映射，方便查找
	fileMap := make(map[string]*zip.File)
	for _, zipFile := range zipReader.File {
		fileMap[zipFile.Name] = zipFile
	}

	// 按顺序导入文件
	for _, fileName := range fileOrder {
		if zipFile, ok := fileMap[fileName]; ok {
			if err := importCSVFile(tx, zipFile); err != nil {
				tx.Rollback()
				h.logger.Printf("导入数据失败: %v", err)
				errorResponse(c, http.StatusInternalServerError, fmt.Sprintf("导入 %s 失败: %v", fileName, err))
				return
			}
		} else {
			h.logger.Printf("警告: 未找到文件 %s", fileName)
		}
	}

	tx.Commit()
	successResponse(c, gin.H{"message": "数据导入成功"})
}

// 修改 importCSVFile 函数，添加错误处理和日志
func importCSVFile(tx *gorm.DB, zipFile *zip.File) error {
	f, err := zipFile.Open()
	if err != nil {
		return fmt.Errorf("打开文件失败: %v", err)
	}
	defer f.Close()

	reader := csv.NewReader(f)

	// 读取表头
	headers, err := reader.Read()
	if err != nil {
		return fmt.Errorf("读取CSV表头失败: %v", err)
	}

	// 逐行读取并导入数据
	lineNum := 1 // 从第一行开始计数（表头算第一行）
	for {
		lineNum++
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("第 %d 行读取失败: %v", lineNum, err)
		}
		if len(record) != len(headers) {
			return fmt.Errorf("第 %d 行数据格式不正确: 期望 %d 列，实际 %d 列",
				lineNum, len(headers), len(record))
		}

		var importErr error
		switch filepath.Base(zipFile.Name) {
		case "users.csv":
			importErr = importUserRecord(tx, record)
		case "products.csv":
			importErr = importProductRecord(tx, record)
		case "tags.csv":
			importErr = importTagRecord(tx, record)
		case "product_tags.csv":
			importErr = importProductTagRecord(tx, record)
		case "orders.csv":
			importErr = importOrderRecord(tx, record)
		case "order_items.csv":
			importErr = importOrderItemRecord(tx, record)
		default:
			return fmt.Errorf("未知的CSV文件: %s", zipFile.Name)
		}

		if importErr != nil {
			return fmt.Errorf("第 %d 行导入失败: %v", lineNum, importErr)
		}
	}

	return nil
}

// 辅助函数：导入用户记录
func importUserRecord(tx *gorm.DB, record []string) error {
	createdAt, _ := time.Parse(time.RFC3339, record[5])
	updatedAt, _ := time.Parse(time.RFC3339, record[6])

	id, err := utils.StringToSnowflakeID(record[0])
	if err != nil {
		return fmt.Errorf("解析用户ID失败: %v", err)
	}
	user := models.User{
		ID:        id,
		Name:      record[1],
		Phone:     record[2],
		Address:   record[3],
		Type:      record[4],
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
	}
	return tx.Create(&user).Error
}

// 辅助函数：导入商品记录
func importProductRecord(tx *gorm.DB, record []string) error {
	createdAt, _ := time.Parse(time.RFC3339, record[6])
	updatedAt, _ := time.Parse(time.RFC3339, record[7])

	id, err := utils.StringToSnowflakeID(record[0])
	if err != nil {
		return fmt.Errorf("解析商品ID失败: %v", err)
	}
	product := models.Product{
		ID:          id,
		Name:        record[1],
		Description: record[2],
		Price:       parseFloat(record[3]),
		Stock:       parseInt(record[4]),
		ImageURL:    record[5],
		CreatedAt:   createdAt,
		UpdatedAt:   updatedAt,
	}
	return tx.Create(&product).Error
}

// 辅助函数：导入标签记录
func importTagRecord(tx *gorm.DB, record []string) error {
	createdAt, _ := time.Parse(time.RFC3339, record[3])
	updatedAt, _ := time.Parse(time.RFC3339, record[4])

	tag := models.Tag{
		ID:          parseInt(record[0]),
		Name:        record[1],
		Description: record[2],

		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
	}
	return tx.Create(&tag).Error
}

// 辅助函数：导入产品标签关联记录
func importProductTagRecord(tx *gorm.DB, record []string) error {
	createdAt, _ := time.Parse(time.RFC3339, record[2])
	updatedAt, _ := time.Parse(time.RFC3339, record[3])

	productId, err := utils.StringToSnowflakeID(record[0])
	if err != nil {
		return fmt.Errorf("解析商品ID失败: %v", err)
	}

	productTag := models.ProductTag{
		ProductID: productId,
		TagID:     parseInt(record[1]),
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
	}
	return tx.Create(&productTag).Error
}

// 辅助函数：导入订单记录
func importOrderRecord(tx *gorm.DB, record []string) error {
	createdAt, _ := time.Parse(time.RFC3339, record[5])
	updatedAt, _ := time.Parse(time.RFC3339, record[6])

	id, err := utils.StringToSnowflakeID(record[0])
	if err != nil {
		return fmt.Errorf("解析订单ID失败: %v", err)
	}

	userID, err := utils.StringToSnowflakeID(record[1])
	if err != nil {
		return fmt.Errorf("解析订单UserID失败: %v", err)
	}
	order := models.Order{
		ID:         id,
		UserID:     userID,
		TotalPrice: models.Price(parseFloat(record[2])),
		Status:     record[3],
		Remark:     record[4],
		CreatedAt:  createdAt,
		UpdatedAt:  updatedAt,
	}
	return tx.Create(&order).Error
}

// 辅助函数：导入订单项记录
func importOrderItemRecord(tx *gorm.DB, record []string) error {

	id, err := utils.StringToSnowflakeID(record[0])
	if err != nil {
		return fmt.Errorf("解析订单项ID失败: %v", err)
	}
	orderId, err := utils.StringToSnowflakeID(record[1])
	if err != nil {
		return fmt.Errorf("解析订单ID失败: %v", err)
	}
	productID, err := utils.StringToSnowflakeID(record[2])
	if err != nil {
		return fmt.Errorf("解析订单ID失败: %v", err)
	}

	orderItem := models.OrderItem{
		ID:        id,
		OrderID:   orderId,
		ProductID: productID,
		Quantity:  parseInt(record[3]),
		Price:     models.Price(parseFloat(record[4])),
	}
	return tx.Create(&orderItem).Error
}

// 辅助函数：解析浮点数
func parseFloat(s string) float64 {
	f, _ := strconv.ParseFloat(s, 64)
	return f
}

// 辅助函数：解析整数
func parseInt(s string) int {
	i, _ := strconv.Atoi(s)
	return i
}
