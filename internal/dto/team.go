package dto

type TeamPaginated struct {
	ID           uint   `json:"id" gorm:"primaryKey"`
	LeaderName   string `json:"leaderName"`
	Number       string `json:"number" gorm:"tinyText"`
	MobileNumber string `json:"mobileNumber" gorm:"tinyText"`
	Company      string `json:"company" gorm:"tinyText"`
}
