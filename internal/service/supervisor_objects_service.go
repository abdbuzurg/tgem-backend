package service

import (
	"backend-v2/internal/repository"
	"backend-v2/model"
)

type supervisorObjectsService struct {
	supervisorObjectsRepo repository.ISupervisorObjectsRepository
}

func InitSupervisorObjectsService(supervisorObjectsRepo repository.ISupervisorObjectsRepository) ISupervisorObjectsService {
	return &supervisorObjectsService{
		supervisorObjectsRepo: supervisorObjectsRepo,
	}
}

type ISupervisorObjectsService interface {
	GetByObjectID(objectID uint) ([]model.SupervisorObjects, error)
	GetBySupervisorWorkerID(workerID uint) ([]model.SupervisorObjects, error)
	CreateBatch(data []model.SupervisorObjects) ([]model.SupervisorObjects, error)
}

func (service *supervisorObjectsService) GetByObjectID(objectID uint) ([]model.SupervisorObjects, error) {
	return service.supervisorObjectsRepo.GetByObjectID(objectID)
}

func (service *supervisorObjectsService) GetBySupervisorWorkerID(workerID uint) ([]model.SupervisorObjects, error) {
	return service.supervisorObjectsRepo.GetBySupervisorWorkerID(workerID)
}

func (service *supervisorObjectsService) CreateBatch(data []model.SupervisorObjects) ([]model.SupervisorObjects, error) {
	return service.supervisorObjectsRepo.CreateBatch(data)
}
