package repository

import (
	"backend-v2/model"

	"gorm.io/gorm"
)

type permissionRepository struct {
	db *gorm.DB
}

func InitPermissionRepository(db *gorm.DB) IPermissionRepository {
	return &permissionRepository{
		db: db,
	}
}

type IPermissionRepository interface {
	GetAll() ([]model.Permission, error)
	GetByRoleID(roleID uint) ([]model.Permission, error)
  GetByResourceURL(resourceURL string, roleID uint) (model.Permission, error)
	Create(data model.Permission) (model.Permission, error)
	CreateBatch(data []model.Permission) error
	Update(data model.Permission) (model.Permission, error)
	Delete(id uint) error
}

func (repo *permissionRepository) GetAll() ([]model.Permission, error) {
	var data []model.Permission
	err := repo.db.Find(&data).Error
	return data, err
}

func (repo *permissionRepository) GetByRoleID(roleID uint) ([]model.Permission, error) {
	var data []model.Permission
	err := repo.db.Find(&data, "role_id = ? AND (r OR w OR u OR d)", roleID).Error
	return data, err
}

func (repo *permissionRepository) Create(data model.Permission) (model.Permission, error) {
	err := repo.db.Create(&data).Error
	return data, err
}

func (repo *permissionRepository) Update(data model.Permission) (model.Permission, error) {
	err := repo.db.Model(model.Permission{}).Select("*").Where("id = ?", data.ID).Updates(&data).Error
	return data, err
}

func (repo *permissionRepository) Delete(id uint) error {
	err := repo.db.Delete(model.Permission{}, "id = ?", id).Error
	return err
}

func (repo *permissionRepository) CreateBatch(data []model.Permission) error {
  err := repo.db.CreateInBatches(&data, 15).Error
  return err
}

func(repo *permissionRepository) GetByResourceURL(resourceURL string, roleID uint) (model.Permission, error) {
  var data model.Permission
  resourceURL = "%" + resourceURL + "%"
  err := repo.db.
    Raw("SELECT * FROM permissions WHERE role_id = ? AND resource_url LIKE ?", roleID, resourceURL).
    Scan(&data).
    Error
  return data, err
}
