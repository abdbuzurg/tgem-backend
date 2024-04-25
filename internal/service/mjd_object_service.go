package service

import (
	"backend-v2/internal/dto"
	"backend-v2/internal/repository"
	"backend-v2/model"
)

type mjdObjectService struct {
	mjdObjectRepo repository.IMJDObjectRepository
}

func InitMJDObjectService(mjdObjectRepo repository.IMJDObjectRepository) IMJDObjectService {
	return &mjdObjectService{
		mjdObjectRepo: mjdObjectRepo,
	}
}

type IMJDObjectService interface {
	GetPaginated(page, limit int, projectID uint) ([]dto.MJDObjectPaginated, error)
	Count(projectID uint) (int64, error)
	Create(data dto.MJDObjectCreate) (model.MJD_Object, error)
	Update(data dto.MJDObjectCreate) (model.MJD_Object, error)
	Delete(id, projectID uint) error
}

func (service *mjdObjectService) GetPaginated(page, limit int, projectID uint) ([]dto.MJDObjectPaginated, error) {

	data, err := service.mjdObjectRepo.GetPaginated(page, limit, projectID)
	if err != nil {
		return []dto.MJDObjectPaginated{}, err
	}

	result := []dto.MJDObjectPaginated{}
	latestEntry := dto.MJDObjectPaginated{}
	for index, object := range data {
		if index == 0 {
			latestEntry = dto.MJDObjectPaginated{
				ObjectID:         object.ObjectID,
				ObjectDetailedID: object.ObjectDetailedID,
				Name:             object.Name,
				Status:           object.Status,
				Model:            object.Model,
				AmountStores:     object.AmountStores,
				AmountEntrances:  object.AmountEntrances,
				HasBasement:      object.HasBasement,
				Supervisors:      []string{},
			}
		}

		if latestEntry.ObjectID == object.ObjectID {
			latestEntry.Supervisors = append(latestEntry.Supervisors, object.SupervisorName)
		} else {

			result = append(result, latestEntry)
			latestEntry = dto.MJDObjectPaginated{
				ObjectID:         object.ObjectID,
				ObjectDetailedID: object.ObjectDetailedID,
				Name:             object.Name,
				Status:           object.Status,
				Model:            object.Model,
				AmountStores:     object.AmountStores,
				AmountEntrances:  object.AmountEntrances,
				HasBasement:      object.HasBasement,
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

func (service *mjdObjectService) Count(projectID uint) (int64, error) {
	return service.mjdObjectRepo.Count(projectID)
}

func (service *mjdObjectService) Create(data dto.MJDObjectCreate) (model.MJD_Object, error) {
	return service.mjdObjectRepo.Create(data)
}

func (service *mjdObjectService) Update(data dto.MJDObjectCreate) (model.MJD_Object, error) {
	return service.mjdObjectRepo.Update(data)
}

func (service *mjdObjectService) Delete(id, projectID uint) error {
	return service.mjdObjectRepo.Delete(id, projectID)
}
