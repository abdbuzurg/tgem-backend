package model

import (
	"time"

	"github.com/shopspring/decimal"
)

type Project struct {
	ID                   uint            `json:"id" gorm:"primaryKey"`
	Name                 string          `json:"name" gorm:"tinyText"`
	Client               string          `json:"client" gorm:"tinyText"`
	Budget               decimal.Decimal `json:"budget" gorm:"type:decimal(20,2)"`
	Description          string          `json:"description"`
	SignedDateOfContract time.Time       `json:"signedDateOfContract"`
	DateStart            time.Time       `json:"dateStart"`
	DateEnd              time.Time       `json:"dateEnd"`

	MaterialsForProject []MaterialForProject `json:"-" gorm:"foreignKey:ProjectID"`
	Invoices            []Invoice            `json:"-" gorm:"foreignKey:ProjectID"`
}
