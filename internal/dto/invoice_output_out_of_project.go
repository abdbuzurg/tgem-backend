package dto

import (
	"backend-v2/model"
	"time"
)

type InvoiceOutputOutOfProjectPaginated struct {
	ID                 uint      `json:"id" gorm:"primaryKey"`
	ProjectID          uint      `json:"-"`
	DeliveryCode       string    `json:"deliveryCode"`
	ReleasedWorkerName uint      `json:"releasedWorkerName"`
	DateOfInvoice      time.Time `json:"dateOfInvoice"`
	Notes              string    `json:"notes"`
	Confirmation       bool      `json:"confirmation"`
}

type InvoiceOutputOutOfProjectCreateQueryData struct {
	Invoice                       model.InvoiceOutputOutOfProject
	InvoiceMaterials              []model.InvoiceMaterials
	SerialNumberMovements         []model.SerialNumberMovement
}

type InvoiceOutputOutOfProjectConfirmationQueryData struct {
  InvoiceData model.InvoiceOutputOutOfProject
  WarehouseMaterials []model.MaterialLocation
  TeamMaterials []model.MaterialLocation
}

type InvoiceOutputOutOfProject struct {
	Details model.InvoiceOutputOutOfProject `json:"details"`
	Items   []InvoiceOutputItem `json:"items"`
}
