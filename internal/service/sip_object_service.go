package service

import (
	"backend-v2/internal/dto"
	"backend-v2/internal/repository"
	"backend-v2/model"
)

type sipObjectService struct {
	sipObjectRepo repository.ISIPObjectRepository
}

func InitSIPObjectService(sipObjectRepo repository.ISIPObjectRepository) ISIPObjectService {
	return &sipObjectService{
		sipObjectRepo: sipObjectRepo,
	}
}

type ISIPObjectService interface {
	GetPaginated(page, limit int, projectID uint) ([]dto.SIPObjectPaginated, error)
	Count(projectID uint) (int64, error)
	Create(data dto.SIPObjectCreate) (model.SIP_Object, error)
	Update(data dto.SIPObjectCreate) (model.SIP_Object, error)
	Delete(id, projectID uint) error
}

func (service *sipObjectService) GetPaginated(page, limit int, projectID uint) ([]dto.SIPObjectPaginated, error) {

	data, err := service.sipObjectRepo.GetPaginated(page, limit, projectID)
	if err != nil {
		return []dto.SIPObjectPaginated{}, err
	}

	result := []dto.SIPObjectPaginated{}
	latestEntry := dto.SIPObjectPaginated{}
	for index, object := range data {
		if index == 0 {
			latestEntry = dto.SIPObjectPaginated{
				ObjectID:         object.ObjectID,
				ObjectDetailedID: object.ObjectDetailedID,
				Name:             object.Name,
				Status:           object.Status,
				AmountFeeders:    object.AmountFeeders,
				Supervisors:      []string{},
			}
		}

		if latestEntry.ObjectID == object.ObjectID {
			latestEntry.Supervisors = append(latestEntry.Supervisors, object.SupervisorName)
		} else {

			result = append(result, latestEntry)
			latestEntry = dto.SIPObjectPaginated{
				ObjectID:         object.ObjectID,
				ObjectDetailedID: object.ObjectDetailedID,
				Name:             object.Name,
				Status:           object.Status,
				AmountFeeders:    object.AmountFeeders,
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

func (service *sipObjectService) Count(projectID uint) (int64, error) {
	return service.sipObjectRepo.Count(projectID)
}

func (service *sipObjectService) Create(data dto.SIPObjectCreate) (model.SIP_Object, error) {
	return service.sipObjectRepo.Create(data)
}

func (service *sipObjectService) Update(data dto.SIPObjectCreate) (model.SIP_Object, error) {
	return service.sipObjectRepo.Update(data)
}

func (service *sipObjectService) Delete(id, projectID uint) error {
	return service.sipObjectRepo.Delete(id, projectID)
}
