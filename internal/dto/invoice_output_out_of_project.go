package dto

import (
	"backend-v2/model"
	"time"
)

type InvoiceOutputOutOfProjectPaginated struct {
	ID                 uint      `json:"id" gorm:"primaryKey"`
	FromProjectID      uint      `json:"fromProjectID"`
	ToProjectID        uint      `json:"toProjectID"`
	ToProjectName      string    `json:"toProjectName"`
	ToProjectManager   string    `json:"toProjectManager"`
	DeliveryCode       string    `json:"deliveryCode"`
	ReleasedWorkerName string    `json:"releasedWorkerName"`
	DateOfInvoice      time.Time `json:"dateOfInvoice"`
	Confirmation       bool      `json:"confirmation"`
}

type InvoiceOutputOutOfProjectCreateQueryData struct {
	Invoice          model.InvoiceOutputOutOfProject
	InvoiceMaterials []model.InvoiceMaterials
}

type InvoiceOutputOutOfProjectConfirmationQueryData struct {
	InvoiceData        model.InvoiceOutputOutOfProject
	WarehouseMaterials []model.MaterialLocation
}

type InvoiceOutputOutOfProject struct {
	Details model.InvoiceOutputOutOfProject `json:"details"`
	Items   []InvoiceOutputItem             `json:"items"`
}

type InvoiceOutputOutOfProjectSearchParameters struct {
	ToProjectID      uint
	FromProjectID    uint
	ReleasedWorkerID uint
}

