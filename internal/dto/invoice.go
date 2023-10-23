package dto

import "backend-v2/model"

type InvoicePaginatedData struct {
	ID                   uint   `json:"id"`
	DeliveryCode         string `json:"deliveryCode"`
	WarehouseManagerName string `json:"warehouseManagerName"`
	ReleasedName         string `json:"releasedName"`
	ObjectName           string `json:"objectName"`
	DateOfInvoice        string `json:"dateOfInvoice"`
}

type InvoiceDetails struct {
	ID               uint   `json:"id"`
	ProjectName      string `json:"projectName"`
	TeamNumber       string `json:"teamNumber"`
	WarehouseManager string `json:"warehouseManager"`
	Released         string `json:"released"`
	Driver           string `json:"driver"`
	Recipient        string `json:"recipient"`
	OperatorAdd      string `json:"operatorAdd"`
	OperatorEdit     string `json:"operatorEdit"`
	ObjectName       string `json:"objectName"`
	DeliveryCode     string `json:"deliveryCode"`
	District         string `json:"district"`
	CarNumber        string `json:"carNumber"`
	Notes            string `json:"notes"`
	DateOfInvoice    string `json:"dateOfInvoice"`
	DateOfAddition   string `json:"dateOfAddition"`
	DateOfEdit       string `json:"dateOfEdit"`
}

type InvoiceMaterial struct {
	Name   string  `json:"name"`
	Code   string  `json:"code"`
	Unit   string  `json:"unit"`
	Amount float64 `json:"amount"`
	Notes  string  `json:"notes"`
}

type InvoiceFullData struct {
	Invoice   InvoiceDetails    `json:"invoice"`
	Materials []InvoiceMaterial `json:"materials"`
}

type InvoiceDataUpdateOrCreate struct {
	Invoice   model.Invoice
	Materials []InvoiceMaterialForUpdateOrCreate
}

type InvoiceMaterialForUpdateOrCreate struct {
	model.Material
	Amount       float64 `json:"amount"`
	InvoiceNotes string  `json:"notes"`
}
