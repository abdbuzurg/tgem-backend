package service

import (
	"backend-v2/internal/repository"
	"backend-v2/model"
	"backend-v2/pkg/utils"
)

type objectOperationService struct {
	objectOperationRepo repository.IObjectOperationRepository
}

func InitObjectOperationService(objectOperationRepo repository.IObjectOperationRepository) IObjectOperationService {
	return &objectOperationService{
		objectOperationRepo: objectOperationRepo,
	}
}

type IObjectOperationService interface {
	GetAll() ([]model.ObjectOperation, error)
	GetPaginated(page, limit int, data model.ObjectOperation) ([]model.ObjectOperation, error)
	GetByID(id uint) (model.ObjectOperation, error)
	Create(data model.ObjectOperation) (model.ObjectOperation, error)
	Update(data model.ObjectOperation) (model.ObjectOperation, error)
	Delete(id uint) error
	Count() (int64, error)
}

func (service *objectOperationService) GetAll() ([]model.ObjectOperation, error) {
	return service.objectOperationRepo.GetAll()
}

func (service *objectOperationService) GetPaginated(page, limit int, data model.ObjectOperation) ([]model.ObjectOperation, error) {
	if !(utils.IsEmptyFields(data)) {
		return service.objectOperationRepo.GetPaginatedFiltered(page, limit, data)
	}

	return service.objectOperationRepo.GetPaginated(page, limit)
}

func (service *objectOperationService) GetByID(id uint) (model.ObjectOperation, error) {
	return service.objectOperationRepo.GetByID(id)
}

func (service *objectOperationService) Create(data model.ObjectOperation) (model.ObjectOperation, error) {
	return service.objectOperationRepo.Create(data)
}

func (service *objectOperationService) Update(data model.ObjectOperation) (model.ObjectOperation, error) {
	return service.objectOperationRepo.Update(data)
}

func (service *objectOperationService) Delete(id uint) error {
	return service.objectOperationRepo.Delete(id)
}

func (service *objectOperationService) Count() (int64, error) {
	return service.objectOperationRepo.Count()
}
