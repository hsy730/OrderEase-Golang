package database

import (
	"orderease/config"
	"sync"

	"gorm.io/gorm"
)

var (
	db   *gorm.DB
	once sync.Once
)

// GetDB 返回数据库连接实例（单例模式）
func GetDB() *gorm.DB {
	once.Do(func() {
		var err error
		db, err = config.InitDB()
		if err != nil {
			panic("failed to connect database: " + err.Error())
		}
	})
	return db
}
