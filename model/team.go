package model

type Team struct {
	ID             uint   `json:"id" gorm:"primaryKey"`
	LeaderWorkerID uint   `json:"leaderWorkerID"`
	Number         string `json:"number" gorm:"tinyText"`
	MobileNumber   string `json:"mobileNumber" gorm:"tinyText"`
	Company        string `json:"company" gorm:"tinyText"`

	Invoices         []Invoice         `json:"-" gorm:"foreignKey:TeamID"`
	ObjectOperations []ObjectOperation `json:"-" gorm:"foreignKey:TeamID"`
}
