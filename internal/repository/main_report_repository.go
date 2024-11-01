package repository

import (
	"backend-v2/internal/dto"

	"gorm.io/gorm"
)

type mainReportRepository struct {
	db *gorm.DB
}

type IMainReportRepository interface {
	MaterialDataForProgressReportInProject(projectID uint) ([]dto.MaterialDataForProgressReportQueryResult, error)
	InvoiceMaterialDataForProgressReport(projectID uint) ([]dto.InvoiceMaterialDataForProgressReportQueryResult, error)
	InvoiceOperationDataForProgressReport(projectID uint) ([]dto.InvoiceOperationDataForProgressReportQueryResult, error)
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
      material_locations.location_type as location_type,
      material_costs.cost_with_customer * material_locations.amount as sum_with_customer_in_location
    FROM material_locations
    INNER JOIN material_costs ON material_locations.material_cost_id = material_costs.id
    INNER JOIN materials ON material_costs.material_id = materials.id
    WHERE
      materials.project_id = ? AND	
      materials.show_planned_amount_in_report = true AND
      (material_locations.location_type = 'warehouse' OR material_locations.location_type = 'team' OR material_locations.location_type = 'object')
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
      invoice_materials.amount * material_costs.cost_with_customer as sum_in_invoice
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
      invoice_objects.confirmed_by_operator = true AND
      invoice_materials.invoice_type = 'object-correction' AND
      materials.project_id = ? AND
      materials.show_planned_amount_in_report = true
    ORDER BY materials.id
    `, projectID).Scan(&result).Error

	return result, err
}
