package repositories

import (
	"database/sql"
	"fmt"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"orderease/models"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func setupTokenTestDB(t *testing.T) (*gorm.DB, sqlmock.Sqlmock, *sql.DB) {
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

func TestNewTokenRepository(t *testing.T) {
	db, _, sqlDB := setupTokenTestDB(t)
	defer sqlDB.Close()

	repo := NewTokenRepository(db)

	assert.NotNil(t, repo)
	assert.Equal(t, db, repo.DB)
}

func TestTokenRepository_CreateBlacklistedToken_Success(t *testing.T) {
	db, mock, sqlDB := setupTokenTestDB(t)
	defer sqlDB.Close()

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO `blacklisted_tokens`")).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	repo := NewTokenRepository(db)
	token := &models.BlacklistedToken{
		Token:     "testtoken123",
		ExpiredAt: time.Now().Add(24 * time.Hour),
	}
	err := repo.CreateBlacklistedToken(token)

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestTokenRepository_CreateBlacklistedToken_DatabaseError(t *testing.T) {
	db, mock, sqlDB := setupTokenTestDB(t)
	defer sqlDB.Close()

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO `blacklisted_tokens`")).
		WillReturnError(fmt.Errorf("database error"))
	mock.ExpectRollback()

	repo := NewTokenRepository(db)
	token := &models.BlacklistedToken{
		Token:     "testtoken123",
		ExpiredAt: time.Now().Add(24 * time.Hour),
	}
	err := repo.CreateBlacklistedToken(token)

	assert.Error(t, err)
	assert.Equal(t, "添加 token 到黑名单失败", err.Error())
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestTokenRepository_IsTokenBlacklisted_True(t *testing.T) {
	db, mock, sqlDB := setupTokenTestDB(t)
	defer sqlDB.Close()

	rows := sqlmock.NewRows([]string{"count"}).AddRow(1)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT count(*) FROM `blacklisted_tokens`")).
		WithArgs("blacklistedtoken").
		WillReturnRows(rows)

	repo := NewTokenRepository(db)
	isBlacklisted, err := repo.IsTokenBlacklisted("blacklistedtoken")

	assert.NoError(t, err)
	assert.True(t, isBlacklisted)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestTokenRepository_IsTokenBlacklisted_False(t *testing.T) {
	db, mock, sqlDB := setupTokenTestDB(t)
	defer sqlDB.Close()

	rows := sqlmock.NewRows([]string{"count"}).AddRow(0)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT count(*) FROM `blacklisted_tokens`")).
		WithArgs("validtoken").
		WillReturnRows(rows)

	repo := NewTokenRepository(db)
	isBlacklisted, err := repo.IsTokenBlacklisted("validtoken")

	assert.NoError(t, err)
	assert.False(t, isBlacklisted)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestTokenRepository_IsTokenBlacklisted_DatabaseError(t *testing.T) {
	db, mock, sqlDB := setupTokenTestDB(t)
	defer sqlDB.Close()

	mock.ExpectQuery(regexp.QuoteMeta("SELECT count(*) FROM `blacklisted_tokens`")).
		WithArgs("sometoken").
		WillReturnError(fmt.Errorf("database error"))

	repo := NewTokenRepository(db)
	isBlacklisted, err := repo.IsTokenBlacklisted("sometoken")

	assert.Error(t, err)
	assert.Equal(t, "检查 token 黑名单失败", err.Error())
	assert.False(t, isBlacklisted)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestTokenRepository_CleanExpiredTokens_Success(t *testing.T) {
	db, mock, sqlDB := setupTokenTestDB(t)
	defer sqlDB.Close()

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM `blacklisted_tokens`")).
		WillReturnResult(sqlmock.NewResult(0, 5))
	mock.ExpectCommit()

	repo := NewTokenRepository(db)
	err := repo.CleanExpiredTokens()

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestTokenRepository_CleanExpiredTokens_DatabaseError(t *testing.T) {
	db, mock, sqlDB := setupTokenTestDB(t)
	defer sqlDB.Close()

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM `blacklisted_tokens`")).
		WillReturnError(fmt.Errorf("database error"))
	mock.ExpectRollback()

	repo := NewTokenRepository(db)
	err := repo.CleanExpiredTokens()

	assert.Error(t, err)
	assert.Equal(t, "清理过期 token 失败", err.Error())
	assert.NoError(t, mock.ExpectationsWereMet())
}
