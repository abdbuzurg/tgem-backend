package service

import (
	"backend-v2/internal/repository"
	"backend-v2/model"
)

type permissionService struct {
	permissionRepo repository.IPermissionRepository
}

func InitPermissionService(
	permissionRepo repository.IPermissionRepository,
) IPermissionService {
  return &permissionService{
    permissionRepo: permissionRepo,
  }
}

type IPermissionService interface{
  GetAll() ([]model.Permission, error)
  GetByRoleID(roleID uint) ([]model.Permission, error)
  Create(data model.Permission) (model.Permission, error)
  Update(data model.Permission) (model.Permission, error)
  Delete(id uint) error
}

func(service *permissionService) GetAll() ([]model.Permission, error) {
  return service.permissionRepo.GetAll()
}

func(service *permissionService) GetByRoleID(roleID uint) ([]model.Permission, error) {
  return service.permissionRepo.GetByRoleID(roleID)
}

func(service *permissionService) Create(data model.Permission) (model.Permission, error) {
  return service.permissionRepo.Create(data)
}

func(service *permissionService) Update(data model.Permission) (model.Permission, error) {
  return service.permissionRepo.Update(data)
}

func(service *permissionService) Delete(id uint) error {
  return service.permissionRepo.Delete(id)
}
