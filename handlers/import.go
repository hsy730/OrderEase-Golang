package handlers

import (
	"archive/zip"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"orderease/models"
	"orderease/utils/log2"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"time"

	"orderease/utils"

	"github.com/bwmarrin/snowflake"
	"github.com/gin-gonic/gin"
	"gorm.io/datatypes"
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
		&models.OrderItem{},
		&models.Order{},
		&models.ProductTag{},
		&models.Tag{},
		&models.Product{},
		&models.User{},
		// 新增导出模块中的表
		&models.Admin{},
		&models.ProductOption{},
		&models.Shop{},
	}

	fileOrder := []string{
		"users.csv",
		"admins.csv",
		"shops.csv",
		"products.csv",                  // 先导入products
		"product_option_categories.csv", // 再导入product_option_categories
		"product_options.csv",
		"tags.csv",
		"product_tags.csv",
		"orders.csv",
		"order_items.csv",
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

	// 动态创建模型实例
	model, converters, err := createModelAndConverters(zipFile.Name)
	if err != nil {
		return err
	}

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("读取记录失败: %v", err)
		}

		elem := reflect.New(reflect.TypeOf(model).Elem()).Interface()
		if err := parseRecordToModel(record, headers, converters, elem); err != nil {
			return err
		}

		if err := tx.Create(elem).Error; err != nil {
			return fmt.Errorf("创建记录失败: %v", err)
		}
	}
	return nil
}

func createImportConverters(model interface{}) map[string]func(string) (interface{}, error) {
	t := reflect.TypeOf(model).Elem()
	converters := make(map[string]func(string) (interface{}, error))

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		colName := parseColumnName(field.Tag.Get("gorm"))
		if colName == "" {
			continue
		}
		log2.Debugf("createImportConverters, colName: %s, fieldName: %s", colName, reflect.TypeOf(model).Elem().Field(i).Name)
		converter, err := createImportConverter(reflect.ValueOf(model).Elem().Field(i))
		if err != nil {
			panic(fmt.Sprintf("创建字段 %s 的转换器失败: %v", colName, err))
		}
		converters[colName] = converter
	}
	return converters
}

func createImportConverter(value reflect.Value) (func(string) (interface{}, error), error) {
	fieldType := value.Type()
	log2.Debugf("createImportConverter, kind: %v, type: %v", value.Kind(), value.Type())
	switch {
	case fieldType == reflect.TypeOf(time.Time{}):
		return parseTime, nil
	case fieldType == reflect.TypeOf(models.Price(0)):
		return parsePrice, nil
	case fieldType == reflect.TypeOf(snowflake.ID(0)):
		return parseSnowflakeID, nil
	case fieldType == reflect.TypeOf(datatypes.JSON{}):
		return func(s string) (interface{}, error) {
			trimmed := strings.TrimSpace(s)
			// 处理空值
			if trimmed == "" || trimmed == "{}" {
				return datatypes.JSON("{}"), nil
			}

			// 验证JSON格式
			if !json.Valid([]byte(trimmed)) {
				return nil, fmt.Errorf("无效的JSON格式: %s", trimmed)
			}

			return datatypes.JSON(trimmed), nil
		}, nil
	case fieldType.Kind() == reflect.Slice:
		elemType := fieldType.Elem()

		// 对于其他slice类型，需要确保元素类型是可寻址的
		var elemConverter func(string) (interface{}, error)
		var elemValue reflect.Value

		// 创建一个该元素类型的实例
		if elemType.Kind() == reflect.Ptr {
			elemValue = reflect.New(elemType.Elem())
		} else {
			elemValue = reflect.New(elemType)
		}

		elemConverter, err := createImportConverter(elemValue.Elem())
		if err != nil {
			return nil, fmt.Errorf("无法创建元素转换器: %w", err)
		}

		return func(s string) (interface{}, error) {
			// 先尝试解析为数组
			var rawArray []json.RawMessage
			if err := json.Unmarshal([]byte(s), &rawArray); err == nil {
				slice := reflect.MakeSlice(fieldType, 0, len(rawArray))
				for i, r := range rawArray {
					elemValue, err := elemConverter(string(r))
					if err != nil {
						return nil, fmt.Errorf("%s[%d]: %w", fieldType.String(), i, err)
					}
					if !reflect.TypeOf(elemValue).AssignableTo(elemType) {
						return nil, fmt.Errorf("%s[%d]类型不匹配: 期望 %s 实际 %T",
							fieldType.String(), i, elemType, elemValue)
					}
					slice = reflect.Append(slice, reflect.ValueOf(elemValue))
				}
				return slice.Interface(), nil
			}

			// 如果不是数组，尝试解析为单个对象
			var rawObject json.RawMessage
			if err := json.Unmarshal([]byte(s), &rawObject); err != nil {
				return nil, fmt.Errorf("JSON解析失败: %w", err)
			}

			// 创建单元素切片
			slice := reflect.MakeSlice(fieldType, 1, 1)
			elemValue, err := elemConverter(string(rawObject))
			if err != nil {
				return nil, fmt.Errorf("%s[0]: %w", fieldType.String(), err)
			}
			if !reflect.TypeOf(elemValue).AssignableTo(elemType) {
				return nil, fmt.Errorf("%s[0]类型不匹配: 期望 %s 实际 %T",
					fieldType.String(), elemType, elemValue)
			}
			slice = reflect.Append(slice, reflect.ValueOf(elemValue))
			return slice.Interface(), nil
		}, nil
	case strings.HasPrefix(fieldType.String(), "datatypes.JSON"):
		return func(s string) (interface{}, error) {
			trimmed := strings.TrimSpace(s)
			// 处理空值
			if trimmed == "" || trimmed == "{}" {
				return datatypes.JSON("{}"), nil
			}

			// 验证JSON格式
			if !json.Valid([]byte(trimmed)) {
				return nil, fmt.Errorf("无效的JSON格式: %s", trimmed)
			}

			return datatypes.JSON(trimmed), nil
		}, nil

	case fieldType.Kind() == reflect.Map:
		return func(s string) (interface{}, error) {
			if fieldType != reflect.TypeOf(datatypes.JSON{}) {
				return nil, fmt.Errorf("暂不支持%s类型的map转换", fieldType)
			}
			var raw map[string]interface{}
			if err := json.Unmarshal([]byte(s), &raw); err != nil {
				return nil, fmt.Errorf("JSON解析失败: %w", err)
			}
			return datatypes.JSON(s), nil
		}, nil
	default:
		return parseBasicType(value), nil
	}
}

// 完善所有模型处理器
func createModelAndConverters(filename string) (interface{}, map[string]func(string) (interface{}, error), error) {
	switch filepath.Base(filename) {
	case "admins.csv":
		return &models.Admin{}, createImportConverters(&models.Admin{}), nil
	case "product_option_categories.csv":
		return &models.ProductOptionCategory{}, createImportConverters(&models.ProductOptionCategory{}), nil
	case "product_options.csv":
		return &models.ProductOption{}, createImportConverters(&models.ProductOption{}), nil
	case "shops.csv":
		return &models.Shop{}, createImportConverters(&models.Shop{}), nil
	case "users.csv":
		return &models.User{}, createImportConverters(&models.User{}), nil
	case "products.csv":
		return &models.Product{}, createImportConverters(&models.Product{}), nil
	case "tags.csv":
		return &models.Tag{}, createImportConverters(&models.Tag{}), nil
	case "product_tags.csv":
		return &models.ProductTag{}, createImportConverters(&models.ProductTag{}), nil
	case "orders.csv":
		return &models.Order{}, createImportConverters(&models.Order{}), nil
	case "order_items.csv":
		return &models.OrderItem{}, createImportConverters(&models.OrderItem{}), nil
	default:
		return nil, nil, fmt.Errorf("未知模型类型")
	}
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

func parseRecordToModel(record []string, headers []string, converters map[string]func(string) (interface{}, error), model interface{}) error {
	v := reflect.ValueOf(model).Elem()

	for i, header := range headers {
		converter, ok := converters[header]
		if !ok {
			return fmt.Errorf("未找到字段转换器: %s", header)
		}

		value, err := converter(record[i])
		if err != nil {
			return fmt.Errorf("%s字段转换失败: %v", header, err)
		}

		field := v.FieldByName(getFieldNameByColumn(header, model))
		if !field.IsValid() {
			return fmt.Errorf("模型字段不存在: %s", header)
		}
		field.Set(reflect.ValueOf(value))
	}
	return nil
}

func getFieldNameByColumn(column string, model interface{}) string {
	t := reflect.TypeOf(model).Elem()
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if parseColumnName(field.Tag.Get("gorm")) == column {
			return field.Name
		}
	}
	return ""
}

func parseTime(s string) (interface{}, error) {
	return time.Parse(time.RFC3339, s)
}

func parsePrice(s string) (interface{}, error) {
	f, err := strconv.ParseFloat(s, 64)
	return models.Price(f), err
}

func parseSnowflakeID(s string) (interface{}, error) {
	return utils.StringToSnowflakeID(s)
}

func parseBasicType(value reflect.Value) func(string) (interface{}, error) {
	return func(s string) (interface{}, error) {
		log2.Debugf("parseBasicType, kind: %v, s: %s", value.Kind(), s)
		switch value.Kind() {
		case reflect.String:
			return s, nil
		case reflect.Int, reflect.Int64:
			return strconv.Atoi(s)
		case reflect.Uint64:
			return strconv.ParseUint(s, 10, 64)
		case reflect.Uint8:
			return strconv.ParseUint(s, 10, 8)
		case reflect.Uint16:
			return strconv.ParseUint(s, 10, 16)
		case reflect.Uint32:
			return strconv.ParseUint(s, 10, 32)
		case reflect.Uint:
			return strconv.ParseUint(s, 10, 0)
		case reflect.Float64:
			return strconv.ParseFloat(s, 64)
		case reflect.Bool:
			return strconv.ParseBool(s)
		default:
			return nil, fmt.Errorf("不支持的字段类型: %s, %s", value.Kind(), value.Type())
		}
	}
}
