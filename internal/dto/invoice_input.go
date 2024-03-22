package dto

import (
	"backend-v2/model"
	"time"
)

type InvoiceInput struct {
	Details model.InvoiceInput     `json:"details"`
	Items   []InvoiceInputMaterial `json:"items"`
}

type InvoiceInputMaterial struct {
	MaterialData  model.InvoiceMaterials `json:"materialData"`
	SerialNumbers []string               `json:"serialNumbers"`
}

type InvoiceInputPaginated struct {
	ID                   uint      `json:"id"`
	WarehouseManagerName string    `json:"warehouseManagerName"`
	ReleasedName         string    `json:"releasedName"`
	OperatorAddName      string    `json:"operatorAddName"`
	OperatorEditName     string    `json:"operatorEditName"`
	DeliveryCode         string    `json:"deliveryCode"`
	Notes                string    `json:"notes"`
	DateOfInvoice        time.Time `json:"dateOfInvoice"`
	DateOfAdd            time.Time `json:"dateOfAdd"`
	DateOfEdit           time.Time `json:"dateOfEdit"`
	Confirmation         bool      `json:"confirmation"`
}

type InvoiceInputReportFilterRequest struct {
	Code             string    `json:"code"`
	WarehouseManager string    `json:"warehouseManager"`
	Released         string    `json:"released"`
	DateFrom         time.Time `json:"dateFrom"`
	DateTo           time.Time `json:"dateTo"`
}

type InvoiceInputReportFilter struct {
	Code               string
	WarehouseManagerID uint
	ReleasedID         uint
	DateFrom           time.Time
	DateTo             time.Time
}
