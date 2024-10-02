package dto

import "backend-v2/model"

type SIPObjectPaginatedQuery struct {
	ObjectID         uint   `json:"objectID"`
	ObjectDetailedID uint   `json:"objectDetailedID"`
	Type             string `json:"type"`
	Name             string `json:"name"`
	Status           string `json:"status"`
	AmountFeeders    uint   `json:"amountFeeders"`
}

type SIPObjectPaginated struct {
	ObjectID         uint     `json:"objectID"`
	ObjectDetailedID uint     `json:"objectDetailedID"`
	Type             string   `json:"type"`
	Name             string   `json:"name"`
	Status           string   `json:"status"`
	AmountFeeders    uint     `json:"amountFeeders"`
	Supervisors      []string `json:"supervisors"`
	Teams            []string `json:"teams"`
}

type SIPObjectCreate struct {
	BaseInfo     model.Object     `json:"baseInfo"`
	DetailedInfo model.SIP_Object `json:"detailedInfo"`
	Supervisors  []uint           `json:"supervisors"`
	Teams        []uint           `json:"teams"`
}

type SIPObjectImportData struct {
	Object            model.Object
	SIP               model.SIP_Object
	ObjectSupervisors model.ObjectSupervisors
	ObjectTeam        model.ObjectTeams
}

type SIPObjectSearchParameters struct {
	ProjectID          uint
	ObjectName         string
	SupervisorWorkerID uint
	TeamID             uint
}
