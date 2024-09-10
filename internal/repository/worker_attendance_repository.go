package repository

import (
	"backend-v2/internal/dto"
	"backend-v2/model"

	"gorm.io/gorm"
)

type workerAttendanceRepository struct {
	db *gorm.DB
}

type IWorkerAttendanceRepository interface {
	CreateBatch(data []model.WorkerAttendance) error
  GetPaginated(projectID uint) ([]dto.WorkerAttendancePaginated, error)
  Count(projectID uint) (int64, error)
}

func InitWorkerAttendanceRepository(db *gorm.DB) IWorkerAttendanceRepository {
	return &workerAttendanceRepository{
		db: db,
	}
}

func (repo *workerAttendanceRepository) CreateBatch(data []model.WorkerAttendance) error {
	return repo.db.CreateInBatches(&data, 50).Error
}

func (repo *workerAttendanceRepository) GetPaginated(projectID uint) ([]dto.WorkerAttendancePaginated, error) {
  var result []dto.WorkerAttendancePaginated
  err := repo.db.Raw(`
    SELECT 
      worker_attendances.id as id,
      workers.name as worker_name,
      workers.company_worker_id  as company_worker_id,
      worker_attendances.start as "start",
      worker_attendances.end as "end"
    FROM worker_attendances
    INNER JOIN workers ON workers.id = worker_attendances.worker_id
    WHERE worker_attendances.project_id = ?
    `, projectID).Scan(&result).Error

  return result, err
}

func (repo *workerAttendanceRepository) Count(projectID uint) (int64, error) {
  var count int64
  err := repo.db.Raw(`SELECT COUNT(*) FROM worker_attendances WHERE project_id = ?`, projectID).Scan(&count).Error
  return count, err
}
