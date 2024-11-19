package middleware

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func Permission(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
    c.Next()
	}
}
