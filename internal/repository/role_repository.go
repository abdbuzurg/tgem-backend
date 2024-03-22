package repository

import (
	"backend-v2/model"

	"gorm.io/gorm"
)

type roleRepository struct {
	db *gorm.DB
}

func InitRoleRepository(db *gorm.DB) IRoleRepository {
	return &roleRepository{
		db: db,
	}
}

type IRoleRepository interface{
  GetAll() ([]model.Role, error)
  Create(data model.Role) (model.Role, error)
  Update(data model.Role) (model.Role, error)
  Delete(id uint) error
}

func(repo *roleRepository) GetAll() ([]model.Role, error) {
  data := []model.Role{}
  err := repo.db.Find(&data).Error
  return data, err
}

func(repo *roleRepository) Create(data model.Role) (model.Role, error){
  err := repo.db.Create(&data).Error
  return data, err
} 

func(repo *roleRepository) Update(data model.Role) (model.Role, error){
  err := repo.db.Model(model.Role{}).Select("*").Where("id = ?", data.ID).Updates(&data).Error  
  return data, err
}

func(repo *roleRepository) Delete(id uint) error {
  err := repo.db.Delete(&model.Role{}, "id = ?", id).Error
  return err
}
