package api

import (
	"backend-v2/api/middleware"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func SetupRouter(db *gorm.DB) *gin.Engine {

	router := gin.Default()

	router.Use(gin.Recovery())

	router.Use(middleware.CORSMiddleware())

	return router

}
