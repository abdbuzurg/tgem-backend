package service

import (
	"backend-v2/internal/repository"
	"backend-v2/model"
	"backend-v2/pkg/utils"
)

type materialCostService struct {
	materialCostRepo repository.IMaterialCostRepository
}

func InitMaterialCostService(materialCostRepo repository.IMaterialCostRepository) IMaterialCostService {
	return &materialCostService{
		materialCostRepo: materialCostRepo,
	}
}

type IMaterialCostService interface {
	GetAll() ([]model.MaterialCost, error)
	GetPaginated(page, limit int, data model.MaterialCost) ([]model.MaterialCost, error)
	GetByID(id uint) (model.MaterialCost, error)
	Create(data model.MaterialCost) (model.MaterialCost, error)
	Update(data model.MaterialCost) (model.MaterialCost, error)
	Delete(id uint) error
	Count() (int64, error)
}

func (service *materialCostService) GetAll() ([]model.MaterialCost, error) {
	return service.materialCostRepo.GetAll()
}

func (service *materialCostService) GetPaginated(page, limit int, data model.MaterialCost) ([]model.MaterialCost, error) {
	if !(utils.IsEmptyFields(data)) {
		return service.materialCostRepo.GetPaginatedFiltered(page, limit, data)
	}

	return service.materialCostRepo.GetPaginated(page, limit)
}

func (service *materialCostService) GetByID(id uint) (model.MaterialCost, error) {
	return service.materialCostRepo.GetByID(id)
}

func (service *materialCostService) Create(data model.MaterialCost) (model.MaterialCost, error) {
	return service.materialCostRepo.Create(data)
}

func (service *materialCostService) Update(data model.MaterialCost) (model.MaterialCost, error) {
	return service.materialCostRepo.Update(data)
}

func (service *materialCostService) Delete(id uint) error {
	return service.materialCostRepo.Delete(id)
}

func (service *materialCostService) Count() (int64, error) {
	return service.materialCostRepo.Count()
}
