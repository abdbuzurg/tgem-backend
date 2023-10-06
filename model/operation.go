package model

import "github.com/shopspring/decimal"

type Operation struct {
	ID               uint            `json:"id" gorm:"primaryKey"`
	Name             string          `json:"name"`
	Code             string          `json:"code" gorm:"tinyText"`
	CostPrime        decimal.Decimal `json:"costPrime" gorm:"type:decimal(20,4)"`
	CostWithCustomer decimal.Decimal `json:"costWithCustomer" gorm:"type:decimal(20,4)"`

	ObjectOperations []ObjectOperation `json:"-" gorm:"foreignKey:OperationID"`
}
