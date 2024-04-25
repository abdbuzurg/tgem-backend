package dto

type TeamPaginated struct {
	ID           uint     `json:"id" gorm:"primaryKey"`
	LeaderName   string   `json:"leaderName"`
	Number       string   `json:"number" gorm:"tinyText"`
	MobileNumber string   `json:"mobileNumber" gorm:"tinyText"`
	Company      string   `json:"company" gorm:"tinyText"`
	Objects      []string `json:"objects"`
}

type TeamPaginatedQuery struct {
	ID               uint   `json:"id"`
	LeaderName       string `json:"leaderName"`
	TeamNumber       string `json:"number"`
	TeamMobileNumber string `json:"mobileNumber"`
	TeamCompany      string `json:"company"`
	ObjectName       string `json:"objectName"`
}

type TeamMutation struct {
	LeaderWorkerID uint   `json:"leaderWorkerID"`
	Number         string `json:"number"`
	MobileNumber   string `json:"mobileNumber"`
	Company        string `json:"company"`
	Objects        []uint `json:"objects"`
}
