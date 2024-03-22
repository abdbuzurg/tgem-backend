package dto

type ReportBalanceFilterRequest struct {
	Type   string `json:"type"`
	Team   string `json:"team"`
	Object string `json:"object"`
}

type ReportBalanceFilter struct {
	LocationType string
	LocationID   uint
}
