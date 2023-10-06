package service

import (
	"backend-v2/internal/repository"
	"backend-v2/model"
	"backend-v2/pkg/utils"
)

type materialService struct {
	materialRepo repository.IMaterialRepository
}

func InitMaterialService(materialRepo repository.IMaterialRepository) IMaterialService {
	return &materialService{
		materialRepo: materialRepo,
	}
}

type IMaterialService interface {
	GetAll() ([]model.Material, error)
	GetPaginated(page, limit int, data model.Material) ([]model.Material, error)
	GetByID(id uint) (model.Material, error)
	Create(data model.Material) (model.Material, error)
	Update(data model.Material) (model.Material, error)
	Delete(id uint) error
	Count() (int64, error)
}

func (service *materialService) GetAll() ([]model.Material, error) {
	return service.materialRepo.GetAll()
}

func (service *materialService) GetPaginated(page, limit int, data model.Material) ([]model.Material, error) {
	if !(utils.IsEmptyFields(data)) {
		return service.materialRepo.GetPaginatedFiltered(page, limit, data)
	}

	return service.materialRepo.GetPaginated(page, limit)
}

func (service *materialService) GetByID(id uint) (model.Material, error) {
	return service.materialRepo.GetByID(id)
}

func (service *materialService) Create(data model.Material) (model.Material, error) {
	return service.materialRepo.Create(data)
}

func (service *materialService) Update(data model.Material) (model.Material, error) {
	return service.materialRepo.Update(data)
}

func (service *materialService) Delete(id uint) error {
	return service.materialRepo.Delete(id)
}

func (service *materialService) Count() (int64, error) {
	return service.materialRepo.Count()
}
