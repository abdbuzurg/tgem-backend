package repository

import (
	"backend-v2/model"

	"gorm.io/gorm"
)

type districtRepository struct {
	db *gorm.DB
}

func InitDistrictRepository(db *gorm.DB) IDistrictRepository {
	return &districtRepository{
		db: db,
	}
}

type IDistrictRepository interface {
	GetAll() ([]model.District, error)
	GetPaginated(page, limit int) ([]model.District, error)
	GetPaginatedFiltered(page, limit int, filter model.District) ([]model.District, error)
	GetByID(id uint) (model.District, error)
	Create(data model.District) (model.District, error)
	Update(data model.District) (model.District, error)
	Delete(id uint) error
	Count() (int64, error)
	GetByName(name string) (model.District, error)
}

func (repo *districtRepository) GetAll() ([]model.District, error) {
	data := []model.District{}
	err := repo.db.Order("id desc").Find(&data).Error
	return data, err
}

func (repo *districtRepository) GetPaginated(page, limit int) ([]model.District, error) {
	data := []model.District{}
	err := repo.db.Order("id desc").Offset((page - 1) * limit).Limit(limit).Find(&data).Error
	return data, err
}

func (repo *districtRepository) GetPaginatedFiltered(page, limit int, filter model.District) ([]model.District, error) {
	data := []model.District{}
	err := repo.db.
		Raw(`SELECT * FROM districts WHERE
			(nullif(?, '') IS NULL OR name = ?) ORDER BY id DESC LIMIT ? OFFSET ?`,
			filter.Name, filter.Name, limit, (page-1)*limit,
		).
		Scan(&data).Error

	return data, err
}

func (repo *districtRepository) GetByID(id uint) (model.District, error) {
	data := model.District{}
	err := repo.db.Find(&data, "id = ?", id).Error
	return data, err
}

func (repo *districtRepository) Create(data model.District) (model.District, error) {
	err := repo.db.Create(&data).Error
	return data, err
}

func (repo *districtRepository) Update(data model.District) (model.District, error) {
	err := repo.db.Model(&model.District{}).Select("*").Where("id = ?", data.ID).Updates(&data).Error
	return data, err
}

func (repo *districtRepository) Delete(id uint) error {
	return repo.db.Delete(&model.District{}, "id = ?", id).Error
}

func (repo *districtRepository) Count() (int64, error) {
	var count int64
	err := repo.db.Model(&model.District{}).Count(&count).Error
	return count, err
}

func (repo *districtRepository) GetByName(name string) (model.District, error) {
	data := model.District{}
	err := repo.db.Find(&data, "name = ?", name).Error
	return data, err
}
