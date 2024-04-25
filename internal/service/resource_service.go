package service

import (
	"backend-v2/internal/repository"
	"backend-v2/model"
)

type resourceService struct {
	resourceRepo repository.IResourceRepository
}

func InitResourceService(
	resourceRepo repository.IResourceRepository,
) IResourceService {
	return &resourceService{
		resourceRepo: resourceRepo,
	}
}

type IResourceService interface {
	GetAll() ([]model.Resource, error)
}

func (service *resourceService) GetAll() ([]model.Resource, error) {
	return service.resourceRepo.GetAll()
}
