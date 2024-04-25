package service

import (
	"backend-v2/internal/dto"
	"backend-v2/internal/repository"
	"backend-v2/model"
	"errors"
	"fmt"

	"gorm.io/gorm"
)

type permissionService struct {
	permissionRepo repository.IPermissionRepository
	roleRepo       repository.IRoleRepository
	resourceRepo   repository.IResourceRepository
}

func InitPermissionService(
	permissionRepo repository.IPermissionRepository,
	roleRepo repository.IRoleRepository,
	resourceRepo repository.IResourceRepository,
) IPermissionService {
	return &permissionService{
		permissionRepo: permissionRepo,
		roleRepo:       roleRepo,
		resourceRepo:   resourceRepo,
	}
}

type IPermissionService interface {
	GetAll() ([]model.Permission, error)
	GetByRoleName(roleName string) ([]dto.UserPermission, error)
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
	return service.permissionRepo.CreateBatch(data)
}

func (service *permissionService) GetByRoleName(roleName string) ([]dto.UserPermission, error) {
	return service.permissionRepo.GetByRoleName(roleName)
}

func (service *permissionService) GetByResourceURL(resourceURL string, roleID uint) error {

	permission, err := service.permissionRepo.GetByResourceURL(resourceURL, roleID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil
	} else if err != nil {
		return err
	}

	if permission.ID == 0 {
		return nil
	}

	if !permission.R && !permission.W && !permission.U && !permission.D {
		return fmt.Errorf("Доступ запрещен")
	}

	return nil
}
