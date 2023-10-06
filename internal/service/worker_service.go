package service

import (
	"backend-v2/internal/repository"
	"backend-v2/model"
	"backend-v2/pkg/utils"
)

type workerService struct {
	workerRepo repository.IWorkerRepository
}

func InitWorkerService(workerRepo repository.IWorkerRepository) IWorkerService {
	return &workerService{
		workerRepo: workerRepo,
	}
}

type IWorkerService interface {
	GetAll() ([]model.Worker, error)
	GetPaginated(page, limit int, data model.Worker) ([]model.Worker, error)
	GetByID(id uint) (model.Worker, error)
	Create(data model.Worker) (model.Worker, error)
	Update(data model.Worker) (model.Worker, error)
	Delete(id uint) error
	Count() (int64, error)
}

func (service *workerService) GetAll() ([]model.Worker, error) {
	return service.workerRepo.GetAll()
}

func (service *workerService) GetPaginated(page, limit int, data model.Worker) ([]model.Worker, error) {
	if !(utils.IsEmptyFields(data)) {
		return service.workerRepo.GetPaginatedFiltered(page, limit, data)
	}

	return service.workerRepo.GetPaginated(page, limit)
}

func (service *workerService) GetByID(id uint) (model.Worker, error) {
	return service.workerRepo.GetByID(id)
}

func (service *workerService) Create(data model.Worker) (model.Worker, error) {
	return service.workerRepo.Create(data)
}

func (service *workerService) Update(data model.Worker) (model.Worker, error) {
	return service.workerRepo.Update(data)
}

func (service *workerService) Delete(id uint) error {
	return service.workerRepo.Delete(id)
}

func (service *workerService) Count() (int64, error) {
	return service.workerRepo.Count()
}
