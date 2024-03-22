package model

type Role struct {
  ID uint `gorm:"primaryKey" json:"id"`
  Name string `json:"string"`
  Description string `json:"name"`

  Permissions []Permission `json:"-" gorm:"foreignKey:RoleID"`
  Users []User `json:"-" gorm:"foreignKey:UserRoleID"`
}
