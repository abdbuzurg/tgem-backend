package model

type ObjectOperation struct {
	ID              uint    `json:"id" gorm:"primaryKey"`
	ProjectID       uint    `json:"projectID"`
	InvoiceObjectID uint    `json:"invoiceObjectID"`
	OperationID     uint    `json:"operationID"`
	Amount          float64 `json:"amount"`
	Notes           string  `json:"notes"`
}
