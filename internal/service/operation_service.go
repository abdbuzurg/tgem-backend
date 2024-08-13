package service

import (
	"backend-v2/internal/dto"
	"backend-v2/internal/repository"
	"backend-v2/model"
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
	GetAll(projectID uint) ([]dto.OperationPaginated, error)
	GetPaginated(page, limit int, filter dto.OperationSearchParameters) ([]dto.OperationPaginated, error)
	GetByID(id uint) (model.Operation, error)
  GetByName(name string, projectID uint) (model.Operation, error)
	Create(data dto.Operation) (model.Operation, error)
	Update(data dto.Operation) (model.Operation, error)
	Delete(id uint) error
	Count(filter dto.OperationSearchParameters) (int64, error)
}

func (service *operationService) GetAll(projectID uint) ([]dto.OperationPaginated, error) {
	return service.operationRepo.GetAll(projectID)
}

func (service *operationService) GetPaginated(page, limit int, filter dto.OperationSearchParameters) ([]dto.OperationPaginated, error) {
	return service.operationRepo.GetPaginated(page, limit, filter)
}

func (service *operationService) GetByID(id uint) (model.Operation, error) {
	return service.operationRepo.GetByID(id)
}

func (service *operationService) Create(data dto.Operation) (model.Operation, error) {
	return service.operationRepo.Create(data)
}

func (service *operationService) Update(data dto.Operation) (model.Operation, error) {
	return service.operationRepo.Update(data)
}

func (service *operationService) Delete(id uint) error {
	return service.operationRepo.Delete(id)
}

func (service *operationService) Count(filter dto.OperationSearchParameters) (int64, error) {
	return service.operationRepo.Count(filter)
}

func (service *operationService) GetByName(name string, projectID uint) (model.Operation, error) {
  return service.operationRepo.GetByName(name, projectID)
}
