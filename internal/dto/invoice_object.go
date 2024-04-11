package dto

import "time"

type InvoiceObjectPaginated struct {
	ID            uint      `json:"id"`
	DeliveryCode  string    `json:"deliveryCode"`
	Supervisor    string    `json:"supervisor"`
	ObjectName    string    `json:"objectName"`
	TeamName      string    `json:"teamName"`
	DateOfInvoice time.Time `json:"dateOfInvoice"`
}
