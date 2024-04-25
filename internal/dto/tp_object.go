package dto

import "backend-v2/model"

type TPObjectPaginatedQuery struct {
	ObjectID         uint   `json:"objectID"`
	ObjectDetailedID uint   `json:"objectDetailedID"`
	Name             string `json:"name"`
	Status           string `json:"status"`
	Model            string `json:"model"`
	VoltageClass     string `json:"voltageClass"`
	Nourashes        string `json:"nourashes"`
	SupervisorName   string `json:"supervisorName"`
}

type TPObjectPaginated struct {
	ObjectID         uint     `json:"objectID"`
	ObjectDetailedID uint     `json:"objectDetailedID"`
	Name             string   `json:"name"`
	Status           string   `json:"status"`
	Model            string   `json:"model"`
	VoltageClass     string   `json:"voltageClass"`
	Nourashes        string   `json:"nourashes"`
	Supervisors      []string `json:"supervisors"`
}

type TPObjectCreate struct {
	BaseInfo     model.Object    `json:"baseInfo"`
	DetailedInfo model.TP_Object `json:"detailedInfo"`
	Supervisors  []uint          `json:"supervisors"`
}
