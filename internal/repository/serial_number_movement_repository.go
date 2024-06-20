package repository

import (
	"backend-v2/model"

	"gorm.io/gorm"
)

type serialNumberMovementRepository struct {
	db *gorm.DB
}

func InitSerialNumberMovementRepository(db *gorm.DB) ISerialNumberMovementRepository {
	return &serialNumberMovementRepository{
		db: db,
	}
}

type ISerialNumberMovementRepository interface {
	CreateInBatches(data []model.SerialNumberMovement) ([]model.SerialNumberMovement, error)
	GetByInvoice(invoiceID uint, invoiceType string) ([]model.SerialNumberMovement, error)
}

func (repo *serialNumberMovementRepository) CreateInBatches(data []model.SerialNumberMovement) ([]model.SerialNumberMovement, error) {
	err := repo.db.CreateInBatches(&data, 15).Error
	return data, err
}

func (repo *serialNumberMovementRepository) GetByInvoice(invoiceID uint, invoiceType string) ([]model.SerialNumberMovement, error) {
	data := []model.SerialNumberMovement{}
	err := repo.db.Find(&data, "invoice_id = ? AND invoice_type = ?", invoiceID, invoiceType).Error
	return data, err
}
