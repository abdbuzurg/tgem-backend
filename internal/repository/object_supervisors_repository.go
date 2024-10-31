package repository

import (
	"backend-v2/internal/dto"
	"backend-v2/model"

	"gorm.io/gorm"
)

type objectSupervisorsRepository struct {
	db *gorm.DB
}

func InitObjectSupervisorsRepository(db *gorm.DB) IObjectSupervisorsRepository {
	return &objectSupervisorsRepository{
		db: db,
	}
}

type IObjectSupervisorsRepository interface {
	GetByObjectID(objectID uint) ([]model.ObjectSupervisors, error)
	GetBySupervisorWorkerID(workerID uint) ([]model.ObjectSupervisors, error)
	CreateBatch(data []model.ObjectSupervisors) ([]model.ObjectSupervisors, error)
	GetSupervisorAndObjectNamesByObjectID(projectID, objectID uint) ([]dto.SupervisorAndObjectNameQueryResult, error)
	GetSupervisorsNameByObjectID(objectID uint) ([]string, error)
}

func (repo *objectSupervisorsRepository) GetByObjectID(objectID uint) ([]model.ObjectSupervisors, error) {
	var data []model.ObjectSupervisors
	err := repo.db.Find(&data, "object_id = ?", objectID).Error
	return data, err
}

func (repo *objectSupervisorsRepository) GetBySupervisorWorkerID(workerID uint) ([]model.ObjectSupervisors, error) {
	var data []model.ObjectSupervisors
	err := repo.db.Find(&data, "supervisor_worker_id = ?", workerID).Error
	return data, err
}

func (repo *objectSupervisorsRepository) CreateBatch(data []model.ObjectSupervisors) ([]model.ObjectSupervisors, error) {
	err := repo.db.CreateInBatches(&data, 10).Error
	return data, err
}

func (repo *objectSupervisorsRepository) GetSupervisorsByObjectID(objectID uint) ([]model.Worker, error) {
	var data []model.Worker
	err := repo.db.
		Raw(`
      SELECT *
      FROM object_supervisors
      INNER JOIN workers ON object_supervisors.supervisor_worker_id = worker.id
    `).
		Scan(&data).Error
	return data, err
}

func (repo *objectSupervisorsRepository) GetSupervisorAndObjectNamesByObjectID(projectID, objectID uint) ([]dto.SupervisorAndObjectNameQueryResult, error) {
	data := []dto.SupervisorAndObjectNameQueryResult{}
	err := repo.db.Raw(`
    SELECT 
      objects.name as object_name,
      objects.type as object_type,
      workers.name as supervisor_name
    FROM object_supervisors
    RIGHT JOIN objects ON objects.id = object_supervisors.object_id
    LEFT JOIN workers ON workers.id = object_supervisors.supervisor_worker_id
    WHERE 
      objects.project_id = ? AND
      objects.id = ?
    `, projectID, objectID).Scan(&data).Error

	return data, err
}

func (repo *objectSupervisorsRepository) GetSupervisorsNameByObjectID(objectID uint) ([]string, error) {
	data := []string{}
	err := repo.db.Raw(`
    SELECT 
      workers.name as supervisor_name
    FROM object_supervisors
    INNER JOIN objects ON objects.id = object_supervisors.object_id
    INNER JOIN workers ON workers.id = object_supervisors.supervisor_worker_id
    WHERE 
      objects.id = ?
    `, objectID).Scan(&data).Error

	return data, err
}
