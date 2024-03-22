package dto

type InvoiceObjectSubmissionPaginated struct {
	ID             uint   `json:"id"`
	Supervisor     string `json:"supervisor"`
	ObjectName     string `json:"objectName"`
	ApprovalStatus bool   `json:"approvalStatus"`
}
