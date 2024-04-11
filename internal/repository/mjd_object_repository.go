package repository

import (
	"backend-v2/model"

	"gorm.io/gorm"
)

type mjdObjectRepository struct {
	db *gorm.DB
}

func InitMJDObjectRepository(db *gorm.DB) IMJDObjectRepository {
	return &mjdObjectRepository{
		db: db,
	}
}

type IMJDObjectRepository interface {
	GetAll() ([]model.MJD_Object, error)
	GetPaginated(page, limit int) ([]model.MJD_Object, error)
	GetPaginatedFiltered(page, limit int, filter model.MJD_Object) ([]model.MJD_Object, error)
	GetByID(id uint) (model.MJD_Object, error)
	Create(data model.MJD_Object) (model.MJD_Object, error)
	Update(data model.MJD_Object) (model.MJD_Object, error)
	Delete(id uint) error
	Count() (int64, error)
}

func (repo *mjdObjectRepository) GetAll() ([]model.MJD_Object, error) {
	data := []model.MJD_Object{}
	err := repo.db.Order("id desc").Find(&data).Error
	return data, err
}

func (repo *mjdObjectRepository) GetPaginated(page, limit int) ([]model.MJD_Object, error) {
	data := []model.MJD_Object{}
	err := repo.db.Order("id desc").Offset((page - 1) * limit).Limit(limit).Find(&data).Error
	return data, err
}

func (repo *mjdObjectRepository) GetPaginatedFiltered(page, limit int, filter model.MJD_Object) ([]model.MJD_Object, error) {
	data := []model.MJD_Object{}
	err := repo.db.
		Raw(`SELECT * FROM mjd_objects WHERE
			(nullif(?, '') IS NULL OR type = ?) AND
			(nullif(?, '') IS NULL OR amount_stores = ?) AND
			(nullif(?, '') IS NULL OR amount_entrances = ?) AND
			(nullif(?, '') IS NULL OR has_basement = ?) ORDER BY id DESC LIMIT ? OFFSET ?`,
			filter.Model, filter.Model, filter.AmountStores, filter.AmountStores, filter.AmountEntraces, filter.HasBasement, filter.HasBasement, limit, (page-1)*limit,
		).
		Scan(&data).Error

	return data, err
}

func (repo *mjdObjectRepository) GetByID(id uint) (model.MJD_Object, error) {
	data := model.MJD_Object{}
	err := repo.db.Find(&data, "id = ?", id).Error
	return data, err
}

func (repo *mjdObjectRepository) Create(data model.MJD_Object) (model.MJD_Object, error) {
	err := repo.db.Create(&data).Error
	return data, err
}

func (repo *mjdObjectRepository) Update(data model.MJD_Object) (model.MJD_Object, error) {
	err := repo.db.Model(&model.MJD_Object{}).Select("*").Updates(&data).Error
	return data, err
}

func (repo *mjdObjectRepository) Delete(id uint) error {
	return repo.db.Delete(&model.MJD_Object{}, "id = ?", id).Error
}

func (repo *mjdObjectRepository) Count() (int64, error) {
	var count int64
	err := repo.db.Model(&model.MJD_Object{}).Count(&count).Error
	return count, err
}
