package service

import (
	"backend-v2/internal/repository"
	"backend-v2/model"
	"backend-v2/pkg/utils"
)

type tpTPObjectService struct {
	tpTPObjectRepo repository.ITPObjectRepository
}

func InitTPObjectService(tpTPObjectRepo repository.ITPObjectRepository) ITPObjectService {
	return &tpTPObjectService{
		tpTPObjectRepo: tpTPObjectRepo,
	}
}

type ITPObjectService interface {
	GetAll() ([]model.TP_Object, error)
	GetPaginated(page, limit int, data model.TP_Object) ([]model.TP_Object, error)
	GetByID(id uint) (model.TP_Object, error)
	Create(data model.TP_Object) (model.TP_Object, error)
	Update(data model.TP_Object) (model.TP_Object, error)
	Delete(id uint) error
	Count() (int64, error)
}

func (service *tpTPObjectService) GetAll() ([]model.TP_Object, error) {
	return service.tpTPObjectRepo.GetAll()
}

func (service *tpTPObjectService) GetPaginated(page, limit int, data model.TP_Object) ([]model.TP_Object, error) {
	if !(utils.IsEmptyFields(data)) {
		return service.tpTPObjectRepo.GetPaginatedFiltered(page, limit, data)
	}

	return service.tpTPObjectRepo.GetPaginated(page, limit)
}

func (service *tpTPObjectService) GetByID(id uint) (model.TP_Object, error) {
	return service.tpTPObjectRepo.GetByID(id)
}

func (service *tpTPObjectService) Create(data model.TP_Object) (model.TP_Object, error) {
	return service.tpTPObjectRepo.Create(data)
}

func (service *tpTPObjectService) Update(data model.TP_Object) (model.TP_Object, error) {
	return service.tpTPObjectRepo.Update(data)
}

func (service *tpTPObjectService) Delete(id uint) error {
	return service.tpTPObjectRepo.Delete(id)
}

func (service *tpTPObjectService) Count() (int64, error) {
	return service.tpTPObjectRepo.Count()
}
