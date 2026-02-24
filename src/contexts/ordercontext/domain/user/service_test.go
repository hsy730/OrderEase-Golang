package user

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

// MockRepository 用于测试的模拟仓储
type MockRepository struct {
	createFunc    func(*User) error
	getByIDFunc   func(UserID) (*User, error)
	getByUsernameFunc func(string) (*User, error)
	phoneExistsFunc   func(string) (bool, error)
	usernameExistsFunc func(string) (bool, error)
	updateFunc    func(*User) error
	deleteFunc    func(*User) error
}

func (m *MockRepository) Create(user *User) error {
	if m.createFunc != nil {
		return m.createFunc(user)
	}
	return nil
}

func (m *MockRepository) GetByID(id UserID) (*User, error) {
	if m.getByIDFunc != nil {
		return m.getByIDFunc(id)
	}
	return nil, errors.New("not found")
}

func (m *MockRepository) GetByUsername(username string) (*User, error) {
	if m.getByUsernameFunc != nil {
		return m.getByUsernameFunc(username)
	}
	return nil, errors.New("not found")
}

func (m *MockRepository) PhoneExists(phone string) (bool, error) {
	if m.phoneExistsFunc != nil {
		return m.phoneExistsFunc(phone)
	}
	return false, nil
}

func (m *MockRepository) UsernameExists(username string) (bool, error) {
	if m.usernameExistsFunc != nil {
		return m.usernameExistsFunc(username)
	}
	return false, nil
}

func (m *MockRepository) Update(user *User) error {
	if m.updateFunc != nil {
		return m.updateFunc(user)
	}
	return nil
}

func (m *MockRepository) Delete(user *User) error {
	if m.deleteFunc != nil {
		return m.deleteFunc(user)
	}
	return nil
}

// ==================== Register Tests ====================

func TestService_Register(t *testing.T) {
	mockRepo := &MockRepository{
		usernameExistsFunc: func(username string) (bool, error) {
			return false, nil
		},
		phoneExistsFunc: func(phone string) (bool, error) {
			return false, nil
		},
		createFunc: func(user *User) error {
			return nil
		},
	}

	service := NewService(mockRepo)

	dto := RegisterUserDTO{
		Username: "testuser",
		Phone:    "13800138000",
		Password: "abc123",
		UserType: string(UserTypeDelivery),
		Role:     string(UserRolePublic),
	}

	user, err := service.Register(dto)
	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, "testuser", user.Name())
	assert.Equal(t, "13800138000", user.Phone())
	assert.Equal(t, UserTypeDelivery, user.UserType())
	assert.Equal(t, UserRolePublic, user.Role())
}

func TestService_Register_UsernameExists(t *testing.T) {
	mockRepo := &MockRepository{
		usernameExistsFunc: func(username string) (bool, error) {
			return true, nil
		},
	}

	service := NewService(mockRepo)

	dto := RegisterUserDTO{
		Username: "existinguser",
		Phone:    "13800138000",
		Password: "abc123",
		UserType: string(UserTypeDelivery),
		Role:     string(UserRolePublic),
	}

	user, err := service.Register(dto)
	assert.Error(t, err)
	assert.Equal(t, ErrUsernameAlreadyExists, err)
	assert.Nil(t, user)
}

func TestService_Register_PhoneExists(t *testing.T) {
	mockRepo := &MockRepository{
		usernameExistsFunc: func(username string) (bool, error) {
			return false, nil
		},
		phoneExistsFunc: func(phone string) (bool, error) {
			return true, nil
		},
	}

	service := NewService(mockRepo)

	dto := RegisterUserDTO{
		Username: "testuser",
		Phone:    "13800138000",
		Password: "abc123",
		UserType: string(UserTypeDelivery),
		Role:     string(UserRolePublic),
	}

	user, err := service.Register(dto)
	assert.Error(t, err)
	assert.Equal(t, ErrPhoneAlreadyExists, err)
	assert.Nil(t, user)
}

func TestService_Register_InvalidPassword(t *testing.T) {
	mockRepo := &MockRepository{
		usernameExistsFunc: func(username string) (bool, error) {
			return false, nil
		},
	}

	service := NewService(mockRepo)

	dto := RegisterUserDTO{
		Username: "testuser",
		Phone:    "13800138000",
		Password: "ab12", // too short
		UserType: string(UserTypeDelivery),
		Role:     string(UserRolePublic),
	}

	user, err := service.Register(dto)
	assert.Error(t, err)
	assert.Equal(t, ErrInvalidPassword, err)
	assert.Nil(t, user)
}

func TestService_Register_InvalidUserType(t *testing.T) {
	mockRepo := &MockRepository{
		usernameExistsFunc: func(username string) (bool, error) {
			return false, nil
		},
	}

	service := NewService(mockRepo)

	dto := RegisterUserDTO{
		Username: "testuser",
		Phone:    "13800138000",
		Password: "abc123",
		UserType: "invalid",
		Role:     string(UserRolePublic),
	}

	user, err := service.Register(dto)
	assert.Error(t, err)
	assert.Equal(t, ErrInvalidUserType, err)
	assert.Nil(t, user)
}

func TestService_Register_InvalidRole(t *testing.T) {
	mockRepo := &MockRepository{
		usernameExistsFunc: func(username string) (bool, error) {
			return false, nil
		},
	}

	service := NewService(mockRepo)

	dto := RegisterUserDTO{
		Username: "testuser",
		Phone:    "13800138000",
		Password: "abc123",
		UserType: string(UserTypeDelivery),
		Role:     "invalid",
	}

	user, err := service.Register(dto)
	assert.Error(t, err)
	assert.Equal(t, ErrInvalidRole, err)
	assert.Nil(t, user)
}

// ==================== SimpleRegister Tests ====================

func TestService_SimpleRegister(t *testing.T) {
	mockRepo := &MockRepository{
		usernameExistsFunc: func(username string) (bool, error) {
			return false, nil
		},
		createFunc: func(user *User) error {
			return nil
		},
	}

	service := NewService(mockRepo)

	dto := SimpleRegisterDTO{
		Username: "testuser",
		Password: "abc123",
	}

	user, err := service.SimpleRegister(dto)
	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, "testuser", user.Name())
	assert.Equal(t, UserTypeDelivery, user.UserType())
	assert.Equal(t, UserRolePublic, user.Role())
}

func TestService_SimpleRegister_UsernameExists(t *testing.T) {
	mockRepo := &MockRepository{
		usernameExistsFunc: func(username string) (bool, error) {
			return true, nil
		},
	}

	service := NewService(mockRepo)

	dto := SimpleRegisterDTO{
		Username: "existinguser",
		Password: "abc123",
	}

	user, err := service.SimpleRegister(dto)
	assert.Error(t, err)
	assert.Equal(t, ErrUsernameAlreadyExists, err)
	assert.Nil(t, user)
}

func TestService_SimpleRegister_InvalidPassword(t *testing.T) {
	mockRepo := &MockRepository{
		usernameExistsFunc: func(username string) (bool, error) {
			return false, nil
		},
	}

	service := NewService(mockRepo)

	dto := SimpleRegisterDTO{
		Username: "testuser",
		Password: "ab12", // too short
	}

	user, err := service.SimpleRegister(dto)
	assert.Error(t, err)
	assert.Equal(t, ErrInvalidPassword, err)
	assert.Nil(t, user)
}

// ==================== RegisterWithPasswordValidation Tests ====================

func TestService_RegisterWithPasswordValidation(t *testing.T) {
	mockRepo := &MockRepository{
		usernameExistsFunc: func(username string) (bool, error) {
			return false, nil
		},
		createFunc: func(user *User) error {
			return nil
		},
	}

	service := NewService(mockRepo)

	dto := RegisterWithPasswordValidationDTO{
		Username: "testuser",
		Password: "abc123",
	}

	user, err := service.RegisterWithPasswordValidation(dto)
	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, "testuser", user.Name())
}

func TestService_RegisterWithPasswordValidation_InvalidPassword(t *testing.T) {
	mockRepo := &MockRepository{
		usernameExistsFunc: func(username string) (bool, error) {
			return false, nil
		},
	}

	service := NewService(mockRepo)

	dto := RegisterWithPasswordValidationDTO{
		Username: "testuser",
		Password: "ab12", // too short
	}

	user, err := service.RegisterWithPasswordValidation(dto)
	assert.Error(t, err)
	assert.Equal(t, ErrInvalidPassword, err)
	assert.Nil(t, user)
}

// ==================== UpdatePhone Tests ====================

func TestService_UpdatePhone(t *testing.T) {
	testUser, _ := NewUser("testuser", "13800138000", "abc123", UserTypeDelivery, UserRolePublic)

	mockRepo := &MockRepository{
		getByIDFunc: func(id UserID) (*User, error) {
			return testUser, nil
		},
		phoneExistsFunc: func(phone string) (bool, error) {
			return false, nil
		},
		updateFunc: func(user *User) error {
			return nil
		},
	}

	service := NewService(mockRepo)

	err := service.UpdatePhone(testUser.ID(), "13900139000")
	assert.NoError(t, err)
	assert.Equal(t, "13900139000", testUser.Phone())
}

func TestService_UpdatePhone_SamePhone(t *testing.T) {
	testUser, _ := NewUser("testuser", "13800138000", "abc123", UserTypeDelivery, UserRolePublic)

	mockRepo := &MockRepository{
		getByIDFunc: func(id UserID) (*User, error) {
			return testUser, nil
		},
	}

	service := NewService(mockRepo)

	err := service.UpdatePhone(testUser.ID(), "13800138000")
	assert.NoError(t, err)
}

func TestService_UpdatePhone_PhoneExists(t *testing.T) {
	testUser, _ := NewUser("testuser", "13800138000", "abc123", UserTypeDelivery, UserRolePublic)

	mockRepo := &MockRepository{
		getByIDFunc: func(id UserID) (*User, error) {
			return testUser, nil
		},
		phoneExistsFunc: func(phone string) (bool, error) {
			return true, nil
		},
	}

	service := NewService(mockRepo)

	err := service.UpdatePhone(testUser.ID(), "13900139000")
	assert.Error(t, err)
	assert.Equal(t, ErrPhoneAlreadyExists, err)
}

// ==================== UpdatePassword Tests ====================

func TestService_UpdatePassword(t *testing.T) {
	testUser, _ := NewUser("testuser", "13800138000", "abc123", UserTypeDelivery, UserRolePublic)

	mockRepo := &MockRepository{
		getByIDFunc: func(id UserID) (*User, error) {
			return testUser, nil
		},
		updateFunc: func(user *User) error {
			return nil
		},
	}

	service := NewService(mockRepo)

	err := service.UpdatePassword(testUser.ID(), "newpass123")
	assert.NoError(t, err)
}

func TestService_UpdatePassword_InvalidPassword(t *testing.T) {
	testUser, _ := NewUser("testuser", "13800138000", "abc123", UserTypeDelivery, UserRolePublic)

	mockRepo := &MockRepository{
		getByIDFunc: func(id UserID) (*User, error) {
			return testUser, nil
		},
	}

	service := NewService(mockRepo)

	err := service.UpdatePassword(testUser.ID(), "ab12") // too short
	assert.Error(t, err)
	assert.Equal(t, ErrInvalidPassword, err)
}

// ==================== DeleteUser Tests ====================

func TestService_DeleteUser(t *testing.T) {
	testUser, _ := NewUser("testuser", "13800138000", "abc123", UserTypeDelivery, UserRolePublic)

	mockRepo := &MockRepository{
		getByIDFunc: func(id UserID) (*User, error) {
			return testUser, nil
		},
		deleteFunc: func(user *User) error {
			return nil
		},
	}

	service := NewService(mockRepo)

	err := service.DeleteUser(testUser.ID())
	assert.NoError(t, err)
}

func TestService_DeleteUser_UserNotFound(t *testing.T) {
	mockRepo := &MockRepository{
		getByIDFunc: func(id UserID) (*User, error) {
			return nil, errors.New("not found")
		},
	}

	service := NewService(mockRepo)

	err := service.DeleteUser(NewUserID())
	assert.Error(t, err)
}

// ==================== GetByID and GetByUsername Tests ====================

func TestService_GetByID(t *testing.T) {
	testUser, _ := NewUser("testuser", "13800138000", "abc123", UserTypeDelivery, UserRolePublic)

	mockRepo := &MockRepository{
		getByIDFunc: func(id UserID) (*User, error) {
			return testUser, nil
		},
	}

	service := NewService(mockRepo)

	user, err := service.GetByID(testUser.ID())
	assert.NoError(t, err)
	assert.Equal(t, testUser.ID(), user.ID())
}

func TestService_GetByID_NotFound(t *testing.T) {
	mockRepo := &MockRepository{
		getByIDFunc: func(id UserID) (*User, error) {
			return nil, errors.New("not found")
		},
	}

	service := NewService(mockRepo)

	user, err := service.GetByID(NewUserID())
	assert.Error(t, err)
	assert.Nil(t, user)
}

func TestService_GetByUsername(t *testing.T) {
	testUser, _ := NewUser("testuser", "13800138000", "abc123", UserTypeDelivery, UserRolePublic)

	mockRepo := &MockRepository{
		getByUsernameFunc: func(username string) (*User, error) {
			return testUser, nil
		},
	}

	service := NewService(mockRepo)

	user, err := service.GetByUsername("testuser")
	assert.NoError(t, err)
	assert.Equal(t, "testuser", user.Name())
}

func TestService_GetByUsername_NotFound(t *testing.T) {
	mockRepo := &MockRepository{
		getByUsernameFunc: func(username string) (*User, error) {
			return nil, errors.New("not found")
		},
	}

	service := NewService(mockRepo)

	user, err := service.GetByUsername("nonexistent")
	assert.Error(t, err)
	assert.Nil(t, user)
}
