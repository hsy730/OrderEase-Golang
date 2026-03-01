package repositories

import (
	"database/sql"
	"fmt"
	"orderease/models"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func setupAdminRepoTestDB(t *testing.T) (*gorm.DB, sqlmock.Sqlmock, *sql.DB) {
	t.Helper()

	sqlDB, mock, err := sqlmock.New()
	assert.NoError(t, err)

	dialector := mysql.Config{
		Conn:                      sqlDB,
		SkipInitializeWithVersion: true,
	}

	db, err := gorm.Open(mysql.New(dialector), &gorm.Config{})
	assert.NoError(t, err)

	return db, mock, sqlDB
}

func TestNewAdminRepository(t *testing.T) {
	db, _, sqlDB := setupAdminRepoTestDB(t)
	defer sqlDB.Close()

	repo := NewAdminRepository(db)

	assert.NotNil(t, repo)
	assert.Equal(t, db, repo.DB)
}

func TestAdminRepository_GetAdminByUsername_Found(t *testing.T) {
	db, mock, sqlDB := setupAdminRepoTestDB(t)
	defer sqlDB.Close()

	now := time.Now()
	rows := sqlmock.NewRows([]string{"id", "username", "password", "created_at", "updated_at"}).
		AddRow(1, "admin", "hashedpassword", now, now)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `admins` WHERE username = ? ORDER BY `admins`.`id` LIMIT ?")).
		WithArgs("admin", 1).
		WillReturnRows(rows)

	repo := NewAdminRepository(db)
	admin, err := repo.GetAdminByUsername("admin")

	assert.NoError(t, err)
	assert.NotNil(t, admin)
	assert.Equal(t, uint64(1), admin.ID)
	assert.Equal(t, "admin", admin.Username)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestAdminRepository_GetAdminByUsername_NotFound(t *testing.T) {
	db, mock, sqlDB := setupAdminRepoTestDB(t)
	defer sqlDB.Close()

	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `admins` WHERE username = ? ORDER BY `admins`.`id` LIMIT ?")).
		WithArgs("nonexistent", 1).
		WillReturnError(gorm.ErrRecordNotFound)

	repo := NewAdminRepository(db)
	admin, err := repo.GetAdminByUsername("nonexistent")

	assert.Error(t, err)
	assert.Equal(t, "管理员不存在", err.Error())
	assert.Nil(t, admin)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestAdminRepository_GetAdminByUsername_DatabaseError(t *testing.T) {
	db, mock, sqlDB := setupAdminRepoTestDB(t)
	defer sqlDB.Close()

	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `admins` WHERE username = ? ORDER BY `admins`.`id` LIMIT ?")).
		WithArgs("admin", 1).
		WillReturnError(fmt.Errorf("database error"))

	repo := NewAdminRepository(db)
	admin, err := repo.GetAdminByUsername("admin")

	assert.Error(t, err)
	assert.Equal(t, "查询管理员失败", err.Error())
	assert.Nil(t, admin)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestAdminRepository_GetFirstAdmin_Found(t *testing.T) {
	db, mock, sqlDB := setupAdminRepoTestDB(t)
	defer sqlDB.Close()

	now := time.Now()
	rows := sqlmock.NewRows([]string{"id", "username", "password", "created_at", "updated_at"}).
		AddRow(1, "admin", "hashedpassword", now, now)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `admins` ORDER BY `admins`.`id` LIMIT ?")).
		WithArgs(1).
		WillReturnRows(rows)

	repo := NewAdminRepository(db)
	admin, err := repo.GetFirstAdmin()

	assert.NoError(t, err)
	assert.NotNil(t, admin)
	assert.Equal(t, uint64(1), admin.ID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestAdminRepository_GetFirstAdmin_NotFound(t *testing.T) {
	db, mock, sqlDB := setupAdminRepoTestDB(t)
	defer sqlDB.Close()

	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `admins` ORDER BY `admins`.`id` LIMIT ?")).
		WithArgs(1).
		WillReturnError(gorm.ErrRecordNotFound)

	repo := NewAdminRepository(db)
	admin, err := repo.GetFirstAdmin()

	assert.Error(t, err)
	assert.Equal(t, "管理员账户不存在", err.Error())
	assert.Nil(t, admin)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestAdminRepository_GetFirstAdmin_DatabaseError(t *testing.T) {
	db, mock, sqlDB := setupAdminRepoTestDB(t)
	defer sqlDB.Close()

	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `admins` ORDER BY `admins`.`id` LIMIT ?")).
		WithArgs(1).
		WillReturnError(fmt.Errorf("database error"))

	repo := NewAdminRepository(db)
	admin, err := repo.GetFirstAdmin()

	assert.Error(t, err)
	assert.Equal(t, "查询管理员失败", err.Error())
	assert.Nil(t, admin)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestAdminRepository_Update_Success(t *testing.T) {
	db, mock, sqlDB := setupAdminRepoTestDB(t)
	defer sqlDB.Close()

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta("UPDATE `admins`")).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	repo := NewAdminRepository(db)
	admin := &models.Admin{ID: 1, Username: "admin", Password: "newhashedpassword"}
	err := repo.Update(admin)

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestAdminRepository_Update_DatabaseError(t *testing.T) {
	db, mock, sqlDB := setupAdminRepoTestDB(t)
	defer sqlDB.Close()

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta("UPDATE `admins`")).
		WillReturnError(fmt.Errorf("database error"))
	mock.ExpectRollback()

	repo := NewAdminRepository(db)
	admin := &models.Admin{ID: 1, Username: "admin", Password: "newhashedpassword"}
	err := repo.Update(admin)

	assert.Error(t, err)
	assert.Equal(t, "更新管理员失败", err.Error())
	assert.NoError(t, mock.ExpectationsWereMet())
}
