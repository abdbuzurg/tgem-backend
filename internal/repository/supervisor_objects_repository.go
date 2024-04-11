package repository

import (
	"backend-v2/model"

	"gorm.io/gorm"
)

type supervisorObjectsRepository struct {
	db *gorm.DB
}

func InitSupervisorObjectsRepository(db *gorm.DB) ISupervisorObjectsRepository {
	return &supervisorObjectsRepository{
		db: db,
	}
}

type ISupervisorObjectsRepository interface {
	GetByObjectID(objectID uint) ([]model.SupervisorObjects, error)
	GetBySupervisorWorkerID(workerID uint) ([]model.SupervisorObjects, error)
	CreateBatch(data []model.SupervisorObjects) ([]model.SupervisorObjects, error)
}

func (repo *supervisorObjectsRepository) GetByObjectID(objectID uint) ([]model.SupervisorObjects, error) {
	var data []model.SupervisorObjects
	err := repo.db.Find(&data, "object_id = ?", objectID).Error
	return data, err
}

func (repo *supervisorObjectsRepository) GetBySupervisorWorkerID(workerID uint) ([]model.SupervisorObjects, error) {
	var data []model.SupervisorObjects
	err := repo.db.Find(&data, "supervisor_worker_id = ?", workerID).Error
	return data, err
}

func (repo *supervisorObjectsRepository) CreateBatch(data []model.SupervisorObjects) ([]model.SupervisorObjects, error) {
	err := repo.db.CreateInBatches(&data, 10).Error
	return data, err
}

func(repo *supervisorObjectsRepository) GetSupervisorsByObjectID(objectID uint) ([]model.Worker, error) {
  var data []model.Worker
  err := repo.db.
    Raw(`
      SELECT *
      FROM supervisor_objects
      INNER JOIN workers ON supervisor_objects.supervisor_worker_id = worker.id
    `).
    Scan(&data).Error
  return data, err
}
