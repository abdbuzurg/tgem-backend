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
	GetCodesByMaterialID(projectID, materialID uint, status string) ([]string, error)
	GetCodesByMaterialIDAndStatus(projectID, materialID uint, status string) ([]string, error)
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

func (repo *serialNumberRepository) GetCodesByMaterialID(projectID, materialID uint, status string) ([]string, error) {
	var data []string
	err := repo.db.Raw(`
    SELECT serial_numbers.code
    FROM serial_numbers
      INNER JOIN material_costs ON material_costs.id = serial_numbers.material_cost_id
      INNER JOIN materials ON materials.id = material_costs.material_id
    WHERE
      materials.project_id = ? AND
      materials.id = ? AND
      serial_numbers.status = ?
    `, projectID, materialID, status).Scan(&data).Error
	return data, err
}

func (repo *serialNumberRepository) GetCodesByMaterialIDAndStatus(projectID, materialID uint, status string) ([]string, error) {
	var data []string
	err := repo.db.Raw(`
    SELECT serial_numbers.code
    FROM serial_numbers
      INNER JOIN material_locations ON serial_numbers.status_id = material_locations.id
      INNER JOIN material_costs ON material_costs.id = material_locations.material_cost_id
      INNER JOIN materials ON materials.id = material_costs.material_id
    WHERE
      materials.project_id = ? AND
      materials.id = ? AND
      material_locations.location_type = ? AND
      serial_numbers.status = ?;
    `, projectID, materialID, status, status).Scan(&data).Error

	return data, err
}
