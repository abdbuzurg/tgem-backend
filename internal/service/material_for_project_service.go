package service

import (
	"backend-v2/internal/repository"
	"backend-v2/model"
)

type materialForProjectService struct {
	materialForProjectRepo repository.IMaterialForProjectRepositry
}

func InitMaterialForProjectService(materialForProjectRepo repository.IMaterialForProjectRepositry) IMaterialForProjectService {
	return &materialForProjectService{
		materialForProjectRepo: materialForProjectRepo,
	}
}

type IMaterialForProjectService interface {
	GetAll() ([]model.MaterialForProject, error)
	GetByProjectID(projectID uint) ([]model.MaterialForProject, error)
	GetByMaterialID(materialID uint) ([]model.MaterialForProject, error)
	Create(data model.MaterialForProject) (model.MaterialForProject, error)
	Update(data model.MaterialForProject) (model.MaterialForProject, error)
	Delete(id uint) error
}

func (service *materialForProjectService) GetAll() ([]model.MaterialForProject, error) {
	return service.materialForProjectRepo.GetAll()
}

func (service *materialForProjectService) GetByProjectID(projectID uint) ([]model.MaterialForProject, error) {
	return service.materialForProjectRepo.GetByProjectID(projectID)
}

func (service *materialForProjectService) GetByMaterialID(materialID uint) ([]model.MaterialForProject, error) {
	return service.materialForProjectRepo.GetByMaterialID(materialID)
}

func (service *materialForProjectService) Create(data model.MaterialForProject) (model.MaterialForProject, error) {
	return service.materialForProjectRepo.Create(data)
}

func (service *materialForProjectService) Update(data model.MaterialForProject) (model.MaterialForProject, error) {
	return service.materialForProjectRepo.Update(data)
}

func (service *materialForProjectService) Delete(id uint) error {
	return service.materialForProjectRepo.Delete(id)
}
