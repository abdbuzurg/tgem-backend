package repository

import (
	"backend-v2/model"

	"gorm.io/gorm"
)

type materialLocationRepository struct {
	db *gorm.DB
}

func InitMaterialLocationRepository(db *gorm.DB) IMaterialLocationRepository {
	return &materialLocationRepository{
		db: db,
	}
}

type IMaterialLocationRepository interface {
	GetAll() ([]model.MaterialLocation, error)
	GetPaginated(page, limit int) ([]model.MaterialLocation, error)
	GetPaginatedFiltered(page, limit int, filter model.MaterialLocation) ([]model.MaterialLocation, error)
	GetByID(id uint) (model.MaterialLocation, error)
	GetByMaterialCostIDOrCreate(projectID, materialCostID uint, locationType string, locationTypeID uint) (model.MaterialLocation, error)
	Create(data model.MaterialLocation) (model.MaterialLocation, error)
	Update(data model.MaterialLocation) (model.MaterialLocation, error)
	Delete(id uint) error
	Count() (int64, error)
	GetUniqueMaterialCostsByLocation(locationType string, locationID uint) ([]uint, error)
	UniqueObjectIDs() ([]uint, error)
	UniqueTeamIDs() ([]uint, error)
	GetByLocationTypeAndID(locationType string, locationID uint) ([]model.MaterialLocation, error)
	GetTotalAmountInWarehouse(projectID, materialID uint) (float64, error)
	GetUniqueMaterialsFromLocation(projectID, locationID uint, locationType string) ([]model.Material, error)
	GetUniqueMaterialCostsFromLocation(projectID, materialID, locationID uint, locationType string) ([]model.MaterialCost, error)
	GetUniqueMaterialTotalAmount(projectID, materialCostID, locationID uint, locationType string) (float64, error)
	GetTotalAmountInLocation(projectID, materialID, locationID uint, locationType string) (float64, error)
	GetTotalAmountInTeamsByTeamNumber(projectID, materialID uint, teamNumber string) (float64, error)
}

func (repo *materialLocationRepository) GetAll() ([]model.MaterialLocation, error) {
	data := []model.MaterialLocation{}
	err := repo.db.Order("id desc").Find(&data).Error
	return data, err
}

func (repo *materialLocationRepository) GetPaginated(page, limit int) ([]model.MaterialLocation, error) {
	data := []model.MaterialLocation{}
	err := repo.db.Order("id desc").Offset((page - 1) * limit).Limit(limit).Find(&data).Error
	return data, err
}

func (repo *materialLocationRepository) GetPaginatedFiltered(page, limit int, filter model.MaterialLocation) ([]model.MaterialLocation, error) {
	data := []model.MaterialLocation{}
	err := repo.db.
		Raw(`SELECT * FROM materials WHERE
			(nullif(?, '') IS NULL OR material_cost_id = ?) AND
			(nullif(?, '') IS NULL OR location_id = ?) AND
			(nullif(?, '') IS NULL OR location_type = ?) AND
			(nullif(?, '') IS NULL OR amount = ?) ORDER BY id DESC LIMIT ? OFFSET ?`,
			filter.MaterialCostID, filter.MaterialCostID,
			filter.LocationID, filter.LocationID,
			filter.LocationType, filter.LocationType,
			filter.Amount, filter.Amount,
			limit, (page-1)*limit,
		).
		Scan(&data).Error

	return data, err
}

func (repo *materialLocationRepository) GetByID(id uint) (model.MaterialLocation, error) {
	data := model.MaterialLocation{}
	err := repo.db.Find(&data, "id = ?", id).Error
	return data, err
}

func (repo *materialLocationRepository) GetByMaterialCostIDOrCreate(projectID, materialCostID uint, locationType string, locationID uint) (model.MaterialLocation, error) {
	data := model.MaterialLocation{}
	err := repo.db.FirstOrCreate(
		&data,
		model.MaterialLocation{
			LocationType:   locationType,
			LocationID:     locationID,
			MaterialCostID: materialCostID,
			ProjectID:      projectID,
		}).
		Error
	return data, err
}

func (repo *materialLocationRepository) Create(data model.MaterialLocation) (model.MaterialLocation, error) {
	err := repo.db.Create(&data).Error
	return data, err
}

func (repo *materialLocationRepository) Update(data model.MaterialLocation) (model.MaterialLocation, error) {
	err := repo.db.Table("material_locations").Updates(&data).Where("id = ?", data.ID).Error
	return data, err
}

func (repo *materialLocationRepository) Delete(id uint) error {
	return repo.db.Delete(&model.MaterialLocation{}, "id = ?", id).Error
}

func (repo *materialLocationRepository) Count() (int64, error) {
	var count int64
	err := repo.db.Model(&model.MaterialLocation{}).Count(&count).Error
	return count, err
}

func (repo *materialLocationRepository) GetUniqueMaterialCostsByLocation(
	locationType string,
	locationID uint,
) ([]uint, error) {
	var data []uint
	err := repo.db.Raw(`
    SELECT DISTINCT(material_cost_id) FROM material_locations WHERE location_type = ? and location_id = ?
    `, locationType, locationID).Scan(&data).Error

	return data, err
}

func (repo *materialLocationRepository) UniqueObjectIDs() ([]uint, error) {
	var data []uint
	err := repo.db.Raw("SELECT DISTINCT(location_id) FROM material_locations WHERE location_type='objects'").Scan(&data).Error
	return data, err
}

func (repo *materialLocationRepository) UniqueTeamIDs() ([]uint, error) {
	var data []uint
	err := repo.db.Raw("SELECT DISTINCT(location_id) FROM material_locations WHERE location_type='teams'").Scan(&data).Error
	return data, err
}

func (repo *materialLocationRepository) GetByLocationTypeAndID(
	locationType string,
	locationID uint,
) ([]model.MaterialLocation, error) {

	var data []model.MaterialLocation
	err := repo.db.Raw(`
    SELECT * 
    FROM material_locations 
    WHERE 
      location_type = ? AND
      (nullif(?, 0) IS NULL OR location_id = ?)
  `, locationType, locationID, locationID).Scan(&data).Error

	return data, err
}

func (repo *materialLocationRepository) GetTotalAmountInWarehouse(projectID, materialID uint) (float64, error) {
	var totalAmount float64
	err := repo.db.Raw(`
    SELECT SUM(material_locations.amount)
    FROM materials
      INNER JOIN material_costs ON material_costs.material_id = materials.id
      INNER JOIN material_locations ON material_locations.material_cost_id = material_costs.id
    WHERE 
      material_locations.project_id = ?
      AND material_locations.location_type = 'warehouse'
      AND materials.id = ?
    `, projectID, materialID).Scan(&totalAmount).Error
	return totalAmount, err
}

func (repo *materialLocationRepository) GetUniqueMaterialsFromLocation(
	projectID, locationID uint,
	locationType string,
) ([]model.Material, error) {
	var data []model.Material
	err := repo.db.Raw(`
    SELECT DISTINCT 
      materials.unit,
      materials.name, 
      materials.category, 
      materials.id, 
      materials.code, 
      materials.has_serial_number
    FROM material_locations
      INNER JOIN material_costs ON material_costs.id = material_locations.material_cost_id
      INNER JOIN materials ON materials.id = material_costs.material_id
    WHERE 
      material_locations.project_id = ? AND
      material_locations.location_id = ? AND
      material_locations.location_type = ?
    `, projectID, locationID, locationType).Scan(&data).Error

	return data, err
}

func (repo *materialLocationRepository) GetUniqueMaterialCostsFromLocation(
	projectID, materialID, locationID uint,
	locationType string,
) ([]model.MaterialCost, error) {
	var data []model.MaterialCost
	err := repo.db.Raw(`
    SELECT DISTINCT 
      material_costs.id,
      material_costs.material_id,
      material_costs.cost_m19,
      material_costs.cost_with_customer,
      material_costs.cost_prime
    FROM material_locations
      INNER JOIN material_costs ON material_costs.id = material_locations.material_cost_id
      INNER JOIN materials ON materials.id = material_costs.material_id
    WHERE 
      material_locations.project_id = ? AND
      material_locations.location_type = ? AND
      material_locations.location_id = ? AND
      materials.id = ? 
    `, projectID, locationType, locationID, materialID).Scan(&data).Error

	return data, err
}

func (repo *materialLocationRepository) GetUniqueMaterialTotalAmount(
	projectID, materialCostID, locationID uint,
	locationType string,
) (float64, error) {
	var amount float64
	err := repo.db.Raw(`
    SELECT material_locations.amount
    FROM material_locations
    WHERE
      material_locations.project_id = ? AND
      material_locations.location_type = ? AND
      material_locations.location_id = ? AND
      material_locations.material_cost_id = ?
    `, projectID, locationType, locationID, materialCostID).Scan(&amount).Error

	return amount, err
}

func (repo *materialLocationRepository) GetTotalAmountInLocation(
	projectID, materialID, locationID uint,
	locationType string,
) (float64, error) {
	var data float64
	err := repo.db.Raw(`
      SELECT SUM(material_locations.amount)
      FROM materials
        INNER JOIN material_costs ON material_costs.material_id = materials.id
        INNER JOIN material_locations ON material_locations.material_cost_id = material_costs.id
      WHERE 
        materials.project_id = ? AND
        materials.id = ? AND
        material_locations.location_type = ? AND
        material_locations.location_id = ?;
    `, projectID, materialID, locationType, locationID).Scan(&data).Error

	return data, err
}

func (repo *materialLocationRepository) GetTotalAmountInTeamsByTeamNumber(
	projectID, materialID uint,
	teamNumber string,
) (float64, error) {
	var data float64
	err := repo.db.Raw(`
    SELECT SUM(material_locations.amount)
    FROM material_locations
      INNER JOIN teams ON material_locations.location_id = teams.id
      INNER JOIN material_costs ON material_costs.id = material_locations.material_cost_id
      INNER JOIN materials ON materials.id = material_costs.material_id
    WHERE 
      materials.project_id = ? AND
      materials.id = ? AND
      material_locations.location_type = 'teams' AND
      teams.number = ?
    `, projectID, materialID, teamNumber).Scan(&data).Error

	return data, err
}
