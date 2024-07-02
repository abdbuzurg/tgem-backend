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
	GetPaginated(page, limit int, filter model.InvoiceOutputOutOfProject) ([]dto.InvoiceOutputOutOfProjectPaginated, error)
	Create(data dto.InvoiceOutputOutOfProjectCreateQueryData) (model.InvoiceOutputOutOfProject, error)
	Count(projectID uint) (int64, error)
	Confirmation(data dto.InvoiceOutputOutOfProjectConfirmationQueryData) error
}

func (repo *invoiceOutputOutOfProjectRepository) GetPaginated(page, limit int, filter model.InvoiceOutputOutOfProject) ([]dto.InvoiceOutputOutOfProjectPaginated, error) {
	data := []dto.InvoiceOutputOutOfProjectPaginated{}
	err := repo.db.Raw(`
    SELECT 
      invoice_output_out_of_projects.id as id,
      invoice_output_out_of_projects.delivery_code as delivery_code,
      released.name as released_name,
      invoice_output_out_of_projects.date_of_invoice as date_of_invoice,
      invoice_output_out_of_projects.confirmation as confirmation,
      invoice_output_out_of_projects.notes as notes
    FROM invoice_outputs
      INNER JOIN workers AS released ON released.id = invoice_outputs.released_worker_id
    WHERE
      invoice_output_out_of_projects.project_id = ? AND
      (nullif(?, 0) IS NULL OR invoice_outputs.released_worker_id = ?) AND
      (nullif(?, '') IS NULL OR invoice_outputs.delivery_code = ?) 
    ORDER BY invoice_output_out_of_projects.id DESC LIMIT ? OFFSET ?;
  `,
		filter.ProjectID,
		filter.ReleasedWorkerID, filter.ReleasedWorkerID,
		filter.DeliveryCode, filter.DeliveryCode,
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

		for index := range data.SerialNumberMovements {
			data.SerialNumberMovements[index].InvoiceID = result.ID
		}

		if err := tx.CreateInBatches(&data.SerialNumberMovements, 15).Error; err != nil {
			return err
		}

		return nil

	})
	return result, err
}

func (repo *invoiceOutputOutOfProjectRepository) Count(projectID uint) (int64, error) {
	var count int64
	err := repo.db.Raw("SELECT COUNT(*) FROM invoice_output_out_of_projects WHERE project_id = ?", projectID).Scan(&count).Error
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

		if err := tx.Exec(`
        UPDATE serial_number_movements
        SET confirmation = true
        WHERE 
          serial_number_movements.invoice_type = 'output-out-of-project' AND
          serial_number_movements.invoice_id = ?
      `, data.InvoiceData.ID).Error; err != nil {
			return err
		}

		return nil
	})
}
