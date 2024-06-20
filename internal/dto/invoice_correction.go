package dto

import "github.com/shopspring/decimal"

type InvoiceCorrectionMaterialsData struct {
	InvoiceMaterialID uint            `json:"invoiceMaterialID"`
	MaterialName      string          `json:"materialName"`
	MaterialCost      decimal.Decimal `json:"materialCost"`
	MaterialAmount    float64         `json:"materialAmount"`
}
