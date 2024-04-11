package repository

import (
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
	GetPaginated(page, limit int, projectID, workerID uint) ([]model.InvoiceObject, error)
	Create(data model.InvoiceObject) (model.InvoiceObject, error)
	Delete(id uint) error
}

func (repo *invoiceObjectRepository) GetPaginated(page, limit int, projectID, workerID uint) ([]model.InvoiceObject, error) {
	var data []model.InvoiceObject
	err := repo.db.
    Order("id desc").
    Offset((page-1)*limit).
    Limit(limit).
    Find(&data, "project_id = ? AND (nullif(?, 0) OR supervisor_worker_id = ?)", projectID, workerID, workerID).
    Error
	return data, err
}

func (repo *invoiceObjectRepository) Create(data model.InvoiceObject) (model.InvoiceObject, error) {
	err := repo.db.Create(&data).Error
	return data, err
}

func (repo *invoiceObjectRepository) Delete(id uint) error {
	err := repo.db.Delete(&model.InvoiceObject{}, "id = ?", id).Error
	return err
}
