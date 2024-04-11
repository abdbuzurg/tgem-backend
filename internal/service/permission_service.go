package service

import (
	"backend-v2/internal/repository"
	"backend-v2/model"
	"backend-v2/pkg/utils"
	"errors"
	"fmt"

	"gorm.io/gorm"
)

type permissionService struct {
	permissionRepo repository.IPermissionRepository
	roleRepo       repository.IRoleRepository
}

func InitPermissionService(
	permissionRepo repository.IPermissionRepository,
	roleRepo repository.IRoleRepository,
) IPermissionService {
	return &permissionService{
		permissionRepo: permissionRepo,
		roleRepo:       roleRepo,
	}
}

type IPermissionService interface {
	GetAll() ([]model.Permission, error)
  GetByRoleName(roleName string) ([]model.Permission, error)
	GetByRoleID(roleID uint) ([]model.Permission, error)
  GetByResourceURL(resourceURL string, roleID uint) error
	Create(data model.Permission) (model.Permission, error)
	CreateBatch(data []model.Permission) error
	Update(data model.Permission) (model.Permission, error)
	Delete(id uint) error
}

func (service *permissionService) GetAll() ([]model.Permission, error) {
	return service.permissionRepo.GetAll()
}

func (service *permissionService) GetByRoleID(roleID uint) ([]model.Permission, error) {
	return service.permissionRepo.GetByRoleID(roleID)
}

func (service *permissionService) Create(data model.Permission) (model.Permission, error) {
	return service.permissionRepo.Create(data)
}

func (service *permissionService) Update(data model.Permission) (model.Permission, error) {
	return service.permissionRepo.Update(data)
}

func (service *permissionService) Delete(id uint) error {
	return service.permissionRepo.Delete(id)
}

func (service *permissionService) CreateBatch(data []model.Permission) error {

	allPermissions, err := utils.AvailablePermissionList()
	if err != nil {
		return err
	}

	for _, defaultPermission := range allPermissions {
		existenceOfDefaultPermission := false
		for index, userPermission := range data {
			if defaultPermission.ResourceUrl == userPermission.ResourceUrl {
		    data[index].ResourceName = defaultPermission.ResourceName	
        existenceOfDefaultPermission = true
				break
			}
		}

		if !existenceOfDefaultPermission {
			data = append(data, defaultPermission)
		}
	}

	role, err := service.roleRepo.GetLast()
	if err != nil {
		return err
	}

	for index := range data {
		data[index].RoleID = role.ID
	}

	return service.permissionRepo.CreateBatch(data)
}

func(service *permissionService) GetByRoleName(roleName string) ([]model.Permission, error) {
  
  role, err := service.roleRepo.GetByName(roleName)
  if err != nil {
    return []model.Permission{}, err
  }

  permissions, err := service.permissionRepo.GetByRoleID(role.ID)
  if err != nil {
    return []model.Permission{}, err
  }

  return permissions, err 
}

func(service *permissionService) GetByResourceURL(resourceURL string, roleID uint) error {
  
  permission, err := service.permissionRepo.GetByResourceURL(resourceURL, roleID)
  if errors.Is(err, gorm.ErrRecordNotFound) {
    return nil
  } else if err != nil {
    return err
  }

  if (permission.ID == 0) {
    return nil
  }

  if (!permission.R && !permission.W && !permission.U && !permission.D) {
    return fmt.Errorf("Доступ запрещен")
  }

  return nil
}
