package dto

import (
	"backend-v2/model"
	"time"

	"github.com/shopspring/decimal"
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
	DeliveryCode         string    `json:"deliveryCode"`
	Notes                string    `json:"notes"`
	DateOfInvoice        time.Time `json:"dateOfInvoice"`
	Confirmation         bool      `json:"confirmation"`
}

type InvoiceInputCreateQueryData struct {
	InvoiceData          model.InvoiceInput
	InvoiceMaterials     []model.InvoiceMaterials
	SerialNumbers        []model.SerialNumber
	SerialNumberMovement []model.SerialNumberMovement
}

type InvoiceInputConfirmationQueryData struct {
	InvoiceData          model.InvoiceInput
	ToBeUpdatedMaterials []model.MaterialLocation
	ToBeCreatedMaterials []model.MaterialLocation
	SerialNumbers        []model.SerialNumberLocation
}

type InvoiceInputReportFilterRequest struct {
	ProjectID        uint      `json:"projectID"`
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

type NewMaterialDataFromInvoiceInput struct {
	Category         string          `json:"category" gorm:"tinyText"`
	Code             string          `json:"code" gorm:"tinyText"`
	Name             string          `json:"name" gorm:"tinyText"`
	Unit             string          `json:"unit" gorm:"tinyText"`
	Notes            string          `json:"notes"`
	ProjectID        uint            `json:"projectID"`
	HasSerialNumber  bool            `json:"hasSerialNumber"`
	CostPrime        decimal.Decimal `json:"costPrime" gorm:"type:decimal(20,4)"`
	CostM19          decimal.Decimal `json:"costM19" gorm:"type:decimal(20,4)"`
	CostWithCustomer decimal.Decimal `json:"costWithCustomer" gorm:"type:decimal(20,4)"`
}

type InvoiceInputMaterialForEdit struct {
	MaterialID      uint            `json:"materialID"`
	MaterialName    string          `json:"materialName"`
	Unit            string          `json:"unit"`
	Amount          float64         `json:"amount"`
	MaterialCostID  uint            `json:"materialCostID"`
	MaterialCost    float64 `json:"materialCost"`
	Notes           string          `json:"notes"`
	HasSerialNumber bool            `json:"hasSerialNumber"`
	SerialNumbers   []string        `json:"serialNumbers"`
}
