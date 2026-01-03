package repositories

import (
	"errors"
	"orderease/domain/shared"
	"orderease/domain/user"
	"orderease/infrastructure/persistence"
	"orderease/models"
	"orderease/utils/log2"

	"gorm.io/gorm"
)

type UserRepositoryImpl struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) user.UserRepository {
	return &UserRepositoryImpl{db: db}
}

func (r *UserRepositoryImpl) Save(u *user.User) error {
	model := persistence.UserToModel(u)
	if err := r.db.Create(model).Error; err != nil {
		log2.Errorf("保存用户失败: %v", err)
		return errors.New("保存用户失败")
	}
	u.ID = shared.ID(model.ID)
	return nil
}

func (r *UserRepositoryImpl) FindByID(id shared.ID) (*user.User, error) {
	var model models.User
	if err := r.db.First(&model, id.Value()).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("用户不存在")
		}
		log2.Errorf("查询用户失败: %v", err)
		return nil, errors.New("查询用户失败")
	}
	return persistence.UserToDomain(model), nil
}

func (r *UserRepositoryImpl) FindByName(name string) (*user.User, error) {
	var model models.User
	if err := r.db.Where("name = ?", name).First(&model).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("用户不存在")
		}
		log2.Errorf("查询用户失败: %v", err)
		return nil, errors.New("查询用户失败")
	}
	return persistence.UserToDomain(model), nil
}

func (r *UserRepositoryImpl) FindAll(page, pageSize int) ([]user.User, int64, error) {
	var total int64
	if err := r.db.Model(&models.User{}).Count(&total).Error; err != nil {
		log2.Errorf("查询用户总数失败: %v", err)
		return nil, 0, errors.New("查询用户总数失败")
	}

	var modelsList []models.User
	offset := (page - 1) * pageSize
	if err := r.db.Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&modelsList).Error; err != nil {
		log2.Errorf("查询用户列表失败: %v", err)
		return nil, 0, errors.New("查询用户列表失败")
	}

	users := make([]user.User, len(modelsList))
	for i, m := range modelsList {
		users[i] = *persistence.UserToDomain(m)
	}
	return users, total, nil
}

func (r *UserRepositoryImpl) Delete(id shared.ID) error {
	if err := r.db.Delete(&models.User{}, id.Value()).Error; err != nil {
		log2.Errorf("删除用户失败: %v", err)
		return errors.New("删除用户失败")
	}
	return nil
}

func (r *UserRepositoryImpl) Update(u *user.User) error {
	model := persistence.UserToModel(u)
	if err := r.db.Save(model).Error; err != nil {
		log2.Errorf("更新用户失败: %v", err)
		return errors.New("更新用户失败")
	}
	return nil
}

func (r *UserRepositoryImpl) Exists(id shared.ID) (bool, error) {
	var count int64
	if err := r.db.Model(&models.User{}).Where("id = ?", id.Value()).Count(&count).Error; err != nil {
		log2.Errorf("检查用户是否存在失败: %v", err)
		return false, errors.New("检查用户是否存在失败")
	}
	return count > 0, nil
}
