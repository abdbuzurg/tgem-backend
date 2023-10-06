package model

type ObjectOperation struct {
	ID             uint `json:"id" gorm:"primaryKey"`
	ObjectID       uint `json:"objectID"`
	MaterialCostID uint `json:"materialCostID"`
	OperationID    uint `json:"operationID"`
	TeamID         uint `json:"teamID"`
}
