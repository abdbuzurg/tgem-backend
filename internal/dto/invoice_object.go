package dto

import (
	"backend-v2/model"
	"time"
)

type InvoiceObjectPaginated struct {
	ID                  uint      `json:"id"`
	DeliveryCode        string    `json:"deliveryCode"`
	SupervisorName      string    `json:"supervisorName"`
	ObjectName          string    `json:"objectName"`
	TeamNumber          string    `json:"teamNumber"`
	DateOfInvoice       time.Time `json:"dateOfInvoice"`
	ConfirmedByOperator bool      `json:"confirmedByOperator"`
}

type InvoiceObjectItem struct {
	MaterialID    uint     `json:"materialID"`
	Amount        float64  `json:"amount"`
	SerialNumbers []string `json:"serialNumbers"`
}

type InvoiceObjectCreate struct {
	Details model.InvoiceObject `json:"details"`
	Items   []InvoiceObjectItem `json:"items"`
}

type InvoiceObjectFullDataItem struct {
	ID           uint    `json:"id"`
	MaterialName string  `json:"materialName"`
	Amount       float64 `json:"amount"`
	Notes        string  `json:"notes"`
}

type InvoiceObjectFullData struct {
	Details InvoiceObjectPaginated      `json:"details"`
	Items   []InvoiceObjectFullDataItem `json:"items"`
}
