package service

import (
	"backend-v2/internal/repository"
	"backend-v2/model"
	"backend-v2/pkg/utils"
)

type sipSIPObjectService struct {
	sipSIPObjectRepo repository.ISIPObjectRepository
}

func InitSIPObjectService(sipSIPObjectRepo repository.ISIPObjectRepository) ISIPObjectService {
	return &sipSIPObjectService{
		sipSIPObjectRepo: sipSIPObjectRepo,
	}
}

type ISIPObjectService interface {
	GetAll() ([]model.SIP_Object, error)
	GetPaginated(page, limit int, data model.SIP_Object) ([]model.SIP_Object, error)
	GetByID(id uint) (model.SIP_Object, error)
	Create(data model.SIP_Object) (model.SIP_Object, error)
	Update(data model.SIP_Object) (model.SIP_Object, error)
	Delete(id uint) error
	Count() (int64, error)
}

func (service *sipSIPObjectService) GetAll() ([]model.SIP_Object, error) {
	return service.sipSIPObjectRepo.GetAll()
}

func (service *sipSIPObjectService) GetPaginated(page, limit int, data model.SIP_Object) ([]model.SIP_Object, error) {
	if !(utils.IsEmptyFields(data)) {
		return service.sipSIPObjectRepo.GetPaginatedFiltered(page, limit, data)
	}

	return service.sipSIPObjectRepo.GetPaginated(page, limit)
}

func (service *sipSIPObjectService) GetByID(id uint) (model.SIP_Object, error) {
	return service.sipSIPObjectRepo.GetByID(id)
}

func (service *sipSIPObjectService) Create(data model.SIP_Object) (model.SIP_Object, error) {
	return service.sipSIPObjectRepo.Create(data)
}

func (service *sipSIPObjectService) Update(data model.SIP_Object) (model.SIP_Object, error) {
	return service.sipSIPObjectRepo.Update(data)
}

func (service *sipSIPObjectService) Delete(id uint) error {
	return service.sipSIPObjectRepo.Delete(id)
}

func (service *sipSIPObjectService) Count() (int64, error) {
	return service.sipSIPObjectRepo.Count()
}
