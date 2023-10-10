package database

import (
	"backend-v2/model"
	"fmt"

	"github.com/casbin/casbin/v2"
	gormadapter "github.com/casbin/gorm-adapter/v3"
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

	a, err := gormadapter.NewAdapterByDB(db)
	if err != nil {
		return nil, err
	}
	gormadapter.TurnOffAutoMigrate(db)

	_, err = casbin.NewEnforcer("./pkg/database/acl_model.conf", a)
	if err != nil {
		return nil, err
	}

	AutoMigrate(db)
	InitialMigration(db)

	return db, nil

}

func AutoMigrate(db *gorm.DB) {
	db.AutoMigrate(
		model.Project{},
		model.Worker{},
		model.User{},
		model.Material{},
		model.MaterialForProject{},
		model.Object{},
		model.Team{},
		model.Invoice{},
		model.InvoiceMaterials{},
	)

	db.Table("kl04kv_objects").AutoMigrate(model.KL04KV_Object{})
	db.Table("mjd_objects").AutoMigrate(model.MJD_Object{})
	db.Table("sip_objects").AutoMigrate(model.SIP_Object{})
	db.Table("stvt_objects").AutoMigrate(model.STVT_Object{})
	db.Table("tp_objects").AutoMigrate(model.TP_Object{})
}
