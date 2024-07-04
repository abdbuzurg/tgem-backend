package service

import (
	"backend-v2/internal/dto"
	"backend-v2/internal/repository"
	"backend-v2/model"
)

type objectService struct {
	objectRepo            repository.IObjectRepository
	objectSupervisorsRepo repository.IObjectSupervisorsRepository
	kl04kvObjectRepo      repository.IKL04KVObjectRepository
	mjdObjectRepo         repository.IMJDObjectRepository
	sipObjectRepo         repository.ISIPObjectRepository
	stvtObjectRepo        repository.ISTVTObjectRepository
	tpObjectRepo          repository.ITPObjectRepository
	objectTeamsRepo       repository.IObjectTeamsRepository
}

func InitObjectService(
	objectRepo repository.IObjectRepository,
	objectSupervisorsRepo repository.IObjectSupervisorsRepository,
	kl04kvObjectRepo repository.IKL04KVObjectRepository,
	mjdObjectRepo repository.IMJDObjectRepository,
	sipObjectRepo repository.ISIPObjectRepository,
	stvtObjectRepo repository.ISTVTObjectRepository,
	tpObjectRepo repository.ITPObjectRepository,
	objectTeamsRepo repository.IObjectTeamsRepository,
) IObjectService {
	return &objectService{
		objectRepo:            objectRepo,
		objectSupervisorsRepo: objectSupervisorsRepo,
		kl04kvObjectRepo:      kl04kvObjectRepo,
		mjdObjectRepo:         mjdObjectRepo,
		sipObjectRepo:         sipObjectRepo,
		stvtObjectRepo:        stvtObjectRepo,
		tpObjectRepo:          tpObjectRepo,
		objectTeamsRepo:       objectTeamsRepo,
	}
}

type IObjectService interface {
	GetAll(projectID uint) ([]model.Object, error)
	GetPaginated(page, limit int, data model.Object) ([]dto.ObjectPaginated, error)
	GetByID(id uint) (model.Object, error)
	Create(data dto.ObjectCreate) (model.Object, error)
	Update(data model.Object) (model.Object, error)
	Delete(id uint) error
	Count() (int64, error)
	GetTeamsByObjectID(objectID uint) ([]dto.TeamDataForSelect, error)
}

func (service *objectService) GetAll(projectID uint) ([]model.Object, error) {
	return service.objectRepo.GetAll(projectID)
}

func (service *objectService) GetPaginated(page, limit int, filter model.Object) ([]dto.ObjectPaginated, error) {

	data, err := service.objectRepo.GetPaginatedFiltered(page, limit, filter)
	if err != nil {
		return []dto.ObjectPaginated{}, err
	}

	result := []dto.ObjectPaginated{}
	latestEntry := dto.ObjectPaginated{}
	for index, object := range data {
		if index == 0 {
			latestEntry = dto.ObjectPaginated{
				ID:          object.ID,
				Type:        object.ObjectType,
				Name:        object.ObjectName,
				Status:      object.ObjectStatus,
				Supervisors: []string{},
			}
		}

		if latestEntry.ID == object.ID {
			latestEntry.Supervisors = append(latestEntry.Supervisors, object.SupervisorName)
		} else {

			result = append(result, latestEntry)
			latestEntry = dto.ObjectPaginated{
				ID:     object.ID,
				Type:   object.ObjectType,
				Name:   object.ObjectName,
				Status: object.ObjectStatus,
				Supervisors: []string{
					object.SupervisorName,
				},
			}

		}
	}

	if len(data) != 0 {
		result = append(result, latestEntry)
	}

	return result, nil

}

func (service *objectService) GetByID(id uint) (model.Object, error) {
	return service.objectRepo.GetByID(id)
}

func (service *objectService) Create(data dto.ObjectCreate) (model.Object, error) {

	object := model.Object{
		ID:               0,
		ObjectDetailedID: 0,
		Type:             "",
		Name:             data.Name,
		Status:           data.Status,
		ProjectID:        data.ProjectID,
	}

	object, err := service.objectRepo.Create(object)
	if err != nil {
		return model.Object{}, err
	}

	return object, nil
}

func (service *objectService) Update(data model.Object) (model.Object, error) {
	return service.objectRepo.Update(data)
}

func (service *objectService) Delete(id uint) error {
	return service.objectRepo.Delete(id)
}

func (service *objectService) Count() (int64, error) {
	return service.objectRepo.Count()
}

func(service *objectService) GetTeamsByObjectID(objectID uint) ([]dto.TeamDataForSelect, error) {
  return service.objectTeamsRepo.GetTeamsByObjectID(objectID)
}
