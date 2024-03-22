package model

type InvoiceObjectSubmission struct {
	ID                 uint `gorm:"primaryKey" json:"id"`
	SuperwisorWorkerID uint `json:"supervisorWorkerID"`
	ObjectID           uint `json:"objectID"`
	ApprovalStatus     bool `json:"approvalStatus"`
}
