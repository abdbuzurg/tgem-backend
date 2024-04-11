package dto

import "time"

type UserActionView struct {
	ID                  uint      `json:"id" gorm:"primaryKey"`
	ActionURL           string    `json:"actionURL"`
	ActionType          string    `json:"actionType"`
	ActionID            uint      `json:"actionID"`
	ActionStatus        bool      `json:"actionStatus"`
	ActionStatusMessage string    `json:"actionStatusMessage"`
	DateOfAction        time.Time `json:"dateOfAction"`
}
