package main

import (
	"backend-v2/api"
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

	port := fmt.Sprintf(":%d", viper.GetInt("App.Port"))

	app := api.SetupRouter(db)
	app.Run(port)
}
