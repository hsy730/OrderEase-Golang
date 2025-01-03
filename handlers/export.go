package handlers

import (
	"fmt"
	"net/http"
	"orderease/database"
	"orderease/models"
	"os"
	"time"
)

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
