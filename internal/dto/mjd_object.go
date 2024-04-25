package dto

import "backend-v2/model"

type MJDObjectPaginatedQuery struct {
	ObjectID         uint   `json:"objectID"`
	ObjectDetailedID uint   `json:"objectDetailedID"`
	Name             string `json:"name" gorm:"tinyText"`
	Status           string `json:"status" gorm:"tinyText"`
	Model            string `json:"model" gorm:"tinyText"`
	AmountStores     uint   `json:"amountStores"`
	AmountEntrances  uint   `json:"amountEntrances"`
	HasBasement      bool   `json:"hasBasement"`
	SupervisorName   string `json:"supervisorName"`
}

type MJDObjectPaginated struct {
	ObjectID         uint     `json:"objectID"`
	ObjectDetailedID uint     `json:"objectDetailedID"`
	Name             string   `json:"name" gorm:"tinyText"`
	Status           string   `json:"status" gorm:"tinyText"`
	Model            string   `json:"model" gorm:"tinyText"`
	AmountStores     uint     `json:"amountStores"`
	AmountEntrances  uint     `json:"amountEntrances"`
	HasBasement      bool     `json:"hasBasement"`
	Supervisors      []string `json:"supervisors"`
}

type MJDObjectCreate struct {
	BaseInfo     model.Object     `json:"baseInfo"`
	DetailedInfo model.MJD_Object `json:"detailedInfo"`
	Supervisors  []uint           `json:"supervisors"`
}
