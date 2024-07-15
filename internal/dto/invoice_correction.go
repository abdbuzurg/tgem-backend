package dto

import (
	"backend-v2/model"
	"time"
)

type InvoiceCorrectionMaterialsData struct {
	InvoiceMaterialID uint    `json:"invoiceMaterialID"`
	MaterialName      string  `json:"materialName"`
	MaterialID        uint    `json:"materialID"`
	MaterialAmount    float64 `json:"materialAmount"`
	Notes             string  `json:"notes"`
}

type InvoiceCorrectionPaginated struct {
	ID                  uint      `json:"id"`
	DeliveryCode        string    `json:"deliveryCode"`
	SupervisorName      string    `json:"supervisorName"`
	ObjectName          string    `json:"objectName"`
	TeamNumber          string    `json:"teamNumber"`
	TeamID              uint      `json:"teamID"`
	DateOfInvoice       time.Time `json:"dateOfInvoice"`
	ConfirmedByOperator bool      `json:"confirmedByOperator"`
}

type InvoiceCorrectionCreateDetails struct {
	ID               uint      `json:"id"`
	DateOfCorrection time.Time `json:"dateOfCorrection"`
	OperatorWorkerID uint      `json:"operatorWorkerID"`
}

type InvoiceCorrectionCreate struct {
	Details InvoiceCorrectionCreateDetails   `json:"details"`
	Items   []InvoiceCorrectionMaterialsData `json:"items"`
}

type InvoiceCorrectionCreateQuery struct {
	Details         model.InvoiceObject
	OperatorDetails model.InvoiceObjectOperator
	Items           []model.InvoiceMaterials
	TeamLocation    []model.MaterialLocation
	ObjectLocation  []model.MaterialLocation
}

type InvoiceCorrectionReportFilter struct {
	ProjectID uint      `json:"projectID"`
	ObjectID  uint      `json:"objectID"`
	TeamID    uint      `json:"teamID"`
	DateFrom  time.Time `json:"dateFrom"`
	DateTo    time.Time `json:"dateTo"`
}

type InvoiceCorrectionReportData struct {
	ID               uint
	DeliveryCode     string
	ObjectName       string
	ObjectType       string
	TeamNumber       string
	TeamLeaderName   string
	DateOfInvoice    time.Time
	OperatorName     string
	DateOfCorrection time.Time
}
