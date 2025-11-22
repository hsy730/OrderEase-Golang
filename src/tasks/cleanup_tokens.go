package tasks

import (
	"orderease/models"
	"orderease/utils/log2"
	"time"

	"gorm.io/gorm"
)

// TokenCleanupTask token清理任务
type TokenCleanupTask struct {
	db     *gorm.DB
	logger *log2.Logger
}

// NewTokenCleanupTask 创建token清理任务
func NewTokenCleanupTask(db *gorm.DB) *TokenCleanupTask {
	return &TokenCleanupTask{
		db:     db,
		logger: log2.GetLogger(),
	}
}

// StartTokenCleanup 启动token清理任务
func (t *TokenCleanupTask) StartTokenCleanup() {
	// 每天凌晨2点执行清理
	ticker := time.NewTicker(24 * time.Hour)
	go func() {
		for {
			now := time.Now()
			next := time.Date(now.Year(), now.Month(), now.Day()+1, 2, 0, 0, 0, now.Location())
			time.Sleep(next.Sub(now))

			t.logger.Infof("开始清理过期token黑名单")
			if err := t.cleanupExpiredTokens(); err != nil {
				t.logger.Errorf("清理过期token失败: %v", err)
			}
			t.logger.Infof("清理过期token完成")

			<-ticker.C
		}
	}()
}

// cleanupExpiredTokens 清理过期的token记录
func (t *TokenCleanupTask) cleanupExpiredTokens() error {
	return t.db.Where("expired_at < ?", time.Now()).Delete(&models.BlacklistedToken{}).Error
}
