package repositories

import (
	"database/sql"
	"fmt"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// setupUserTestDB 创建测试用的 mock 数据库
func setupUserTestDB(t *testing.T) (*gorm.DB, sqlmock.Sqlmock, *sql.DB) {
	t.Helper()

	sqlDB, mock, err := sqlmock.New()
	assert.NoError(t, err)

	dialector := mysql.New(mysql.Config{
		Conn:                      sqlDB,
		SkipInitializeWithVersion: true,
	})

	db, err := gorm.Open(dialector, &gorm.Config{})
	assert.NoError(t, err)

	return db, mock, sqlDB
}

// ==================== CheckUsernameExists Tests ====================

func TestUserRepository_CheckUsernameExists_UsernameExists(t *testing.T) {
	db, mock, sqlDB := setupUserTestDB(t)
	defer sqlDB.Close()

	rows := sqlmock.NewRows([]string{"count"}).AddRow(1)
	mock.ExpectQuery("SELECT count\\(\\*\\) FROM `users`").
		WithArgs("testuser").
		WillReturnRows(rows)

	repo := NewUserRepository(db)

	exists, err := repo.CheckUsernameExists("testuser")

	assert.NoError(t, err)
	assert.True(t, exists)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_CheckUsernameExists_UsernameDoesNotExist(t *testing.T) {
	db, mock, sqlDB := setupUserTestDB(t)
	defer sqlDB.Close()

	rows := sqlmock.NewRows([]string{"count"}).AddRow(0)
	mock.ExpectQuery("SELECT count\\(\\*\\) FROM `users`").
		WithArgs("nonexistent").
		WillReturnRows(rows)

	repo := NewUserRepository(db)

	exists, err := repo.CheckUsernameExists("nonexistent")

	assert.NoError(t, err)
	assert.False(t, exists)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_CheckUsernameExists_DatabaseError(t *testing.T) {
	db, mock, sqlDB := setupUserTestDB(t)
	defer sqlDB.Close()

	mock.ExpectQuery("SELECT count\\(\\*\\) FROM `users`").
		WithArgs("testuser").
		WillReturnError(fmt.Errorf("database error"))

	repo := NewUserRepository(db)

	exists, err := repo.CheckUsernameExists("testuser")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "检查用户名失败")
	assert.False(t, exists)
}

// ==================== CheckPhoneExists Tests ====================

func TestUserRepository_CheckPhoneExists_PhoneExists(t *testing.T) {
	db, mock, sqlDB := setupUserTestDB(t)
	defer sqlDB.Close()

	rows := sqlmock.NewRows([]string{"count"}).AddRow(1)
	mock.ExpectQuery("SELECT count\\(\\*\\) FROM `users`").
		WithArgs("13800138000").
		WillReturnRows(rows)

	repo := NewUserRepository(db)

	exists, err := repo.CheckPhoneExists("13800138000")

	assert.NoError(t, err)
	assert.True(t, exists)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_CheckPhoneExists_PhoneDoesNotExist(t *testing.T) {
	db, mock, sqlDB := setupUserTestDB(t)
	defer sqlDB.Close()

	rows := sqlmock.NewRows([]string{"count"}).AddRow(0)
	mock.ExpectQuery("SELECT count\\(\\*\\) FROM `users`").
		WithArgs("13900139000").
		WillReturnRows(rows)

	repo := NewUserRepository(db)

	exists, err := repo.CheckPhoneExists("13900139000")

	assert.NoError(t, err)
	assert.False(t, exists)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_CheckPhoneExists_DatabaseError(t *testing.T) {
	db, mock, sqlDB := setupUserTestDB(t)
	defer sqlDB.Close()

	mock.ExpectQuery("SELECT count\\(\\*\\) FROM `users`").
		WithArgs("13800138000").
		WillReturnError(fmt.Errorf("database error"))

	repo := NewUserRepository(db)

	exists, err := repo.CheckPhoneExists("13800138000")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "检查手机号失败")
	assert.False(t, exists)
}
