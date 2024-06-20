package repository

import (
	"backend-v2/internal/dto"
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
	GetByCode(code string) (model.SerialNumber, error)
	GetByMaterialCostID(materialCostID uint) ([]model.SerialNumber, error)
	GetMaterialCostIDsByCodesInLocation(materialID uint, codes []string, locationType string, locationID uint) ([]dto.MaterialCostIDAndSNLocationIDQueryResult, error)
	GetSerialNumberIDsBySerialNumberCodes(codes []string) ([]model.SerialNumber, error)
	Create(data model.SerialNumber) (model.SerialNumber, error)
	CreateInBatches(data []model.SerialNumber) ([]model.SerialNumber, error)
	Update(data model.SerialNumber) (model.SerialNumber, error)
	Delete(id uint) error
	GetCodesByMaterialID(projectID, materialID uint, status string) ([]string, error)
	GetCodesByMaterialIDAndLocation(projectID, materialID uint, locationType string, locationID uint) ([]string, error)
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

func (repo *serialNumberRepository) GetMaterialCostIDsByCodesInLocation(materialID uint, codes []string, locationType string, locationID uint) ([]dto.MaterialCostIDAndSNLocationIDQueryResult, error) {
	data := []dto.MaterialCostIDAndSNLocationIDQueryResult{}
	err := repo.db.Raw(`
      SELECT
        material_costs.id as material_cost_id,
        serial_numbers.id as serial_number_id,
        serial_number_locations.id as serial_number_location_id
      FROM serial_numbers
      INNER JOIN material_costs ON material_costs.id = serial_numbers.material_cost_id
      INNER JOIN materials ON materials.id = material_costs.material_id
      INNER JOIN serial_number_locations ON serial_number_locations.serial_number_id = serial_numbers.id
      WHERE 	
        materials.id = ? AND
        serial_number_locations.location_type = ? AND
        serial_number_locations.location_id = ? AND
        serial_numbers.code IN ?
      ORDER BY material_costs.id
    `, materialID, locationType, locationID, codes).Scan(&data).Error

	return data, err
}

func (repo *serialNumberRepository) GetSerialNumberIDsBySerialNumberCodes(codes []string) ([]model.SerialNumber, error) {
	data := []model.SerialNumber{}
	err := repo.db.Find(&data, "code IN ?", codes).Error
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
    FROM materials
    INNER JOIN material_costs ON material_costs.material_id = materials.id
    INNER JOIN serial_numbers ON material_costs.id = serial_numbers.material_cost_id
    INNER JOIN serial_number_locations ON serial_number_locations.serial_number_id = serial_numbers.id
    WHERE
      materials.project_id = serial_numbers.project_id AND
      materials.project_id = serial_number_locations.project_id AND
      materials.project_id = ? AND
      materials.id = ? AND
      serial_number_locations.location_type = ? AND
      serial_number_locations.location_id = 0;
    `, projectID, materialID, status).Scan(&data).Error
	return data, err
}

func (repo *serialNumberRepository) GetCodesByMaterialIDAndLocation(projectID, materialID uint, locationType string, locationID uint) ([]string, error) {
	var data []string
	err := repo.db.Raw(`
    SELECT serial_numbers.code
    FROM material_locations
    INNER JOIN serial_numbers ON serial_numbers.material_cost_id = material_locations.material_cost_id
    INNER JOIN serial_number_locations ON serial_number_locations.serial_number_id = serial_numbers.id
    INNER JOIN material_costs ON material_locations.material_cost_id = material_costs.id
    INNER JOIN materials ON materials.id = material_costs.material_id
    WHERE
      materials.project_id = ? AND
      materials.id = ? AND
      material_locations.location_type = serial_number_locations.location_type AND
      material_locations.location_type = ? AND
      material_locations.location_id = serial_number_locations.location_id AND
      material_locations.location_id = ?;
    `, projectID, materialID, locationType, locationID).Scan(&data).Error

	return data, err
}
