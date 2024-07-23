package model

import "time"

type InvoiceOutputOutOfProject struct {
	ID               uint      `json:"id" gorm:"primaryKey"`
	ToProjectID      uint      `json:"toProjectID"`
	FromProjectID    uint      `json:"fromProjectID"`
	DeliveryCode     string    `json:"deliveryCode"`
	ReleasedWorkerID uint      `json:"releasedWorkerID"`
	DateOfInvoice    time.Time `json:"dateOfInvoice"`
	Notes            string    `json:"notes"`
	Confirmation     bool      `json:"confirmation"`
}
