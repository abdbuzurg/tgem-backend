package service

import (
	"backend-v2/internal/dto"
	"backend-v2/internal/repository"
	"backend-v2/model"
	"backend-v2/pkg/jwt"
	"backend-v2/pkg/security"
	"backend-v2/pkg/utils"
	"fmt"
)

type userService struct {
	userRepo repository.IUserRepository
}

func InitUserService(userRepo repository.IUserRepository) IUserService {
	return &userService{
		userRepo: userRepo,
	}
}

type IUserService interface {
	GetAll() ([]model.User, error)
	GetPaginated(page, limit int, data model.User) ([]model.User, error)
	GetByID(id uint) (model.User, error)
	Create(data model.User) (model.User, error)
	Update(data model.User) (model.User, error)
	Delete(id uint) error
	Count() (int64, error)
	Login(data dto.LoginData) (string, error)
	GetPermissions(username string) ([]dto.UserPermission, error)
}

func (service *userService) GetAll() ([]model.User, error) {
	return service.userRepo.GetAll()
}

func (service *userService) GetPaginated(page, limit int, data model.User) ([]model.User, error) {
	if !(utils.IsEmptyFields(data)) {
		return service.userRepo.GetPaginatedFiltered(page, limit, data)
	}

	return service.userRepo.GetPaginated(page, limit)
}

func (service *userService) GetByID(id uint) (model.User, error) {
	return service.userRepo.GetByID(id)
}

func (service *userService) Create(data model.User) (model.User, error) {
	return service.userRepo.Create(data)
}

func (service *userService) Update(data model.User) (model.User, error) {
	return service.userRepo.Update(data)
}

func (service *userService) Delete(id uint) error {
	return service.userRepo.Delete(id)
}

func (service *userService) Count() (int64, error) {
	return service.userRepo.Count()
}

func (service *userService) Login(data dto.LoginData) (string, error) {
	user, err := service.userRepo.GetByUsername(data.Username)
	if err != nil {
		return "", err
	}

	err = security.VerifyPassword(user.Password, data.Password)
	if err != nil {
		return "", fmt.Errorf("incorrect password")
	}

	token, err := jwt.CreateToken(user.Username)
	if err != nil {
		return "", err
	}

	return token, nil
}

func (service *userService) GetPermissions(username string) ([]dto.UserPermission, error) {
	user, err := service.userRepo.GetByUsername(username)
	if err != nil {
		return []dto.UserPermission{}, err
	}

	permissionsRaw, err := service.userRepo.GetPermissions(user.ID)
	if err != nil {
		return []dto.UserPermission{}, err
	}

	data := []dto.UserPermission{}
	for _, permission := range permissionsRaw {
		data = append(data, dto.UserPermission{
			ResourceName:   permission.V1,
			ResourceAction: permission.V2,
		})
	}

	return data, nil
}
