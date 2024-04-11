package repository

import (
	"backend-v2/model"

	"gorm.io/gorm"
)

type tpObjectRepository struct {
	db *gorm.DB
}

func InitTPObjectRepository(db *gorm.DB) ITPObjectRepository {
	return &tpObjectRepository{
		db: db,
	}
}

type ITPObjectRepository interface {
	GetAll() ([]model.TP_Object, error)
	GetPaginated(page, limit int) ([]model.TP_Object, error)
	GetPaginatedFiltered(page, limit int, filter model.TP_Object) ([]model.TP_Object, error)
	GetByID(id uint) (model.TP_Object, error)
	Create(data model.TP_Object) (model.TP_Object, error)
	Update(data model.TP_Object) (model.TP_Object, error)
	Delete(id uint) error
	Count() (int64, error)
}

func (repo *tpObjectRepository) GetAll() ([]model.TP_Object, error) {
	data := []model.TP_Object{}
	err := repo.db.Order("id desc").Find(&data).Error
	return data, err
}

func (repo *tpObjectRepository) GetPaginated(page, limit int) ([]model.TP_Object, error) {
	data := []model.TP_Object{}
	err := repo.db.Order("id desc").Offset((page - 1) * limit).Limit(limit).Find(&data).Error
	return data, err
}

func (repo *tpObjectRepository) GetPaginatedFiltered(page, limit int, filter model.TP_Object) ([]model.TP_Object, error) {
	data := []model.TP_Object{}
	err := repo.db.
		Raw(`SELECT * FROM tp_objects WHERE
			(nullif(?, '') IS NULL OR model = ?) AND
			(nullif(?, '') IS NULL OR voltage_class = ?) AND
			(nullif(?, '') IS NULL OR nourashes = ?) ORDER BY id DESC LIMIT ? OFFSET ?`,
			filter.Model, filter.Model, filter.VoltageClass, filter.VoltageClass, filter.Nourashes, filter.Nourashes, limit, (page-1)*limit,
		).
		Scan(&data).Error

	return data, err
}

func (repo *tpObjectRepository) GetByID(id uint) (model.TP_Object, error) {
	data := model.TP_Object{}
	err := repo.db.Find(&data, "id = ?", id).Error
	return data, err
}

func (repo *tpObjectRepository) Create(data model.TP_Object) (model.TP_Object, error) {
	err := repo.db.Create(&data).Error
	return data, err
}

func (repo *tpObjectRepository) Update(data model.TP_Object) (model.TP_Object, error) {
	err := repo.db.Model(&model.TP_Object{}).Select("*").Updates(&data).Error
	return data, err
}

func (repo *tpObjectRepository) Delete(id uint) error {
	return repo.db.Delete(&model.TP_Object{}, "id = ?", id).Error
}

func (repo *tpObjectRepository) Count() (int64, error) {
	var count int64
	err := repo.db.Model(&model.TP_Object{}).Count(&count).Error
	return count, err
}
