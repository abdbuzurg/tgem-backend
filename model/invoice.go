package model

import "time"

type Invoice struct {
	ID                       uint      `json:"id" gorm:"primaryKey"`
	ProjectID                uint      `json:"projectID"`
	TeamID                   uint      `json:"teamID"`
	WarehouseManagerWorkerID uint      `json:"warehouseManagerWorkerID"`
	ReleasedWorkerID         uint      `json:"releasedWorkerID"`
	DriverWorkerID           uint      `json:"driverWorkerID"`
	RecipientWorkerID        uint      `json:"recipientWorkerID"`
	OperatorAddWorkerID      uint      `json:"operatorAddWorkerID"`
	OperatorEditWorkerID     uint      `json:"operatorEditWorkerID"`
	ObjectID                 uint      `json:"objectID"`
	InvoiceType              string    `json:"invoiceType" gorm:"tinyText"`
	DeliveryCode             string    `json:"deliveryCode" gorm:"tinyText"`
	District                 string    `json:"district" gorm:"tinyText"`
	CarNumber                string    `json:"carNumber" gorm:"tinyText"`
	Notes                    string    `json:"notes"`
	DateOfInvoice            time.Time `json:"dateOfInvoice"`
	DateOfAddition           time.Time `json:"dateOfAddition"`
	DateOfEdit               time.Time `json:"dateOfEdit"`

	InvoiceMaterials []InvoiceMaterials `json:"-" gorm:"foreignKey:InvoiceID"`
}
