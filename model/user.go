package model

type User struct {
	ID       uint   `json:"id" gorm:"primaryKey"`
	WorkerID uint   `json:"workerID"`
	Username string `json:"username" gorm:"tinyText"`
	Password string `json:"password"`
}
