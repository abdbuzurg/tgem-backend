package dto

type UserPermission struct {
	ResourceName string `json:"resourceName"`
	R            bool   `json:"r"`
	W            bool   `json:"w"`
	U            bool   `json:"u"`
	D            bool   `json:"d"`
}
