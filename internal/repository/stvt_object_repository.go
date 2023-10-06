package repository

import (
	"backend-v2/model"

	"gorm.io/gorm"
)

type stvtObjectRepository struct {
	db *gorm.DB
}

func InitSTVTObjectRepository(db *gorm.DB) ISTVTObjectRepository {
	return &stvtObjectRepository{
		db: db,
	}
}

type ISTVTObjectRepository interface {
	GetAll() ([]model.STVT_Object, error)
	GetPaginated(page, limit int) ([]model.STVT_Object, error)
	GetPaginatedFiltered(page, limit int, filter model.STVT_Object) ([]model.STVT_Object, error)
	GetByID(id uint) (model.STVT_Object, error)
	Create(data model.STVT_Object) (model.STVT_Object, error)
	Update(data model.STVT_Object) (model.STVT_Object, error)
	Delete(id uint) error
	Count() (int64, error)
}

func (repo *stvtObjectRepository) GetAll() ([]model.STVT_Object, error) {
	data := []model.STVT_Object{}
	err := repo.db.Order("id desc").Find(&data).Error
	return data, err
}

func (repo *stvtObjectRepository) GetPaginated(page, limit int) ([]model.STVT_Object, error) {
	data := []model.STVT_Object{}
	err := repo.db.Order("id desc").Offset((page - 1) * limit).Limit(limit).Find(&data).Error
	return data, err
}

func (repo *stvtObjectRepository) GetPaginatedFiltered(page, limit int, filter model.STVT_Object) ([]model.STVT_Object, error) {
	data := []model.STVT_Object{}
	err := repo.db.
		Raw(`SELECT * FROM stvt_objects WHERE
			(nullif(?, '') IS NULL OR voltage_class = ?) AND
			(nullif(?, '') IS NULL OR tt_coefficient = ?) ORDER BY id DESC LIMIT ? OFFSET ?`,
			filter.VoltageClass, filter.VoltageClass, filter.TTCoefficient, filter.TTCoefficient, limit, (page-1)*limit,
		).
		Scan(&data).Error

	return data, err
}

func (repo *stvtObjectRepository) GetByID(id uint) (model.STVT_Object, error) {
	data := model.STVT_Object{}
	err := repo.db.Find(&data, "id = ?", id).Error
	return data, err
}

func (repo *stvtObjectRepository) Create(data model.STVT_Object) (model.STVT_Object, error) {
	err := repo.db.Create(&data).Error
	return data, err
}

func (repo *stvtObjectRepository) Update(data model.STVT_Object) (model.STVT_Object, error) {
	err := repo.db.Model(&model.STVT_Object{}).Select("*").Updates(&data).Error
	return data, err
}

func (repo *stvtObjectRepository) Delete(id uint) error {
	return repo.db.Delete(&model.STVT_Object{}, "id = ?", id).Error
}

func (repo *stvtObjectRepository) Count() (int64, error) {
	var count int64
	err := repo.db.Model(&model.STVT_Object{}).Count(&count).Error
	return count, err
}
