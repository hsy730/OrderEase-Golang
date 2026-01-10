package user

import (
	"testing"
	"time"

	"orderease/domain/shared"

	"github.com/stretchr/testify/assert"
)

func TestUserRole_IsValid(t *testing.T) {
	tests := []struct {
		name string
		role UserRole
		want bool
	}{
		{"private user", UserRolePrivate, true},
		{"public user", UserRolePublic, true},
		{"invalid role", UserRole("unknown"), false},
		{"empty role", UserRole(""), false},
		{"random role", UserRole("admin"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.role.IsValid()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestUserType_IsValid(t *testing.T) {
	tests := []struct {
		name string
		typ  UserType
		want bool
	}{
		{"delivery type", UserTypeDelivery, true},
		{"pickup type", UserTypePickup, true},
		{"system type", UserTypeSystem, true},
		{"invalid type", UserType("unknown"), false},
		{"empty type", UserType(""), false},
		{"random type", UserType("admin"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.typ.IsValid()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestNewUser(t *testing.T) {
	tests := []struct {
		name      string
		userName  string
		role      UserRole
		userType  UserType
		password  string
		wantErr   bool
		errMsg    string
		validate  func(*testing.T, *User)
	}{
		{
			name:     "valid private delivery user",
			userName: "张三",
			role:     UserRolePrivate,
			userType: UserTypeDelivery,
			password: "password123",
			wantErr:  false,
			validate: func(t *testing.T, u *User) {
				assert.Equal(t, shared.ID(0), u.ID)
				assert.Equal(t, "张三", u.Name)
				assert.Equal(t, UserRolePrivate, u.Role)
				assert.Equal(t, UserTypeDelivery, u.Type)
				assert.Equal(t, "password123", u.Password)
				assert.False(t, u.CreatedAt.IsZero())
				assert.False(t, u.UpdatedAt.IsZero())
			},
		},
		{
			name:     "valid public pickup user",
			userName: "李四",
			role:     UserRolePublic,
			userType: UserTypePickup,
			password: "pass456",
			wantErr:  false,
		},
		{
			name:     "valid system user",
			userName: "系统用户",
			role:     UserRolePublic,
			userType: UserTypeSystem,
			password: "admin",
			wantErr:  false,
		},
		{
			name:     "empty name",
			userName: "",
			role:     UserRolePrivate,
			userType: UserTypeDelivery,
			password: "password123",
			wantErr:  true,
			errMsg:   "用户名不能为空",
		},
		{
			name:     "invalid role",
			userName: "张三",
			role:     UserRole("unknown"),
			userType: UserTypeDelivery,
			password: "password123",
			wantErr:  true,
			errMsg:   "无效的用户角色",
		},
		{
			name:     "invalid type",
			userName: "张三",
			role:     UserRolePrivate,
			userType: UserType("unknown"),
			password: "password123",
			wantErr:  true,
			errMsg:   "无效的用户类型",
		},
		{
			name:     "empty password valid",
			userName: "张三",
			role:     UserRolePrivate,
			userType: UserTypeDelivery,
			password: "",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewUser(tt.userName, tt.role, tt.userType, tt.password)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
				assert.Nil(t, got)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, got)
				if tt.validate != nil {
					tt.validate(t, got)
				}
			}
		})
	}
}

func TestUser_UpdateBasicInfo(t *testing.T) {
	tests := []struct {
		name     string
		user     *User
		userName string
		phone    string
		address  string
		wantErr  bool
		validate func(*testing.T, *User)
	}{
		{
			name: "update all fields",
			user: &User{
				Name:    "原姓名",
				Phone:   "",
				Address: "",
			},
			userName: "新姓名",
			phone:   "123456789",
			address: "新地址",
			wantErr: false,
			validate: func(t *testing.T, u *User) {
				assert.Equal(t, "新姓名", u.Name)
				assert.Equal(t, "123456789", u.Phone)
				assert.Equal(t, "新地址", u.Address)
			},
		},
		{
			name: "update only some fields",
			user: &User{
				Name:    "原姓名",
				Phone:   "原电话",
				Address: "原地址",
			},
			userName: "",
			phone:   "新电话",
			address: "",
			wantErr: false,
			validate: func(t *testing.T, u *User) {
				assert.Equal(t, "原姓名", u.Name, "name should not change")
				assert.Equal(t, "新电话", u.Phone)
				assert.Equal(t, "原地址", u.Address, "address should not change")
			},
		},
		{
			name: "update with empty strings keeps original",
			user: &User{
				Name:    "原姓名",
				Phone:   "原电话",
				Address: "原地址",
			},
			userName: "",
			phone:   "",
			address: "",
			wantErr: false,
			validate: func(t *testing.T, u *User) {
				assert.Equal(t, "原姓名", u.Name)
				assert.Equal(t, "原电话", u.Phone)
				assert.Equal(t, "原地址", u.Address)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			oldUpdatedAt := tt.user.UpdatedAt
			err := tt.user.UpdateBasicInfo(tt.userName, tt.phone, tt.address)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.True(t, tt.user.UpdatedAt.After(oldUpdatedAt) || tt.user.UpdatedAt.Equal(oldUpdatedAt))
				if tt.validate != nil {
					tt.validate(t, tt.user)
				}
			}
		})
	}
}

func TestUser_UpdatePassword(t *testing.T) {
	tests := []struct {
		name        string
		user        *User
		newPassword string
		wantErr     bool
		errMsg      string
	}{
		{
			name:        "valid password",
			user:        &User{Password: "oldpass"},
			newPassword: "newpass123",
			wantErr:     false,
		},
		{
			name:        "empty password",
			user:        &User{Password: "oldpass"},
			newPassword: "",
			wantErr:     true,
			errMsg:      "新密码不能为空",
		},
		{
			name:        "password with special characters",
			user:        &User{Password: "oldpass"},
			newPassword: "new@#$pass",
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			oldUpdatedAt := tt.user.UpdatedAt
			err := tt.user.UpdatePassword(tt.newPassword)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
				assert.True(t, tt.user.UpdatedAt.After(oldUpdatedAt) || tt.user.UpdatedAt.Equal(oldUpdatedAt))
				assert.Equal(t, tt.newPassword, tt.user.Password)
			}
		})
	}
}

func TestUser_IsSystemUser(t *testing.T) {
	tests := []struct {
		name     string
		userType UserType
		want     bool
	}{
		{"system user", UserTypeSystem, true},
		{"delivery user", UserTypeDelivery, false},
		{"pickup user", UserTypePickup, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := &User{Type: tt.userType}
			got := u.IsSystemUser()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestUser_IsPublicUser(t *testing.T) {
	tests := []struct {
		name string
		role UserRole
		want bool
	}{
		{"public user", UserRolePublic, true},
		{"private user", UserRolePrivate, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := &User{Role: tt.role}
			got := u.IsPublicUser()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestUser_IsPrivateUser(t *testing.T) {
	tests := []struct {
		name string
		role UserRole
		want bool
	}{
		{"private user", UserRolePrivate, true},
		{"public user", UserRolePublic, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := &User{Role: tt.role}
			got := u.IsPrivateUser()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestUser_Timestamps(t *testing.T) {
	before := time.Now()
	user, err := NewUser("张三", UserRolePrivate, UserTypeDelivery, "password123")
	after := time.Now()

	assert.NoError(t, err)
	assert.True(t, user.CreatedAt.After(before) || user.CreatedAt.Equal(before))
	assert.True(t, user.CreatedAt.Before(after) || user.CreatedAt.Equal(after))
	assert.True(t, user.UpdatedAt.After(before) || user.UpdatedAt.Equal(before))
	assert.True(t, user.UpdatedAt.Before(after) || user.UpdatedAt.Equal(after))
}

func TestUser_AllRolesAndTypes(t *testing.T) {
	// Verify all defined roles are valid
	roles := []UserRole{UserRolePrivate, UserRolePublic}
	for _, role := range roles {
		t.Run(string(role), func(t *testing.T) {
			assert.True(t, role.IsValid(), "%s should be valid", role)
		})
	}

	// Verify all defined types are valid
	types := []UserType{UserTypeDelivery, UserTypePickup, UserTypeSystem}
	for _, typ := range types {
		t.Run(string(typ), func(t *testing.T) {
			assert.True(t, typ.IsValid(), "%s should be valid", typ)
		})
	}
}
