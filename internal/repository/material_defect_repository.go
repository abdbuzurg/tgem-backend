package repository

import (
	"backend-v2/model"
	"errors"

	"gorm.io/gorm"
)

type materialDefectRepository struct {
	db *gorm.DB
}

func InitMaterialDefectRepository(db *gorm.DB) IMaterialDefectRepository {
	return &materialDefectRepository{
		db: db,
	}
}

type IMaterialDefectRepository interface {
	GetByMaterialLocationID(materialLocationID uint) (model.MaterialDefect, error)
	Create(data model.MaterialDefect) (model.MaterialDefect, error)
	Update(data model.MaterialDefect) (model.MaterialDefect, error)
	FindOrCreate(materialLocationID uint) (model.MaterialDefect, error)
}

func (repo *materialDefectRepository) GetByMaterialLocationID(materialLocationID uint) (model.MaterialDefect, error) {
	var data model.MaterialDefect
	err := repo.
		db.
		Raw("SELECT * FROM material_defects WHERE material_location_id = ?", materialLocationID).
		Scan(&data).
		Error

  if errors.Is(err, gorm.ErrRecordNotFound) {
    return model.MaterialDefect{}, nil
  }

	return data, err
}

func (repo *materialDefectRepository) Create(data model.MaterialDefect) (model.MaterialDefect, error) {
	err := repo.db.Create(&data).Error
	return data, err
}

func (repo *materialDefectRepository) Update(data model.MaterialDefect) (model.MaterialDefect, error) {
	err := repo.db.Table("material_defects").Where("id = ?", data.ID).Updates(&data).Error
	return data, err
}

func (repo *materialDefectRepository) FindOrCreate(materialLocationID uint) (model.MaterialDefect, error) {
	var data model.MaterialDefect
	err := repo.db.FirstOrCreate(&data, model.MaterialDefect{MaterialLocationID: materialLocationID}).Error
	return data, err
}
