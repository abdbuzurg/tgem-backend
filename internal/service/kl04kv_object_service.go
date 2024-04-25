package service

import (
	"backend-v2/internal/dto"
	"backend-v2/internal/repository"
	"backend-v2/model"
)

type kl04kvObjectService struct {
	kl04kvObjectRepo repository.IKL04KVObjectRepository
	objectRepo       repository.IObjectRepository
}

func InitKL04KVObjectService(
	kl04kvObjectRepo repository.IKL04KVObjectRepository,
	objectRepo repository.IObjectRepository,
) IKL04KVObjectService {
	return &kl04kvObjectService{
		kl04kvObjectRepo: kl04kvObjectRepo,
	}
}

type IKL04KVObjectService interface {
	GetPaginated(page, limit int, projectID uint) ([]dto.KL04KVObjectPaginated, error)
	Count(projectID uint) (int64, error)
	Create(data dto.KL04KVObjectCreate) (model.KL04KV_Object, error)
	Delete(projectID, id uint) error
	Update(data dto.KL04KVObjectCreate) (model.KL04KV_Object, error)
}

func (service *kl04kvObjectService) GetPaginated(page, limit int, projectID uint) ([]dto.KL04KVObjectPaginated, error) {

	data, err := service.kl04kvObjectRepo.GetPaginated(page, limit, projectID)
	if err != nil {
		return []dto.KL04KVObjectPaginated{}, err
	}

	result := []dto.KL04KVObjectPaginated{}
	latestEntry := dto.KL04KVObjectPaginated{}
	for index, object := range data {
		if index == 0 {
			latestEntry = dto.KL04KVObjectPaginated{
				ObjectID:         object.ObjectID,
				ObjectDetailedID: object.ObjectDetailedID,
				Name:             object.Name,
				Status:           object.Status,
				Length:           object.Length,
				Nourashes:        object.Nourashes,
				Supervisors:      []string{},
			}
		}

		if latestEntry.ObjectID == object.ObjectID {
			latestEntry.Supervisors = append(latestEntry.Supervisors, object.SupervisorName)
		} else {

			result = append(result, latestEntry)
			latestEntry = dto.KL04KVObjectPaginated{
				ObjectID:         object.ObjectID,
				ObjectDetailedID: object.ObjectDetailedID,
				Name:             object.Name,
				Status:           object.Status,
				Length:           object.Length,
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

func (service *kl04kvObjectService) Count(projectID uint) (int64, error) {
	return service.kl04kvObjectRepo.Count(projectID)
}

func (service *kl04kvObjectService) Create(data dto.KL04KVObjectCreate) (model.KL04KV_Object, error) {
	return service.kl04kvObjectRepo.Create(data)
}

func (service *kl04kvObjectService) Delete(projectID, id uint) error {
	return service.kl04kvObjectRepo.Delete(projectID, id)
}

func (service *kl04kvObjectService) Update(data dto.KL04KVObjectCreate) (model.KL04KV_Object, error) {
	return service.kl04kvObjectRepo.Update(data)
}
