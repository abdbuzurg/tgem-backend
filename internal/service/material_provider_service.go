package service

import (
	"backend-v2/internal/repository"
	"backend-v2/model"
)

type materialProviderService struct {
	materialProviderRepo repository.IMaterialProviderRepository
}

func InitMaterialProviderService(materialProviderRepo repository.IMaterialProviderRepository) IMaterialProviderService {
	return &materialProviderService{
		materialProviderRepo: materialProviderRepo,
	}
}

type IMaterialProviderService interface {
	GetPaginated(page, limit int, projectID uint) ([]model.MaterialProvider, error)
	Count(projectID uint) (int64, error)
	Create(data model.MaterialProvider) (model.MaterialProvider, error)
	Update(data model.MaterialProvider) (model.MaterialProvider, error)
	Delete(id uint) error
}

func (service *materialProviderService) GetPaginated(page, limit int, projectID uint) ([]model.MaterialProvider, error) {
	return service.materialProviderRepo.GetPaginated(page, limit, projectID)
}

func (service *materialProviderService) Count(projectID uint) (int64, error) {
	return service.materialProviderRepo.Count(projectID)
}

func (service *materialProviderService) Create(data model.MaterialProvider) (model.MaterialProvider, error) {
	return service.materialProviderRepo.Create(data)
}

func (service *materialProviderService) Update(data model.MaterialProvider) (model.MaterialProvider, error) {
	return service.materialProviderRepo.Update(data)
}

func (service *materialProviderService) Delete(id uint) error {
	return service.materialProviderRepo.Delete(id)
}
