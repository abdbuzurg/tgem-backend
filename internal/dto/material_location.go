package dto

import "github.com/shopspring/decimal"

type ReportBalanceFilterRequest struct {
	Type   string `json:"type"`
	Team   string `json:"team"`
	Object string `json:"object"`
}

type ReportBalanceFilter struct {
	LocationType string
	LocationID   uint
}

type BalanceReportQueryResult struct {
	LocationID      uint
	MaterialCode    string
	MaterialName    string
	MaterialUnit    string
	TotalAmount     float64
	DefectAmount    float64
	MaterialCostM19 decimal.Decimal
	TotalCost       decimal.Decimal
	TotalDefectCost decimal.Decimal
}
