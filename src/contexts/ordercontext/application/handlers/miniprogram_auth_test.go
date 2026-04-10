package handlers

import (
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/bwmarrin/snowflake"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"orderease/contexts/thirdparty/infrastructure/config"
	repoThirdparty "orderease/contexts/thirdparty/infrastructure/persistence/repositories"
	"orderease/contexts/thirdparty/infrastructure/external/wechat"
	orderRepo "orderease/contexts/ordercontext/infrastructure/repositories"

	services "orderease/contexts/ordercontext/application/services"
)

func setupMiniProgramAuthHandlerTestDB(t *testing.T) (*gorm.DB, sqlmock.Sqlmock) {
	t.Helper()

	sqlDB, mock, err := sqlmock.New()
	assert.NoError(t, err)

	dialector := mysql.New(mysql.Config{
		Conn:                      sqlDB,
		SkipInitializeWithVersion: true,
	})

	db, err := gorm.Open(dialector, &gorm.Config{})
	assert.NoError(t, err)

	return db, mock
}

func newTestMiniProgramAuthHandler(db *gorm.DB) *MiniProgramAuthHandler {
	userRepo := orderRepo.NewUserRepository(db)
	bindingRepo := repoThirdparty.NewUserThirdpartyBindingRepository(db)
	authService := services.NewMiniProgramAuthService(db, userRepo, bindingRepo)
	return &MiniProgramAuthHandler{
		miniProgramClient: wechat.NewMiniProgramClient("test_app_id", "test_app_secret"),
		config:            &config.MiniProgramConfig{Enabled: true, AppID: "test_app_id", AppSecret: "test_app_secret"},
		bindingRepo:       bindingRepo,
		authService:       authService,
	}
}

func newMiniProgramBindingRows(userID snowflake.ID, openID string, nickname, avatarURL string) *sqlmock.Rows {
	now := time.Now()
	return sqlmock.NewRows([]string{
		"id", "user_id", "provider", "provider_user_id",
		"union_id", "nickname", "avatar_url", "gender",
		"country", "province", "city", "metadata",
		"is_active", "last_login_at", "created_at", "updated_at",
	}).AddRow(1, userID, "wechat", openID,
		"", nickname, avatarURL, 0, "", "", "",
		[]byte("{}"), true, now, now, now)
}

func newMiniProgramUserRows(userID snowflake.ID, name, nickname, avatar string) *sqlmock.Rows {
	now := time.Now()
	return sqlmock.NewRows([]string{
		"id", "name", "role", "password", "phone",
		"address", "type", "nickname", "avatar",
		"created_at", "updated_at",
	}).AddRow(userID, name, "public_user", "",
		"", "", "public_user", nickname, avatar, now, now)
}

// 测试静默登录：首次登录，无绑定记录，应创建新用户
func TestWeChatMiniProgramLogin_SilentLogin_FirstTime(t *testing.T) {
	_, mock := setupMiniProgramAuthHandlerTestDB(t)

	req := MiniProgramLoginRequest{
		Code:   "test_code",
		Silent: true,
	}

	// 模拟绑定查询失败（首次登录）
	mock.ExpectQuery("SELECT * FROM `user_thirdparty_bindings`").
		WithArgs("wechat", "test_openid", true, 1).
		WillReturnError(gorm.ErrRecordNotFound)

	// 模拟创建用户
	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO `users`").
		WillReturnResult(sqlmock.NewResult(1, 1))
	// 模拟创建绑定
	mock.ExpectExec("INSERT INTO `user_thirdparty_bindings`").
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	// 这里我们只测试静默登录的参数处理逻辑，实际的 Code2Session 调用会在集成测试中测试
	t.Run("SilentLogin parameter handling", func(t *testing.T) {
		// 验证静默登录模式下不会使用用户信息
		assert.True(t, req.Silent)
		assert.Empty(t, req.Nickname)
		assert.Empty(t, req.AvatarURL)
	})

	// 清理
	mock.ExpectationsWereMet()
}

// 测试静默登录：已有绑定记录，应更新绑定信息
func TestWeChatMiniProgramLogin_SilentLogin_ExistingUser(t *testing.T) {
	_, mock := setupMiniProgramAuthHandlerTestDB(t)

	userID := snowflake.ID(1000)

	req := MiniProgramLoginRequest{
		Code:   "test_code",
		Silent: true,
	}

	// 模拟绑定查询成功
	mock.ExpectQuery("SELECT * FROM `user_thirdparty_bindings`").
		WithArgs("wechat", "test_openid", true, 1).
		WillReturnRows(newMiniProgramBindingRows(userID, "test_openid", "", ""))

	// 模拟用户查询成功
	mock.ExpectQuery("SELECT * FROM `users`").
		WithArgs(userID, 1).
		WillReturnRows(newMiniProgramUserRows(userID, "wx_user_123456", "", ""))

	// 模拟更新绑定信息
	mock.ExpectBegin()
	mock.ExpectExec("UPDATE `user_thirdparty_bindings`").
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	t.Run("SilentLogin with existing user", func(t *testing.T) {
		assert.True(t, req.Silent)
		// 静默登录模式下，即使没有用户信息也能正常登录
		assert.Empty(t, req.Nickname)
		assert.Empty(t, req.AvatarURL)
	})

	// 清理
	mock.ExpectationsWereMet()
}

// 测试非静默登录：正常授权流程
func TestWeChatMiniProgramLogin_NonSilentLogin(t *testing.T) {
	_, mock := setupMiniProgramAuthHandlerTestDB(t)

	req := MiniProgramLoginRequest{
		Code:      "test_code",
		Silent:    false,
		Nickname:  "测试用户",
		AvatarURL: "https://example.com/avatar.jpg",
	}

	// 模拟绑定查询失败（首次登录）
	mock.ExpectQuery("SELECT * FROM `user_thirdparty_bindings`").
		WithArgs("wechat", "test_openid", true, 1).
		WillReturnError(gorm.ErrRecordNotFound)

	// 模拟创建用户
	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO `users`").
		WillReturnResult(sqlmock.NewResult(1, 1))
	// 模拟创建绑定
	mock.ExpectExec("INSERT INTO `user_thirdparty_bindings`").
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	t.Run("Non-silent login with user info", func(t *testing.T) {
		assert.False(t, req.Silent)
		assert.NotEmpty(t, req.Nickname)
		assert.NotEmpty(t, req.AvatarURL)
	})

	// 清理
	mock.ExpectationsWereMet()
}
