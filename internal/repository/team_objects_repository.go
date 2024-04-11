package repository

import (
	"backend-v2/model"

	"gorm.io/gorm"
)

type teamObjectsRepository struct {
	db *gorm.DB
}

func InitTeamObjectsRepository(db *gorm.DB) ITeamObjectsRepository {
	return &teamObjectsRepository{
		db: db,
	}
}

type ITeamObjectsRepository interface {
	GetByObjectID(objectID uint) ([]model.TeamObjects, error)
	GetByTeamID(teamID uint) ([]model.TeamObjects, error)
	CreateBatch(data []model.TeamObjects) ([]model.TeamObjects, error)
}

func (repo *teamObjectsRepository) GetByObjectID(objectID uint) ([]model.TeamObjects, error) {
	var data []model.TeamObjects
	err := repo.db.Find(&data, "object_id = ?", objectID).Error
	return data, err
}

func (repo *teamObjectsRepository) GetByTeamID(teamID uint) ([]model.TeamObjects, error) {
	var data []model.TeamObjects
	err := repo.db.Find(&data, "team_id = ?", teamID).Error
	return data, err
}

func (repo *teamObjectsRepository) CreateBatch(data []model.TeamObjects) ([]model.TeamObjects, error) {
	err := repo.db.CreateInBatches(&data, 10).Error
	return data, err
}
