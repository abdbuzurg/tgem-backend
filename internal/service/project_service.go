package service

import (
	"backend-v2/internal/repository"
	"backend-v2/model"
)

type projectService struct {
	projectRepo repository.IProjectRepository
}

func InitProjectService(projectRepo repository.IProjectRepository) IProjectService {
	return &projectService{
		projectRepo: projectRepo,
	}
}

type IProjectService interface {
	GetAll() ([]model.Project, error)
	GetPaginated(page, limit int) ([]model.Project, error)
	GetByID(id uint) (model.Project, error)
	Create(data model.Project) (model.Project, error)
	Update(data model.Project) (model.Project, error)
	Delete(id uint) error
	Count() (int64, error)
  GetProjectName(projectID uint) (string, error)
}

func (service *projectService) GetAll() ([]model.Project, error) {
	return service.projectRepo.GetAll()
}

func (service *projectService) GetPaginated(page, limit int) ([]model.Project, error) {
	return service.projectRepo.GetPaginated(page, limit)
}

func (service *projectService) GetByID(id uint) (model.Project, error) {
	return service.projectRepo.GetByID(id)
}

func (service *projectService) Create(data model.Project) (model.Project, error) {
	return service.projectRepo.Create(data)
}

func (service *projectService) Update(data model.Project) (model.Project, error) {
	return service.projectRepo.Update(data)
}

func (service *projectService) Delete(id uint) error {
	return service.projectRepo.Delete(id)
}

func (service *projectService) Count() (int64, error) {
	return service.projectRepo.Count()
}

func (service *projectService) GetProjectName(projectID uint) (string, error) {
  return service.projectRepo.GetProjectName(projectID)
}
