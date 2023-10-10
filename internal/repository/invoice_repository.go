package repository

import (
	"backend-v2/model"

	"gorm.io/gorm"
)

type invoiceRepository struct {
	db *gorm.DB
}

func InitInvoiceRepository(db *gorm.DB) IInvoiceRepository {
	return &invoiceRepository{
		db: db,
	}
}

type IInvoiceRepository interface {
	GetAll(invoiceType string) ([]model.Invoice, error)
	GetPaginated(invoiceType string, page, limit int) ([]model.Invoice, error)
	GetPaginatedFiltered(invoiceType string, page, limit int, filter model.Invoice) ([]model.Invoice, error)
	GetByID(id uint) (model.Invoice, error)
	Create(data model.Invoice) (model.Invoice, error)
	Update(data model.Invoice) (model.Invoice, error)
	Delete(id uint) error
	Count(invoiceType string) (int64, error)
}

func (repo *invoiceRepository) GetAll(invoiceType string) ([]model.Invoice, error) {
	data := []model.Invoice{}
	err := repo.db.Order("id desc").Find(&data, "invoice_type = ?", invoiceType).Error
	return data, err
}

func (repo *invoiceRepository) GetPaginated(invoiceType string, page, limit int) ([]model.Invoice, error) {
	data := []model.Invoice{}
	err := repo.db.Order("id desc").Offset((page-1)*limit).Limit(limit).Find(&data, "invoice_type = ?", invoiceType).Error
	return data, err
}

func (repo *invoiceRepository) GetPaginatedFiltered(invoiceType string, page, limit int, filter model.Invoice) ([]model.Invoice, error) {
	data := []model.Invoice{}
	err := repo.db.
		Raw(`SELECT * FROM invoices WHERE
			(nullif(?, '') IS NULL OR project_id = ?) AND
			(nullif(?, '') IS NULL OR team_id = ?) AND
			(nullif(?, '') IS NULL OR warehouse_manager_worker_id = ?) AND
			(nullif(?, '') IS NULL OR released_worker_id = ?) AND
			(nullif(?, '') IS NULL OR driver_worker_id = ?) AND
			(nullif(?, '') IS NULL OR recipient_worker_id = ?) AND
			(nullif(?, '') IS NULL OR operator_add_worker_id = ?) AND
			(nullif(?, '') IS NULL OR operator_edit_worker_id = ?) AND
			(nullif(?, '') IS NULL OR object_id = ?) AND
			(nullif(?, '') IS NULL OR invoice_type = ?) AND
			(nullif(?, '') IS NULL OR delivery_code = ?) AND
			(nullif(?, '') IS NULL OR district = ?) AND
			(nullif(?, '') IS NULL OR car_number = ?) AND
			(nullif(?, '') IS NULL OR notes = ?)  ORDER BY id DESC LIMIT ? OFFSET ?`,
			filter.ProjectID, filter.ProjectID,
			filter.TeamID, filter.TeamID,
			filter.WarehouseManagerWorkerID, filter.WarehouseManagerWorkerID,
			filter.ReleasedWorkerID, filter.ReleasedWorkerID,
			filter.DriverWorkerID, filter.DriverWorkerID,
			filter.RecipientWorkerID, filter.RecipientWorkerID,
			filter.OperatorAddWorkerID, filter.OperatorAddWorkerID,
			filter.OperatorEditWorkerID, filter.OperatorEditWorkerID,
			filter.ObjectID, filter.ObjectID,
			invoiceType, invoiceType,
			filter.DeliveryCode, filter.DeliveryCode,
			filter.District, filter.District,
			filter.CarNumber, filter.CarNumber,
			filter.Notes, filter.Notes,
			limit, (page-1)*limit,
		).
		Scan(&data).Error

	return data, err
}

func (repo *invoiceRepository) GetByID(id uint) (model.Invoice, error) {
	data := model.Invoice{}
	err := repo.db.Find(&data, "id = ?", id).Error
	return data, err
}

func (repo *invoiceRepository) Create(data model.Invoice) (model.Invoice, error) {
	err := repo.db.Create(&data).Error
	return data, err
}

func (repo *invoiceRepository) Update(data model.Invoice) (model.Invoice, error) {
	err := repo.db.Model(&model.Invoice{}).Select("*").Updates(&data).Error
	return data, err
}

func (repo *invoiceRepository) Delete(id uint) error {
	return repo.db.Delete(&model.Invoice{}, "id = ?", id).Error
}

func (repo *invoiceRepository) Count(invoiceType string) (int64, error) {
	var count int64
	err := repo.db.Raw("SELECT COUNT(*) FROM invoices WHERE invoice_type = ?", invoiceType).Scan(&count).Error
	return count, err
}
