package model

import "time"

type InvoiceReturn struct {
	ID                   uint      `json:"ID" gorm:"primaryKey"`
	ProjectID            uint      `json:"projectID"`
	ReturnerType         string    `json:"returnerType"`
	ReturnerID           uint      `json:"returnerID"`
	DateOfInvoice        time.Time `json:"dateOfInvoice"`
	Notes                string    `json:"notes"`
	DeliveryCode         string    `json:"deliveryCode"`
	Confirmation         bool      `json:"confirmation"`
}
