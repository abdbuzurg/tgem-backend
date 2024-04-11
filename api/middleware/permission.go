package middleware

import (
	"backend-v2/model"
	"backend-v2/pkg/response"
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func Permission(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {

		roleID := c.GetUint("roleID")

		var permissions []model.Permission
		err := db.Find(&permissions, "role_id = ?", roleID).Error
		if err != nil {
			response.ResponseError(c, fmt.Sprintf("Доступ запрещен"))
			c.Abort()
			return
		}

		path := c.Request.URL.Path
		splittedpath := strings.Split(path, "/")
		resourceUrl := "/" + splittedpath[1]

		var indexOfThePermission int = -1
		for index, permission := range permissions {
			if permission.ResourceUrl == resourceUrl {
				indexOfThePermission = index
				break
			}
		}

		requestMethod := c.Request.Method

		if requestMethod == "GET" && !permissions[indexOfThePermission].R {
			response.ResponseError(c, fmt.Sprintf("Доступ запрещен"))
			c.Abort()
			return
		}

		if requestMethod == "POST" && !permissions[indexOfThePermission].W {
			response.ResponseError(c, fmt.Sprintf("Доступ запрещен"))
			c.Abort()
			return
		}

		if requestMethod == "PATCH" && !permissions[indexOfThePermission].U {
			response.ResponseError(c, fmt.Sprintf("Доступ запрещен"))
			c.Abort()
			return
		}

		if requestMethod == "DELETE" && !permissions[indexOfThePermission].D {
			response.ResponseError(c, fmt.Sprintf("Доступ запрещен"))
			c.Abort()
			return
		}

		c.Next()
	}
}
