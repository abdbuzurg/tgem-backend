package database

import (
	"backend-v2/model"
	"fmt"

	"github.com/spf13/viper"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func InitDB() (*gorm.DB, error) {
	username := viper.GetString("Database.Username")
	password := viper.GetString("Database.Password")
	host := viper.GetString("Database.Host")
	port := viper.GetInt("Database.Port")
	dbname := viper.GetString("Database.DBName")

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=disable", host, username, password, dbname, port)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})

	if err != nil {
		return nil, err
	}

	test, err := db.DB()
	if err != nil {
		return nil, err
	}

	if err := test.Ping(); err != nil {
		return nil, err
	}

	AutoMigrate(db)
	InitialMigration(db)

	return db, nil

}

func AutoMigrate(db *gorm.DB) {
	db.AutoMigrate(
		model.District{},
		model.Role{},
    model.Resource{},
		model.Project{},
		model.Worker{},
		model.User{},
		model.UserAction{},
		model.UserInProject{},
		model.Material{},
		model.MaterialCost{},
		model.MaterialLocation{},
		model.MaterialDefect{},
		model.Object{},
		model.SupervisorObjects{},
		model.Operation{},
		model.ObjectOperation{},
		model.Permission{},
		model.SerialNumber{},
		model.Team{},
		model.TeamObjects{},
		model.InvoiceMaterials{},
		model.InvoiceInput{},
		model.InvoiceOutput{},
		model.InvoiceReturn{},
		model.InvoiceObject{},
		model.InvoiceWriteOff{},
		model.OperatorErrorFound{},
		model.KL04KV_Object{},
		model.MJD_Object{},
		model.SIP_Object{},
		model.STVT_Object{},
		model.TP_Object{},
	)
}
