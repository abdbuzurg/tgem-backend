package service

import (
	"backend-v2/internal/repository"
	"backend-v2/model"
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
	GetAll(projectID uint) ([]model.District, error)
	GetPaginated(page, limit int, projectID uint) ([]model.District, error)
	GetByID(id uint) (model.District, error)
	Create(data model.District) (model.District, error)
	Update(data model.District) (model.District, error)
	Delete(id uint) error
	Count(projectID uint) (int64, error)
}

func (service *districtService) GetAll(projectID uint) ([]model.District, error) {
	return service.districtRepo.GetAll(projectID)
}

func (service *districtService) GetPaginated(page, limit int, projectID uint) ([]model.District, error) {
	return service.districtRepo.GetPaginated(page, limit, projectID)
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

func (service *districtService) Count(projectID uint) (int64, error) {
	return service.districtRepo.Count(projectID)
}
