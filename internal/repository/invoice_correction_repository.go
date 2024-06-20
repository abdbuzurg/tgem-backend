package repository

import (
	"backend-v2/internal/dto"

	"gorm.io/gorm"
)

type invoiceCorrectionRepository struct {
	db *gorm.DB
}

func InitInvoiceCorrectionRepository(db *gorm.DB) IInvoiceCorrectionRepository {
	return &invoiceCorrectionRepository{
		db: db,
	}
}

type IInvoiceCorrectionRepository interface {
	GetInvoiceMaterialsDataByInvoiceObjectID(id uint) ([]dto.InvoiceCorrectionMaterialsData, error)
  GetSerialNumberOfMaterialInTeam(projectID uint, materialID uint, teamNumber string) ([]string, error)
}

func (repo *invoiceCorrectionRepository) GetInvoiceMaterialsDataByInvoiceObjectID(id uint) ([]dto.InvoiceCorrectionMaterialsData, error) {
	data := []dto.InvoiceCorrectionMaterialsData{}
	err := repo.db.Raw(`
    SELECT 
      invoice_materials.id as invoice_material_id,
      materials.name as material_name,
      material_costs.cost_m19 as material_cost,
      invoice_materials.amount as material_amount
    FROM invoice_materials
    INNER JOIN material_costs ON material_costs.id = invoice_materials.material_cost_id
    INNER JOIN materials ON materials.id = material_costs.material_id
    WHERE 
      invoice_materials.invoice_type = 'object' AND
      invoice_materials.invoice_id = ?
    `, id).Scan(&data).Error

	return data, err
}

func (repo *invoiceCorrectionRepository) GetSerialNumberOfMaterialInTeam(projectID uint, materialID uint, teamNumber string) ([]string, error) {
	data := []string{}
	err := repo.db.Raw(`
    SELECT serial_numbers.code
    FROM material_locations
    INNER JOIN teams ON team.id = material_locations.location_id
    INNER JOIN serial_numbers ON serial_numbers.material_cost_id = material_locations.material_cost_id
    INNER JOIN serial_number_locations ON serial_number_locations.serial_number_id = serial_numbers.id
    INNER JOIN material_costs ON material_locations.material_cost_id = material_costs.id
    INNER JOIN materials ON materials.id = material_costs.material_id
    WHERE
      materials.project_id = ? AND
      materials.id = ? AND
      teams.number = ? AND
      material_locations.location_type = serial_number_locations.location_type AND
      material_locations.location_id = serial_number_locations.location_id AND
    `, projectID, materialID, teamNumber).Scan(&data).Error

	return data, err
}
