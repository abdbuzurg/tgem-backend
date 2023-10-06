package repository

import (
	"backend-v2/model"

	"gorm.io/gorm"
)

type kl04kvObjectRepository struct {
	db *gorm.DB
}

func InitKL04KVObjectRepository(db *gorm.DB) IKL04KVObjectRepository {
	return &kl04kvObjectRepository{
		db: db,
	}
}

type IKL04KVObjectRepository interface {
	GetAll() ([]model.KL04KV_Object, error)
	GetPaginated(page, limit int) ([]model.KL04KV_Object, error)
	GetPaginatedFiltered(page, limit int, filter model.KL04KV_Object) ([]model.KL04KV_Object, error)
	GetByID(id uint) (model.KL04KV_Object, error)
	Create(data model.KL04KV_Object) (model.KL04KV_Object, error)
	Update(data model.KL04KV_Object) (model.KL04KV_Object, error)
	Delete(id uint) error
	Count() (int64, error)
}

func (repo *kl04kvObjectRepository) GetAll() ([]model.KL04KV_Object, error) {
	data := []model.KL04KV_Object{}
	err := repo.db.Order("id desc").Find(&data).Error
	return data, err
}

func (repo *kl04kvObjectRepository) GetPaginated(page, limit int) ([]model.KL04KV_Object, error) {
	data := []model.KL04KV_Object{}
	err := repo.db.Order("id desc").Offset((page - 1) * limit).Limit(limit).Find(&data).Error
	return data, err
}

func (repo *kl04kvObjectRepository) GetPaginatedFiltered(page, limit int, filter model.KL04KV_Object) ([]model.KL04KV_Object, error) {
	data := []model.KL04KV_Object{}
	err := repo.db.
		Raw(`SELECT * FROM kl04kv_objects WHERE
			(nullif(?, '') IS NULL OR length = ?) AND
			(nullif(?, '') IS NULL OR nourashes = ?) ORDER BY id DESC LIMIT ? OFFSET ?`,
			filter.Length, filter.Length, filter.Nourashes, filter.Nourashes, limit, (page-1)*limit,
		).
		Scan(&data).Error

	return data, err
}

func (repo *kl04kvObjectRepository) GetByID(id uint) (model.KL04KV_Object, error) {
	data := model.KL04KV_Object{}
	err := repo.db.Find(&data, "id = ?", id).Error
	return data, err
}

func (repo *kl04kvObjectRepository) Create(data model.KL04KV_Object) (model.KL04KV_Object, error) {
	err := repo.db.Create(&data).Error
	return data, err
}

func (repo *kl04kvObjectRepository) Update(data model.KL04KV_Object) (model.KL04KV_Object, error) {
	err := repo.db.Model(&model.KL04KV_Object{}).Select("*").Updates(&data).Error
	return data, err
}

func (repo *kl04kvObjectRepository) Delete(id uint) error {
	return repo.db.Delete(&model.KL04KV_Object{}, "id = ?", id).Error
}

func (repo *kl04kvObjectRepository) Count() (int64, error) {
	var count int64
	err := repo.db.Model(&model.KL04KV_Object{}).Count(&count).Error
	return count, err
}
