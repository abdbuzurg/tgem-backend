package service

import (
	"backend-v2/internal/repository"
	"backend-v2/model"
	"backend-v2/pkg/utils"
)

type mjdObjectService struct {
	mjdObjectRepo repository.IMJDObjectRepository
}

func InitMJDObjectService(mjdObjectRepo repository.IMJDObjectRepository) IMJDObjectService {
	return &mjdObjectService{
		mjdObjectRepo: mjdObjectRepo,
	}
}

type IMJDObjectService interface {
	GetAll() ([]model.MJD_Object, error)
	GetPaginated(page, limit int, data model.MJD_Object) ([]model.MJD_Object, error)
	GetByID(id uint) (model.MJD_Object, error)
	Create(data model.MJD_Object) (model.MJD_Object, error)
	Update(data model.MJD_Object) (model.MJD_Object, error)
	Delete(id uint) error
	Count() (int64, error)
}

func (service *mjdObjectService) GetAll() ([]model.MJD_Object, error) {
	return service.mjdObjectRepo.GetAll()
}

func (service *mjdObjectService) GetPaginated(page, limit int, data model.MJD_Object) ([]model.MJD_Object, error) {
	if !(utils.IsEmptyFields(data)) {
		return service.mjdObjectRepo.GetPaginatedFiltered(page, limit, data)
	}

	return service.mjdObjectRepo.GetPaginated(page, limit)
}

func (service *mjdObjectService) GetByID(id uint) (model.MJD_Object, error) {
	return service.mjdObjectRepo.GetByID(id)
}

func (service *mjdObjectService) Create(data model.MJD_Object) (model.MJD_Object, error) {
	return service.mjdObjectRepo.Create(data)
}

func (service *mjdObjectService) Update(data model.MJD_Object) (model.MJD_Object, error) {
	return service.mjdObjectRepo.Update(data)
}

func (service *mjdObjectService) Delete(id uint) error {
	return service.mjdObjectRepo.Delete(id)
}

func (service *mjdObjectService) Count() (int64, error) {
	return service.mjdObjectRepo.Count()
}
