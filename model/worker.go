package model

type Worker struct {
	ID           uint   `json:"id" gorm:"primaryKey"`
	Name         string `json:"name" gorm:"tinyText"`
	JobTitle     string `json:"jobTitle" gorm:"tinyText"`
	MobileNumber string `json:"mobileNumber" gorm:"tinyText"`

	User                     User      `json:"-" gorm:"foreignKey:WorkerID"`
	Teams                    []Team    `json:"-" gorm:"foreignKey:LeaderWorkerID"`
	Objects                  []Object  `json:"-" gorm:"foreignKey:SupervisorWorkerID"`
	WarehouseManagerInvoices []Invoice `json:"-" gorm:"foreignKey:WarehouseManagerWorkerID"`
	ReleasedInvoices         []Invoice `json:"-" gorm:"foreignKey:ReleasedWorkerID"`
	DriverInvoices           []Invoice `json:"-" gorm:"foreignKey:DriverWorkerID"`
	RecipientInvoices        []Invoice `json:"-" gorm:"foreignKey:RecipientWorkerID"`
	OperatorAddInvoices      []Invoice `json:"-" gorm:"foreignKey:OperatorAddWorkerID"`
	OperatorEditInvoices     []Invoice `json:"-" gorm:"foreignKey:OperatorEditWorkerID"`
}
