package dto

import "backend-v2/model"

type SIPObjectPaginatedQuery struct {
	ObjectID         uint   `json:"objectID"`
	ObjectDetailedID uint   `json:"objectDetailedID"`
	Type             string `json:"type"`
	Name             string `json:"name"`
	Status           string `json:"status"`
	AmountFeeders    uint   `json:"amountFeeders"`
	SupervisorName   string `json:"supervisorName"`
}

type SIPObjectPaginated struct {
	ObjectID         uint     `json:"objectID"`
	ObjectDetailedID uint     `json:"objectDetailedID"`
	Type             string   `json:"type"`
	Name             string   `json:"name"`
	Status           string   `json:"status"`
	AmountFeeders    uint     `json:"amountFeeders"`
	Supervisors      []string `json:"supervisors"`
}

type SIPObjectCreate struct {
	BaseInfo     model.Object     `json:"baseInfo"`
	DetailedInfo model.SIP_Object `json:"detailedInfo"`
	Supervisors  []uint           `json:"supervisors"`
}
