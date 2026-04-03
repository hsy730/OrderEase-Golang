package handlers

import (
	"database/sql"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/bwmarrin/snowflake"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	thirdpartyuser "orderease/contexts/thirdparty/domain/user"
	"orderease/contexts/thirdparty/domain/oauth"
	"orderease/contexts/thirdparty/infrastructure/persistence/repositories"
)

func setupWechatHandlerTestDB(t *testing.T) (*gorm.DB, sqlmock.Sqlmock, *sql.DB) {
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

func newTestWechatHandler(db *gorm.DB) *WeChatHandler {
	bindingRepo := repositories.NewUserThirdpartyBindingRepository(db)
	userService := thirdpartyuser.NewService(db, bindingRepo)
	return &WeChatHandler{
		userService: userService,
		jwtService:  NewJWTService(),
	}
}

func newOAuthResult(openID string, rawData map[string]interface{}) *oauth.OAuthResult {
	return &oauth.OAuthResult{
		OpenID:      openID,
		UnionID:     "union_123",
		AccessToken: "access_token_1",
		RawData:     rawData,
	}
}

func newBindingRows(userID snowflake.ID, openID string, nickname, avatarURL string) *sqlmock.Rows {
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

func newUserRows(userID snowflake.ID, name, nickname, avatar string) *sqlmock.Rows {
	now := time.Now()
	return sqlmock.NewRows([]string{
		"id", "name", "role", "password", "phone",
		"address", "type", "nickname", "avatar",
		"created_at", "updated_at",
	}).AddRow(userID, name, "public_user", "",
		"", "", "public_user", nickname, avatar, now, now)
}

// ==================== FindOrCreateByOpenID Tests ====================

// 首次授权：无绑定记录，应创建新用户和新绑定
func TestFindOrCreateUser_FirstTimeAuthorization(t *testing.T) {
	db, mock, sqlDB := setupWechatHandlerTestDB(t)
	defer sqlDB.Close()

	h := newTestWechatHandler(db)

	result := newOAuthResult("wx_openid_001", map[string]interface{}{
		"nickname":   "微信昵称_A",
		"headimgurl": "https://wx.qq.com/avatar/a.jpg",
	})

	mock.ExpectQuery("SELECT \\* FROM `user_thirdparty_bindings`").
		WithArgs("wechat", "wx_openid_001", true, 1).
		WillReturnError(gorm.ErrRecordNotFound)

	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO `users`").
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("INSERT INTO `user_thirdparty_bindings`").
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	user, err := h.userService.FindOrCreateByOpenID(result)

	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, "微信昵称_A", user.Name)
	assert.Equal(t, "public_user", user.Type)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// 反复授权：已有绑定且用户信息全部变化，应更新 User 表的 Name/Nickname/Avatar
func TestFindOrCreateUser_ReAuthorization_AllFieldsUpdated(t *testing.T) {
	db, mock, sqlDB := setupWechatHandlerTestDB(t)
	defer sqlDB.Close()

	h := newTestWechatHandler(db)

	userID := snowflake.ID(1000)

	result := newOAuthResult("wx_openid_002", map[string]interface{}{
		"nickname":   "新微信昵称",
		"headimgurl": "https://wx.qq.com/avatar/new.jpg",
	})

	mock.ExpectQuery("SELECT \\* FROM `user_thirdparty_bindings`").
		WithArgs("wechat", "wx_openid_002", true, 1).
		WillReturnRows(newBindingRows(userID, "wx_openid_002", "旧昵称", ""))

	mock.ExpectQuery("SELECT \\* FROM `users`").
		WithArgs(userID, 1).
		WillReturnRows(newUserRows(userID, "微信用户_did_002", "旧昵称", ""))

	mock.ExpectBegin()
	mock.ExpectExec("UPDATE `user_thirdparty_bindings`").
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	mock.ExpectBegin()
	mock.ExpectExec("UPDATE `users`").
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	user, err := h.userService.FindOrCreateByOpenID(result)

	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, userID, user.ID)
	assert.Equal(t, "新微信昵称", user.Name)
	assert.Equal(t, "新微信昵称", user.Nickname)
	assert.Equal(t, "https://wx.qq.com/avatar/new.jpg", user.Avatar)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// 反复授权：已有绑定但用户信息无变化，不应调用 Save 更新 User 表
func TestFindOrCreateUser_ReAuthorization_NoChanges(t *testing.T) {
	db, mock, sqlDB := setupWechatHandlerTestDB(t)
	defer sqlDB.Close()

	h := newTestWechatHandler(db)

	userID := snowflake.ID(2000)

	result := newOAuthResult("wx_openid_003", map[string]interface{}{
		"nickname":   "相同昵称",
		"headimgurl": "https://wx.qq.com/avatar/same.jpg",
	})

	mock.ExpectQuery("SELECT \\* FROM `user_thirdparty_bindings`").
		WithArgs("wechat", "wx_openid_003", true, 1).
		WillReturnRows(newBindingRows(userID, "wx_openid_003", "相同昵称", "https://wx.qq.com/avatar/same.jpg"))

	mock.ExpectQuery("SELECT \\* FROM `users`").
		WithArgs(userID, 1).
		WillReturnRows(newUserRows(userID, "相同昵称", "相同昵称", "https://wx.qq.com/avatar/same.jpg"))

	mock.ExpectBegin()
	mock.ExpectExec("UPDATE `user_thirdparty_bindings`").
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	user, err := h.userService.FindOrCreateByOpenID(result)

	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, userID, user.ID)
	assert.Equal(t, "相同昵称", user.Name)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// 反复授权：仅有部分字段变化（仅昵称变化），应只更新变化的字段
func TestFindOrCreateUser_ReAuthorization_PartialUpdate_NicknameOnly(t *testing.T) {
	db, mock, sqlDB := setupWechatHandlerTestDB(t)
	defer sqlDB.Close()

	h := newTestWechatHandler(db)

	userID := snowflake.ID(3000)

	result := newOAuthResult("wx_openid_004", map[string]interface{}{
		"nickname":   "新昵称_仅改此项",
		"headimgurl": "https://wx.qq.com/avatar/exist.jpg",
	})

	mock.ExpectQuery("SELECT \\* FROM `user_thirdparty_bindings`").
		WithArgs("wechat", "wx_openid_004", true, 1).
		WillReturnRows(newBindingRows(userID, "wx_openid_004", "旧昵称", "https://wx.qq.com/avatar/exist.jpg"))

	mock.ExpectQuery("SELECT \\* FROM `users`").
		WithArgs(userID, 1).
		WillReturnRows(newUserRows(userID, "已同步名称", "旧昵称", "https://wx.qq.com/avatar/exist.jpg"))

	mock.ExpectBegin()
	mock.ExpectExec("UPDATE `user_thirdparty_bindings`").
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	mock.ExpectBegin()
	mock.ExpectExec("UPDATE `users`").
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	user, err := h.userService.FindOrCreateByOpenID(result)

	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, "新昵称_仅改此项", user.Nickname)
	assert.Equal(t, "https://wx.qq.com/avatar/exist.jpg", user.Avatar)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// 反复授权：绑定存在但用户记录被删除，应返回错误
func TestFindOrCreateUser_ReAuthorization_UserDeleted(t *testing.T) {
	db, mock, sqlDB := setupWechatHandlerTestDB(t)
	defer sqlDB.Close()

	h := newTestWechatHandler(db)

	userID := snowflake.ID(9999)

	result := newOAuthResult("wx_openid_deleted", map[string]interface{}{
		"nickname":   "测试用户",
		"headimgurl": "https://wx.qq.com/avatar/test.jpg",
	})

	mock.ExpectQuery("SELECT \\* FROM `user_thirdparty_bindings`").
		WithArgs("wechat", "wx_openid_deleted", true, 1).
		WillReturnRows(newBindingRows(userID, "wx_openid_deleted", "孤儿绑定", ""))

	mock.ExpectQuery("SELECT \\* FROM `users`").
		WithArgs(userID, 1).
		WillReturnError(gorm.ErrRecordNotFound)

	user, err := h.userService.FindOrCreateByOpenID(result)

	assert.Error(t, err)
	assert.Nil(t, user)
	assert.Contains(t, err.Error(), "find user by binding failed")
	assert.NoError(t, mock.ExpectationsWereMet())
}

// 首次授权：无 nickname 和 headimgurl 回退为 OpenID 后缀
func TestFindOrCreateUser_FirstTime_NoNicknameOrAvatar(t *testing.T) {
	db, mock, sqlDB := setupWechatHandlerTestDB(t)
	defer sqlDB.Close()

	h := newTestWechatHandler(db)

	result := newOAuthResult("wx_openid_noinfo", map[string]interface{}{})

	mock.ExpectQuery("SELECT \\* FROM `user_thirdparty_bindings`").
		WithArgs("wechat", "wx_openid_noinfo", true, 1).
		WillReturnError(gorm.ErrRecordNotFound)

	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO `users`").
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("INSERT INTO `user_thirdparty_bindings`").
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	user, err := h.userService.FindOrCreateByOpenID(result)

	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Contains(t, user.Name, "noinfo")
	assert.NoError(t, mock.ExpectationsWereMet())
}
