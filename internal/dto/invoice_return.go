package dto

import (
	"backend-v2/model"
	"time"
)

type InvoiceReturnPaginated struct {
	ID               uint      `json:"id"`
	DeliveryCode     string    `json:"deliveryCode"`
	ProjectName      string    `json:"projectName"`
	OperatorAddName  string    `json:"operatorAddName"`
	OperatorEditName string    `json:"operatorEditName"`
	ReturnerType     string    `json:"returnerType"`
	ReturnerName     string    `json:"returnerName"`
	DateOfInvoice    time.Time `json:"dateOfInvoice"`
	DateOfAdd        time.Time `json:"dateOfAdd"`
	DateOfEdit       time.Time `json:"dateOfEdit"`
	Notes            string    `json:"notes"`
	Confirmation     bool      `json:"confirmation"`
}

type InvoiceReturnItem struct {
	MaterialCostID uint    `json:"materialCostID"`
	Amount         float64 `json:"amount"`
	IsDefected     bool    `json:"isDefected"`
}

type InvoiceReturn struct {
	Details model.InvoiceReturn `json:"details"`
	Items   []InvoiceReturnItem `json:"items"`
}

type InvoiceReturnReportFilterRequest struct {
	Code         string    `json:"code"`
	ReturnerType string    `json:"returnerType"`
	Returner     string    `json:"returner"`
	DateFrom     time.Time `json:"dateFrom"`
	DateTo       time.Time `json:"dateTo"`
}

type InvoiceReturnReportFilter struct {
	Code         string
	ReturnerType string
	ReturnerID   uint
	DateFrom     time.Time
	DateTo       time.Time
}
