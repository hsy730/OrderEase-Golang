package services

import (
	"errors"
	"orderease/application/dto"
	"orderease/domain/user"
	"orderease/domain/shared"
	"orderease/utils/log2"

	"gorm.io/gorm"
)

type UserService struct {
	userRepo user.UserRepository
	db       *gorm.DB
}

func NewUserService(
	userRepo user.UserRepository,
	db *gorm.DB,
) *UserService {
	return &UserService{
		userRepo: userRepo,
		db:       db,
	}
}

func (s *UserService) CreateUser(req *dto.CreateUserRequest) (*dto.UserResponse, error) {
	userEntity, err := user.NewUser(req.Name, req.Role, req.Type, req.Password)
	if err != nil {
		return nil, err
	}

	userEntity.Phone = req.Phone
	userEntity.Address = req.Address

	if err := s.userRepo.Save(userEntity); err != nil {
		log2.Errorf("保存用户失败: %v", err)
		return nil, errors.New("保存用户失败")
	}

	return s.toUserResponse(userEntity), nil
}

func (s *UserService) GetUser(id shared.ID) (*dto.UserResponse, error) {
	userEntity, err := s.userRepo.FindByID(id)
	if err != nil {
		return nil, err
	}

	return s.toUserResponse(userEntity), nil
}

func (s *UserService) GetUserByName(name string) (*dto.UserResponse, error) {
	userEntity, err := s.userRepo.FindByName(name)
	if err != nil {
		return nil, err
	}

	return s.toUserResponse(userEntity), nil
}

func (s *UserService) GetUsers(page, pageSize int) (*dto.UserListResponse, error) {
	users, total, err := s.userRepo.FindAll(page, pageSize)
	if err != nil {
		return nil, err
	}

	data := make([]dto.UserResponse, len(users))
	for i, userEntity := range users {
		data[i] = *s.toUserResponse(&userEntity)
	}

	return &dto.UserListResponse{
		Total:    total,
		Page:     page,
		PageSize: pageSize,
		Data:     data,
	}, nil
}

func (s *UserService) UpdateUser(id shared.ID, name, phone, address string) error {
	userEntity, err := s.userRepo.FindByID(id)
	if err != nil {
		return err
	}

	if err := userEntity.UpdateBasicInfo(name, phone, address); err != nil {
		return err
	}

	if err := s.userRepo.Update(userEntity); err != nil {
		return errors.New("更新用户失败")
	}

	return nil
}

func (s *UserService) DeleteUser(id shared.ID) error {
	if err := s.userRepo.Delete(id); err != nil {
		return errors.New("删除用户失败")
	}

	return nil
}

func (s *UserService) toUserResponse(userEntity *user.User) *dto.UserResponse {
	return &dto.UserResponse{
		ID:        userEntity.ID,
		Name:      userEntity.Name,
		Role:      userEntity.Role,
		Type:      userEntity.Type,
		Phone:     userEntity.Phone,
		Address:   userEntity.Address,
		CreatedAt: userEntity.CreatedAt,
		UpdatedAt: userEntity.UpdatedAt,
	}
}
