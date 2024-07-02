package repository

import (
	"backend-v2/internal/dto"
	"backend-v2/model"

	"gorm.io/gorm"
)

type invoiceObjectRepository struct {
	db *gorm.DB
}

func InitInvoiceObjectRepository(db *gorm.DB) IInvoiceObjectRepository {
	return &invoiceObjectRepository{
		db: db,
	}
}

type IInvoiceObjectRepository interface {
	Create(data dto.InvoiceObjectCreateQueryData) (model.InvoiceObject, error)
	Delete(id uint) error
	Count(projectID uint) (int64, error)
	GetInvoiceObjectDescriptiveDataByID(id uint) (dto.InvoiceObjectPaginated, error)
	GetPaginated(page, limit int, projectID uint) ([]dto.InvoiceObjectPaginated, error)
	GetByID(id uint) (model.InvoiceObject, error)
	GetForCorrection(projectID uint) ([]dto.InvoiceCorrectionPaginated, error)
	GetTeamsFromObjectID(objectID uint) ([]model.Team, error)
}

func (repo *invoiceObjectRepository) GetInvoiceObjectDescriptiveDataByID(id uint) (dto.InvoiceObjectPaginated, error) {
	data := dto.InvoiceObjectPaginated{}
	err := repo.db.Raw(`
    SELECT 
      invoice_objects.id as id,
      workers.name as supervisor_name,
      objects.name as object_name,
      teams.number as team_number,
      invoice_objects.date_of_invoice as date_of_invoice,
      invoice_objects.delivery_code as delivery_code,
      invoice_objects.confirmed_by_operator as confirmed_by_operator  
    FROM invoice_objects
      INNER JOIN workers ON workers.id = invoice_objects.supervisor_worker_id
      INNER JOIN objects ON objects.id = invoice_objects.object_id
      INNER JOIN teams ON teams.id = invoice_objects.team_id
    WHERE
      invoice_objects.id = ?
    `, id).Scan(&data).Error

	return data, err
}

func (repo *invoiceObjectRepository) Create(data dto.InvoiceObjectCreateQueryData) (model.InvoiceObject, error) {
	invoice := data.Invoice
	err := repo.db.Transaction(func(tx *gorm.DB) error {

		if err := tx.Create(&invoice).Error; err != nil {
			return err
		}

		for index := range data.InvoiceMaterials {
			data.InvoiceMaterials[index].InvoiceID = invoice.ID
		}

		if err := tx.CreateInBatches(&data.InvoiceMaterials, 15).Error; err != nil {
			return err
		}

		for index := range data.SerialNumberMovements {
			data.SerialNumberMovements[index].InvoiceID = invoice.ID
		}

		if err := tx.CreateInBatches(&data.SerialNumberMovements, 15).Error; err != nil {
			return err
		}

		return nil
	})

	return invoice, err
}

func (repo *invoiceObjectRepository) Delete(id uint) error {
	err := repo.db.Delete(&model.InvoiceObject{}, "id = ?", id).Error
	return err
}

func (repo *invoiceObjectRepository) Count(projectID uint) (int64, error) {
	var count int64
	err := repo.db.Raw("SELECT COUNT(*) FROM invoice_objects WHERE project_id = ?", projectID).Scan(&count).Error
	return count, err
}

func (repo *invoiceObjectRepository) GetPaginated(page, limit int, projectID uint) ([]dto.InvoiceObjectPaginated, error) {
	data := []dto.InvoiceObjectPaginated{}
	err := repo.db.Raw(`
    SELECT 
      invoice_objects.id as id,
      workers.name as supervisor_name,
      objects.name as object_name,
      teams.number as team_number,
      invoice_objects.date_of_invoice as date_of_invoice,
      invoice_objects.delivery_code as delivery_code,
      invoice_objects.confirmed_by_operator as confirmed_by_operator  

    FROM invoice_objects
      INNER JOIN workers ON workers.id = invoice_objects.supervisor_worker_id
      INNER JOIN objects ON objects.id = invoice_objects.object_id
      INNER JOIN teams ON teams.id = invoice_objects.team_id
    WHERE
      invoice_objects.project_id = ?
      ORDER BY invoice_objects.id DESC LIMIT ? OFFSET ?
    `, projectID, limit, (page-1)*limit).Scan(&data).Error

	return data, err
}

func (repo *invoiceObjectRepository) GetByID(id uint) (model.InvoiceObject, error) {
	data := model.InvoiceObject{}
	err := repo.db.Raw(`
    SELECT *    
    FROM invoice_objects
    WHERE invoice_objects.id = ?
  `, id).Scan(&data).Error

	return data, err
}

func (repo *invoiceObjectRepository) GetForCorrection(projectID uint) ([]dto.InvoiceCorrectionPaginated, error) {
	data := []dto.InvoiceCorrectionPaginated{}
	err := repo.db.Raw(`
    SELECT 
      invoice_objects.id as id,
      workers.name as supervisor_name,
      objects.name as object_name,
      teams.id as team_id,
      teams.number as team_number,
      invoice_objects.date_of_invoice as date_of_invoice,
      invoice_objects.delivery_code as delivery_code,
      invoice_objects.confirmed_by_operator as confirmed_by_operator  

    FROM invoice_objects
      INNER JOIN workers ON workers.id = invoice_objects.supervisor_worker_id
      INNER JOIN objects ON objects.id = invoice_objects.object_id
      INNER JOIN teams ON teams.id = invoice_objects.team_id
    WHERE
      invoice_objects.project_id = ? AND
      invoice_objects.confirmed_by_operator = false
    `, projectID).Scan(&data).Error

	return data, err
}

func (repo *invoiceObjectRepository) GetTeamsFromObjectID(objectID uint) ([]model.Team, error) {
	data := []model.Team{}
	err := repo.db.Raw(`
    SELECT 
      teams.id,
      teams.project_id,
      teams.number,
      teams.mobile_number,
      teams.company 
    FROM object_teams
    INNER JOIN teams ON teams.id = object_teams.team_id
    WHERE object_teams.object_id = ?;
    `, objectID).Scan(&data).Error

	return data, err
}
