package model

type MaterialForProject struct {
	ID         uint `json:"id" gorm:"primaryKey"`
	MaterialID uint `json:"materialID"`
	ProjectID  uint `json:"projectID"`
}
