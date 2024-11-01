package dto

import "time"

type MaterialDataForProgressReportQueryResult struct {
	ID                        uint
	Code                      string
	Name                      string
	Unit                      string
	PlannedAmountForProject   float64
	LocationAmount            float64
	LocationType              string
	SumWithCustomerInLocation float64
}

type InvoiceMaterialDataForProgressReportQueryResult struct {
	MaterialID   uint
	Amount       float64
	InvoiceType  string
	SumInInvoice float64
}

type MaterialDataForRemainingMaterialAnalysisQueryResult struct {
	ID                      uint
	Code                    string
	Name                    string
	Unit                    string
	PlannedAmountForProject float64
	LocationAmount          float64
	LocationType            string
}

type MaterialsInstalledOnObjectForRemainingMaterialAnalysisQueryResult struct {
	ID               uint
	Amount           float64
	DateOfCorrection time.Time
}
