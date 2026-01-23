package user

// Repository 用户仓储接口（领域层定义接口）
type Repository interface {
	Create(user *User) error
	GetByID(id UserID) (*User, error)
	GetByUsername(username string) (*User, error)
	PhoneExists(phone string) (bool, error)
	UsernameExists(username string) (bool, error)
	Update(user *User) error
	Delete(user *User) error
}

// RepositoryAdapter 仓储适配器（将现有的 UserRepository 适配到领域接口）
// 注意：这里使用类型别名避免循环引用
type UserRepositoryAdapter struct {
	// 这里需要注入 repositories.UserRepository
	// 由于跨包引用，将在 service 层通过依赖注入处理
	createFunc    func(*User) error
	getByIDFunc   func(UserID) (*User, error)
	getByUsernameFunc func(string) (*User, error)
	phoneExistsFunc   func(string) (bool, error)
	usernameExistsFunc func(string) (bool, error)
	updateFunc    func(*User) error
	deleteFunc    func(*User) error
}

// NewRepositoryAdapter 创建仓储适配器
func NewRepositoryAdapter(
	createFunc func(*User) error,
	getByIDFunc func(UserID) (*User, error),
	getByUsernameFunc func(string) (*User, error),
	phoneExistsFunc func(string) (bool, error),
	usernameExistsFunc func(string) (bool, error),
	updateFunc func(*User) error,
	deleteFunc func(*User) error,
) *UserRepositoryAdapter {
	return &UserRepositoryAdapter{
		createFunc:    createFunc,
		getByIDFunc:   getByIDFunc,
		getByUsernameFunc: getByUsernameFunc,
		phoneExistsFunc:   phoneExistsFunc,
		usernameExistsFunc: usernameExistsFunc,
		updateFunc:    updateFunc,
		deleteFunc:    deleteFunc,
	}
}

// Create 创建用户
func (a *UserRepositoryAdapter) Create(user *User) error {
	return a.createFunc(user)
}

// GetByID 根据ID获取用户
func (a *UserRepositoryAdapter) GetByID(id UserID) (*User, error) {
	return a.getByIDFunc(id)
}

// GetByUsername 根据用户名获取用户
func (a *UserRepositoryAdapter) GetByUsername(username string) (*User, error) {
	return a.getByUsernameFunc(username)
}

// PhoneExists 检查手机号是否存在
func (a *UserRepositoryAdapter) PhoneExists(phone string) (bool, error) {
	return a.phoneExistsFunc(phone)
}

// UsernameExists 检查用户名是否存在
func (a *UserRepositoryAdapter) UsernameExists(username string) (bool, error) {
	return a.usernameExistsFunc(username)
}

// Update 更新用户
func (a *UserRepositoryAdapter) Update(user *User) error {
	return a.updateFunc(user)
}

// Delete 删除用户
func (a *UserRepositoryAdapter) Delete(user *User) error {
	return a.deleteFunc(user)
}
