package model

type Permission struct {
  ID  uint  `gorm:"primaryKey" json:"id"`
  RoleID  uint  `json:"roleID"`
  Resource  string  `json:"resource"`
  R bool  `json:"r"`
  W bool  `json:"w"`
  X bool  `json:"x"`
  D bool  `json:"d"`
}
