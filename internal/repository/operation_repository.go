package repository

import (
	"backend-v2/model"

	"gorm.io/gorm"
)

type operationRepository struct {
	db *gorm.DB
}

func InitOperationRepository(db *gorm.DB) IOperationRepository {
	return &operationRepository{
		db: db,
	}
}

type IOperationRepository interface {
	GetAll() ([]model.Operation, error)
	GetPaginated(page, limit int) ([]model.Operation, error)
	GetPaginatedFiltered(page, limit int, filter model.Operation) ([]model.Operation, error)
	GetByID(id uint) (model.Operation, error)
	Create(data model.Operation) (model.Operation, error)
	Update(data model.Operation) (model.Operation, error)
	Delete(id uint) error
	Count() (int64, error)
}

func (repo *operationRepository) GetAll() ([]model.Operation, error) {
	data := []model.Operation{}
	err := repo.db.Order("id desc").Find(&data).Error
	return data, err
}

func (repo *operationRepository) GetPaginated(page, limit int) ([]model.Operation, error) {
	data := []model.Operation{}
	err := repo.db.Order("id desc").Offset((page - 1) * limit).Limit(limit).Find(&data).Error
	return data, err
}

func (repo *operationRepository) GetPaginatedFiltered(page, limit int, filter model.Operation) ([]model.Operation, error) {
	data := []model.Operation{}
	err := repo.db.
		Raw(`SELECT * FROM materials WHERE
			(nullif(?, '') IS NULL OR name = ?) AND
			(nullif(?, '') IS NULL OR code = ?) AND
			(nullif(?, '') IS NULL OR cost_primer = ?) AND
			(nullif(?, '') IS NULL OR cost_with_customer = ?) ORDER BY id DESC LIMIT ? OFFSET ?`,
			filter.Name, filter.Name,
			filter.Code, filter.Code,
			filter.CostPrime, filter.CostPrime,
			filter.CostWithCustomer, filter.CostWithCustomer,
			limit, (page-1)*limit,
		).
		Scan(&data).Error

	return data, err
}

func (repo *operationRepository) GetByID(id uint) (model.Operation, error) {
	data := model.Operation{}
	err := repo.db.Find(&data, "id = ?", id).Error
	return data, err
}

func (repo *operationRepository) Create(data model.Operation) (model.Operation, error) {
	err := repo.db.Create(&data).Error
	return data, err
}

func (repo *operationRepository) Update(data model.Operation) (model.Operation, error) {
	err := repo.db.Model(&model.Operation{}).Select("*").Updates(&data).Error
	return data, err
}

func (repo *operationRepository) Delete(id uint) error {
	return repo.db.Delete(&model.Operation{}, "id = ?", id).Error
}

func (repo *operationRepository) Count() (int64, error) {
	var count int64
	err := repo.db.Model(&model.Operation{}).Count(&count).Error
	return count, err
}
