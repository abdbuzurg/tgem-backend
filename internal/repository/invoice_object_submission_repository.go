package repository

import (
	"backend-v2/model"

	"gorm.io/gorm"
)

type invoiceObjectSubmissionRepository struct {
	db *gorm.DB
}

func InitInvoiceObjectSubmissionRepository(db *gorm.DB) IInvoiceObjectSubmissionRepository {
	return &invoiceObjectSubmissionRepository{
		db: db,
	}
}

type IInvoiceObjectSubmissionRepository interface {
	GetPaginated(page, limit int, projectID uint) ([]model.InvoiceObjectSubmission, error)
	Create(data model.InvoiceObjectSubmission) (model.InvoiceObjectSubmission, error)
	Delete(id uint) error
}

func (repo *invoiceObjectSubmissionRepository) GetPaginated(page, limit int, projectID uint) ([]model.InvoiceObjectSubmission, error) {
	var data []model.InvoiceObjectSubmission
	err := repo.db.Order("id desc").Offset((page-1)*limit).Limit(limit).Find(&data, "project_id = ?", projectID).Error
	return data, err
}

func (repo *invoiceObjectSubmissionRepository) Create(data model.InvoiceObjectSubmission) (model.InvoiceObjectSubmission, error) {
	err := repo.db.Create(&data).Error
	return data, err
}

func (repo *invoiceObjectSubmissionRepository) Delete(id uint) error {
	err := repo.db.Delete(&model.InvoiceObjectSubmission{}, "id = ?", id).Error
	return err
}
