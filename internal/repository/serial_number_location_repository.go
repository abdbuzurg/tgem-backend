package repository

import (
	"backend-v2/model"

	"gorm.io/gorm"
)

type serialNumberLocationRepository struct {
	db *gorm.DB
}

func InitSerialNumberLocationRepository(db *gorm.DB) ISerialNumberLocationRepository {
	return &serialNumberLocationRepository{
		db: db,
	}
}

type ISerialNumberLocationRepository interface{}

func(repo *serialNumberLocationRepository) CreateInBatches(data []model.SerialNumberLocation) ([]model.SerialNumberLocation, error) {
  err := repo.db.CreateInBatches(&data, 15).Error
  return data, err
}
