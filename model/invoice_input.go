package model

import "time"

type InvoiceInput struct {
	ID                       uint      `json:"id" gorm:"primaryKey"`
	ProjectID                uint      `json:"projectID"`
	WarehouseManagerWorkerID uint      `json:"warehouseManagerWorkerID"`
	ReleasedWorkerID         uint      `json:"releasedWorkerID"`
	OperatorAddWorkerID      uint      `json:"operatorAddWorkerID"`
	OperatorEditWorkerID     uint      `json:"operatorEditWorkerID"`
	DeliveryCode             string    `json:"deliveryCode" gorm:"tinyText"`
	Notes                    string    `json:"notes"`
	DateOfInvoice            time.Time `json:"dateOfInvoice"`
	DateOfAdd                time.Time `json:"dateOfAdd"`
	DateOfEdit               time.Time `json:"dateOfEdit"`
	Confirmed                bool      `json:"confirmation"`
}
