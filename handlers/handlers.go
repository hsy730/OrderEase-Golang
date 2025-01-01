package handlers

import (
	"OrderEase/models"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"OrderEase/utils"

	"gorm.io/gorm"
)

type Handler struct {
	DB *gorm.DB
}

// 创建商品
func (h *Handler) CreateProduct(w http.ResponseWriter, r *http.Request) {
	var product models.Product
	if err := json.NewDecoder(r.Body).Decode(&product); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// 清理输入数据
	utils.SanitizeProduct(&product)

	if err := h.DB.Create(&product).Error; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(product)
}

// 获取商品列表
func (h *Handler) GetProducts(w http.ResponseWriter, r *http.Request) {
	var products []models.Product

	// 支持分页
	pageStr := r.URL.Query().Get("page")
	pageSizeStr := r.URL.Query().Get("pageSize")

	page := 1
	pageSize := 10

	if pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	if pageSizeStr != "" {
		if ps, err := strconv.Atoi(pageSizeStr); err == nil && ps > 0 {
			pageSize = ps
		}
	}

	offset := (page - 1) * pageSize

	if err := h.DB.Offset(offset).Limit(pageSize).Find(&products).Error; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(products)
}

// 获取单个商品详情
func (h *Handler) GetProduct(w http.ResponseWriter, r *http.Request) {
	productID := r.URL.Query().Get("id")
	var product models.Product

	if err := h.DB.First(&product, productID).Error; err != nil {
		http.Error(w, "商品未找到", http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(product)
}

// 更新商品信息
func (h *Handler) UpdateProduct(w http.ResponseWriter, r *http.Request) {
	productID := r.URL.Query().Get("id")
	var product models.Product

	// 先检查商品是否存在
	if err := h.DB.First(&product, productID).Error; err != nil {
		http.Error(w, "商品未找到", http.StatusNotFound)
		return
	}

	// 解析更新数据
	var updateData models.Product
	if err := json.NewDecoder(r.Body).Decode(&updateData); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// 清理更新数据
	utils.SanitizeProduct(&updateData)

	// 更新商品信息
	if err := h.DB.Model(&product).Updates(updateData).Error; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(product)
}

// 删除商品
func (h *Handler) DeleteProduct(w http.ResponseWriter, r *http.Request) {
	productID := r.URL.Query().Get("id")

	// 先获取商品信息
	var product models.Product
	if err := h.DB.First(&product, productID).Error; err != nil {
		http.Error(w, "商品不存在", http.StatusNotFound)
		return
	}

	// 如果有图片，删除图片文件
	if product.ImageURL != "" {
		imagePath := strings.TrimPrefix(product.ImageURL, "/")
		if err := os.Remove(imagePath); err != nil && !os.IsNotExist(err) {
			log.Printf("删除商品图片失败: %v", err)
		}
	}

	// 删除商品记录
	if err := h.DB.Delete(&product).Error; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "商品删除成功"})
}

// 创建订单
func (h *Handler) CreateOrder(w http.ResponseWriter, r *http.Request) {
	var order models.Order
	if err := json.NewDecoder(r.Body).Decode(&order); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// 清理订单数据
	utils.SanitizeOrder(&order)

	// 开启事务
	tx := h.DB.Begin()

	if err := tx.Create(&order).Error; err != nil {
		tx.Rollback()
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// 更新商品库存
	for _, item := range order.Items {
		var product models.Product
		if err := tx.First(&product, item.ProductID).Error; err != nil {
			tx.Rollback()
			http.Error(w, "商品不存在", http.StatusBadRequest)
			return
		}

		if product.Stock < item.Quantity {
			tx.Rollback()
			http.Error(w, "库存不足", http.StatusBadRequest)
			return
		}

		product.Stock -= item.Quantity
		if err := tx.Save(&product).Error; err != nil {
			tx.Rollback()
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	tx.Commit()
	json.NewEncoder(w).Encode(order)
}

// 更新订单
func (h *Handler) UpdateOrder(w http.ResponseWriter, r *http.Request) {
	orderID := r.URL.Query().Get("id")
	var order models.Order

	if err := h.DB.First(&order, orderID).Error; err != nil {
		http.Error(w, "订单未找到", http.StatusNotFound)
		return
	}

	var updateData models.Order
	if err := json.NewDecoder(r.Body).Decode(&updateData); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.DB.Model(&order).Updates(updateData).Error; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(order)
}

// 获取订单列表
func (h *Handler) GetOrders(w http.ResponseWriter, r *http.Request) {
	var orders []models.Order

	// 支持分页
	pageStr := r.URL.Query().Get("page")
	pageSizeStr := r.URL.Query().Get("pageSize")

	page := 1
	pageSize := 10

	if pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	if pageSizeStr != "" {
		if ps, err := strconv.Atoi(pageSizeStr); err == nil && ps > 0 {
			pageSize = ps
		}
	}

	offset := (page - 1) * pageSize

	if err := h.DB.Offset(offset).Limit(pageSize).
		Preload("Items").
		Preload("Items.Product").
		Find(&orders).Error; err != nil {
		utils.Logger.Printf("获取订单列表失败: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(orders)
}

// 获取单个订单详情
func (h *Handler) GetOrder(w http.ResponseWriter, r *http.Request) {
	orderID := r.URL.Query().Get("id")
	var order models.Order

	if err := h.DB.Preload("Items").
		Preload("Items.Product").
		First(&order, orderID).Error; err != nil {
		utils.Logger.Printf("获取订单详情失败, ID: %s, 错误: %v", orderID, err)
		http.Error(w, "订单未找到", http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(order)
}

// 删除订单
func (h *Handler) DeleteOrder(w http.ResponseWriter, r *http.Request) {
	orderID := r.URL.Query().Get("id")

	if err := h.DB.Delete(&models.Order{}, orderID).Error; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "订单删除成功"})
}

// 上传商品图片
func (h *Handler) UploadProductImage(w http.ResponseWriter, r *http.Request) {
	productID := r.URL.Query().Get("id")
	if productID == "" {
		utils.Logger.Printf("上传图片失败: 缺少商品ID")
		http.Error(w, "缺少商品ID", http.StatusBadRequest)
		return
	}

	var product models.Product
	if err := h.DB.First(&product, productID).Error; err != nil {
		utils.Logger.Printf("上传图片失败: 商品不存在, ID: %s, 错误: %v", productID, err)
		http.Error(w, "商品不存在", http.StatusNotFound)
		return
	}

	// 解析多部分表单，32 << 20 指定最大文件大小为32MB
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// 获取上传的文件
	file, handler, err := r.FormFile("image")
	if err != nil {
		http.Error(w, "获取上传文件失败", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// 检查文件类型
	if !isValidImageType(handler.Header.Get("Content-Type")) {
		http.Error(w, "不支持的文件类型", http.StatusBadRequest)
		return
	}

	// 创建上传目录
	uploadDir := "./uploads/products"
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		utils.Logger.Printf("创建上传目录失败: %v", err)
		http.Error(w, "创建上传目录失败", http.StatusInternalServerError)
		return
	}

	// 如果已有图片，先删除旧图片
	if product.ImageURL != "" {
		oldImagePath := strings.TrimPrefix(product.ImageURL, "/")
		if err := os.Remove(oldImagePath); err != nil && !os.IsNotExist(err) {
			log.Printf("删除旧图片失败: %v", err)
		}
	}

	// 生成文件名和路径
	filename := fmt.Sprintf("product_%d_%d%s",
		product.ID,
		time.Now().Unix(),
		filepath.Ext(handler.Filename))

	// 数据库中存储的URL（相对路径）
	imageURL := fmt.Sprintf("/uploads/products/%s", filename)

	// 实际文件系统路径
	filePath := fmt.Sprintf("./uploads/products/%s", filename)

	// 创建目标文件
	dst, err := os.Create(filePath)
	if err != nil {
		utils.Logger.Printf("创建文件失败: %v", err)
		http.Error(w, "创建文件失败", http.StatusInternalServerError)
		return
	}
	defer dst.Close()

	// 复制文件内容
	if _, err = io.Copy(dst, file); err != nil {
		utils.Logger.Printf("保存文件失败: %v", err)
		http.Error(w, "保存文件失败", http.StatusInternalServerError)
		return
	}

	// 生成图片URL并验证
	imageURL = fmt.Sprintf("/uploads/products/%s", filename)
	if err := utils.ValidateImageURL(imageURL); err != nil {
		utils.Logger.Printf("图片URL验证失败: %v", err)
		http.Error(w, "无效的图片格式", http.StatusBadRequest)
		return
	}

	// 更新商品的图片URL（存储相对路径）
	if err := h.DB.Model(&product).Update("image_url", imageURL).Error; err != nil {
		utils.Logger.Printf("更新商品图片URL失败: %v", err)
		http.Error(w, "更新商品图片失败", http.StatusInternalServerError)
		return
	}

	// 根据是新增还是更新返回不同的消息
	message := "图片更新成功"
	if product.ImageURL == "" {
		message = "图片上传成功"
	}

	operationType := "update"
	if message == "图片上传成功" {
		operationType = "create"
	}

	// 响应中返回相对路径
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": message,
		"url":     imageURL, // 返回相对路径
		"type":    operationType,
	})
}

// 检查文件类型是否为图片
func isValidImageType(contentType string) bool {
	validTypes := map[string]bool{
		"image/jpeg": true,
		"image/png":  true,
		"image/gif":  true,
	}
	return validTypes[contentType]
}

// 获取商品图片
func (h *Handler) GetProductImage(w http.ResponseWriter, r *http.Request) {
	imagePath := r.URL.Query().Get("path")
	if imagePath == "" {
		http.Error(w, "缺少图片路径", http.StatusBadRequest)
		return
	}

	if !strings.HasPrefix(imagePath, "/") {
		imagePath = "/" + imagePath
	}

	// 验证图片路径
	if err := utils.ValidateImageURL(imagePath); err != nil {
		utils.Logger.Printf("无效的图片路径请求: %v", err)
		http.Error(w, "无效的图片路径", http.StatusBadRequest)
		return
	}

	// 确保路径以 /uploads/ 开头
	if !strings.HasPrefix(imagePath, "/uploads/") {
		imagePath = "/uploads/" + imagePath
	}

	// 移除开头的斜杠，因为我们要从当前目录访问
	imagePath = "." + imagePath

	// 检查文件是否存在
	if _, err := os.Stat(imagePath); os.IsNotExist(err) {
		http.Error(w, "图片不存在", http.StatusNotFound)
		return
	}

	// 设置正确的 Content-Type
	contentType := "image/jpeg"
	if strings.HasSuffix(imagePath, ".png") {
		contentType = "image/png"
	} else if strings.HasSuffix(imagePath, ".gif") {
		contentType = "image/gif"
	}
	w.Header().Set("Content-Type", contentType)

	// 提供文件
	http.ServeFile(w, r, imagePath)
}
