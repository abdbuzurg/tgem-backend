package repository

import (
	"backend-v2/model"

	"gorm.io/gorm"
)

type materialRepository struct {
	db *gorm.DB
}

func InitMaterialRepository(db *gorm.DB) IMaterialRepository {
	return &materialRepository{
		db: db,
	}
}

type IMaterialRepository interface {
	GetAll(projectID uint) ([]model.Material, error)
	GetPaginated(page, limit int) ([]model.Material, error)
	GetPaginatedFiltered(page, limit int, filter model.Material) ([]model.Material, error)
	GetByID(id uint) (model.Material, error)
	Create(data model.Material) (model.Material, error)
  CreateInBatches(data []model.Material) ([]model.Material, error)
	Update(data model.Material) (model.Material, error)
	Delete(id uint) error
	Count() (int64, error)
}

func (repo *materialRepository) GetAll(projectID uint) ([]model.Material, error) {
	data := []model.Material{}
	err := repo.db.Order("id desc").Find(&data, "project_id = ?", projectID).Error
	return data, err
}

func (repo *materialRepository) GetPaginated(page, limit int) ([]model.Material, error) {
	data := []model.Material{}
	err := repo.db.Order("id desc").Offset((page - 1) * limit).Limit(limit).Find(&data).Error
	return data, err
}

func (repo *materialRepository) GetPaginatedFiltered(page, limit int, filter model.Material) ([]model.Material, error) {
	data := []model.Material{}
	err := repo.db.
		Raw(`
    SELECT * 
    FROM materials 
    WHERE
      project_id = ? AND
			(nullif(?, '') IS NULL OR category = ?) AND
			(nullif(?, '') IS NULL OR code = ?) AND
			(nullif(?, '') IS NULL OR name = ?) AND
			(nullif(?, '') IS NULL OR unit = ?) 
    ORDER BY id DESC LIMIT ? OFFSET ?`,
    filter.ProjectID,
		filter.Category, filter.Category, 
    filter.Code, filter.Code, 
    filter.Name, filter.Name, 
    filter.Unit, filter.Unit, 
    limit, (page-1)*limit,
		).Scan(&data).Error

	return data, err
}

func (repo *materialRepository) GetByID(id uint) (model.Material, error) {
	data := model.Material{}
	err := repo.db.Find(&data, "id = ?", id).Error
	return data, err
}

func (repo *materialRepository) Create(data model.Material) (model.Material, error) {
	err := repo.db.Create(&data).Error
	return data, err
}

func(repo *materialRepository) CreateInBatches(data []model.Material) ([]model.Material, error) {
  err := repo.db.CreateInBatches(&data, 20).Error
  return data, err
}

func (repo *materialRepository) Update(data model.Material) (model.Material, error) {
	err := repo.db.Model(&model.Material{}).Select("*").Where("id = ?", data.ID).Updates(&data).Error
	return data, err
}

func (repo *materialRepository) Delete(id uint) error {
	return repo.db.Delete(&model.Material{}, "id = ?", id).Error
}

func (repo *materialRepository) Count() (int64, error) {
	var count int64
	err := repo.db.Model(&model.Material{}).Count(&count).Error
	return count, err
}
