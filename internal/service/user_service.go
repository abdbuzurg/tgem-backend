package service

import (
	"backend-v2/internal/repository"
	"backend-v2/model"
	"backend-v2/pkg/utils"
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
