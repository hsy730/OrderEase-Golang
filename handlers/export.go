package handlers

import (
	"archive/zip"
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"net/http"
	"orderease/models"
	"strconv"
	"strings"
	"time"

	"reflect"

	"github.com/gin-gonic/gin"
	"gorm.io/datatypes"
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

	// 导出管理员数据
	if err := exportTableToCSV(h.DB, zipWriter, "admins.csv", &[]models.Admin{}); err != nil {
		h.logger.Printf("导出管理员数据失败: %v", err)
		errorResponse(c, http.StatusInternalServerError, "导出失败")
		return
	}

	// 导出商品选项数据
	if err := exportTableToCSV(h.DB, zipWriter, "product_options.csv", &[]models.ProductOption{}); err != nil {
		h.logger.Printf("导出商品选项数据失败: %v", err)
		errorResponse(c, http.StatusInternalServerError, "导出失败")
		return
	}

	// 导出店铺数据
	if err := exportTableToCSV(h.DB, zipWriter, "shops.csv", &[]models.Shop{}); err != nil {
		h.logger.Printf("导出店铺数据失败: %v", err)
		errorResponse(c, http.StatusInternalServerError, "导出失败")
		return
	}

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

// 辅助函数：获取 CSV 表头
func getCSVHeaders(model interface{}) ([]string, error) {
	// 由于报错显示 reflect 未定义，需要添加 reflect 包的导入
	// 此处仅展示选择部分的修改，实际使用时需要在文件开头添加 import "reflect"

	// 原代码保持不变，需要在文件开头补充 import 语句
	t := reflect.TypeOf(model).Elem().Elem()
	var headers []string
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		gormTag := field.Tag.Get("gorm")
		if column := parseColumnName(gormTag); column != "" {
			headers = append(headers, column)
		}
	}
	return headers, nil
}

func parseColumnName(tag string) string {
	for _, part := range strings.Split(tag, ";") {
		if strings.HasPrefix(part, "column:") {
			return strings.TrimPrefix(part, "column:")
		}
	}
	return ""
}

// 辅助函数：获取 CSV 记录
func getCSVRecords(model interface{}) ([][]string, error) {
	var records [][]string

	headers, _ := getCSVHeaders(model)
	converters := getCSVColumnConverter(model, headers)

	v := reflect.ValueOf(model).Elem()

	for i := 0; i < v.Len(); i++ {
		var record []string
		elem := v.Index(i)

		for _, header := range headers {
			fieldValue, err := converters[header](elem)
			if err != nil {
				return nil, err
			}
			record = append(record, fieldValue)
		}
		records = append(records, record)
	}
	return records, nil
}

type fieldConverter func(v reflect.Value) (string, error)

func getCSVColumnConverter(model interface{}, headers []string) map[string]fieldConverter {
	v := reflect.ValueOf(model).Elem()

	converters := make(map[string]fieldConverter, len(headers))

	if v.Len() != 0 {
		elem := v.Index(0)

		for _, header := range headers {
			converters[header] = getFieldValueByColumn(elem, header)
		}
	}

	return converters
}

func getFieldValueByColumn(v reflect.Value, column string) fieldConverter {
	t := v.Type()
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if parseColumnName(field.Tag.Get("gorm")) == column {
			return func(inputParam reflect.Value) (string, error) {
				converter := convertValueToString(v.Field(i))
				return converter(inputParam.FieldByName(field.Name))
			}
		}
	}
	return func(v reflect.Value) (string, error) {
		return "", fmt.Errorf("column %s not found", column)
	}
}

func convertValueToString(fieldValue reflect.Value) fieldConverter {
	switch fieldValue.Kind() {
	case reflect.String:
		return func(v reflect.Value) (string, error) {
			return v.String(), nil
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return func(v reflect.Value) (string, error) {
			return strconv.FormatInt(v.Int(), 10), nil
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return func(v reflect.Value) (string, error) {
			return strconv.FormatUint(v.Uint(), 10), nil
		}
	case reflect.Float32, reflect.Float64:
		return func(v reflect.Value) (string, error) {
			return strconv.FormatFloat(v.Float(), 'f', -1, 64), nil
		}
	case reflect.Slice, reflect.Map:
		return func(v reflect.Value) (string, error) {
			jsonData, err := json.Marshal(v.Interface())
			if err != nil {
				return "", err
			}
			return string(jsonData), nil
		}
	case reflect.Struct:
		if t := fieldValue.Type(); t == reflect.TypeOf(time.Time{}) {
			return func(v reflect.Value) (string, error) {
				return v.Interface().(time.Time).Format(time.RFC3339), nil
			}
		}
		if t := fieldValue.Type(); t == reflect.TypeOf(datatypes.JSON{}) {
			return func(v reflect.Value) (string, error) {
				jsonData, err := json.Marshal(v.Interface())
				if err != nil {
					return "", err
				}
				return string(jsonData), nil
			}
		}
		if stringer, ok := fieldValue.Interface().(fmt.Stringer); ok {
			return func(v reflect.Value) (string, error) {
				return stringer.String(), nil
			}
		}
	case reflect.Bool:
		return func(v reflect.Value) (string, error) {
			return strconv.FormatBool(v.Bool()), nil
		}
	}
	return func(v reflect.Value) (string, error) {
		return "", fmt.Errorf("unsupported type: %s, kind: %v", fieldValue.Type(), fieldValue.Kind())
	}
}
