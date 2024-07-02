package repository

import (
	"backend-v2/internal/dto"
	"backend-v2/model"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
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
	GetSerialNumberOfMaterialInTeam(projectID uint, materialID uint, teamID uint) ([]string, error)
	Create(data dto.InvoiceCorrectionCreateQuery) (model.InvoiceObject, error)
}

func (repo *invoiceCorrectionRepository) GetInvoiceMaterialsDataByInvoiceObjectID(id uint) ([]dto.InvoiceCorrectionMaterialsData, error) {
	data := []dto.InvoiceCorrectionMaterialsData{}
	err := repo.db.Raw(`
    SELECT 
      invoice_materials.id as invoice_material_id,
      materials.name as material_name,
      materials.id as material_id,
      invoice_materials.notes as notes,
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

func (repo *invoiceCorrectionRepository) GetSerialNumberOfMaterialInTeam(projectID uint, materialID uint, teamID uint) ([]string, error) {
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
      teams.id = ? AND
      material_locations.location_type = serial_number_locations.location_type AND
      material_locations.location_id = serial_number_locations.location_id
    `, projectID, materialID, teamID).Scan(&data).Error

	return data, err
}

func (repo *invoiceCorrectionRepository) Create(data dto.InvoiceCorrectionCreateQuery) (model.InvoiceObject, error) {
	result := data.Details
	err := repo.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&model.InvoiceObject{}).Select("*").Where("id = ?", result.ID).Updates(&result).Error; err != nil {
			return err
		}

		for index := range data.Items {
			data.Items[index].InvoiceID = result.ID
		}

		if err := tx.CreateInBatches(&data.Items, 15).Error; err != nil {
			return err
		}

		if err := tx.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "id"}},
			DoUpdates: clause.AssignmentColumns([]string{"amount"}),
		}).Create(&data.TeamLocation).Error; err != nil {
			return err
		}

		if err := tx.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "id"}},
			DoUpdates: clause.AssignmentColumns([]string{"amount"}),
		}).Create(&data.ObjectLocation).Error; err != nil {
			return err
		}

		return nil
	})

	return result, err
}
