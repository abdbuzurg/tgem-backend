package service

import (
	"backend-v2/internal/repository"
	"backend-v2/model"
	"backend-v2/pkg/utils"
)

type materialLocationService struct {
	materialLocationRepo repository.IMaterialLocationRepository
}

func InitMaterialLocationService(materialLocationRepo repository.IMaterialLocationRepository) IMaterialLocationService {
	return &materialLocationService{
		materialLocationRepo: materialLocationRepo,
	}
}

type IMaterialLocationService interface {
	GetAll() ([]model.MaterialLocation, error)
	GetPaginated(page, limit int, data model.MaterialLocation) ([]model.MaterialLocation, error)
	GetByID(id uint) (model.MaterialLocation, error)
	Create(data model.MaterialLocation) (model.MaterialLocation, error)
	Update(data model.MaterialLocation) (model.MaterialLocation, error)
	Delete(id uint) error
	Count() (int64, error)
}

func (service *materialLocationService) GetAll() ([]model.MaterialLocation, error) {
	return service.materialLocationRepo.GetAll()
}

func (service *materialLocationService) GetPaginated(page, limit int, data model.MaterialLocation) ([]model.MaterialLocation, error) {
	if !(utils.IsEmptyFields(data)) {
		return service.materialLocationRepo.GetPaginatedFiltered(page, limit, data)
	}

	return service.materialLocationRepo.GetPaginated(page, limit)
}

func (service *materialLocationService) GetByID(id uint) (model.MaterialLocation, error) {
	return service.materialLocationRepo.GetByID(id)
}

func (service *materialLocationService) Create(data model.MaterialLocation) (model.MaterialLocation, error) {
	return service.materialLocationRepo.Create(data)
}

func (service *materialLocationService) Update(data model.MaterialLocation) (model.MaterialLocation, error) {
	return service.materialLocationRepo.Update(data)
}

func (service *materialLocationService) Delete(id uint) error {
	return service.materialLocationRepo.Delete(id)
}

func (service *materialLocationService) Count() (int64, error) {
	return service.materialLocationRepo.Count()
}
