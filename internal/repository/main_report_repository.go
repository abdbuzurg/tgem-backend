package repository

import (
	"backend-v2/internal/dto"
	"time"

	"gorm.io/gorm"
)

type mainReportRepository struct {
	db *gorm.DB
}

type IMainReportRepository interface {
	MaterialDataForProgressReportInProject(projectID uint) ([]dto.MaterialDataForProgressReportQueryResult, error)
	InvoiceMaterialDataForProgressReport(projectID uint) ([]dto.InvoiceMaterialDataForProgressReportQueryResult, error)
	MaterialDataForProgressReportInProjectInGivenDate(projectID uint, date time.Time) ([]dto.MaterialDataForProgressReportInGivenDateQueryResult, error)
	InvoiceOperationDataForProgressReport(projectID uint) ([]dto.InvoiceOperationDataForProgressReportQueryResult, error)
	InvoiceOperationDataForProgressReportInGivenDate(projectID uint, date time.Time) ([]dto.InvoiceOperationDataForProgressReportInGivenDataQueryResult, error)
	MaterialDataForRemainingMaterialAnalysis(projectID uint) ([]dto.MaterialDataForRemainingMaterialAnalysisQueryResult, error)
	MaterialsInstalledOnObjectForRemainingMaterialAnalysis(projectID uint) ([]dto.MaterialsInstalledOnObjectForRemainingMaterialAnalysisQueryResult, error)
}

func InitMainReportRepository(db *gorm.DB) IMainReportRepository {
	return &mainReportRepository{
		db: db,
	}
}

func (repo *mainReportRepository) MaterialDataForProgressReportInProject(projectID uint) ([]dto.MaterialDataForProgressReportQueryResult, error) {
	result := []dto.MaterialDataForProgressReportQueryResult{}
	err := repo.db.Raw(`
    SELECT 
      materials.id as id,
      materials.code as code,
      materials.name as name,
      materials.unit as unit,
      materials.planned_amount_for_project as planned_amount_for_project,
      material_locations.amount as location_amount,
      material_locations.location_type as location_type
    FROM material_locations
    INNER JOIN material_costs ON material_locations.material_cost_id = material_costs.id
    INNER JOIN materials ON material_costs.material_id = materials.id
    WHERE
      materials.project_id = ? AND	
      materials.show_planned_amount_in_report = true
    ORDER BY materials.id
    `, projectID).Scan(&result).Error

	return result, err
}

func (repo *mainReportRepository) InvoiceMaterialDataForProgressReport(projectID uint) ([]dto.InvoiceMaterialDataForProgressReportQueryResult, error) {
	result := []dto.InvoiceMaterialDataForProgressReportQueryResult{}
	err := repo.db.Raw(`
    SELECT 
      materials.id as material_id,
      invoice_materials.amount as amount,
      invoice_materials.invoice_type as invoice_type,
      material_costs.cost_with_customer as cost_with_customer
    FROM materials
    INNER JOIN material_costs ON material_costs.material_id = materials.id
    RIGHT JOIN invoice_materials ON invoice_materials.material_cost_id = material_costs.id
    WHERE 
      materials.project_id = ? AND
      materials.show_planned_amount_in_report = true AND
      (invoice_materials.invoice_type = 'input' OR invoice_materials.invoice_type = 'object-correction')
    ORDER BY materials.id
    `, projectID).Scan(&result).Error

	return result, err
}

func (repo *mainReportRepository) InvoiceOperationDataForProgressReport(projectID uint) ([]dto.InvoiceOperationDataForProgressReportQueryResult, error) {
	result := []dto.InvoiceOperationDataForProgressReportQueryResult{}
	err := repo.db.Raw(`
    SELECT 
      operations.id as id,
      operations.code as code,
      operations.name as name,
      operations.cost_with_customer as cost_with_customer,
      operations.planned_amount_for_project as planned_amount_for_project,
      invoice_operations.amount as amount_in_invoice
    FROM invoice_objects
    INNER JOIN invoice_operations ON invoice_operations.invoice_id = invoice_objects.id
    INNER JOIN operations ON operations.id = invoice_operations.operation_id
    WHERE
      invoice_objects.confirmed_by_operator = true AND
      invoice_operations.invoice_type = 'object-correction' AND
      operations.project_id = ? AND
      operations.show_planned_amount_in_report = true
    ORDER BY operations.id
    `, projectID).Scan(&result).Error

	return result, err
}

func (repo *mainReportRepository) MaterialDataForProgressReportInProjectInGivenDate(projectID uint, date time.Time) ([]dto.MaterialDataForProgressReportInGivenDateQueryResult, error) {
	loc, _ := time.LoadLocation("Asia/Dushanbe")
	year, month, day := date.Date()
	midnightOfGivenDate := time.Date(year, month, day, 0, 0, 0, 0, loc)
	midnightOfTomorrowFromGivenDate := midnightOfGivenDate.AddDate(0, 0, 1)
	midnightOfGivenDateStr := midnightOfGivenDate.Format("2006-01-02")
	midnightOfTomorrowFromGivenDateStr := midnightOfTomorrowFromGivenDate.Format("2006-01-02")

	result := []dto.MaterialDataForProgressReportInGivenDateQueryResult{}
	err := repo.db.Raw(`
      SELECT 
        materials.id as id,
        materials.code as code,
        materials.name as name,
        materials.unit as unit,
        materials.planned_amount_for_project as amount_planned_for_project,
        project_progress_materials.received as amount_received,
        project_progress_materials.installed as amount_installed,
        project_progress_materials.amount_in_warehouse as amount_in_warehouse,
        project_progress_materials.amount_in_teams as amount_in_teams,
        project_progress_materials.amount_in_objects as amount_in_objects,
        project_progress_materials.amount_write_off as amount_write_off,
        material_costs.cost_with_customer as cost_with_customer
      FROM project_progress_materials
      INNER JOIN material_costs ON material_costs.id = project_progress_materials.material_cost_id
      INNER JOIN materials ON materials.id = material_costs.material_id
      WHERE 
        project_progress_materials.project_id = ? AND
        ? < project_progress_materials.date AND project_progress_materials.date < ?
      ORDER BY materials.id
    `, projectID, midnightOfGivenDateStr, midnightOfTomorrowFromGivenDateStr).Scan(&result).Error

	return result, err
}

func (repo *mainReportRepository) InvoiceOperationDataForProgressReportInGivenDate(projectID uint, date time.Time) ([]dto.InvoiceOperationDataForProgressReportInGivenDataQueryResult, error) {
	loc, _ := time.LoadLocation("Asia/Dushanbe")
	year, month, day := date.Date()
	midnightOfGivenDate := time.Date(year, month, day, 0, 0, 0, 0, loc)
	midnightOfTomorrowFromGivenDate := midnightOfGivenDate.AddDate(0, 0, 1)
	midnightOfGivenDateStr := midnightOfGivenDate.Format("2006-01-02")
	midnightOfTomorrowFromGivenDateStr := midnightOfTomorrowFromGivenDate.Format("2006-01-02")
	result := []dto.InvoiceOperationDataForProgressReportInGivenDataQueryResult{}
	err := repo.db.Raw(`
    SELECT 
      operations.code as code,
      operations.name as name,
      operations.cost_with_customer as cost_with_customer,
      operations.planned_amount_for_project as amount_planned_for_project,
      project_progress_operations.installed as amount_installed
    FROM project_progress_operations
    INNER JOIN operations ON project_progress_operations.operation_id = operations.id
    WHERE 
      project_progress_operations.project_id = ? AND
      ? < project_progress_operations.date AND project_progress_operations.date < ?
    `, projectID, midnightOfGivenDateStr, midnightOfTomorrowFromGivenDateStr).Scan(&result).Error

	return result, err
}

func (repo *mainReportRepository) MaterialDataForRemainingMaterialAnalysis(projectID uint) ([]dto.MaterialDataForRemainingMaterialAnalysisQueryResult, error) {
	result := []dto.MaterialDataForRemainingMaterialAnalysisQueryResult{}
	err := repo.db.Raw(`
      SELECT 
        materials.id as id,
        materials.code as code,
        materials.name as name,
        materials.unit as unit,
        materials.planned_amount_for_project as planned_amount_for_project,
        material_locations.amount as location_amount,
        material_locations.location_type as location_type
      FROM material_locations
      INNER JOIN material_costs ON material_locations.material_cost_id = material_costs.id
      INNER JOIN materials ON material_costs.material_id = materials.id
      WHERE
        materials.project_id = ? AND
        materials.show_planned_amount_in_report = true AND
        (material_locations.location_type = 'warehouse' OR material_locations.location_type = 'team')
      ORDER BY materials.id
    `, projectID).Scan(&result).Error

	return result, err
}

func (repo *mainReportRepository) MaterialsInstalledOnObjectForRemainingMaterialAnalysis(projectID uint) ([]dto.MaterialsInstalledOnObjectForRemainingMaterialAnalysisQueryResult, error) {
	result := []dto.MaterialsInstalledOnObjectForRemainingMaterialAnalysisQueryResult{}
	err := repo.db.Raw(`
    SELECT 
      materials.id as id,
      invoice_materials.amount as amount,
      invoice_objects.date_of_correction as date_of_correction
    FROM invoice_objects
    INNER JOIN invoice_materials ON invoice_objects.id = invoice_materials.invoice_id
    INNER JOIN material_costs ON invoice_materials.material_cost_id = material_costs.id
    INNER JOIN materials ON material_costs.material_id = materials.id
    WHERE 
      materials.project_id = ? AND
      materials.show_planned_amount_in_report = true
      invoice_objects.confirmed_by_operator = true AND
      invoice_materials.invoice_type = 'object-correction' 
    ORDER BY materials.id
    `, projectID).Scan(&result).Error

	return result, err
}
