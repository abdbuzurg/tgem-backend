package service

import (
	"backend-v2/internal/repository"
	"backend-v2/model"
	"backend-v2/pkg/utils"
)

type districtService struct {
	districtRepo repository.IDistrictRepository
}

func InitDistrictService(districtRepo repository.IDistrictRepository) IDistrictService {
	return &districtService{
		districtRepo: districtRepo,
	}
}

type IDistrictService interface {
	GetAll() ([]model.District, error)
	GetPaginated(page, limit int, data model.District) ([]model.District, error)
	GetByID(id uint) (model.District, error)
	Create(data model.District) (model.District, error)
	Update(data model.District) (model.District, error)
	Delete(id uint) error
	Count() (int64, error)
}

func (service *districtService) GetAll() ([]model.District, error) {
	return service.districtRepo.GetAll()
}

func (service *districtService) GetPaginated(page, limit int, data model.District) ([]model.District, error) {
	if !(utils.IsEmptyFields(data)) {
		return service.districtRepo.GetPaginatedFiltered(page, limit, data)
	}

	return service.districtRepo.GetPaginated(page, limit)
}

func (service *districtService) GetByID(id uint) (model.District, error) {
	return service.districtRepo.GetByID(id)
}

func (service *districtService) Create(data model.District) (model.District, error) {
	return service.districtRepo.Create(data)
}

func (service *districtService) Update(data model.District) (model.District, error) {
	return service.districtRepo.Update(data)
}

func (service *districtService) Delete(id uint) error {
	return service.districtRepo.Delete(id)
}

func (service *districtService) Count() (int64, error) {
	return service.districtRepo.Count()
}
