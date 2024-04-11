package model

type Worker struct {
	ID           uint   `json:"id" gorm:"primaryKey"`
	Name         string `json:"name" gorm:"tinyText"`
	JobTitle     string `json:"jobTitle" gorm:"tinyText"`
	MobileNumber string `json:"mobileNumber" gorm:"tinyText"`

	User    User     `json:"-" gorm:"foreignKey:WorkerID"`
	Teams   []Team   `json:"-" gorm:"foreignKey:LeaderWorkerID"`

  //Object Workers
  SupervisorObjectss []SupervisorObjects `json:"-" gorm:"foreignKey:SupervisorWorkerID"`

	//Invoice Input Workers
	InvoiceInputsWarehouseManager []InvoiceInput `json:"-" gorm:"foreignKey:WarehouseManagerWorkerID"`
	InvoiceInputsReleased         []InvoiceInput `json:"-" gorm:"foreignKey:ReleasedWorkerID"`

	//Invoice Object
	InvoiceObjectsSupervisor []InvoiceObject `json:"-" gorm:"foreignKey:SupervisorWorkerID"`

	//Invoice Output
	InvoiceOutputsWarehouseManager []InvoiceOutput `json:"-" gorm:"foreignKey:WarehouseManagerWorkerID"`
	InvoiceOutputsReleased         []InvoiceOutput `json:"-" gorm:"foreignKey:ReleasedWorkerID"`
	InvoiceOutputsRecipient        []InvoiceOutput `json:"-" gorm:"foreignKey:RecipientWorkerID"`
}
