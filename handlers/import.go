package handlers

import (
	"encoding/csv"
	"fmt"
	"io"
	"net/http"
	"orderease/database"
	"orderease/models"
	"strconv"
	"strings"
)

func ImportData(w http.ResponseWriter, r *http.Request) {
	// 解析上传的文件
	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "获取上传文件失败: "+err.Error(), http.StatusBadRequest)
		return
	}
	defer file.Close()

	// 检查文件类型
	if !strings.HasSuffix(header.Filename, ".csv") {
		http.Error(w, "只支持 CSV 文件格式", http.StatusBadRequest)
		return
	}

	// 获取数据库连接
	db := database.GetDB()

	// 创建 CSV reader
	reader := csv.NewReader(file)

	// 跳过表头
	if _, err := reader.Read(); err != nil {
		http.Error(w, "读取CSV表头失败: "+err.Error(), http.StatusBadRequest)
		return
	}

	// 开启事务
	tx := db.Begin()

	var count int64
	// 逐行读取并导入数据
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			tx.Rollback()
			http.Error(w, "读取CSV数据失败: "+err.Error(), http.StatusBadRequest)
			return
		}

		// 解析数据
		price, _ := strconv.ParseFloat(record[3], 64)
		stock, _ := strconv.Atoi(record[4])

		product := models.Product{
			Name:        record[1],
			Description: record[2],
			Price:       price,
			Stock:       stock,
			ImageURL:    record[5],
		}

		// 保存到数据库
		if err := tx.Create(&product).Error; err != nil {
			tx.Rollback()
			http.Error(w, "导入数据失败: "+err.Error(), http.StatusInternalServerError)
			return
		}

		count++
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		http.Error(w, "提交事务失败: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// 返回成功信息
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"message": "数据导入成功", "count": %d}`, count)
}
