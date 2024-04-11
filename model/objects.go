package model

type Object struct {
	ID uint `json:"id" gorm:"primaryKey"`

	//Custom field that references different type of Object tables in the database
	//The type will be determined by type filed on the application level not database
	ObjectDetailedID uint   `json:"objectDetailedID"`
	Type             string `json:"type" gorm:"tinyText"`
	Name             string `json:"name" gorm:"tinyText"`
	Status           string `json:"status" gorm:"tinyText"`

	SupervisorObjectss []SupervisorObjects `json:"-" gorm:"foreignKey:ObjectID"`
	TeamObjectss       []TeamObjects       `json:"-" gorm:"foreignKey:ObjectID"`
	ObjectOperations   []ObjectOperation   `json:"-" gorm:"foreignKey:ObjectID"`
	InvoiceOutputs     []InvoiceOutput     `json:"-" gorm:"foreignKey:ObjectID"`
	InvoiceObject      []InvoiceObject     `json:"-" gorm:"foreignKey:ObjectID"`
}

type MJD_Object struct {
	ID             uint   `json:"id" gorm:"primaryKey"`
	Model          string `json:"model" gorm:"tinyText"`
	AmountStores   uint   `json:"amountStores"`
	AmountEntraces uint   `json:"amountEntraces"`
	HasBasement    bool   `json:"hasBasement"`
}

type TP_Object struct {
	ID           uint   `json:"id" gorm:"primaryKey"`
	Model        string `json:"model" gorm:"tinyText"`
	VoltageClass string `json:"voltageClass" gorm:"tinyText"`
	Nourashes    string `json:"nourashes" gorm:"tinyText"`
}

type STVT_Object struct {
	ID            uint   `json:"id" gorm:"primaryKey"`
	VoltageClass  string `json:"voltageClass" gorm:"tinyText"`
	TTCoefficient string `json:"ttCoefficient" gorm:"tinyText"`
}

type SIP_Object struct {
	ID            uint `json:"id" gorm:"primaryKey"`
	AmountFeeders uint `json:"amountFeeders"`
}

type KL04KV_Object struct {
	ID        uint    `json:"id" gorm:"primaryKey"`
	Length    float64 `json:"length"`
	Nourashes string  `json:"nourashes" gorm:"tinyText"`
}
