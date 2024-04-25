package dto

import (
	"backend-v2/model"
	"time"
)

type InvoiceOutputPaginated struct {
	ID                   uint      `json:"id"`
	WarehouseManagerName string    `json:"warehouseManagerName"`
	ReleasedName         string    `json:"releasedName"`
	RecipientName        string    `json:"recipientName"`
	TeamName             string    `json:"teamName"`
	ObjectName           string    `json:"objectName"`
	DistrictName         string    `json:"districtName"`
	DeliveryCode         string    `json:"deliveryCode"`
	Notes                string    `json:"notes"`
	DateOfInvoice        time.Time `json:"dateOfInvoice"`
	Confirmation         bool      `json:"confirmation"`
}

type InvoiceOutputItem struct {
	MaterialID    uint     `json:"materialID"`
	Amount        float64  `json:"amount"`
	SerialNumbers []string `json:"serialNumbers"`
}

type InvoiceOutput struct {
	Details model.InvoiceOutput `json:"details"`
	Items   []InvoiceOutputItem `json:"items"`
}

type InvoiceObject struct {
	ID             uint   `json:"id"`
	TeamLeaderName string `json:"teamLeaderName"`
	TeamNumber     string `json:"teamNumber"`
	ObjectName     string `json:"objectName"`
}

type InvoiceOutputReportFilterRequest struct {
	ProjectID        uint      `json:"projectID"`
	Code             string    `json:"code"`
	WarehouseManager string    `json:"warehouseManager"`
	Received         string    `json:"recieved"`
	Object           string    `json:"object"`
	Team             string    `json:"team"`
	District         string    `json:"district"`
	DateFrom         time.Time `json:"dateFrom"`
	DateTo           time.Time `json:"dateTo"`
}

type InvoiceOutputReportFilter struct {
	Code               string
	WarehouseManagerID uint
	ReceivedID         uint
	ObjectID           uint
	TeamID             uint
	DistrictID         uint
	DateFrom           time.Time
	DateTo             time.Time
}
