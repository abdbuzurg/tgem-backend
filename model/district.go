package model

type District struct {
	ID   uint   `json:"id" gorm:"primaryKey"`
	Name string `json:"name"`

	InvoiceOutput []InvoiceOutput `json:"-" gorm:"foreignKey:DistrictID"`
}
