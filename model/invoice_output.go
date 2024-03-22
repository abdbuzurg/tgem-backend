package model

import (
	"time"
)

type InvoiceOutput struct {
	ID                       uint      `json:"id" gorm:"primaryKey"`
	DistrictID               uint      `json:"districtID"`
	ProjectID                uint      `json:"projectID"`
	WarehouseManagerWorkerID uint      `json:"warehouseManagerWorkerID"`
	ReleasedWorkerID         uint      `json:"releasedWorkerID"`
	RecipientWorkerID        uint      `json:"recipientWorkerID"`
	TeamID                   uint      `json:"teamID"`
	ObjectID                 uint      `json:"ObjectID"`
	DeliveryCode             string    `json:"deliveryCode"`
	OperatorAddWorkerID      uint      `json:"operatorAddWorkerID"`
	OperatorEditWorkerID     uint      `json:"operatorEditWorkerID"`
	DateOfInvoice            time.Time `json:"dateOfInvoice"`
	DateOfAdd                time.Time `json:"dateOfAdd"`
	DateOfEdit               time.Time `json:"dateOfEdit"`
	Notes                    string    `json:"notes"`
	Confirmation             bool      `json:"confirmation"`
	ObjectConfirmation       bool      `json:"objectConfirmation"`
}
