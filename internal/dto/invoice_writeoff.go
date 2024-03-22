package dto

import (
	"backend-v2/model"
	"time"
)

type InvoiceWriteOffPaginated struct {
	ID               uint      `json:"id"`
	WriteOffType     string    `json:"writeOffType"`
	DeliveryCode     string    `json:"deliveryCode"`
	OperatorAddName  string    `json:"operatorAddName"`
	OperatorEditName string    `json:"operatorEditName"`
	DateOfInvoice    time.Time `json:"dateOfInvoice"`
	DateOfAdd        time.Time `json:"dateOfAdd"`
	DateOfEdit       time.Time `json:"dateOfEdit"`
	Notes            string    `json:"notes"`
}

type InvoiceWriteOffItem struct {
	MaterialCostID uint    `json:"materialCostID"`
	Amount         float64 `json:"amount"`
}

type InvoiceWriteOff struct {
	Details model.InvoiceWriteOff `json:"details"`
	Items   []InvoiceWriteOffItem `json:"items"`
}
