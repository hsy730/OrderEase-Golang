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
	"orderease/contexts/thirdparty/domain/oauth"
	"orderease/models"
)

func setupThirdpartyTestDB(t *testing.T) (*gorm.DB, sqlmock.Sqlmock, *sql.DB) {
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

func TestNewUserThirdpartyBindingRepository(t *testing.T) {
	db, _, sqlDB := setupThirdpartyTestDB(t)
	defer sqlDB.Close()

	repo := NewUserThirdpartyBindingRepository(db)

	assert.NotNil(t, repo)
}

func TestProviderWeChat(t *testing.T) {
	assert.Equal(t, "wechat", oauth.ProviderWeChat.String())
}

func TestProvider(t *testing.T) {
	provider := oauth.Provider("alipay")
	assert.Equal(t, "alipay", provider.String())
}

// ==================== FindByProviderAndUserID Tests ====================

func TestUserThirdpartyBindingRepository_FindByProviderAndUserID_Success(t *testing.T) {
	db, mock, sqlDB := setupThirdpartyTestDB(t)
	defer sqlDB.Close()

	now := time.Now()
	rows := sqlmock.NewRows([]string{"id", "user_id", "provider", "provider_user_id", "provider_nickname", "is_active", "last_login_at", "created_at", "updated_at"}).
		AddRow(1, 100, "wechat", "wx123", "TestUser", true, now, now, now)

	mock.ExpectQuery("SELECT \\* FROM `user_thirdparty_bindings`").
		WithArgs("wechat", "wx123", true, 1).
		WillReturnRows(rows)

	repo := NewUserThirdpartyBindingRepository(db)
	binding, err := repo.FindByProviderAndUserID(oauth.ProviderWeChat, "wx123")

	assert.NoError(t, err)
	assert.NotNil(t, binding)
	assert.Equal(t, snowflake.ID(100), binding.UserID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserThirdpartyBindingRepository_FindByProviderAndUserID_NotFound(t *testing.T) {
	db, mock, sqlDB := setupThirdpartyTestDB(t)
	defer sqlDB.Close()

	mock.ExpectQuery("SELECT \\* FROM `user_thirdparty_bindings`").
		WithArgs("wechat", "nonexistent", true, 1).
		WillReturnError(gorm.ErrRecordNotFound)

	repo := NewUserThirdpartyBindingRepository(db)
	binding, err := repo.FindByProviderAndUserID(oauth.ProviderWeChat, "nonexistent")

	assert.Error(t, err)
	assert.Nil(t, binding)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// ==================== FindByUserID Tests ====================

func TestUserThirdpartyBindingRepository_FindByUserID_Success(t *testing.T) {
	db, mock, sqlDB := setupThirdpartyTestDB(t)
	defer sqlDB.Close()

	now := time.Now()
	rows := sqlmock.NewRows([]string{"id", "user_id", "provider", "provider_user_id", "provider_nickname", "is_active", "last_login_at", "created_at", "updated_at"}).
		AddRow(1, 100, "wechat", "wx123", "TestUser1", true, now, now, now).
		AddRow(2, 100, "alipay", "alipay456", "TestUser2", true, now, now, now)

	mock.ExpectQuery("SELECT \\* FROM `user_thirdparty_bindings`").
		WithArgs(uint64(100), true).
		WillReturnRows(rows)

	repo := NewUserThirdpartyBindingRepository(db)
	bindings, err := repo.FindByUserID(100)

	assert.NoError(t, err)
	assert.Len(t, bindings, 2)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserThirdpartyBindingRepository_FindByUserID_Empty(t *testing.T) {
	db, mock, sqlDB := setupThirdpartyTestDB(t)
	defer sqlDB.Close()

	rows := sqlmock.NewRows([]string{"id", "user_id", "provider", "provider_user_id", "provider_nickname", "is_active", "last_login_at", "created_at", "updated_at"})

	mock.ExpectQuery("SELECT \\* FROM `user_thirdparty_bindings`").
		WithArgs(uint64(999), true).
		WillReturnRows(rows)

	repo := NewUserThirdpartyBindingRepository(db)
	bindings, err := repo.FindByUserID(999)

	assert.NoError(t, err)
	assert.Empty(t, bindings)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// ==================== FindByUserIDAndProvider Tests ====================

func TestUserThirdpartyBindingRepository_FindByUserIDAndProvider_Success(t *testing.T) {
	db, mock, sqlDB := setupThirdpartyTestDB(t)
	defer sqlDB.Close()

	now := time.Now()
	rows := sqlmock.NewRows([]string{"id", "user_id", "provider", "provider_user_id", "provider_nickname", "is_active", "last_login_at", "created_at", "updated_at"}).
		AddRow(1, 100, "wechat", "wx123", "TestUser", true, now, now, now)

	mock.ExpectQuery("SELECT \\* FROM `user_thirdparty_bindings`").
		WithArgs(uint64(100), "wechat", true, 1).
		WillReturnRows(rows)

	repo := NewUserThirdpartyBindingRepository(db)
	binding, err := repo.FindByUserIDAndProvider(100, oauth.ProviderWeChat)

	assert.NoError(t, err)
	assert.NotNil(t, binding)
	assert.Equal(t, "wechat", binding.Provider)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserThirdpartyBindingRepository_FindByUserIDAndProvider_NotFound(t *testing.T) {
	db, mock, sqlDB := setupThirdpartyTestDB(t)
	defer sqlDB.Close()

	mock.ExpectQuery("SELECT \\* FROM `user_thirdparty_bindings`").
		WithArgs(uint64(999), "wechat", true, 1).
		WillReturnError(gorm.ErrRecordNotFound)

	repo := NewUserThirdpartyBindingRepository(db)
	binding, err := repo.FindByUserIDAndProvider(999, oauth.ProviderWeChat)

	assert.Error(t, err)
	assert.Nil(t, binding)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// ==================== Create Tests ====================

func TestUserThirdpartyBindingRepository_Create_Success(t *testing.T) {
	db, mock, sqlDB := setupThirdpartyTestDB(t)
	defer sqlDB.Close()

	binding := &models.UserThirdpartyBinding{
		UserID:         100,
		Provider:       "wechat",
		ProviderUserID: "wx123",
		Nickname:       "TestUser",
		IsActive:       true,
	}

	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO `user_thirdparty_bindings`").
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	repo := NewUserThirdpartyBindingRepository(db)
	err := repo.Create(binding)

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserThirdpartyBindingRepository_Create_DatabaseError(t *testing.T) {
	db, mock, sqlDB := setupThirdpartyTestDB(t)
	defer sqlDB.Close()

	binding := &models.UserThirdpartyBinding{
		UserID:         100,
		Provider:       "wechat",
		ProviderUserID: "wx123",
	}

	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO `user_thirdparty_bindings`").
		WillReturnError(fmt.Errorf("database error"))
	mock.ExpectRollback()

	repo := NewUserThirdpartyBindingRepository(db)
	err := repo.Create(binding)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "create binding failed")
	assert.NoError(t, mock.ExpectationsWereMet())
}

// ==================== Update Tests ====================

func TestUserThirdpartyBindingRepository_Update_Success(t *testing.T) {
	db, mock, sqlDB := setupThirdpartyTestDB(t)
	defer sqlDB.Close()

	binding := &models.UserThirdpartyBinding{
		ID:             1,
		UserID:         100,
		Provider:       "wechat",
		ProviderUserID: "wx123",
		Nickname:       "UpdatedUser",
		IsActive:       true,
	}

	mock.ExpectBegin()
	mock.ExpectExec("UPDATE `user_thirdparty_bindings`").
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	repo := NewUserThirdpartyBindingRepository(db)
	err := repo.Update(binding)

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserThirdpartyBindingRepository_Update_DatabaseError(t *testing.T) {
	db, mock, sqlDB := setupThirdpartyTestDB(t)
	defer sqlDB.Close()

	binding := &models.UserThirdpartyBinding{
		ID:             1,
		UserID:         100,
		Provider:       "wechat",
		ProviderUserID: "wx123",
	}

	mock.ExpectBegin()
	mock.ExpectExec("UPDATE `user_thirdparty_bindings`").
		WillReturnError(fmt.Errorf("database error"))
	mock.ExpectRollback()

	repo := NewUserThirdpartyBindingRepository(db)
	err := repo.Update(binding)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "update binding failed")
	assert.NoError(t, mock.ExpectationsWereMet())
}

// ==================== UpdateLastLogin Tests ====================

func TestUserThirdpartyBindingRepository_UpdateLastLogin_Success(t *testing.T) {
	db, mock, sqlDB := setupThirdpartyTestDB(t)
	defer sqlDB.Close()

	mock.ExpectBegin()
	mock.ExpectExec("UPDATE `user_thirdparty_bindings`").
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	repo := NewUserThirdpartyBindingRepository(db)
	err := repo.UpdateLastLogin(1)

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserThirdpartyBindingRepository_UpdateLastLogin_DatabaseError(t *testing.T) {
	db, mock, sqlDB := setupThirdpartyTestDB(t)
	defer sqlDB.Close()

	mock.ExpectBegin()
	mock.ExpectExec("UPDATE `user_thirdparty_bindings`").
		WillReturnError(fmt.Errorf("database error"))
	mock.ExpectRollback()

	repo := NewUserThirdpartyBindingRepository(db)
	err := repo.UpdateLastLogin(1)

	assert.Error(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// ==================== Deactivate Tests ====================

func TestUserThirdpartyBindingRepository_Deactivate_Success(t *testing.T) {
	db, mock, sqlDB := setupThirdpartyTestDB(t)
	defer sqlDB.Close()

	mock.ExpectBegin()
	mock.ExpectExec("UPDATE `user_thirdparty_bindings`").
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	repo := NewUserThirdpartyBindingRepository(db)
	err := repo.Deactivate(1)

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserThirdpartyBindingRepository_Deactivate_DatabaseError(t *testing.T) {
	db, mock, sqlDB := setupThirdpartyTestDB(t)
	defer sqlDB.Close()

	mock.ExpectBegin()
	mock.ExpectExec("UPDATE `user_thirdparty_bindings`").
		WillReturnError(fmt.Errorf("database error"))
	mock.ExpectRollback()

	repo := NewUserThirdpartyBindingRepository(db)
	err := repo.Deactivate(1)

	assert.Error(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// ==================== Delete Tests ====================

func TestUserThirdpartyBindingRepository_Delete_Success(t *testing.T) {
	db, mock, sqlDB := setupThirdpartyTestDB(t)
	defer sqlDB.Close()

	mock.ExpectBegin()
	mock.ExpectExec("DELETE FROM `user_thirdparty_bindings`").
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	repo := NewUserThirdpartyBindingRepository(db)
	err := repo.Delete(1)

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserThirdpartyBindingRepository_Delete_DatabaseError(t *testing.T) {
	db, mock, sqlDB := setupThirdpartyTestDB(t)
	defer sqlDB.Close()

	mock.ExpectBegin()
	mock.ExpectExec("DELETE FROM `user_thirdparty_bindings`").
		WillReturnError(fmt.Errorf("database error"))
	mock.ExpectRollback()

	repo := NewUserThirdpartyBindingRepository(db)
	err := repo.Delete(1)

	assert.Error(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// ==================== ListActive Tests ====================

func TestUserThirdpartyBindingRepository_ListActive_Success(t *testing.T) {
	db, mock, sqlDB := setupThirdpartyTestDB(t)
	defer sqlDB.Close()

	now := time.Now()
	rows := sqlmock.NewRows([]string{"id", "user_id", "provider", "provider_user_id", "provider_nickname", "is_active", "last_login_at", "created_at", "updated_at"}).
		AddRow(1, 100, "wechat", "wx123", "TestUser1", true, now, now, now).
		AddRow(2, 101, "alipay", "alipay456", "TestUser2", true, now, now, now)

	mock.ExpectQuery("SELECT \\* FROM `user_thirdparty_bindings`").
		WithArgs(true, 10).
		WillReturnRows(rows)

	repo := NewUserThirdpartyBindingRepository(db)
	bindings, err := repo.ListActive(10, 0)

	assert.NoError(t, err)
	assert.Len(t, bindings, 2)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserThirdpartyBindingRepository_ListActive_Empty(t *testing.T) {
	db, mock, sqlDB := setupThirdpartyTestDB(t)
	defer sqlDB.Close()

	rows := sqlmock.NewRows([]string{"id", "user_id", "provider", "provider_user_id", "provider_nickname", "is_active", "last_login_at", "created_at", "updated_at"})

	mock.ExpectQuery("SELECT \\* FROM `user_thirdparty_bindings`").
		WithArgs(true).
		WillReturnRows(rows)

	repo := NewUserThirdpartyBindingRepository(db)
	bindings, err := repo.ListActive(0, 0)

	assert.NoError(t, err)
	assert.Empty(t, bindings)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// ==================== CountByProvider Tests ====================

func TestUserThirdpartyBindingRepository_CountByProvider_Success(t *testing.T) {
	db, mock, sqlDB := setupThirdpartyTestDB(t)
	defer sqlDB.Close()

	rows := sqlmock.NewRows([]string{"provider", "count"}).
		AddRow("wechat", 50).
		AddRow("alipay", 30)

	mock.ExpectQuery("SELECT provider, COUNT\\(\\*\\) as count FROM `user_thirdparty_bindings`").
		WillReturnRows(rows)

	repo := NewUserThirdpartyBindingRepository(db)
	counts, err := repo.CountByProvider()

	assert.NoError(t, err)
	assert.Equal(t, int64(50), counts["wechat"])
	assert.Equal(t, int64(30), counts["alipay"])
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserThirdpartyBindingRepository_CountByProvider_DatabaseError(t *testing.T) {
	db, mock, sqlDB := setupThirdpartyTestDB(t)
	defer sqlDB.Close()

	mock.ExpectQuery("SELECT provider, COUNT\\(\\*\\) as count FROM `user_thirdparty_bindings`").
		WillReturnError(fmt.Errorf("database error"))

	repo := NewUserThirdpartyBindingRepository(db)
	counts, err := repo.CountByProvider()

	assert.Error(t, err)
	assert.Nil(t, counts)
	assert.NoError(t, mock.ExpectationsWereMet())
}
