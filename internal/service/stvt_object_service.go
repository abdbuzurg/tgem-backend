package service

import (
	"backend-v2/internal/dto"
	"backend-v2/internal/repository"
	"backend-v2/model"
)

type stvtObjectService struct {
	stvtObjectRepo repository.ISTVTObjectRepository
}

func InitSTVTObjectService(stvtObjectRepo repository.ISTVTObjectRepository) ISTVTObjectService {
	return &stvtObjectService{
		stvtObjectRepo: stvtObjectRepo,
	}
}

type ISTVTObjectService interface{
	GetPaginated(page, limit int, projectID uint) ([]dto.STVTObjectPaginated, error)
	Count(projectID uint) (int64, error)
	Create(data dto.STVTObjectCreate) (model.STVT_Object, error)
	Update(data dto.STVTObjectCreate) (model.STVT_Object, error)
	Delete(id, projectID uint) error
}

func (service *stvtObjectService) GetPaginated(page, limit int, projectID uint) ([]dto.STVTObjectPaginated, error) {

	data, err := service.stvtObjectRepo.GetPaginated(page, limit, projectID)
	if err != nil {
		return []dto.STVTObjectPaginated{}, err
	}

	result := []dto.STVTObjectPaginated{}
	latestEntry := dto.STVTObjectPaginated{}
	for index, object := range data {
		if index == 0 {
			latestEntry = dto.STVTObjectPaginated{
				ObjectID:         object.ObjectID,
				ObjectDetailedID: object.ObjectDetailedID,
				Name:             object.Name,
				Status:           object.Status,
				VoltageClass:     object.VoltageClass,
				TTCoefficient:    object.TTCoefficient,
				Supervisors:      []string{},
			}
		}

		if latestEntry.ObjectID == object.ObjectID {
			latestEntry.Supervisors = append(latestEntry.Supervisors, object.SupervisorName)
		} else {

			result = append(result, latestEntry)
			latestEntry = dto.STVTObjectPaginated{
				ObjectID:         object.ObjectID,
				ObjectDetailedID: object.ObjectDetailedID,
				Name:             object.Name,
				Status:           object.Status,
				VoltageClass:     object.VoltageClass,
				TTCoefficient:    object.TTCoefficient,
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

func (service *stvtObjectService) Count(projectID uint) (int64, error) {
	return service.stvtObjectRepo.Count(projectID)
}

func (service *stvtObjectService) Create(data dto.STVTObjectCreate) (model.STVT_Object, error) {
	return service.stvtObjectRepo.Create(data)
}

func (service *stvtObjectService) Update(data dto.STVTObjectCreate) (model.STVT_Object, error) {
	return service.stvtObjectRepo.Update(data)
}

func (service *stvtObjectService) Delete(id, projectID uint) error {
	return service.stvtObjectRepo.Delete(id, projectID)
}
