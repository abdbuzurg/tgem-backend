package service

import (
	"backend-v2/internal/dto"
	"backend-v2/internal/repository"
	"backend-v2/model"
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
	GetPaginated(page, limit int, projectID uint) ([]dto.MaterialCostView, error)
	GetByID(id uint) (model.MaterialCost, error)
	Create(data model.MaterialCost) (model.MaterialCost, error)
	Update(data model.MaterialCost) (model.MaterialCost, error)
	Delete(id uint) error
	Count() (int64, error)
	GetByMaterialID(materialID uint) ([]model.MaterialCost, error)
}

func (service *materialCostService) GetAll() ([]model.MaterialCost, error) {
	return service.materialCostRepo.GetAll()
}

func (service *materialCostService) GetPaginated(page, limit int, projectID uint) ([]dto.MaterialCostView, error) {
	return service.materialCostRepo.GetPaginatedFiltered(page, limit, projectID)
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

func (service *materialCostService) GetByMaterialID(materialID uint) ([]model.MaterialCost, error) {
	return service.materialCostRepo.GetByMaterialIDSorted(materialID)
}
