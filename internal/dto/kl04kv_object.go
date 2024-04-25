package dto

import "backend-v2/model"

type KL04KVObjectPaginated struct {
	ObjectID         uint     `json:"objectID"`
	ObjectDetailedID uint     `json:"objectDetailedID"`
	Name             string   `json:"name"`
	Status           string   `json:"status"`
	Length           float64  `json:"length"`
	Nourashes        string   `json:"nourashes"`
	Supervisors      []string `json:"supervisors"`
}

type KL04KVObjectPaginatedQuery struct {
	ObjectID         uint    `json:"objectID"`
	ObjectDetailedID uint    `json:"objectDetailedID"`
	Name             string  `json:"name"`
	Status           string  `json:"status"`
	Length           float64 `json:"length"`
	Nourashes        string  `json:"nourashes"`
	SupervisorName   string  `json:"supervisorName"`
}

type KL04KVObjectCreate struct {
	BaseInfo     model.Object        `json:"baseInfo"`
	DetailedInfo model.KL04KV_Object `json:"detailedInfo"`
	Supervisors  []uint              `json:"supervisors"`
}
