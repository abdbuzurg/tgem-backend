package repository

import (
	"backend-v2/model"

	"gorm.io/gorm"
)

type invoiceWriteOffRepository struct {
	db *gorm.DB
}

func InitInvoiceWriteOffRepository(db *gorm.DB) IInvoiceWriteOffRepository {
	return &invoiceWriteOffRepository{
		db: db,
	}
}

type IInvoiceWriteOffRepository interface {
	GetAll() ([]model.InvoiceWriteOff, error)
	GetPaginated(page, limit int) ([]model.InvoiceWriteOff, error)
	GetPaginatedFiltered(page, limit int, filter model.InvoiceWriteOff) ([]model.InvoiceWriteOff, error)
	GetByID(id uint) (model.InvoiceWriteOff, error)
	Create(data model.InvoiceWriteOff) (model.InvoiceWriteOff, error)
	Update(data model.InvoiceWriteOff) (model.InvoiceWriteOff, error)
	Delete(id uint) error
	Count() (int64, error)
}

func (repo *invoiceWriteOffRepository) GetAll() ([]model.InvoiceWriteOff, error) {
	data := []model.InvoiceWriteOff{}
	err := repo.db.Order("id desc").Find(&data).Error
	return data, err
}

func (repo *invoiceWriteOffRepository) GetPaginated(page, limit int) ([]model.InvoiceWriteOff, error) {
	data := []model.InvoiceWriteOff{}
	err := repo.db.Order("id desc").Offset((page - 1) * limit).Limit(limit).Find(&data).Error
	return data, err
}

func (repo *invoiceWriteOffRepository) GetPaginatedFiltered(page, limit int, filter model.InvoiceWriteOff) ([]model.InvoiceWriteOff, error) {
	data := []model.InvoiceWriteOff{}
	err := repo.db.
		Raw(`SELECT * FROM invoice_write_offs WHERE
			nullif(?, '') IS NULL OR delivery_code = ?) AND
			(nullif(?, '') IS NULL OR type = ?) AND
			(nullif(?, 0) IS NULL OR operator_add_worker_id = ?) AND
			(nullif(?, 0) IS NULL OR operator_edit_worker_id = ?) ORDER BY id DESC LIMIT ? OFFSET ?`,
			filter.DeliveryCode, filter.DeliveryCode,
			filter.WriteOffType, filter.WriteOffType,
			filter.OperatorAddWorkerID, filter.OperatorAddWorkerID,
			filter.OperatorEditWorkerID, filter.OperatorEditWorkerID,
			limit, (page-1)*limit,
		).
		Scan(&data).Error

	return data, err
}

func (repo *invoiceWriteOffRepository) GetByID(id uint) (model.InvoiceWriteOff, error) {
	data := model.InvoiceWriteOff{}
	err := repo.db.Find(&data, "id = ?", id).Error
	return data, err
}

func (repo *invoiceWriteOffRepository) Create(data model.InvoiceWriteOff) (model.InvoiceWriteOff, error) {
	err := repo.db.Create(&data).Error
	return data, err
}

func (repo *invoiceWriteOffRepository) Update(data model.InvoiceWriteOff) (model.InvoiceWriteOff, error) {
	err := repo.db.Model(&model.InvoiceWriteOff{}).Select("*").Where("id = ?", data.ID).Updates(&data).Error
	return data, err
}

func (repo *invoiceWriteOffRepository) Delete(id uint) error {
	return repo.db.Delete(&model.InvoiceWriteOff{}, "id = ?", id).Error
}

func (repo *invoiceWriteOffRepository) Count() (int64, error) {
	var count int64
	err := repo.db.Model(&model.InvoiceWriteOff{}).Count(&count).Error
	return count, err
}
