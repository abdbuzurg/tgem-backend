package model

import "time"

type InvoiceReturn struct {
	ID                   uint      `json:"ID" gorm:"primaryKey"`
	ProjectID            uint      `json:"projectID"`
	OperatorAddWorkerID  uint      `json:"operatorAddWorkerID"`
	OperatorEditWorkerID uint      `json:"operatorEditWorkerID"`
	ReturnerType         string    `json:"returnerType"`
	ReturnerID           uint      `json:"returnerID"`
	DateOfInvoice        time.Time `json:"dateOfInvoice"`
	DateOfAdd            time.Time `json:"dateOfAdd"`
	DateOfEdit           time.Time `json:"DateOfEdit"`
	Notes                string    `json:"notes"`
	DeliveryCode         string    `json:"deliveryCode"`
	Confirmation         bool      `json:"confirmation"`
}
