package dto

import "github.com/shopspring/decimal"

type InvoiceMaterial struct {
	ID            uint            `json:"id" gorm:"primaryKey"`
	MaterialName  string          `json:"materialName"`
	MaterialPrice decimal.Decimal `json:"materialPrice"`
	InvoiceID     uint            `json:"invoiceID"`
	InvoiceType   string          `json:"invoiceType" gorm:"tinyText"`
	Amount        float64         `json:"amount"`
	Notes         string          `json:"notes"`
	Unit          string          `json:"unit"`
}
