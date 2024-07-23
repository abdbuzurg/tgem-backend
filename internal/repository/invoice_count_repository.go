package repository

import "gorm.io/gorm"

type invoiceCountRepository struct {
	db *gorm.DB
}

func InitInvoiceCountRepository(db *gorm.DB) IInvoiceCountRepository {
	return &invoiceCountRepository{
		db: db,
	}
}

type IInvoiceCountRepository interface {
	CountInvoice(invoiceType string, projectID uint) (uint, error)
}

func (repo *invoiceCountRepository) CountInvoice(invoiceType string, projectID uint) (uint, error) {
	result := uint(0)
	err := repo.db.Raw(`
    SELECT count
    FROM invoice_counts
    WHERE 
      invoice_type = ? AND
      project_id = ?
  `, invoiceType, projectID).Scan(&result).Error

	return result, err
}
