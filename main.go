package main

import (
	"backend-v2/api"
	"backend-v2/internal/jobs"
	"backend-v2/pkg/config"
	"backend-v2/pkg/database"
	"fmt"
	"log"

	"github.com/spf13/viper"
)

func init() {
	config.GetConfig()
}

func main() {
	db, err := database.InitDB()
	if err != nil {
		log.Fatal(err)
		return
	}

  go jobs.Run()

	port := fmt.Sprintf("127.0.0.1:%d", viper.GetInt("App.Port"))
	app := api.SetupRouter(db)
	if err := app.Run(port); err != nil {
		panic(err)
	}
}
