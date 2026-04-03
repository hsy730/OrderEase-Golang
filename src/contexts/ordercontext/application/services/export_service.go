package services

import (
	"archive/zip"
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/datatypes"
	"gorm.io/gorm"
	"orderease/models"
)

type ExportService struct {
	db *gorm.DB
}

func NewExportService(db *gorm.DB) *ExportService {
	return &ExportService{db: db}
}

// ExportAllData 导出所有数据为 ZIP 文件（含 CSV + 上传文件）
func (s *ExportService) ExportAllData(c *gin.Context) error {
	timestamp := time.Now().Format("20060102_150405")
	zipFilename := fmt.Sprintf("export_%s.zip", timestamp)

	buf := new(bytes.Buffer)
	zipWriter := zip.NewWriter(buf)

	tables := []struct {
		filename string
		model    interface{}
	}{
		{"admins.csv", &[]models.Admin{}},
		{"product_option_categories.csv", &[]models.ProductOptionCategory{}},
		{"product_options.csv", &[]models.ProductOption{}},
		{"shops.csv", &[]models.Shop{}},
		{"users.csv", &[]models.User{}},
		{"products.csv", &[]models.Product{}},
		{"tags.csv", &[]models.Tag{}},
		{"product_tags.csv", &[]models.ProductTag{}},
		{"orders.csv", &[]models.Order{}},
		{"order_items.csv", &[]models.OrderItem{}},
		{"order_item_options.csv", &[]models.OrderItemOption{}},
		{"user_thirdparty_bindings.csv", &[]models.UserThirdpartyBinding{}},
	}

	for _, table := range tables {
		if err := s.exportTableToCSV(zipWriter, table.filename, table.model); err != nil {
			return fmt.Errorf("导出 %s 失败: %w", table.filename, err)
		}
	}

	if err := s.addUploadsToZip(zipWriter); err != nil {
		return fmt.Errorf("图片打包失败: %w", err)
	}

	if err := zipWriter.Close(); err != nil {
		return fmt.Errorf("关闭 ZIP writer 失败: %w", err)
	}

	c.Header("Content-Type", "application/zip")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", zipFilename))
	c.Writer.Write(buf.Bytes())
	return nil
}

func (s *ExportService) exportTableToCSV(zipWriter *zip.Writer, filename string, model interface{}) error {
	w, err := zipWriter.Create(filename)
	if err != nil {
		return err
	}
	csvWriter := csv.NewWriter(w)
	defer csvWriter.Flush()

	if err := s.db.Find(model).Error; err != nil {
		return err
	}

	headers, err := getCSVHeaders(model)
	if err != nil {
		return err
	}
	if err := csvWriter.Write(headers); err != nil {
		return err
	}

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

func (s *ExportService) addUploadsToZip(zipWriter *zip.Writer) error {
	basePath := filepath.Join("uploads")
	return filepath.Walk(basePath, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return err
		}
		relPath, _ := filepath.Rel(basePath, path)
		zipPath := filepath.Join("uploads", relPath)
		zipHeader, _ := zip.FileInfoHeader(info)
		zipHeader.Name = zipPath
		writer, err := zipWriter.CreateHeader(zipHeader)
		if err != nil {
			return err
		}
		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()
		_, err = io.Copy(writer, file)
		return err
	})
}

func getCSVHeaders(model interface{}) ([]string, error) {
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
		return func(v reflect.Value) (string, error) { return v.String(), nil }
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return func(v reflect.Value) (string, error) { return strconv.FormatInt(v.Int(), 10), nil }
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return func(v reflect.Value) (string, error) { return strconv.FormatUint(v.Uint(), 10), nil }
	case reflect.Float32, reflect.Float64:
		return func(v reflect.Value) (string, error) { return strconv.FormatFloat(v.Float(), 'f', -1, 64), nil }
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
			return func(v reflect.Value) (string, error) { return v.Interface().(time.Time).Format(time.RFC3339), nil }
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
			return func(v reflect.Value) (string, error) { return stringer.String(), nil }
		}
	case reflect.Bool:
		return func(v reflect.Value) (string, error) { return strconv.FormatBool(v.Bool()), nil }
	}
	return func(v reflect.Value) (string, error) {
		return "", fmt.Errorf("unsupported type: %s, kind: %v", fieldValue.Type(), fieldValue.Kind())
	}
}
