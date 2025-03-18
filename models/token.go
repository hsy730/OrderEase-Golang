package models

import "time"

// BlacklistedToken token黑名单
type BlacklistedToken struct {
	// 请求黑名单数据量小，可以不使用雪花ID
	ID        uint      `gorm:"primarykey" json:"id"`
	Token     string    `gorm:"type:varchar(500);not null;uniqueIndex" json:"token"`
	ExpiredAt time.Time `gorm:"not null;index" json:"expired_at"` // token原本的过期时间
	CreatedAt time.Time `gorm:"not null" json:"created_at"`       // 加入黑名单的时间
}
