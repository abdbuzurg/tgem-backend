package service

import (
	"backend-v2/internal/repository"
	"backend-v2/model"
)

type roleService struct {
	roleRepo repository.IRoleRepository 
}

func InitRoleService(
  roleRepo repository.IRoleRepository,
) IRoleService {
  return &roleService{
    roleRepo: roleRepo,
  }
}

type IRoleService interface {
  GetAll() ([]model.Role, error)
  Create(data model.Role) (model.Role, error)
  Update(data model.Role) (model.Role, error)
  Delete(id uint) error
}

func(service *roleService) GetAll() ([]model.Role, error) {
  return service.roleRepo.GetAll()
} 

func(service *roleService) Create(data model.Role) (model.Role, error) {
  return service.roleRepo.Create(data)
}

func(service *roleService) Update(data model.Role) (model.Role, error) {
  return service.roleRepo.Update(data)
}

func(service *roleService) Delete(id uint) error {
  return service.roleRepo.Delete(id)
}
