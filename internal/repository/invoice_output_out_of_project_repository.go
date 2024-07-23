package repository

import (
	"backend-v2/internal/dto"
	"backend-v2/model"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type invoiceOutputOutOfProjectRepository struct {
	db *gorm.DB
}

func InitInvoiceOutputOutOfProjectRepository(db *gorm.DB) IInvoiceOutputOutOfProjectRepository {
	return &invoiceOutputOutOfProjectRepository{
		db: db,
	}
}

type IInvoiceOutputOutOfProjectRepository interface {
	GetPaginated(page, limit int, filter dto.InvoiceOutputOutOfProjectSearchParameters) ([]dto.InvoiceOutputOutOfProjectPaginated, error)
	GetByID(id uint) (model.InvoiceOutputOutOfProject, error)
	Create(data dto.InvoiceOutputOutOfProjectCreateQueryData) (model.InvoiceOutputOutOfProject, error)
	Count(filter dto.InvoiceOutputOutOfProjectSearchParameters) (int64, error)
	Confirmation(data dto.InvoiceOutputOutOfProjectConfirmationQueryData) error
	Delete(id uint) error
	Update(data dto.InvoiceOutputOutOfProjectCreateQueryData) (model.InvoiceOutputOutOfProject, error)
	GetMaterialsForEdit(id uint) ([]dto.InvoiceOutputMaterialsForEdit, error)
}

func (repo *invoiceOutputOutOfProjectRepository) GetPaginated(page, limit int, filter dto.InvoiceOutputOutOfProjectSearchParameters) ([]dto.InvoiceOutputOutOfProjectPaginated, error) {
	data := []dto.InvoiceOutputOutOfProjectPaginated{}
	err := repo.db.Raw(`
    SELECT 
      invoice_output_out_of_projects.id as id,
      invoice_output_out_of_projects.from_project_id as from_project_id,
      invoice_output_out_of_projects.to_project_id as to_project_id,
      projects.name as to_project_name,    
      projects.project_manager as to_project_manager,
      invoice_output_out_of_projects.delivery_code as delivery_code,
      workers.name as released_worker_name,
      invoice_output_out_of_projects.date_of_invoice as date_of_invoice,
      invoice_output_out_of_projects.confirmation as confirmation
    FROM invoice_output_out_of_projects 
    INNER JOIN projects ON invoice_output_out_of_projects.to_project_id = projects.id
    INNER JOIN workers ON invoice_output_out_of_projects.released_worker_id = workers.id
    WHERE 
      from_project_id = ? AND
      (nullif(?, 0) IS NULL OR to_project_id = ?) AND
			(nullif(?, 0) IS NULL OR released_worker_id = ?)
    ORDER BY invoice_output_out_of_projects.id 
    LIMIT ? 
    OFFSET ?
    `,
		filter.FromProjectID,
		filter.ToProjectID, filter.ToProjectID,
		filter.ReleasedWorkerID, filter.ReleasedWorkerID,
		limit, (page-1)*limit,
	).Scan(&data).Error

	return data, err
}

func (repo *invoiceOutputOutOfProjectRepository) Create(data dto.InvoiceOutputOutOfProjectCreateQueryData) (model.InvoiceOutputOutOfProject, error) {
	result := data.Invoice
	err := repo.db.Transaction(func(tx *gorm.DB) error {

		if err := tx.Create(&result).Error; err != nil {
			return err
		}

		for index := range data.InvoiceMaterials {
			data.InvoiceMaterials[index].InvoiceID = result.ID
		}

		if err := tx.CreateInBatches(&data.InvoiceMaterials, 15).Error; err != nil {
			return err
		}

		err := tx.Exec(`
      UPDATE invoice_counts
      SET count = count + 1
      WHERE
        invoice_type = 'output' AND
        project_id = ?
      `, result.FromProjectID).Error
		if err != nil {
			return err
		}

		// for index := range data.SerialNumberMovements {
		// 	data.SerialNumberMovements[index].InvoiceID = result.ID
		// }

		// if err := tx.CreateInBatches(&data.SerialNumberMovements, 15).Error; err != nil {
		// 	return err
		// }

		return nil

	})
	return result, err
}

func (repo *invoiceOutputOutOfProjectRepository) Count(filter dto.InvoiceOutputOutOfProjectSearchParameters) (int64, error) {
	var count int64
	err := repo.db.Raw(`
    SELECT COUNT(*) 
    FROM invoice_output_out_of_projects 
    WHERE 
      from_project_id = ? AND
      (nullif(?, 0) IS NULL OR to_project_id = ?) AND
			(nullif(?, 0) IS NULL OR released_worker_id = ?)
    `,
		filter.FromProjectID,
		filter.ToProjectID, filter.ToProjectID,
		filter.ReleasedWorkerID, filter.ReleasedWorkerID,
	).Scan(&count).Error
	return count, err
}

func (repo *invoiceOutputOutOfProjectRepository) Confirmation(data dto.InvoiceOutputOutOfProjectConfirmationQueryData) error {
	return repo.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&model.InvoiceOutputOutOfProject{}).Select("*").Where("id = ?", data.InvoiceData.ID).Updates(&data.InvoiceData).Error; err != nil {
			return err
		}

		if err := tx.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "id"}},
			DoUpdates: clause.AssignmentColumns([]string{"amount"}),
		}).Create(&data.WarehouseMaterials).Error; err != nil {
			return err
		}

		return nil
	})
}

func (repo *invoiceOutputOutOfProjectRepository) Delete(id uint) error {
	return repo.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Delete(&model.InvoiceMaterials{}, "invoice_type = 'output-out-of-project' AND invoice_id = ?", id).Error; err != nil {
			return err
		}

		if err := tx.Delete(&model.InvoiceOutputOutOfProject{}, "id = ?", id).Error; err != nil {
			return err
		}

		return nil
	})
}

func (repo *invoiceOutputOutOfProjectRepository) Update(data dto.InvoiceOutputOutOfProjectCreateQueryData) (model.InvoiceOutputOutOfProject, error) {
	result := data.Invoice
	err := repo.db.Transaction(func(tx *gorm.DB) error {
		err := tx.Model(&model.InvoiceOutput{}).Select("*").Where("id = ?", result.ID).Updates(&result).Error
		if err != nil {
			return err
		}

		if err = tx.Delete(model.InvoiceMaterials{}, "invoice_id = ? AND invoice_type='output-out-of-project'", result.ID).Error; err != nil {
			return nil
		}

		for index := range data.InvoiceMaterials {
			data.InvoiceMaterials[index].InvoiceID = result.ID
		}

		if err := tx.CreateInBatches(&data.InvoiceMaterials, 15).Error; err != nil {
			return err
		}

		return nil

	})

	return result, err
}

func (repo *invoiceOutputOutOfProjectRepository) GetByID(id uint) (model.InvoiceOutputOutOfProject, error) {
	result := model.InvoiceOutputOutOfProject{}
	err := repo.db.First(&result, "id = ?", id).Error
	return result, err
}

func (repo *invoiceOutputOutOfProjectRepository) GetMaterialsForEdit(id uint) ([]dto.InvoiceOutputMaterialsForEdit, error) {
	result := []dto.InvoiceOutputMaterialsForEdit{}
	err := repo.db.Raw(`
    SELECT 
      materials.id as material_id,
      materials.name as material_name,
      materials.unit as material_unit,
      material_locations.amount as warehouse_amount,
      invoice_materials.amount as amount,
      invoice_materials.notes as notes,
      materials.has_serial_number as has_serial_number
    FROM invoice_materials
    INNER JOIN material_costs ON material_costs.id = invoice_materials.material_cost_id
    INNER JOIN materials ON materials.id = material_costs.material_id
    INNER JOIN material_locations ON material_locations.material_cost_id = invoice_materials.material_cost_id
    WHERE
      material_locations.location_type = 'warehouse' AND
      invoice_materials.invoice_type = 'output-out-of-project' AND
      invoice_materials.invoice_id = ?
    `, id).Scan(&result).Error

	return result, err
}
