package repository

import (
	"backend-v2/model"

	"gorm.io/gorm"
)

type sipObjectRepository struct {
	db *gorm.DB
}

func InitSIPObjectRepository(db *gorm.DB) ISIPObjectRepository {
	return &sipObjectRepository{
		db: db,
	}
}

type ISIPObjectRepository interface {
	GetAll() ([]model.SIP_Object, error)
	GetPaginated(page, limit int) ([]model.SIP_Object, error)
	GetPaginatedFiltered(page, limit int, filter model.SIP_Object) ([]model.SIP_Object, error)
	GetByID(id uint) (model.SIP_Object, error)
	Create(data model.SIP_Object) (model.SIP_Object, error)
	Update(data model.SIP_Object) (model.SIP_Object, error)
	Delete(id uint) error
	Count() (int64, error)
}

func (repo *sipObjectRepository) GetAll() ([]model.SIP_Object, error) {
	data := []model.SIP_Object{}
	err := repo.db.Order("id desc").Find(&data).Error
	return data, err
}

func (repo *sipObjectRepository) GetPaginated(page, limit int) ([]model.SIP_Object, error) {
	data := []model.SIP_Object{}
	err := repo.db.Order("id desc").Offset((page - 1) * limit).Limit(limit).Find(&data).Error
	return data, err
}

func (repo *sipObjectRepository) GetPaginatedFiltered(page, limit int, filter model.SIP_Object) ([]model.SIP_Object, error) {
	data := []model.SIP_Object{}
	err := repo.db.
		Raw(`SELECT * FROM sip_objects WHERE
			(nullif(?, '') IS NULL OR amount_feeders = ?) ORDER BY id DESC LIMIT ? OFFSET ?`,
			filter.AmountFeeders, filter.AmountFeeders, limit, (page-1)*limit,
		).
		Scan(&data).Error

	return data, err
}

func (repo *sipObjectRepository) GetByID(id uint) (model.SIP_Object, error) {
	data := model.SIP_Object{}
	err := repo.db.Find(&data, "id = ?", id).Error
	return data, err
}

func (repo *sipObjectRepository) Create(data model.SIP_Object) (model.SIP_Object, error) {
	err := repo.db.Create(&data).Error
	return data, err
}

func (repo *sipObjectRepository) Update(data model.SIP_Object) (model.SIP_Object, error) {
	err := repo.db.Model(&model.SIP_Object{}).Select("*").Updates(&data).Error
	return data, err
}

func (repo *sipObjectRepository) Delete(id uint) error {
	return repo.db.Delete(&model.SIP_Object{}, "id = ?", id).Error
}

func (repo *sipObjectRepository) Count() (int64, error) {
	var count int64
	err := repo.db.Model(&model.SIP_Object{}).Count(&count).Error
	return count, err
}
