package model

type SupervisorObjects struct {
	ID                 uint `json:"id" gorm:"primaryKey"`
	SupervisorWorkerID uint `json:"supervisorWorkerID"`
	ObjectID           uint `json:"objectID"`
}
