package dto

import "backend-v2/model"

type ObjectPaginated struct {
	model.Object   `json:"object"`
	SupervisorName string `json:"supervisorName"`
}
