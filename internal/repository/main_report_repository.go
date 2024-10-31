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
      material_locations.amount <> 0 
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
