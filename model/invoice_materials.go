package model

type InvoiceMaterials struct {
	ID             uint    `json:"id" gorm:"primaryKey"`
	MaterialCostID uint    `json:"materialID"`
	InvoiceID      uint    `json:"invoiceID"`
	Amount         float64 `json:"amount"`
	Notes          string  `json:"notes"`
}
