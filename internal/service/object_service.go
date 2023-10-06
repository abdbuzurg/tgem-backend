package service

import (
	"backend-v2/internal/repository"
	"backend-v2/model"
	"backend-v2/pkg/utils"
)

type objectService struct {
	objectRepo repository.IObjectRepository
}

func InitObjectService(objectRepo repository.IObjectRepository) IObjectService {
	return &objectService{
		objectRepo: objectRepo,
	}
}

type IObjectService interface {
	GetAll() ([]model.Object, error)
	GetPaginated(page, limit int, data model.Object) ([]model.Object, error)
	GetByID(id uint) (model.Object, error)
	Create(data model.Object) (model.Object, error)
	Update(data model.Object) (model.Object, error)
	Delete(id uint) error
	Count() (int64, error)
}

func (service *objectService) GetAll() ([]model.Object, error) {
	return service.objectRepo.GetAll()
}

func (service *objectService) GetPaginated(page, limit int, data model.Object) ([]model.Object, error) {
	if !(utils.IsEmptyFields(data)) {
		return service.objectRepo.GetPaginatedFiltered(page, limit, data)
	}

	return service.objectRepo.GetPaginated(page, limit)
}

func (service *objectService) GetByID(id uint) (model.Object, error) {
	return service.objectRepo.GetByID(id)
}

func (service *objectService) Create(data model.Object) (model.Object, error) {
	return service.objectRepo.Create(data)
}

func (service *objectService) Update(data model.Object) (model.Object, error) {
	return service.objectRepo.Update(data)
}

func (service *objectService) Delete(id uint) error {
	return service.objectRepo.Delete(id)
}

func (service *objectService) Count() (int64, error) {
	return service.objectRepo.Count()
}
