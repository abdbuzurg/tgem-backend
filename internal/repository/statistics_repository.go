package repository

import (
	"backend-v2/internal/dto"

	"gorm.io/gorm"
)

type statisticsRepository struct {
	db *gorm.DB
}

type IStatisticsRepository interface {
	CountInvoiceInputs(projectID uint) (int, error)
	CountInvoiceOutputs(projectID uint) (int, error)
	CountInvoiceReturns(projectID uint) (int, error)
	CountInvoiceWriteOffs(projectID uint) (int, error)
	CountInvoiceInputUniqueCreators(projectID uint) ([]int, error)
	CountInvoiceInputCreatorInvoices(projectID, workerID uint) (int, error)
	CountInvoiceOutputUniqueCreators(projectID uint) ([]int, error)
	CountInvoiceOutputCreatorInvoices(projectID, workerID uint) (int, error)
	CountMaterialInInvoices(materialID uint) ([]dto.InvoiceMaterialStats, error)
	CountMaterialInLocations(materialID uint) ([]dto.LocationMaterialStats, error)
}

func NewStatisticsRepository(db *gorm.DB) IStatisticsRepository {
	return &statisticsRepository{
		db: db,
	}
}

func (repo *statisticsRepository) CountInvoiceInputs(projectID uint) (int, error) {
	count := 0
	err := repo.db.Raw(`SELECT COUNT(*) FROM invoice_inputs WHERE project_id = ?`, projectID).Scan(&count).Error
	return count, err
}

func (repo *statisticsRepository) CountInvoiceOutputs(projectID uint) (int, error) {
	count := 0
	err := repo.db.Raw(`SELECT COUNT(*) FROM invoice_outputs WHERE project_id = ?`, projectID).Scan(&count).Error
	return count, err
}

func (repo *statisticsRepository) CountInvoiceReturns(projectID uint) (int, error) {
	count := 0
	err := repo.db.Raw(`SELECT COUNT(*) FROM invoice_returns WHERE project_id = ?`, projectID).Scan(&count).Error
	return count, err
}

func (repo *statisticsRepository) CountInvoiceWriteOffs(projectID uint) (int, error) {
	count := 0
	err := repo.db.Raw(`SELECT COUNT(*) FROM invoice_write_offs WHERE project_id = ?`, projectID).Scan(&count).Error
	return count, err
}

func (repo *statisticsRepository) CountInvoiceInputUniqueCreators(projectID uint) ([]int, error) {
	uniqueCreators := []int{}
	err := repo.db.Raw(`SELECT DISTINCT released_worker_id FROM invoice_inputs WHERE project_id = ?`, projectID).Scan(&uniqueCreators).Error
	return uniqueCreators, err
}

func (repo *statisticsRepository) CountInvoiceInputCreatorInvoices(projectID, workerID uint) (int, error) {
	count := 0
	err := repo.db.Raw(`SELECT COUNT(*) FROM invoice_inputs WHERE  project_id = ? AND released_worker_id = ?`, projectID, workerID).Scan(&count).Error
	return count, err
}

func (repo *statisticsRepository) CountInvoiceOutputUniqueCreators(projectID uint) ([]int, error) {
	uniqueCreators := []int{}
	err := repo.db.Raw(`SELECT DISTINCT released_worker_id FROM invoice_outputs WHERE project_id = ?`, projectID).Scan(&uniqueCreators).Error
	return uniqueCreators, err
}

func (repo *statisticsRepository) CountInvoiceOutputCreatorInvoices(projectID, workerID uint) (int, error) {
	count := 0
	err := repo.db.Raw(`SELECT COUNT(*) FROM invoice_outputs WHERE  project_id = ? AND released_worker_id = ?`, projectID, workerID).Scan(&count).Error
	return count, err
}

func (repo *statisticsRepository) CountMaterialInInvoices(materialID uint) ([]dto.InvoiceMaterialStats, error) {
	result := []dto.InvoiceMaterialStats{}
	err := repo.db.Raw(`
    SELECT 
      invoice_materials.amount as amount,
      invoice_materials.invoice_type as invoice_type
    FROM invoice_materials
    INNER JOIN material_costs ON invoice_materials.material_cost_id = material_costs.id
    INNER JOIN materials ON materials.id = material_costs.material_id 
    WHERE materials.id = ?;
    `, materialID).Scan(&result).Error

	return result, err
}

func (repo *statisticsRepository) CountMaterialInLocations(materialID uint) ([]dto.LocationMaterialStats, error) {
	result := []dto.LocationMaterialStats{}
	err := repo.db.Raw(`
    SELECT 
      material_locations.amount as amount,
      material_locations.location_type as location_type
    FROM material_locations
    INNER JOIN material_costs ON material_locations.material_cost_id = material_costs.id
    INNER JOIN materials ON materials.id = material_costs.material_id 
    WHERE materials.id = ?;
    `, materialID).Scan(&result).Error

  return result, err
}
