package services

import (
	"database/sql"
	"fmt"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/bwmarrin/snowflake"
	"github.com/stretchr/testify/assert"
	"orderease/models"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func setupTempTokenTestDB(t *testing.T) (*gorm.DB, sqlmock.Sqlmock, *sql.DB) {
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

func TestNewTempTokenService(t *testing.T) {
	db, _, sqlDB := setupTempTokenTestDB(t)
	defer sqlDB.Close()

	service := NewTempTokenService(db)

	assert.NotNil(t, service)
	assert.Equal(t, db, service.db)
}

func TestTempTokenService_CreateShopSystemUser_NewUser(t *testing.T) {
	db, mock, sqlDB := setupTempTokenTestDB(t)
	defer sqlDB.Close()

	shopID := snowflake.ID(123)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `users` WHERE type = ? AND name = ? ORDER BY `users`.`id` LIMIT ?")).
		WithArgs("system", "shop_123_system", 1).
		WillReturnError(gorm.ErrRecordNotFound)

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO `users`")).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	service := NewTempTokenService(db)
	user, err := service.CreateShopSystemUser(shopID)

	assert.NoError(t, err)
	assert.Equal(t, "shop_123_system", user.Name)
	assert.Equal(t, models.UserRolePublic, user.Role)
	assert.Equal(t, "system", user.Type)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestTempTokenService_CreateShopSystemUser_ExistingUser(t *testing.T) {
	db, mock, sqlDB := setupTempTokenTestDB(t)
	defer sqlDB.Close()

	shopID := snowflake.ID(123)
	existingUser := models.User{
		Name:  "shop_123_system",
		Role:  models.UserRolePublic,
		Type:  "system",
		Phone: "",
	}

	rows := sqlmock.NewRows([]string{"id", "name", "role", "type", "phone", "address", "password", "nickname", "created_at", "updated_at"}).
		AddRow(1, existingUser.Name, existingUser.Role, existingUser.Type, existingUser.Phone, "", "", "", time.Now(), time.Now())

	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `users` WHERE type = ? AND name = ? ORDER BY `users`.`id` LIMIT ?")).
		WithArgs("system", "shop_123_system", 1).
		WillReturnRows(rows)

	service := NewTempTokenService(db)
	user, err := service.CreateShopSystemUser(shopID)

	assert.NoError(t, err)
	assert.Equal(t, "shop_123_system", user.Name)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestTempTokenService_CreateShopSystemUser_DatabaseError(t *testing.T) {
	db, mock, sqlDB := setupTempTokenTestDB(t)
	defer sqlDB.Close()

	shopID := snowflake.ID(123)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `users` WHERE type = ? AND name = ? ORDER BY `users`.`id` LIMIT ?")).
		WithArgs("system", "shop_123_system", 1).
		WillReturnError(gorm.ErrRecordNotFound)

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO `users`")).
		WillReturnError(fmt.Errorf("database error"))
	mock.ExpectRollback()

	service := NewTempTokenService(db)
	user, err := service.CreateShopSystemUser(shopID)

	assert.Error(t, err)
	assert.Equal(t, models.User{}, user)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestTempTokenService_GenerateTempToken_NewToken(t *testing.T) {
	db, mock, sqlDB := setupTempTokenTestDB(t)
	defer sqlDB.Close()

	shopID := snowflake.ID(123)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `users` WHERE type = ? AND name = ? ORDER BY `users`.`id` LIMIT ?")).
		WithArgs("system", "shop_123_system", 1).
		WillReturnError(gorm.ErrRecordNotFound)

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO `users`")).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `temp_tokens` WHERE shop_id = ? ORDER BY `temp_tokens`.`id` LIMIT ?")).
		WithArgs(shopID, 1).
		WillReturnError(gorm.ErrRecordNotFound)

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO `temp_tokens`")).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	service := NewTempTokenService(db)
	token, err := service.GenerateTempToken(shopID)

	assert.NoError(t, err)
	assert.Equal(t, shopID, token.ShopID)
	assert.Len(t, token.Token, 6)
	assert.True(t, token.ExpiresAt.After(time.Now()))
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestTempTokenService_GenerateTempToken_ExistingToken(t *testing.T) {
	db, mock, sqlDB := setupTempTokenTestDB(t)
	defer sqlDB.Close()

	shopID := snowflake.ID(123)
	existingToken := models.TempToken{
		ShopID:    shopID,
		UserID:    1,
		Token:     "123456",
		ExpiresAt: time.Now().Add(-1 * time.Hour),
	}

	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `users` WHERE type = ? AND name = ? ORDER BY `users`.`id` LIMIT ?")).
		WithArgs("system", "shop_123_system", 1).
		WillReturnError(gorm.ErrRecordNotFound)

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO `users`")).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	rows := sqlmock.NewRows([]string{"id", "shop_id", "user_id", "token", "expires_at", "created_at"}).
		AddRow(1, existingToken.ShopID, existingToken.UserID, existingToken.Token, existingToken.ExpiresAt, time.Now())

	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `temp_tokens` WHERE shop_id = ? ORDER BY `temp_tokens`.`id` LIMIT ?")).
		WithArgs(shopID, 1).
		WillReturnRows(rows)

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta("UPDATE `temp_tokens`")).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	service := NewTempTokenService(db)
	token, err := service.GenerateTempToken(shopID)

	assert.NoError(t, err)
	assert.Equal(t, shopID, token.ShopID)
	assert.Len(t, token.Token, 6)
	assert.NotEqual(t, "123456", token.Token)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestTempTokenService_GetValidTempToken_ExistingValid(t *testing.T) {
	db, mock, sqlDB := setupTempTokenTestDB(t)
	defer sqlDB.Close()

	shopID := snowflake.ID(123)
	existingToken := models.TempToken{
		ShopID:    shopID,
		UserID:    1,
		Token:     "123456",
		ExpiresAt: time.Now().Add(1 * time.Hour),
	}

	rows := sqlmock.NewRows([]string{"id", "shop_id", "user_id", "token", "expires_at", "created_at"}).
		AddRow(1, existingToken.ShopID, existingToken.UserID, existingToken.Token, existingToken.ExpiresAt, time.Now())

	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `temp_tokens` WHERE shop_id = ? ORDER BY `temp_tokens`.`id` LIMIT ?")).
		WithArgs(shopID, 1).
		WillReturnRows(rows)

	service := NewTempTokenService(db)
	token, err := service.GetValidTempToken(shopID)

	assert.NoError(t, err)
	assert.Equal(t, "123456", token.Token)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestTempTokenService_GetValidTempToken_Expired(t *testing.T) {
	db, mock, sqlDB := setupTempTokenTestDB(t)
	defer sqlDB.Close()

	shopID := snowflake.ID(123)
	expiredToken := models.TempToken{
		ID:        1,
		ShopID:    shopID,
		UserID:    1,
		Token:     "123456",
		ExpiresAt: time.Now().Add(-1 * time.Hour),
	}

	rows := sqlmock.NewRows([]string{"id", "shop_id", "user_id", "token", "expires_at", "created_at"}).
		AddRow(expiredToken.ID, expiredToken.ShopID, expiredToken.UserID, expiredToken.Token, expiredToken.ExpiresAt, time.Now())

	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `temp_tokens` WHERE shop_id = ? ORDER BY `temp_tokens`.`id` LIMIT ?")).
		WithArgs(shopID, 1).
		WillReturnRows(rows)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `users` WHERE type = ? AND name = ? ORDER BY `users`.`id` LIMIT ?")).
		WithArgs("system", "shop_123_system", 1).
		WillReturnError(gorm.ErrRecordNotFound)

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO `users`")).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `temp_tokens` WHERE shop_id = ? ORDER BY `temp_tokens`.`id` LIMIT ?")).
		WithArgs(shopID, 1).
		WillReturnError(gorm.ErrRecordNotFound)

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO `temp_tokens`")).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	service := NewTempTokenService(db)
	token, err := service.GetValidTempToken(shopID)

	assert.NoError(t, err)
	assert.Len(t, token.Token, 6)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestTempTokenService_ValidateTempToken_Valid(t *testing.T) {
	db, mock, sqlDB := setupTempTokenTestDB(t)
	defer sqlDB.Close()

	shopID := snowflake.ID(123)
	tokenStr := "123456"
	userID := uint64(1)

	tokenRows := sqlmock.NewRows([]string{"id", "shop_id", "user_id", "token", "expires_at", "created_at"}).
		AddRow(1, shopID, userID, tokenStr, time.Now().Add(1*time.Hour), time.Now())

	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `temp_tokens` WHERE shop_id = ? AND token = ? ORDER BY `temp_tokens`.`id` LIMIT ?")).
		WithArgs(shopID, tokenStr, 1).
		WillReturnRows(tokenRows)

	userRows := sqlmock.NewRows([]string{"id", "name", "role", "type", "phone", "address", "password", "nickname", "created_at", "updated_at"}).
		AddRow(userID, "shop_123_system", models.UserRolePublic, "system", "", "", "", "", time.Now(), time.Now())

	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `users` WHERE id = ? ORDER BY `users`.`id` LIMIT ?")).
		WithArgs(userID, 1).
		WillReturnRows(userRows)

	service := NewTempTokenService(db)
	valid, user, err := service.ValidateTempToken(shopID, tokenStr)

	assert.NoError(t, err)
	assert.True(t, valid)
	assert.Equal(t, "shop_123_system", user.Name)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestTempTokenService_ValidateTempToken_NotFound(t *testing.T) {
	db, mock, sqlDB := setupTempTokenTestDB(t)
	defer sqlDB.Close()

	shopID := snowflake.ID(123)
	tokenStr := "123456"

	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `temp_tokens` WHERE shop_id = ? AND token = ? ORDER BY `temp_tokens`.`id` LIMIT ?")).
		WithArgs(shopID, tokenStr, 1).
		WillReturnError(gorm.ErrRecordNotFound)

	service := NewTempTokenService(db)
	valid, user, err := service.ValidateTempToken(shopID, tokenStr)

	assert.Error(t, err)
	assert.False(t, valid)
	assert.Equal(t, models.User{}, user)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestTempTokenService_ValidateTempToken_Expired(t *testing.T) {
	db, mock, sqlDB := setupTempTokenTestDB(t)
	defer sqlDB.Close()

	shopID := snowflake.ID(123)
	tokenStr := "123456"
	userID := uint64(1)

	tokenRows := sqlmock.NewRows([]string{"id", "shop_id", "user_id", "token", "expires_at", "created_at"}).
		AddRow(1, shopID, userID, tokenStr, time.Now().Add(-1*time.Hour), time.Now())

	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `temp_tokens` WHERE shop_id = ? AND token = ? ORDER BY `temp_tokens`.`id` LIMIT ?")).
		WithArgs(shopID, tokenStr, 1).
		WillReturnRows(tokenRows)

	service := NewTempTokenService(db)
	valid, user, err := service.ValidateTempToken(shopID, tokenStr)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "过期")
	assert.False(t, valid)
	assert.Equal(t, models.User{}, user)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestTempTokenService_ValidateTempToken_UserNotFound(t *testing.T) {
	db, mock, sqlDB := setupTempTokenTestDB(t)
	defer sqlDB.Close()

	shopID := snowflake.ID(123)
	tokenStr := "123456"
	userID := uint64(1)

	tokenRows := sqlmock.NewRows([]string{"id", "shop_id", "user_id", "token", "expires_at", "created_at"}).
		AddRow(1, shopID, userID, tokenStr, time.Now().Add(1*time.Hour), time.Now())

	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `temp_tokens` WHERE shop_id = ? AND token = ? ORDER BY `temp_tokens`.`id` LIMIT ?")).
		WithArgs(shopID, tokenStr, 1).
		WillReturnRows(tokenRows)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `users` WHERE id = ? ORDER BY `users`.`id` LIMIT ?")).
		WithArgs(userID, 1).
		WillReturnError(gorm.ErrRecordNotFound)

	service := NewTempTokenService(db)
	valid, user, err := service.ValidateTempToken(shopID, tokenStr)

	assert.Error(t, err)
	assert.False(t, valid)
	assert.Equal(t, models.User{}, user)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestTempTokenService_RefreshAllTempTokens_DatabaseError(t *testing.T) {
	db, mock, sqlDB := setupTempTokenTestDB(t)
	defer sqlDB.Close()

	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `shops`")).
		WillReturnError(fmt.Errorf("database error"))

	service := NewTempTokenService(db)
	err := service.RefreshAllTempTokens()

	assert.Error(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}
