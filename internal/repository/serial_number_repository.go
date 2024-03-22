package repository

import (
	"backend-v2/model"

	"gorm.io/gorm"
)

type serialNumberRepository struct {
	db *gorm.DB
}

func InitSerialNumberRepository(db *gorm.DB) ISerialNumberRepository {
	return &serialNumberRepository{
		db: db,
	}
}

type ISerialNumberRepository interface {
	GetAll() ([]model.SerialNumber, error)
	GetByID(id uint) (model.SerialNumber, error)
	GetByStatus(status string, statusID uint) ([]model.SerialNumber, error)
	GetByCode(code string) (model.SerialNumber, error)
	GetByMaterialCostID(materialCostID uint) ([]model.SerialNumber, error)
	Create(data model.SerialNumber) (model.SerialNumber, error)
	CreateInBatches(data []model.SerialNumber) ([]model.SerialNumber, error)
	Update(data model.SerialNumber) (model.SerialNumber, error)
	Delete(id uint) error
}

func (repo *serialNumberRepository) GetAll() ([]model.SerialNumber, error) {
	data := []model.SerialNumber{}
	err := repo.db.Find(&data).Error
	return data, err
}

func (repo *serialNumberRepository) GetByID(id uint) (model.SerialNumber, error) {
	data := model.SerialNumber{}
	err := repo.db.Find(&data, "id = ?", id).Error
	return data, err
}

func (repo *serialNumberRepository) GetByStatus(status string, statusID uint) ([]model.SerialNumber, error) {
	var data []model.SerialNumber
	err := repo.db.Find(&data, "status = ? AND status_id = ?", status, statusID).Error
	return data, err
}

func (repo *serialNumberRepository) GetByCode(code string) (model.SerialNumber, error) {
	data := model.SerialNumber{}
	err := repo.db.Find(&data, "code = ?", code).Error
	return data, err
}

func (repo *serialNumberRepository) GetByMaterialCostID(materialCostID uint) ([]model.SerialNumber, error) {
	data := []model.SerialNumber{}
	err := repo.db.Find(&data, "material_cost_id = ?", materialCostID).Error
	return data, err
}

func (repo *serialNumberRepository) Create(data model.SerialNumber) (model.SerialNumber, error) {
	err := repo.db.Create(&data).Error
	return data, err
}

func (repo *serialNumberRepository) CreateInBatches(data []model.SerialNumber) ([]model.SerialNumber, error) {
	err := repo.db.CreateInBatches(&data, 100).Error
	return data, err
}

func (repo *serialNumberRepository) Update(data model.SerialNumber) (model.SerialNumber, error) {
	err := repo.db.Model(model.SerialNumber{}).Select("*").Where("id = ?", data.ID).Updates(&data).Error
	return data, err
}

func (repo *serialNumberRepository) Delete(id uint) error {
	err := repo.db.Delete(model.SerialNumber{}, "id = ?", id).Error
	return err
}
