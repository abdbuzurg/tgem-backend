package model

type SerialNumber struct {
	ID             uint   `json:"id" gorm:"primaryKey"`
	Code           string `json:"code" gorm:"text"`
	MaterialCostID uint   `json:"materialCostID"`

	//STATUS CAN BE SERVER THINGS
	//WAREHOUSE, TEAMS, OBJECT, WRITEOFF --> LOOK FOR ID IN MATERIAL LOCATION
	//PENDING	--> LOOK FOR ID IN INVOICE MATERIALS
	Status   string `json:"status"`
	StatusID uint   `json:"statusID"`
}
