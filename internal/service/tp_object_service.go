package service

import (
	"backend-v2/internal/dto"
	"backend-v2/internal/repository"
	"backend-v2/model"
)

type tpObjectService struct {
	tpObjectRepo repository.ITPObjectRepository
}

func InitTPObjectService(tpObjectRepo repository.ITPObjectRepository) ITPObjectService {
	return &tpObjectService{
		tpObjectRepo: tpObjectRepo,
	}
}

type ITPObjectService interface {
	GetPaginated(page, limit int, projectID uint) ([]dto.TPObjectPaginated, error)
	Count(projectID uint) (int64, error)
	Create(data dto.TPObjectCreate) (model.TP_Object, error)
	Update(data dto.TPObjectCreate) (model.TP_Object, error)
	Delete(id, projectID uint) error
}

func (service *tpObjectService) GetPaginated(page, limit int, projectID uint) ([]dto.TPObjectPaginated, error) {

	data, err := service.tpObjectRepo.GetPaginated(page, limit, projectID)
	if err != nil {
		return []dto.TPObjectPaginated{}, err
	}

	result := []dto.TPObjectPaginated{}
	latestEntry := dto.TPObjectPaginated{}
	for index, object := range data {
		if index == 0 {
			latestEntry = dto.TPObjectPaginated{
				ObjectID:         object.ObjectID,
				ObjectDetailedID: object.ObjectDetailedID,
				Name:             object.Name,
				Status:           object.Status,
				Model:            object.Model,
				VoltageClass:     object.VoltageClass,
				Nourashes:        object.Nourashes,
				Supervisors:      []string{},
			}
		}

		if latestEntry.ObjectID == object.ObjectID {
			latestEntry.Supervisors = append(latestEntry.Supervisors, object.SupervisorName)
		} else {

			result = append(result, latestEntry)
			latestEntry = dto.TPObjectPaginated{
				ObjectID:         object.ObjectID,
				ObjectDetailedID: object.ObjectDetailedID,
				Name:             object.Name,
				Status:           object.Status,
				Model:            object.Model,
				VoltageClass:     object.VoltageClass,
				Nourashes:        object.Nourashes,
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

func (service *tpObjectService) Count(projectID uint) (int64, error) {
	return service.tpObjectRepo.Count(projectID)
}

func (service *tpObjectService) Create(data dto.TPObjectCreate) (model.TP_Object, error) {
	return service.tpObjectRepo.Create(data)
}

func (service *tpObjectService) Update(data dto.TPObjectCreate) (model.TP_Object, error) {
	return service.tpObjectRepo.Update(data)
}

func (service *tpObjectService) Delete(id, projectID uint) error {
	return service.tpObjectRepo.Delete(id, projectID)
}
