package dto

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
