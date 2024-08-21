package repository

import (
	"backend-v2/internal/dto"
	"backend-v2/model"
	"errors"

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
  GetByLocationType(locationType string) ([]model.MaterialLocation, error)
	Create(data model.MaterialLocation) (model.MaterialLocation, error)
	Update(data model.MaterialLocation) (model.MaterialLocation, error)
	Delete(id uint) error
	Count() (int64, error)
	GetUniqueMaterialCostsByLocation(locationType string, locationID uint) ([]uint, error)
	UniqueObjects(projectID uint) ([]dto.ObjectDataForSelect, error)
	UniqueTeams(projectID uint) ([]dto.TeamDataForSelect, error)
	GetByLocationTypeAndID(locationType string, locationID uint) ([]model.MaterialLocation, error)
	GetTotalAmountInWarehouse(projectID, materialID uint) (float64, error)
	GetUniqueMaterialsFromLocation(projectID, locationID uint, locationType string) ([]model.Material, error)
	GetUniqueMaterialCostsFromLocation(projectID, materialID, locationID uint, locationType string) ([]model.MaterialCost, error)
	GetUniqueMaterialTotalAmount(projectID, materialCostID, locationID uint, locationType string) (float64, error)
	GetTotalAmountInLocation(projectID, materialID, locationID uint, locationType string) (float64, error)
	GetTotalAmountInTeamsByTeamNumber(projectID, materialID uint, teamNumber string) (float64, error)
	GetDataForBalanceReport(projectID uint, locationType string, locationID uint) ([]dto.BalanceReportQueryResult, error)
	GetMaterialAmountSortedByCostM19InLocation(projectID, materialID uint, locationType string, locationID uint) ([]dto.MaterialAmountSortedByCostM19QueryResult, error)
	GetMaterialsInLocationBasedOnInvoiceID(locationID uint, locationType string, invoiceID uint, invoiceType string) ([]model.MaterialLocation, error)
	Live(data dto.MaterialLocationLiveSearchParameters) ([]dto.MaterialLocationLiveView, error)
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

func (repo *materialLocationRepository) UniqueObjects(projectID uint) ([]dto.ObjectDataForSelect, error) {
	data := []dto.ObjectDataForSelect{}
	err := repo.db.Raw(`
    SELECT 
      objects.id as id,
      objects.name as object_name,
      objects.type as object_type
    FROM objects
    WHERE objects.id IN (
      SELECT DISTINCT(location_id)
      FROM material_locations
      WHERE 
      location_type='object' AND 
      amount > 0 AND
      material_locations.project_id = ?
    )
    `, projectID).Scan(&data).Error

	return data, err
}

func (repo *materialLocationRepository) UniqueTeams(projectID uint) ([]dto.TeamDataForSelect, error) {
	data := []dto.TeamDataForSelect{}
	err := repo.db.Raw(`
      SELECT 
        teams.id,
        teams.number as team_number,
        workers.name as team_leader_name
      FROM teams
      INNER JOIN team_leaders ON team_leaders.team_id = teams.id
      INNER JOIN workers ON team_leaders.leader_worker_id = workers.id
      WHERE teams.id IN (
        SELECT DISTINCT(location_id)
        FROM material_locations
        WHERE 
          location_type='team' AND 
          amount > 0 AND
          material_locations.project_id = ?
      )
    `, projectID).Scan(&data).Error

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
      location_id = ?  
    
  `, locationType, locationID).Scan(&data).Error

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
      material_locations.location_type = ? AND
      material_locations.amount > 0
    `, projectID, locationID, locationType).Scan(&data).Error

	return data, err
}

func (repo *materialLocationRepository) GetUniqueMaterialCostsFromLocation(projectID, materialID, locationID uint, locationType string) ([]model.MaterialCost, error) {
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
	data := float64(0)
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

func (repo *materialLocationRepository) GetDataForBalanceReport(projectID uint, locationType string, locationID uint) ([]dto.BalanceReportQueryResult, error) {
	data := []dto.BalanceReportQueryResult{}
	err := repo.db.Raw(`
    SELECT 
      material_locations.location_id AS location_id,
      materials.code AS material_code,
      materials.name AS material_name,
      materials.unit AS material_unit,
      material_locations.amount AS total_amount,
      material_defects.amount AS defect_amount,
      material_costs.cost_m19 AS material_cost_m19,
      material_locations.amount * material_costs.cost_m19 AS total_cost,
      material_defects.amount * material_costs.cost_m19 AS total_defect_cost
    FROM material_locations
    INNER JOIN material_costs ON material_costs.id = material_locations.material_cost_id
    INNER JOIN materials ON materials.id = material_costs.material_id
    LEFT JOIN material_defects ON material_defects.material_location_id = material_locations.id
    WHERE 
      material_locations.project_id = ? AND
      material_locations.location_type = ? AND
      (nullif(?, 0) IS NULL OR material_locations.location_id = ?) AND
      material_locations.amount <> 0
    ORDER BY material_locations.id
    `,
		projectID,
		locationType,
		locationID, locationID).Scan(&data).Error

	return data, err
}

func (repo *materialLocationRepository) GetMaterialAmountSortedByCostM19InLocation(projectID, materialID uint, locationType string, locationID uint) ([]dto.MaterialAmountSortedByCostM19QueryResult, error) {
	data := []dto.MaterialAmountSortedByCostM19QueryResult{}
	err := repo.db.Raw(`
    SELECT 
      materials.id AS material_id,
      material_costs.id AS material_cost_id,
      material_costs.cost_m19 AS material_cost_m19,
      material_locations.amount AS material_amount
    FROM material_locations
    INNER JOIN material_costs ON material_costs.id = material_locations.material_cost_id
    INNER JOIN materials ON materials.id = material_costs.material_id
    WHERE 
      material_locations.project_id = ? AND
      material_locations.location_type = ? AND
      material_locations.location_id = ? AND
      materials.id = ? AND
      material_locations.amount > 0
    ORDER BY material_costs.cost_m19 DESC;
  `, projectID, locationType, locationID, materialID).Scan(&data).Error

	return data, err
}

func (repo *materialLocationRepository) GetMaterialsInLocationBasedOnInvoiceID(locationID uint, locationType string, invoiceID uint, invoiceType string) ([]model.MaterialLocation, error) {
	data := []model.MaterialLocation{}
	err := repo.db.Raw(`
    SELECT *
    FROM material_locations
    WHERE 
      material_locations.location_type = ? AND
      material_locations.location_id = ? AND
      material_locations.material_cost_id IN (
        SELECT invoice_materials.material_cost_id
        FROM invoice_materials
        WHERE 
          invoice_materials.invoice_type = ? AND
          invoice_materials.invoice_id = ?
      )
    `, locationType, locationID, invoiceType, invoiceID).Scan(&data).Error

	return data, err
}

func (repo *materialLocationRepository) Live(data dto.MaterialLocationLiveSearchParameters) ([]dto.MaterialLocationLiveView, error) {
	result := []dto.MaterialLocationLiveView{}
	err := repo.db.Raw(`
    SELECT 
      materials.id as material_id,
      materials.name as material_name,
      materials.unit as material_unit,
      material_costs.id as material_cost_id,
      material_costs.cost_m19 as material_cost_m19,
      material_locations.location_type as location_type,
      material_locations.location_id as location_id,
      material_locations.amount as amount
    FROM material_locations
    INNER JOIN material_costs ON material_costs.id = material_locations.material_cost_id
    INNER JOIN materials ON materials.id = material_costs.material_id
    WHERE 
      material_locations.location_type = ? AND
      material_locations.project_id = ? AND
      (NULLIF(?, 0) IS NULL OR material_locations.location_id = ?) AND
      (NULLIF(?, 0) IS NULL OR materials.id = ?)
    `,
		data.LocationType,
		data.ProjectID,
		data.LocationID, data.LocationID,
		data.MaterialID, data.MaterialID,
	).Scan(&result).Error

	return result, err
}

func (repo *materialLocationRepository) GetByLocationType(locationType string) ([]model.MaterialLocation, error) {
  result := []model.MaterialLocation{}
  err := repo.db.Find(&result, "location_type = ?", locationType).Error
  if errors.Is(err, gorm.ErrRecordNotFound)  {
    return result, nil
  }

  return result, err
}
