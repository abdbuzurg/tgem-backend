package model

import "time"

type InvoiceWriteOff struct {
	ID            uint      `json:"id" gorm:"primaryKey"`
	ProjectID     uint      `json:"projectID"`
	WriteOffType  string    `json:"writeOffType"`
	DeliveryCode  string    `json:"deliveryCode"`
	DateOfInvoice time.Time `json:"dateOfInvoice"`
	DateOfAdd     time.Time `json:"dateOfAdd"`
	DateOfEdit    time.Time `json:"dateOfEdit"`
	Notes         string    `json:"notes"`
}
