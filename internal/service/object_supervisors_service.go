package service

import (
	"backend-v2/internal/repository"
	"backend-v2/model"
)

type objectSupervisorsService struct {
	objectSupervisorsRepo repository.IObjectSupervisorsRepository
}

func InitObjectSupervisorsService(objectSupervisorsRepo repository.IObjectSupervisorsRepository) IObjectSupervisorsService {
	return &objectSupervisorsService{
		objectSupervisorsRepo: objectSupervisorsRepo,
	}
}

type IObjectSupervisorsService interface {
	GetByObjectID(objectID uint) ([]model.ObjectSupervisors, error)
	GetBySupervisorWorkerID(workerID uint) ([]model.ObjectSupervisors, error)
	CreateBatch(data []model.ObjectSupervisors) ([]model.ObjectSupervisors, error)
}

func (service *objectSupervisorsService) GetByObjectID(objectID uint) ([]model.ObjectSupervisors, error) {
	return service.objectSupervisorsRepo.GetByObjectID(objectID)
}

func (service *objectSupervisorsService) GetBySupervisorWorkerID(workerID uint) ([]model.ObjectSupervisors, error) {
	return service.objectSupervisorsRepo.GetBySupervisorWorkerID(workerID)
}

func (service *objectSupervisorsService) CreateBatch(data []model.ObjectSupervisors) ([]model.ObjectSupervisors, error) {
	return service.objectSupervisorsRepo.CreateBatch(data)
}
