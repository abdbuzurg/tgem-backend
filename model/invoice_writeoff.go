package model

import "time"

type InvoiceWriteOff struct {
	ID                   uint      `json:"id" gorm:"primaryKey"`
	WriteOffType         string    `json:"writeOffType"`
	DeliveryCode         string    `json:"deliveryCode"`
	OperatorAddWorkerID  uint      `json:"operatorAddWorkerID"`
	OperatorEditWorkerID uint      `json:"operatorEditWorkerID"`
	DateOfInvoice        time.Time `json:"dateOfInvoice"`
	DateOfAdd            time.Time `json:"dateOfAdd"`
	DateOfEdit           time.Time `json:"dateOfEdit"`
	Notes                string    `json:"notes"`
}
