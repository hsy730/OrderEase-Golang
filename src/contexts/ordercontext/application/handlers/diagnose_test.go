package handlers

import (
	"archive/zip"
	"bytes"
	"encoding/csv"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/bwmarrin/snowflake"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"orderease/models"

	services "orderease/contexts/ordercontext/application/services"
)

func TestDiagnoseOrderStatusFlowFormat(t *testing.T) {
	t.Helper()

	gin.SetMode(gin.TestMode)

	db, srv, cleanup := setupTestMySQLServer(t)
	defer cleanup()
	defer func() {
		if srv != nil {
			srv.Close()
		}
	}()

	err := autoMigrateAllTables(db)
	require.NoError(t, err)

	now := time.Now().UTC().Truncate(time.Second)

	testShop := models.Shop{
		ID:          snowflake.ID(1001),
		Name:        "测试店铺",
		Description: "这是一个用于诊断的店铺",
		CreatedAt:   now,
		UpdatedAt:   now,
		OrderStatusFlow: models.OrderStatusFlow{
			Statuses: []models.OrderStatus{
				{
					Value:   0,
					Label:   "待处理",
					Type:    "warning",
					IsFinal: false,
					Actions: []models.OrderStatusAction{
						{Name: "接单", NextStatus: 1, NextStatusLabel: "已接单"},
						{Name: "取消", NextStatus: 10, NextStatusLabel: "已取消"},
					},
				},
				{
					Value:   1,
					Label:   "已接单",
					Type:    "primary",
					IsFinal: false,
					Actions: []models.OrderStatusAction{
						{Name: "完成", NextStatus: 9, NextStatusLabel: "已完成"},
					},
				},
			},
		},
	}

	err = db.Create(&testShop).Error
	require.NoError(t, err)

	t.Log("✓ 测试数据插入成功")

	var shopsFromDB []models.Shop
	err = db.Find(&shopsFromDB).Error
	require.NoError(t, err)
	require.Len(t, shopsFromDB, 1)

	t.Logf("📊 从数据库读取的 OrderStatusFlow (Go struct):")
	t.Logf("   Type: %T", shopsFromDB[0].OrderStatusFlow)
	t.Logf("   Statuses count: %d", len(shopsFromDB[0].OrderStatusFlow.Statuses))

	if len(shopsFromDB[0].OrderStatusFlow.Statuses) > 0 {
		t.Logf("   First status label: %s", shopsFromDB[0].OrderStatusFlow.Statuses[0].Label)
	}

	jsonBytes, err := json.Marshal(shopsFromDB[0].OrderStatusFlow)
	require.NoError(t, err)

	t.Logf("\n📝 JSON Marshal 结果:")
	t.Logf("   %s", string(jsonBytes))

	valuerResult, err := shopsFromDB[0].OrderStatusFlow.Value()
	require.NoError(t, err)

	t.Logf("\n💾 driver.Valuer.Value() 结果:")
	t.Logf("   Type: %T", valuerResult)
	t.Logf("   Value: %s", string(valuerResult.([]byte)))

	exportService := services.NewExportService(db)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/export", nil)

	err = exportService.ExportAllData(c)
	require.NoError(t, err, "导出失败")

	zipContent := w.Body.Bytes()
	reader, err := zip.NewReader(bytes.NewReader(zipContent), int64(len(zipContent)))
	require.NoError(t, err)

	var shopsCSVContent string
	for _, file := range reader.File {
		if file.Name == "shops.csv" {
			rc, err := file.Open()
			require.NoError(t, err)
			content, err := io.ReadAll(rc)
			rc.Close()
			require.NoError(t, err)
			shopsCSVContent = string(content)
			break
		}
	}

	require.NotEmpty(t, shopsCSVContent, "未找到 shops.csv")

	t.Logf("\n📄 shops.csv 内容 (前2000字符):")
	if len(shopsCSVContent) > 2000 {
		t.Logf("%s...", shopsCSVContent[:2000])
	} else {
		t.Logf("%s", shopsCSVContent)
	}

	csvReader := csv.NewReader(strings.NewReader(shopsCSVContent))
	headers, err := csvReader.Read()
	require.NoError(t, err)

	t.Logf("\n🔍 CSV 表头:")
	for i, header := range headers {
		t.Logf("   [%d] %s", i, header)
	}

	record, err := csvReader.Read()
	require.NoError(t, err)

	t.Logf("\n📋 第一行数据:")
	for i, header := range headers {
		if i < len(record) {
			valuePreview := record[i]
			if len(valuePreview) > 80 {
				valuePreview = valuePreview[:80] + "..."
			}
			t.Logf("   %s = %s", header, valuePreview)
		}
	}

	for i, header := range headers {
		if i < len(record) && strings.Contains(strings.ToLower(header), "order_status") {
			t.Logf("\n⚠️  找到 order_status_flow 字段:")
			t.Logf("   列索引: %d", i)
			t.Logf("   列名: %s", header)
			t.Logf("   值长度: %d 字符", len(record[i]))
			t.Logf("   完整值:\n%s", record[i])

			var testOSF models.OrderStatusFlow
			unmarshalErr := json.Unmarshal([]byte(record[i]), &testOSF)
			if unmarshalErr != nil {
				t.Logf("   ❌ JSON 反序列化失败: %v", unmarshalErr)
				t.Logf("   🔍 问题分析:")

				pos := 0
				count := 0
				for pos < len(record[i]) && count < 5 {
					nextInvalid := -1
					for j := pos; j < len(record[i]); j++ {
						if record[i][j] >= 0x80 || (record[i][j] < 32 && record[i][j] != '\n' && record[i][j] != '\t' && record[i][j] != '\r') {
							nextInvalid = j
							break
						}
					}

					if nextInvalid == -1 {
						break
					}

					start := max(0, nextInvalid-20)
					end := min(len(record[i]), nextInvalid+20)
					t.Logf("      位置 %d 发现异常字符 (0x%02x): ...%s...", nextInvalid, record[i][nextInvalid], record[i][start:end])
					pos = nextInvalid + 1
					count++
				}

				t.Logf("\n   🧪 尝试不同的修复方案:")

				testJSON1 := strings.ReplaceAll(record[i], `"`, `\"`)
				unmarshalErr1 := json.Unmarshal([]byte(testJSON1), &testOSF)
				t.Logf("      方案1 (转义双引号): %v", unmarshalErr1)

				var testOSF2 models.OrderStatusFlow
				unmarshalErr2 := json.Unmarshal([]byte(testJSON1), &testOSF2)
				if unmarshalErr2 == nil {
					t.Logf("      ✅ 方案1 成功! Statuses数量: %d", len(testOSF2.Statuses))
				}

				testJSON3 := strings.ReplaceAll(record[i], `\"`, `"`)
				var testOSF3 models.OrderStatusFlow
				unmarshalErr3 := json.Unmarshal([]byte(testJSON3), &testOSF3)
				t.Logf("      方案3 (反转义双引号): %v", unmarshalErr3)
			} else {
				t.Logf("   ✅ JSON 反序列化成功!")
				t.Logf("   Statuses 数量: %d", len(testOSF.Statuses))
				if len(testOSF.Statuses) > 0 {
					t.Logf("   第一个状态: %s", testOSF.Statuses[0].Label)
				}
			}
		}
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
