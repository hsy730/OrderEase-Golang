package models

import "time"

// BlacklistedToken token黑名单
type BlacklistedToken struct {
	// 请求黑名单数据量小，可以不使用雪花ID
	ID        uint      `gorm:"column:id;primarykey" json:"id"`
	Token     string    `gorm:"column:token;type:varchar(500);not null;uniqueIndex" json:"token"`
	ExpiredAt time.Time `gorm:"column:expired_at;not null;index" json:"expired_at"` // token原本的过期时间
	CreatedAt time.Time `gorm:"column:created_at;not null" json:"created_at"`       // 加入黑名单的时间
}
