package service

import (
	"backend-v2/internal/repository"
	"backend-v2/model"
	"backend-v2/pkg/utils"
)

type operationService struct {
	operationRepo repository.IOperationRepository
}

func InitOperationService(operationRepo repository.IOperationRepository) IOperationService {
	return &operationService{
		operationRepo: operationRepo,
	}
}

type IOperationService interface {
	GetAll() ([]model.Operation, error)
	GetPaginated(page, limit int, data model.Operation) ([]model.Operation, error)
	GetByID(id uint) (model.Operation, error)
	Create(data model.Operation) (model.Operation, error)
	Update(data model.Operation) (model.Operation, error)
	Delete(id uint) error
	Count() (int64, error)
}

func (service *operationService) GetAll() ([]model.Operation, error) {
	return service.operationRepo.GetAll()
}

func (service *operationService) GetPaginated(page, limit int, data model.Operation) ([]model.Operation, error) {
	if !(utils.IsEmptyFields(data)) {
		return service.operationRepo.GetPaginatedFiltered(page, limit, data)
	}

	return service.operationRepo.GetPaginated(page, limit)
}

func (service *operationService) GetByID(id uint) (model.Operation, error) {
	return service.operationRepo.GetByID(id)
}

func (service *operationService) Create(data model.Operation) (model.Operation, error) {
	return service.operationRepo.Create(data)
}

func (service *operationService) Update(data model.Operation) (model.Operation, error) {
	return service.operationRepo.Update(data)
}

func (service *operationService) Delete(id uint) error {
	return service.operationRepo.Delete(id)
}

func (service *operationService) Count() (int64, error) {
	return service.operationRepo.Count()
}
