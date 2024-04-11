package dto

type LoginData struct {
	Username string `json:"username"`
	Password string `json:"password"`
  ProjectID uint `json:"projectID"`
}

type UserPermission struct {
	ResourceName   string `json:"resourceName"`
	ResourceAction string `json:"resourceAction"`
}
