package repository

import (
	"backend-v2/model"

	"gorm.io/gorm"
)

type objectOperationRepository struct {
	db *gorm.DB
}

func InitObjectOperationRepository(db *gorm.DB) IObjectOperationRepository {
	return &objectOperationRepository{
		db: db,
	}
}

type IObjectOperationRepository interface {
	GetAll() ([]model.ObjectOperation, error)
	GetPaginated(page, limit int) ([]model.ObjectOperation, error)
	GetPaginatedFiltered(page, limit int, filter model.ObjectOperation) ([]model.ObjectOperation, error)
	GetByID(id uint) (model.ObjectOperation, error)
	Create(data model.ObjectOperation) (model.ObjectOperation, error)
	Update(data model.ObjectOperation) (model.ObjectOperation, error)
	Delete(id uint) error
	Count() (int64, error)
}

func (repo *objectOperationRepository) GetAll() ([]model.ObjectOperation, error) {
	data := []model.ObjectOperation{}
	err := repo.db.Order("id desc").Find(&data).Error
	return data, err
}

func (repo *objectOperationRepository) GetPaginated(page, limit int) ([]model.ObjectOperation, error) {
	data := []model.ObjectOperation{}
	err := repo.db.Order("id desc").Offset((page - 1) * limit).Limit(limit).Find(&data).Error
	return data, err
}

func (repo *objectOperationRepository) GetPaginatedFiltered(page, limit int, filter model.ObjectOperation) ([]model.ObjectOperation, error) {
	data := []model.ObjectOperation{}
	err := repo.db.
		Raw(`SELECT * FROM materials WHERE
			(nullif(?, '') IS NULL OR object_id = ?) AND
			(nullif(?, '') IS NULL OR material_cost_id = ?) AND
			(nullif(?, '') IS NULL OR operation_id = ?) AND
			(nullif(?, '') IS NULL OR team_id = ?) ORDER BY id DESC LIMIT ? OFFSET ?`,
			filter.ObjectID, filter.ObjectID,
			filter.MaterialCostID, filter.MaterialCostID,
			filter.OperationID, filter.OperationID,
			filter.TeamID, filter.TeamID,
			limit, (page-1)*limit,
		).
		Scan(&data).Error

	return data, err
}

func (repo *objectOperationRepository) GetByID(id uint) (model.ObjectOperation, error) {
	data := model.ObjectOperation{}
	err := repo.db.Find(&data, "id = ?", id).Error
	return data, err
}

func (repo *objectOperationRepository) Create(data model.ObjectOperation) (model.ObjectOperation, error) {
	err := repo.db.Create(&data).Error
	return data, err
}

func (repo *objectOperationRepository) Update(data model.ObjectOperation) (model.ObjectOperation, error) {
	err := repo.db.Model(&model.ObjectOperation{}).Select("*").Updates(&data).Error
	return data, err
}

func (repo *objectOperationRepository) Delete(id uint) error {
	return repo.db.Delete(&model.ObjectOperation{}, "id = ?", id).Error
}

func (repo *objectOperationRepository) Count() (int64, error) {
	var count int64
	err := repo.db.Model(&model.ObjectOperation{}).Count(&count).Error
	return count, err
}
