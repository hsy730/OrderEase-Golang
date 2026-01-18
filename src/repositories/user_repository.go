package repositories

import (
	"errors"
	"orderease/models"
	"orderease/utils/log2"

	"gorm.io/gorm"
)

// UserRepository 用户数据访问层
type UserRepository struct {
	DB *gorm.DB
}

// NewUserRepository 创建UserRepository实例
func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{DB: db}
}

// GetUserByID 根据ID获取用户
func (r *UserRepository) GetUserByID(id string) (*models.User, error) {
	var user models.User
	err := r.DB.First(&user, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.New("用户不存在")
	}
	if err != nil {
		log2.Errorf("GetUserByID failed: %v", err)
		return nil, errors.New("查询用户失败")
	}
	return &user, nil
}

// CheckUsernameExists 检查用户名是否存在
func (r *UserRepository) CheckUsernameExists(username string) (bool, error) {
	var count int64
	err := r.DB.Model(&models.User{}).Where("name = ?", username).Count(&count).Error
	if err != nil {
		log2.Errorf("CheckUsernameExists failed: %v", err)
		return false, errors.New("检查用户名失败")
	}
	return count > 0, nil
}

// GetUsers 获取用户列表（分页+搜索）
func (r *UserRepository) GetUsers(page, pageSize int, search string) ([]models.User, int64, error) {
	var users []models.User
	var total int64

	baseQuery := r.DB.Model(&models.User{})

	// 如果提供了用户名参数，则添加模糊匹配条件
	if search != "" {
		baseQuery = baseQuery.Where("name LIKE ?", "%"+search+"%")
	}

	// 获取总数
	if err := baseQuery.Count(&total).Error; err != nil {
		log2.Errorf("GetUsers count failed: %v", err)
		return nil, 0, errors.New("获取用户总数失败")
	}

	// 分页查询
	offset := (page - 1) * pageSize
	if err := baseQuery.Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&users).Error; err != nil {
		log2.Errorf("GetUsers query failed: %v", err)
		return nil, 0, errors.New("查询用户列表失败")
	}

	return users, total, nil
}

// GetUserSimpleList 获取简单用户列表（仅ID和名称，支持分页和搜索）
func (r *UserRepository) GetUserSimpleList(page, pageSize int, search string) ([]map[string]interface{}, int64, error) {
	var users []models.User
	var total int64

	baseQuery := r.DB.Model(&models.User{}).Select("id", "name")

	// 如果有搜索关键词，添加模糊搜索条件
	if search != "" {
		baseQuery = baseQuery.Where("name LIKE ?", "%"+search+"%")
	}

	// 获取总数
	if err := baseQuery.Count(&total).Error; err != nil {
		log2.Errorf("GetUserSimpleList count failed: %v", err)
		return nil, 0, errors.New("查询用户总数失败")
	}

	// 计算偏移量并查询
	offset := (page - 1) * pageSize
	if err := baseQuery.Offset(offset).Limit(pageSize).Find(&users).Error; err != nil {
		log2.Errorf("GetUserSimpleList query failed: %v", err)
		return nil, 0, errors.New("查询用户列表失败")
	}

	result := make([]map[string]interface{}, len(users))
	for i, user := range users {
		result[i] = map[string]interface{}{
			"id":   user.ID,
			"name": user.Name,
		}
	}
	return result, total, nil
}
