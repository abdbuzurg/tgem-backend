package model

import "time"

type InvoiceOutputOutOfProject struct {
	ID               uint      `json:"id" gorm:"primaryKey"`
	ProjectID        uint      `json:"projectID"`
	DeliveryCode     string    `json:"deliveryCode"`
	ReleasedWorkerID uint      `json:"releasedWorkerID"`
	DateOfInvoice    time.Time `json:"dateOfInvoice"`
	Notes            string    `json:"notes"`
	Confirmation     bool      `json:"confirmation"`
}
