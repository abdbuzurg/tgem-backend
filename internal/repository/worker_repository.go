package repository

import (
	"backend-v2/model"

	"gorm.io/gorm"
)

type workerRepository struct {
	db *gorm.DB
}

func InitWorkerRepository(db *gorm.DB) IWorkerRepository {
	return &workerRepository{
		db: db,
	}
}

type IWorkerRepository interface {
	GetAll() ([]model.Worker, error)
	GetPaginated(page, limit int) ([]model.Worker, error)
	GetPaginatedFiltered(page, limit int, filter model.Worker) ([]model.Worker, error)
	GetByJobTitle(jobTitle string) ([]model.Worker, error)
  GetByName(name string) (model.Worker, error)
	GetByID(id uint) (model.Worker, error)
	Create(data model.Worker) (model.Worker, error)
	Update(data model.Worker) (model.Worker, error)
	Delete(id uint) error
	Count() (int64, error)
}

func (repo *workerRepository) GetAll() ([]model.Worker, error) {
	data := []model.Worker{}
	err := repo.db.Order("id DESC").Find(&data).Error
	return data, err
}

func (repo *workerRepository) GetPaginated(page, limit int) ([]model.Worker, error) {
	data := []model.Worker{}
	err := repo.db.Order("id DESC").Offset((page - 1) * limit).Limit(limit).Find(&data).Error
	return data, err
}

func (repo *workerRepository) GetPaginatedFiltered(page, limit int, filter model.Worker) ([]model.Worker, error) {
	data := []model.Worker{}
	err := repo.db.
		Raw(`SELECT * FROM workers WHERE
			(nullif(?, '') IS NULL OR name = ?) AND
			(nullif(?, '') IS NULL OR job_title = ?) AND
			(nullif(?, '') IS NULL OR mobile_number = ?) ORDER BY id DESC LIMIT ? OFFSET ?`,
			filter.Name, filter.Name, filter.JobTitle, filter.JobTitle, filter.MobileNumber, filter.MobileNumber, limit, (page-1)*limit,
		).Scan(&data).Error

	return data, err
}

func(repo *workerRepository) GetByName(name string) (model.Worker, error) {
  data := model.Worker{}
  err := repo.db.First(&data, "name = ?", name).Error
  return data, err
}

func (repo *workerRepository) GetByJobTitle(jobTitle string) ([]model.Worker, error) {
	data := []model.Worker{}
	err := repo.db.Order("id DESC").Find(&data, "job_title = ?", jobTitle).Error
	return data, err
}

func (repo *workerRepository) GetByID(id uint) (model.Worker, error) {
	data := model.Worker{}
	err := repo.db.Find(&data, "id = ?", id).Error
	return data, err
}

func (repo *workerRepository) Create(data model.Worker) (model.Worker, error) {
	err := repo.db.Create(&data).Error
	return data, err
}

func (repo *workerRepository) Update(data model.Worker) (model.Worker, error) {
	err := repo.db.Model(&model.Worker{}).Select("*").Where("id = ?", data.ID).Updates(&data).Error
	return data, err
}

func (repo *workerRepository) Delete(id uint) error {
	return repo.db.Delete(&model.Worker{}, "id = ?", id).Error
}

func (repo *workerRepository) Count() (int64, error) {
	var count int64
	err := repo.db.Model(&model.Worker{}).Count(&count).Error
	return count, err
}
