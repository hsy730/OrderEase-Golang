package tasks

import (
	"orderease/models"
	"orderease/utils/log2"
	"os"
	"strings"
	"time"

	"gorm.io/gorm"
)

type CleanupTask struct {
	db     *gorm.DB
	logger *log2.Logger
}

func NewCleanupTask(db *gorm.DB) *CleanupTask {
	return &CleanupTask{
		db:     db,
		logger: log2.GetLogger(),
	}
}

// StartCleanupTask 启动定时清理任务
func (t *CleanupTask) StartCleanupTask() {
	// 每天凌晨3点执行清理
	ticker := time.NewTicker(24 * time.Hour)
	go func() {
		for {
			now := time.Now()
			next := time.Date(now.Year(), now.Month(), now.Day()+1, 3, 0, 0, 0, now.Location())
			time.Sleep(next.Sub(now))

			t.logger.Infof("开始执行清理任务")
			if err := t.Cleanup(); err != nil {
				t.logger.Errorf("清理任务执行失败: %v", err)
			}
			t.logger.Infof("清理任务执行完成")

			<-ticker.C
		}
	}()
}

// Cleanup 执行清理任务
func (t *CleanupTask) Cleanup() error {
	// 开启事务
	return t.db.Transaction(func(tx *gorm.DB) error {
		// 1. 清理订单
		if err := t.cleanupOrders(tx); err != nil {
			return err
		}

		// 2. 清理商品
		if err := t.cleanupProducts(tx); err != nil {
			return err
		}

		return nil
	})
}

// cleanupOrders 清理3个月前的已完成订单
func (t *CleanupTask) cleanupOrders(tx *gorm.DB) error {
	threeMonthsAgo := time.Now().AddDate(0, -3, 0)

	// 查找需要删除的订单
	var orders []models.Order
	if err := tx.Where("status = ? AND updated_at < ?",
		models.OrderStatusComplete, threeMonthsAgo).
		Find(&orders).Error; err != nil {
		return err
	}

	// 记录要删除的订单数量
	t.logger.Infof("将删除 %d 个过期订单", len(orders))

	// 删除订单项
	if err := tx.Where("order_id IN (?)",
		tx.Model(&orders).Select("id")).
		Delete(&models.OrderItem{}).Error; err != nil {
		return err
	}

	// 删除订单
	if err := tx.Delete(&orders).Error; err != nil {
		return err
	}

	return nil
}

// cleanupProducts 清理不关联订单的下架商品
func (t *CleanupTask) cleanupProducts(tx *gorm.DB) error {
	// 查找下架且不关联订单的商品
	var products []models.Product
	if err := tx.Where("status = ? AND NOT EXISTS (SELECT 1 FROM order_items WHERE order_items.product_id = products.id)",
		models.ProductStatusOffline).
		Find(&products).Error; err != nil {
		return err
	}

	t.logger.Infof("将删除 %d 个下架商品", len(products))

	// 删除商品图片
	for _, product := range products {
		if product.ImageURL != "" {
			imagePath := strings.TrimPrefix(product.ImageURL, "/")
			if err := os.Remove(imagePath); err != nil && !os.IsNotExist(err) {
				t.logger.Errorf("删除商品图片失败: %v", err)
				// 继续执行，不中断流程
			}
		}
	}

	// 删除商品记录
	if err := tx.Delete(&products).Error; err != nil {
		return err
	}

	return nil
}
