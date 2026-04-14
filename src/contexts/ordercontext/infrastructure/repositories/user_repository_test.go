package repositories

import (
	"database/sql"
	"fmt"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/bwmarrin/snowflake"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"orderease/models"
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

// ==================== Constructor Tests ====================

func TestNewUserRepository(t *testing.T) {
	db, _, sqlDB := setupUserTestDB(t)
	defer sqlDB.Close()

	repo := NewUserRepository(db)

	assert.NotNil(t, repo)
	assert.Equal(t, db, repo.DB)
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

// ==================== GetUserByID Tests ====================

func TestUserRepository_GetUserByID_Success(t *testing.T) {
	db, mock, sqlDB := setupUserTestDB(t)
	defer sqlDB.Close()

	now := time.Now()
	userID := snowflake.ID(123)
	rows := sqlmock.NewRows([]string{"id", "name", "role", "password", "phone", "address", "type", "nickname", "created_at", "updated_at"}).
		AddRow(userID, "testuser", "user", "hashedpass", "13800138000", "Test Address", "delivery", "Test Nick", now, now)

	mock.ExpectQuery("SELECT \\* FROM `users`").
		WithArgs(int64(123), 1).
		WillReturnRows(rows)

	repo := NewUserRepository(db)
	user, err := repo.GetUserByID(fmt.Sprintf("%d", userID))

	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, "testuser", user.Name)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_GetUserByID_NotFound(t *testing.T) {
	db, mock, sqlDB := setupUserTestDB(t)
	defer sqlDB.Close()

	mock.ExpectQuery("SELECT \\* FROM `users`").
		WithArgs(int64(999), 1).
		WillReturnError(gorm.ErrRecordNotFound)

	repo := NewUserRepository(db)
	user, err := repo.GetUserByID("999")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "用户不存在")
	assert.Nil(t, user)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_GetUserByID_DatabaseError(t *testing.T) {
	db, mock, sqlDB := setupUserTestDB(t)
	defer sqlDB.Close()

	mock.ExpectQuery("SELECT \\* FROM `users`").
		WithArgs(int64(123), 1).
		WillReturnError(fmt.Errorf("database error"))

	repo := NewUserRepository(db)
	user, err := repo.GetUserByID("123")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "查询用户失败")
	assert.Nil(t, user)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// ==================== GetByUsername Tests ====================

func TestUserRepository_GetByUsername_Success(t *testing.T) {
	db, mock, sqlDB := setupUserTestDB(t)
	defer sqlDB.Close()

	now := time.Now()
	userID := snowflake.ID(123)
	rows := sqlmock.NewRows([]string{"id", "name", "role", "password", "phone", "address", "type", "nickname", "created_at", "updated_at"}).
		AddRow(userID, "testuser", "user", "hashedpass", "13800138000", "Test Address", "delivery", "Test Nick", now, now)

	mock.ExpectQuery("SELECT \\* FROM `users`").
		WithArgs("testuser", 1).
		WillReturnRows(rows)

	repo := NewUserRepository(db)
	user, err := repo.GetByUsername("testuser")

	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, "testuser", user.Name)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_GetByUsername_NotFound(t *testing.T) {
	db, mock, sqlDB := setupUserTestDB(t)
	defer sqlDB.Close()

	mock.ExpectQuery("SELECT \\* FROM `users`").
		WithArgs("nonexistent", 1).
		WillReturnError(gorm.ErrRecordNotFound)

	repo := NewUserRepository(db)
	user, err := repo.GetByUsername("nonexistent")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "用户不存在")
	assert.Nil(t, user)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_GetByUsername_DatabaseError(t *testing.T) {
	db, mock, sqlDB := setupUserTestDB(t)
	defer sqlDB.Close()

	mock.ExpectQuery("SELECT \\* FROM `users`").
		WithArgs("testuser", 1).
		WillReturnError(fmt.Errorf("database error"))

	repo := NewUserRepository(db)
	user, err := repo.GetByUsername("testuser")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "查询用户失败")
	assert.Nil(t, user)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// ==================== GetUsers Tests ====================

func TestUserRepository_GetUsers_Success(t *testing.T) {
	db, mock, sqlDB := setupUserTestDB(t)
	defer sqlDB.Close()

	now := time.Now()
	countRows := sqlmock.NewRows([]string{"count"}).AddRow(2)
	mock.ExpectQuery("SELECT count\\(\\*\\) FROM `users`").
		WillReturnRows(countRows)

	userRows := sqlmock.NewRows([]string{"id", "name", "role", "phone", "created_at", "updated_at"}).
		AddRow(snowflake.ID(1), "user1", "user", "13800138001", now, now).
		AddRow(snowflake.ID(2), "user2", "user", "13800138002", now, now)
	mock.ExpectQuery("SELECT \\* FROM `users`").
		WithArgs(10).
		WillReturnRows(userRows)

	repo := NewUserRepository(db)
	users, total, err := repo.GetUsers(1, 10, "")

	assert.NoError(t, err)
	assert.Equal(t, int64(2), total)
	assert.Len(t, users, 2)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_GetUsers_WithSearch(t *testing.T) {
	db, mock, sqlDB := setupUserTestDB(t)
	defer sqlDB.Close()

	now := time.Now()
	countRows := sqlmock.NewRows([]string{"count"}).AddRow(1)
	mock.ExpectQuery("SELECT count\\(\\*\\) FROM `users`").
		WillReturnRows(countRows)

	userRows := sqlmock.NewRows([]string{"id", "name", "role", "phone", "created_at", "updated_at"}).
		AddRow(snowflake.ID(1), "testuser", "user", "13800138001", now, now)
	mock.ExpectQuery("SELECT \\* FROM `users`").
		WithArgs("%test%", 10).
		WillReturnRows(userRows)

	repo := NewUserRepository(db)
	users, total, err := repo.GetUsers(1, 10, "test")

	assert.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Len(t, users, 1)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_GetUsers_DatabaseError(t *testing.T) {
	db, mock, sqlDB := setupUserTestDB(t)
	defer sqlDB.Close()

	mock.ExpectQuery("SELECT count\\(\\*\\) FROM `users`").
		WillReturnError(fmt.Errorf("database error"))

	repo := NewUserRepository(db)
	users, total, err := repo.GetUsers(1, 10, "")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "获取用户总数失败")
	assert.Nil(t, users)
	assert.Equal(t, int64(0), total)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// ==================== GetUserSimpleList Tests ====================

func TestUserRepository_GetUserSimpleList_Success(t *testing.T) {
	db, mock, sqlDB := setupUserTestDB(t)
	defer sqlDB.Close()

	countRows := sqlmock.NewRows([]string{"count"}).AddRow(2)
	mock.ExpectQuery("SELECT count\\(\\*\\) FROM `users`").
		WillReturnRows(countRows)

	userRows := sqlmock.NewRows([]string{"id", "name"}).
		AddRow(snowflake.ID(1), "user1").
		AddRow(snowflake.ID(2), "user2")
	mock.ExpectQuery("SELECT `id`,`name` FROM `users`").
		WithArgs(10).
		WillReturnRows(userRows)

	repo := NewUserRepository(db)
	result, total, err := repo.GetUserSimpleList(1, 10, "")

	assert.NoError(t, err)
	assert.Equal(t, int64(2), total)
	assert.Len(t, result, 2)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_GetUserSimpleList_DatabaseError(t *testing.T) {
	db, mock, sqlDB := setupUserTestDB(t)
	defer sqlDB.Close()

	mock.ExpectQuery("SELECT count\\(\\*\\) FROM `users`").
		WillReturnError(fmt.Errorf("database error"))

	repo := NewUserRepository(db)
	result, total, err := repo.GetUserSimpleList(1, 10, "")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "查询用户总数失败")
	assert.Nil(t, result)
	assert.Equal(t, int64(0), total)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// ==================== Create Tests ====================

func TestUserRepository_Create_Success(t *testing.T) {
	db, mock, sqlDB := setupUserTestDB(t)
	defer sqlDB.Close()

	user := &models.User{
		Name:     "newuser",
		Role:     "user",
		Password: "hashedpass",
		Phone:    "13800138000",
	}

	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO `users`").
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	repo := NewUserRepository(db)
	err := repo.Create(user)

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_Create_DatabaseError(t *testing.T) {
	db, mock, sqlDB := setupUserTestDB(t)
	defer sqlDB.Close()

	user := &models.User{
		Name: "newuser",
	}

	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO `users`").
		WillReturnError(fmt.Errorf("database error"))
	mock.ExpectRollback()

	repo := NewUserRepository(db)
	err := repo.Create(user)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "创建用户失败")
	assert.NoError(t, mock.ExpectationsWereMet())
}

// ==================== Update Tests ====================

func TestUserRepository_Update_Success(t *testing.T) {
	db, mock, sqlDB := setupUserTestDB(t)
	defer sqlDB.Close()

	user := &models.User{
		ID:       123,
		Name:     "updateduser",
		Role:     "user",
		Password: "newhashedpass",
		Phone:    "13900139000",
	}

	mock.ExpectBegin()
	mock.ExpectExec("UPDATE `users`").
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	repo := NewUserRepository(db)
	err := repo.Update(user)

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_Update_DatabaseError(t *testing.T) {
	db, mock, sqlDB := setupUserTestDB(t)
	defer sqlDB.Close()

	user := &models.User{
		ID:   123,
		Name: "updateduser",
	}

	mock.ExpectBegin()
	mock.ExpectExec("UPDATE `users`").
		WillReturnError(fmt.Errorf("database error"))
	mock.ExpectRollback()

	repo := NewUserRepository(db)
	err := repo.Update(user)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "更新用户失败")
	assert.NoError(t, mock.ExpectationsWereMet())
}

// ==================== Delete Tests ====================

func TestUserRepository_Delete_Success(t *testing.T) {
	db, mock, sqlDB := setupUserTestDB(t)
	defer sqlDB.Close()

	user := &models.User{
		ID: 123,
	}

	mock.ExpectBegin()
	mock.ExpectExec("DELETE FROM `users`").
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	repo := NewUserRepository(db)
	err := repo.Delete(user)

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_Delete_DatabaseError(t *testing.T) {
	db, mock, sqlDB := setupUserTestDB(t)
	defer sqlDB.Close()

	user := &models.User{
		ID: 123,
	}

	mock.ExpectBegin()
	mock.ExpectExec("DELETE FROM `users`").
		WillReturnError(fmt.Errorf("database error"))
	mock.ExpectRollback()

	repo := NewUserRepository(db)
	err := repo.Delete(user)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "删除用户失败")
	assert.NoError(t, mock.ExpectationsWereMet())
}
