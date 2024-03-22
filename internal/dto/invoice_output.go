package dto

import (
	"backend-v2/model"
	"time"
)

type InvoiceOutputPaginated struct {
	ID                   uint      `json:"id"`
	WarehouseManagerName string    `json:"warehouseManagerName"`
	ReceivedName         string    `json:"releasedName"`
	RecipientName        string    `json:"recipientName"`
	TeamName             string    `json:"teamName"`
	ObjectName           string    `json:"objectName"`
	OperatorAddName      string    `json:"operatorAddName"`
	DistrictName         string    `json:"districtName"`
	OperatorEditName     string    `json:"operatorEditName"`
	DeliveryCode         string    `json:"deliveryCode"`
	Notes                string    `json:"notes"`
	DateOfInvoice        time.Time `json:"dateOfInvoice"`
	DateOfAdd            time.Time `json:"dateOfAdd"`
	DateOfEdit           time.Time `json:"dateOfEdit"`
	Confirmation         bool      `json:"confirmation"`
	ObjectConfirmation   bool      `json:"objectConfirmation"`
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
