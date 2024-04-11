package repository

import (
	"backend-v2/model"

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

type IMaterialDefectRepository interface{
  GetByMaterialLocationID(materialLocationID uint) (model.MaterialDefect, error)
}

func(repo *materialDefectRepository) GetByMaterialLocationID(materialLocationID uint) (model.MaterialDefect, error) {
  var data model.MaterialDefect
  err := repo.
    db.
    Raw("SELECT * FROM material_defects WHERE material_location_id = ?", materialLocationID).
    Scan(&data).
    Error

  return data, err
}
