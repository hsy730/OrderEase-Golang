package handlers

import (
	"archive/zip"
	"bytes"
	"encoding/csv"
	"fmt"
	"net/http"
	"orderease/database"
	"orderease/models"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// 导出数据
func (h *Handler) ExportData(c *gin.Context) {
	// 生成带时间戳的文件名
	timestamp := time.Now().Format("20060102_150405")
	zipFilename := fmt.Sprintf("export_%s.zip", timestamp)

	// 创建一个缓冲区来保存 ZIP 文件
	buf := new(bytes.Buffer)
	zipWriter := zip.NewWriter(buf)

	// 导出用户数据
	if err := exportTableToCSV(h.DB, zipWriter, "users.csv", &[]models.User{}); err != nil {
		h.logger.Printf("导出用户数据失败: %v", err)
		errorResponse(c, http.StatusInternalServerError, "导出失败")
		return
	}

	// 导出商品数据
	if err := exportTableToCSV(h.DB, zipWriter, "products.csv", &[]models.Product{}); err != nil {
		h.logger.Printf("导出商品数据失败: %v", err)
		errorResponse(c, http.StatusInternalServerError, "导出失败")
		return
	}

	// 导出标签数据
	if err := exportTableToCSV(h.DB, zipWriter, "tags.csv", &[]models.Tag{}); err != nil {
		h.logger.Printf("导出标签数据失败: %v", err)
		errorResponse(c, http.StatusInternalServerError, "导出失败")
		return
	}

	// 导出产品标签关联数据
	if err := exportTableToCSV(h.DB, zipWriter, "product_tags.csv", &[]models.ProductTag{}); err != nil {
		h.logger.Printf("导出产品标签关联数据失败: %v", err)
		errorResponse(c, http.StatusInternalServerError, "导出失败")
		return
	}

	// 导出订单数据
	if err := exportTableToCSV(h.DB, zipWriter, "orders.csv", &[]models.Order{}); err != nil {
		h.logger.Printf("导出订单数据失败: %v", err)
		errorResponse(c, http.StatusInternalServerError, "导出失败")
		return
	}

	// 导出订单项数据
	if err := exportTableToCSV(h.DB, zipWriter, "order_items.csv", &[]models.OrderItem{}); err != nil {
		h.logger.Printf("导出订单项数据失败: %v", err)
		errorResponse(c, http.StatusInternalServerError, "导出失败")
		return
	}

	// 关闭 ZIP writer
	if err := zipWriter.Close(); err != nil {
		h.logger.Printf("关闭 ZIP writer 失败: %v", err)
		errorResponse(c, http.StatusInternalServerError, "导出失败")
		return
	}

	// 设置响应头为 ZIP
	c.Header("Content-Type", "application/zip")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", zipFilename))

	// 发送 ZIP 文件
	c.Writer.Write(buf.Bytes())
}

// 辅助函数：导出表数据到 CSV 文件
func exportTableToCSV(db *gorm.DB, zipWriter *zip.Writer, filename string, model interface{}) error {
	// 创建 CSV 文件
	w, err := zipWriter.Create(filename)
	if err != nil {
		return err
	}

	// 创建 CSV writer
	csvWriter := csv.NewWriter(w)
	defer csvWriter.Flush()

	// 查询数据
	if err := db.Find(model).Error; err != nil {
		return err
	}

	// 获取表头
	headers, err := getCSVHeaders(model)
	if err != nil {
		return err
	}

	// 写入表头
	if err := csvWriter.Write(headers); err != nil {
		return err
	}

	// 写入数据
	records, err := getCSVRecords(model)
	if err != nil {
		return err
	}

	for _, record := range records {
		if err := csvWriter.Write(record); err != nil {
			return err
		}
	}

	return nil
}

func ExportData(w http.ResponseWriter, r *http.Request) {
	// 获取数据库连接
	db := database.GetDB()

	// 创建临时文件路径
	timestamp := time.Now().Format("20060102150405")
	filePath := fmt.Sprintf("/tmp/export_%s.csv", timestamp)

	// 查询商品数据
	var products []models.Product
	if err := db.Find(&products).Error; err != nil {
		http.Error(w, "查询数据失败: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// 创建并写入 CSV 文件
	file, err := os.Create(filePath)
	if err != nil {
		http.Error(w, "创建文件失败: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer file.Close()

	// 写入 CSV 头
	fmt.Fprintf(file, "id,name,description,price,stock,image_url,created_at,updated_at\n")

	// 写入数据
	for _, p := range products {
		fmt.Fprintf(file, "%d,%s,%s,%.2f,%d,%s,%s,%s\n",
			p.ID, p.Name, p.Description, p.Price, p.Stock, p.ImageURL,
			p.CreatedAt.Format("2006-01-02 15:04:05"),
			p.UpdatedAt.Format("2006-01-02 15:04:05"))
	}

	// 读取文件内容
	data, err := os.ReadFile(filePath)
	if err != nil {
		http.Error(w, "读取导出文件失败: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// 设置响应头
	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=products_%s.csv", timestamp))

	// 发送文件内容
	w.Write(data)

	// 清理临时文件
	os.Remove(filePath)
}

// 辅助函数：获取 CSV 表头
func getCSVHeaders(model interface{}) ([]string, error) {
	// 根据模型类型返回表头
	switch model.(type) {
	case *[]models.User:
		return []string{"id", "name", "phone", "address", "type", "created_at", "updated_at"}, nil
	case *[]models.Product:
		return []string{"id", "name", "description", "price", "stock", "image_url", "created_at", "updated_at"}, nil
	case *[]models.Order:
		return []string{"id", "user_id", "total_price", "status", "remark", "created_at", "updated_at"}, nil
	case *[]models.OrderItem:
		return []string{"id", "order_id", "product_id", "quantity", "price"}, nil
	case *[]models.Tag:
		return []string{"id", "name", "description", "created_at", "updated_at"}, nil
	case *[]models.ProductTag:
		return []string{"id", "product_id", "tag_id", "created_at", "updated_at"}, nil
	default:
		return nil, fmt.Errorf("unsupported model type")
	}
}

// 辅助函数：获取 CSV 记录
func getCSVRecords(model interface{}) ([][]string, error) {
	var records [][]string

	switch v := model.(type) {
	case *[]models.User:
		for _, u := range *v {
			record := []string{
				strconv.FormatUint(uint64(u.ID), 10),
				u.Name,
				u.Phone,
				u.Address,
				u.Type,
				u.CreatedAt.Format(time.RFC3339),
				u.UpdatedAt.Format(time.RFC3339),
			}
			records = append(records, record)
		}
	case *[]models.Product:
		for _, p := range *v {
			record := []string{
				strconv.FormatUint(uint64(p.ID), 10),
				p.Name,
				p.Description,
				strconv.FormatFloat(p.Price, 'f', 2, 64),
				strconv.Itoa(p.Stock),
				p.ImageURL,
				p.CreatedAt.Format(time.RFC3339),
				p.UpdatedAt.Format(time.RFC3339),
			}
			records = append(records, record)
		}
	case *[]models.Order:
		for _, o := range *v {
			record := []string{
				strconv.FormatUint(uint64(o.ID), 10),
				strconv.FormatUint(uint64(o.UserID), 10),
				strconv.FormatFloat(float64(o.TotalPrice), 'f', 2, 64),
				o.Status,
				o.Remark,
				o.CreatedAt.Format(time.RFC3339),
				o.UpdatedAt.Format(time.RFC3339),
			}
			records = append(records, record)
		}
	case *[]models.OrderItem:
		for _, item := range *v {
			record := []string{
				strconv.FormatUint(uint64(item.ID), 10),
				strconv.FormatUint(uint64(item.OrderID), 10),
				strconv.FormatUint(uint64(item.ProductID), 10),
				strconv.Itoa(item.Quantity),
				strconv.FormatFloat(float64(item.Price), 'f', 2, 64),
			}
			records = append(records, record)
		}
	case *[]models.Tag:
		for _, t := range *v {
			record := []string{
				strconv.FormatUint(uint64(t.ID), 10),
				t.Name,
				t.Description,
				t.CreatedAt.Format(time.RFC3339),
				t.UpdatedAt.Format(time.RFC3339),
			}
			records = append(records, record)
		}
	case *[]models.ProductTag:
		for _, pt := range *v {
			record := []string{
				strconv.FormatUint(uint64(pt.ProductID), 10),
				strconv.FormatUint(uint64(pt.TagID), 10),
				pt.CreatedAt.Format(time.RFC3339),
				pt.UpdatedAt.Format(time.RFC3339),
			}
			records = append(records, record)
		}
	default:
		return nil, fmt.Errorf("unsupported model type")
	}

	return records, nil
}
