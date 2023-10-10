package model

type MaterialLocation struct {
	ID             uint    `json:"id" gorm:"primaryKey"`
	MaterialCostID uint    `json:"materialCostID"`
	LocationID     uint    `json:"materialDetailLocationID"`
	LocationType   string  `json:"locationType" gorm:"tinyText"`
	Amount         float64 `json:"amount"`
}
