package repository

import (
	"backend-v2/model"

	"gorm.io/gorm"
)

type resourceRepositry struct {
	db *gorm.DB
}

func InitResourceRepository(db *gorm.DB) IResourceRepository {
	return &resourceRepositry{
		db: db,
	}
}

type IResourceRepository interface {
	GetAll() ([]model.Resource, error)
}

func (repo *resourceRepositry) GetAll() ([]model.Resource, error) {
	data := []model.Resource{}
	err := repo.db.Find(&data).Error
	return data, err
}
