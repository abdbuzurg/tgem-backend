package service

import (
	"backend-v2/internal/dto"
	"backend-v2/internal/repository"
	"backend-v2/model"
	"fmt"
)

type teamService struct {
	teamRepo        repository.ITeamRepository
	teamObjectsRepo repository.ITeamObjectsRepository
}

func InitTeamService(
	teamRepo repository.ITeamRepository,
	teamObjectsRepo repository.ITeamObjectsRepository,
) ITeamService {
	return &teamService{
		teamRepo: teamRepo,
    teamObjectsRepo: teamObjectsRepo,
	}
}

type ITeamService interface {
	GetAll() ([]model.Team, error)
	GetPaginated(page, limit int, data model.Team) ([]dto.TeamPaginated, error)
	GetByID(id uint) (model.Team, error)
	Create(data dto.TeamMutation) (model.Team, error)
	Update(data model.Team) (model.Team, error)
	Delete(id uint) error
	Count() (int64, error)
}

func (service *teamService) GetAll() ([]model.Team, error) {
	return service.teamRepo.GetAll()
}

func (service *teamService) GetPaginated(page, limit int, data model.Team) ([]dto.TeamPaginated, error) {
  teamPaginatedQueryData, err := service.teamRepo.GetPaginatedFiltered(page, limit, data)
  if err != nil {
    return []dto.TeamPaginated{}, err
  }
  fmt.Println(teamPaginatedQueryData)

  result := []dto.TeamPaginated{}
  latestEntry := dto.TeamPaginated{}
	for index, team := range teamPaginatedQueryData {
    if index == 0 {
      latestEntry = dto.TeamPaginated{
        ID: team.ID,
        Number: team.TeamNumber,
        LeaderName: team.LeaderName,
        MobileNumber: team.TeamMobileNumber,
        Company: team.TeamCompany,
        Objects: []string{},
      }
    }

    if (latestEntry.ID == team.ID) {
      latestEntry.Objects = append(latestEntry.Objects, team.ObjectName)
    } else {
      result = append(result, latestEntry)
      latestEntry = dto.TeamPaginated{
        ID: team.ID,
        Number: team.TeamNumber,
        LeaderName: team.LeaderName,
        MobileNumber: team.TeamMobileNumber,
        Company: team.TeamCompany,
        Objects: []string{
          team.ObjectName,
        },
      }
    }
	}

  if len(teamPaginatedQueryData) > 0 {
    result = append(result, latestEntry)
  }

  fmt.Println(result)

  return result, nil
}

func (service *teamService) GetByID(id uint) (model.Team, error) {
	return service.teamRepo.GetByID(id)
}

func (service *teamService) Create(data dto.TeamMutation) (model.Team, error) {
	team, err := service.teamRepo.Create(model.Team{
		ID:             0,
		LeaderWorkerID: data.LeaderWorkerID,
		MobileNumber:   data.MobileNumber,
		Company:        data.Company,
		Number:         data.Number,
	})
	if err != nil {
		return model.Team{}, err
	}

	teamObjects := []model.TeamObjects{}
	for _, objectID := range data.Objects {
		teamObjects = append(teamObjects, model.TeamObjects{
			ObjectID: objectID,
			TeamID:   team.ID,
		})
	}

  _, err = service.teamObjectsRepo.CreateBatch(teamObjects)
  if err != nil {
    return model.Team{}, err
  }

	return team, nil
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
