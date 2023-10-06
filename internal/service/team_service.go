package service

import (
	"backend-v2/internal/repository"
	"backend-v2/model"
	"backend-v2/pkg/utils"
)

type teamService struct {
	teamRepo repository.ITeamRepository
}

func InitTeamService(teamRepo repository.ITeamRepository) ITeamService {
	return &teamService{
		teamRepo: teamRepo,
	}
}

type ITeamService interface {
	GetAll() ([]model.Team, error)
	GetPaginated(page, limit int, data model.Team) ([]model.Team, error)
	GetByID(id uint) (model.Team, error)
	Create(data model.Team) (model.Team, error)
	Update(data model.Team) (model.Team, error)
	Delete(id uint) error
	Count() (int64, error)
}

func (service *teamService) GetAll() ([]model.Team, error) {
	return service.teamRepo.GetAll()
}

func (service *teamService) GetPaginated(page, limit int, data model.Team) ([]model.Team, error) {
	if !(utils.IsEmptyFields(data)) {
		return service.teamRepo.GetPaginatedFiltered(page, limit, data)
	}

	return service.teamRepo.GetPaginated(page, limit)
}

func (service *teamService) GetByID(id uint) (model.Team, error) {
	return service.teamRepo.GetByID(id)
}

func (service *teamService) Create(data model.Team) (model.Team, error) {
	return service.teamRepo.Create(data)
}

func (service *teamService) Update(data model.Team) (model.Team, error) {
	return service.teamRepo.Update(data)
}

func (service *teamService) Delete(id uint) error {
	return service.teamRepo.Delete(id)
}

func (service *teamService) Count() (int64, error) {
	return service.teamRepo.Count()
}
