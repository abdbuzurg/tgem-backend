package repository

import (
	"backend-v2/model"

	"gorm.io/gorm"
)

type materialForProjectRepository struct {
	db *gorm.DB
}

func InitMaterialForProjectRepository(db *gorm.DB) IMaterialForProjectRepositry {
	return &materialForProjectRepository{
		db: db,
	}
}

type IMaterialForProjectRepositry interface {
	GetAll() ([]model.MaterialForProject, error)
	GetByProjectID(projectID uint) ([]model.MaterialForProject, error)
	GetByMaterialID(materialID uint) ([]model.MaterialForProject, error)
	Create(data model.MaterialForProject) (model.MaterialForProject, error)
	Update(data model.MaterialForProject) (model.MaterialForProject, error)
	Delete(id uint) error
}

func (repo *materialForProjectRepository) GetAll() ([]model.MaterialForProject, error) {
	data := []model.MaterialForProject{}
	err := repo.db.Find(&data).Error
	return data, err
}

func (repo *materialForProjectRepository) GetByProjectID(projectID uint) ([]model.MaterialForProject, error) {
	data := []model.MaterialForProject{}
	err := repo.db.Find(&data, "project_id = ?", projectID).Error
	return data, err
}

func (repo *materialForProjectRepository) GetByMaterialID(materialID uint) ([]model.MaterialForProject, error) {
	data := []model.MaterialForProject{}
	err := repo.db.Find(&data, "material_id = ?", materialID).Error
	return data, err
}

func (repo *materialForProjectRepository) Create(data model.MaterialForProject) (model.MaterialForProject, error) {
	err := repo.db.Create(&data).Error
	return data, err
}

func (repo *materialForProjectRepository) Update(data model.MaterialForProject) (model.MaterialForProject, error) {
	err := repo.db.Model(&model.MaterialForProject{}).Select("*").Updates(&data).Error
	return data, err
}

func (repo *materialForProjectRepository) Delete(id uint) error {
	return repo.db.Delete(&model.MaterialForProject{}, "id = ?", id).Error
}
