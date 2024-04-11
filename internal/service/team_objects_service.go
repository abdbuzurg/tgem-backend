package service

import (
	"backend-v2/internal/repository"
	"backend-v2/model"
)

type teamObjectsService struct {
	teamObjectsRepo repository.ITeamObjectsRepository
}

func InitTeamObjectsService(teamObjectsRepo repository.ITeamObjectsRepository) ITeamObjectsService {
	return &teamObjectsService{
		teamObjectsRepo: teamObjectsRepo,
	}
}

type ITeamObjectsService interface {
	GetByObjectID(objectID uint) ([]model.TeamObjects, error)
	GetByTeamID(teamID uint) ([]model.TeamObjects, error)
	CreateBatch(data []model.TeamObjects) ([]model.TeamObjects, error)
}

func (service *teamObjectsService) GetByObjectID(objectID uint) ([]model.TeamObjects, error) {
	return service.teamObjectsRepo.GetByObjectID(objectID)
}

func (service *teamObjectsService) GetByTeamID(teamID uint) ([]model.TeamObjects, error) {
	return service.teamObjectsRepo.GetByTeamID(teamID)
}

func (service *teamObjectsService) CreateBatch(data []model.TeamObjects) ([]model.TeamObjects, error) {
	return service.teamObjectsRepo.CreateBatch(data)
}
