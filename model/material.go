package model

type Material struct {
	ID              uint   `json:"id" gorm:"primaryKey"`
	Category        string `json:"category" gorm:"tinyText"`
	Code            string `json:"code" gorm:"tinyText"`
	Name            string `json:"name" gorm:"tinyText"`
	Unit            string `json:"unit" gorm:"tinyText"`
	Notes           string `json:"notes"`
	HasSerialNumber bool   `json:"hasSerialNumber"`
	Article         string `json:"article"`
	ProjectID       uint   `json:"projectID"`

	MaterialCosts      []MaterialCost      `json:"-" gorm:"foreignKey:MaterialID"`
	OperationMaterials []OperationMaterial `json:"-" gorm:"foreignKey:MaterialID"`
}
