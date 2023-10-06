package service

import (
	"backend-v2/internal/repository"
	"backend-v2/model"
	"backend-v2/pkg/utils"
)

type stvtObjectService struct {
	stvtObjectRepo repository.ISTVTObjectRepository
}

func InitSTVTObjectService(stvtObjectRepo repository.ISTVTObjectRepository) ISTVTObjectService {
	return &stvtObjectService{
		stvtObjectRepo: stvtObjectRepo,
	}
}

type ISTVTObjectService interface {
	GetAll() ([]model.STVT_Object, error)
	GetPaginated(page, limit int, data model.STVT_Object) ([]model.STVT_Object, error)
	GetByID(id uint) (model.STVT_Object, error)
	Create(data model.STVT_Object) (model.STVT_Object, error)
	Update(data model.STVT_Object) (model.STVT_Object, error)
	Delete(id uint) error
	Count() (int64, error)
}

func (service *stvtObjectService) GetAll() ([]model.STVT_Object, error) {
	return service.stvtObjectRepo.GetAll()
}

func (service *stvtObjectService) GetPaginated(page, limit int, data model.STVT_Object) ([]model.STVT_Object, error) {
	if !(utils.IsEmptyFields(data)) {
		return service.stvtObjectRepo.GetPaginatedFiltered(page, limit, data)
	}

	return service.stvtObjectRepo.GetPaginated(page, limit)
}

func (service *stvtObjectService) GetByID(id uint) (model.STVT_Object, error) {
	return service.stvtObjectRepo.GetByID(id)
}

func (service *stvtObjectService) Create(data model.STVT_Object) (model.STVT_Object, error) {
	return service.stvtObjectRepo.Create(data)
}

func (service *stvtObjectService) Update(data model.STVT_Object) (model.STVT_Object, error) {
	return service.stvtObjectRepo.Update(data)
}

func (service *stvtObjectService) Delete(id uint) error {
	return service.stvtObjectRepo.Delete(id)
}

func (service *stvtObjectService) Count() (int64, error) {
	return service.stvtObjectRepo.Count()
}
