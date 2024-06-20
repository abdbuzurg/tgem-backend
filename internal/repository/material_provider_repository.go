package repository

import (
	"backend-v2/model"

	"gorm.io/gorm"
)

type materialProviderRepository struct {
	db *gorm.DB
}

func InitMaterialProviderRepository(db *gorm.DB) IMaterialProviderRepository {
	return &materialProviderRepository{
		db: db,
	}
}

type IMaterialProviderRepository interface {
	GetPaginated(page, limit int, projectID uint) ([]model.MaterialProvider, error)
	Count(projectID uint) (int64, error)
	Create(data model.MaterialProvider) (model.MaterialProvider, error)
	Update(data model.MaterialProvider) (model.MaterialProvider, error)
	Delete(id uint) error
}

func (repo *materialProviderRepository) GetPaginated(page, limit int, projectID uint) ([]model.MaterialProvider, error) {
	data := []model.MaterialProvider{}
	err := repo.db.Find(&data, "WHERE project_id = ? ORDER BY id DESC LIMIT ? OFFSET ?", projectID, limit, (page-1)*limit).Error
	return data, err
}

func (repo *materialProviderRepository) Count(projectID uint) (int64, error) {
	var count int64
	err := repo.db.Raw(`SELECT COUNT(material_providers.id) FROM material_providers`).Scan(&count).Error
	return count, err
}

func (repo *materialProviderRepository) Create(data model.MaterialProvider) (model.MaterialProvider, error) {
	err := repo.db.Create(&data).Error
	return data, err
}

func (repo *materialProviderRepository) Update(data model.MaterialProvider) (model.MaterialProvider, error) {
	err := repo.db.Model(&model.MaterialProvider{}).Where("id = ?", data.ID).Updates(&data).Error
	return data, err
}

func (repo *materialProviderRepository) Delete(id uint) error {
	return repo.db.Delete(&model.MaterialProvider{}, "id = ?", id).Error
}
