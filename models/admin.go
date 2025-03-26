package models

import (
	"time"

	"golang.org/x/crypto/bcrypt"
)

type Admin struct {
	// 管理员只有几个，可以不使用雪花算法
	ID        uint64    `gorm:"primarykey" json:"id"`
	Username  string    `gorm:"unique" json:"username"`
	Password  string    `json:"-"` // 密码不会在JSON中返回
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// 密码加密
func (a *Admin) HashPassword() error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(a.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	a.Password = string(hashedPassword)
	return nil
}

// 验证密码
func (a *Admin) CheckPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(a.Password), []byte(password))
	return err == nil
}

type UserInfo struct {
	UserID   uint64
	Username string
	IsAdmin  bool
}
