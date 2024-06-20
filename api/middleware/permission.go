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

		path := c.Request.URL.Path
		splittedpath := strings.Split(path, "/")
		resourceUrl := "/" + splittedpath[2]

    fmt.Println(splittedpath)
    fmt.Println(resourceUrl)

		var permission model.Permission
		err := db.Raw(`
      SELECT 
        permissions.role_id as role_id,
        permissions.resource_id as resource_id,
        permissions.r as r,
        permissions.w as w,
        permissions.u as u,
        permissions.d as d
      FROM permissions
        INNER JOIN resources ON permissions.resource_id = resources.id
        INNER JOIN roles ON permissions.role_id = roles.id
      WHERE
        resources.url = ?
        AND roles.id = ?
    `, resourceUrl, roleID).
			Scan(&permission).
			Error
		if err != nil {
			response.ResponseError(c, fmt.Sprintf("Доступ запрещен"))
			c.Abort()
			return
		}

		requestMethod := c.Request.Method

		if requestMethod == "GET" && !permission.R {
			response.ResponseError(c, fmt.Sprintf("Доступ запрещен"))
			c.Abort()
			return
		}

		if requestMethod == "POST" && !permission.W {
			response.ResponseError(c, fmt.Sprintf("Доступ запрещен"))
			c.Abort()
			return
		}

		if requestMethod == "PATCH" && !permission.U {
			response.ResponseError(c, fmt.Sprintf("Доступ запрещен"))
			c.Abort()
			return
		}

		if requestMethod == "DELETE" && !permission.D {
			response.ResponseError(c, fmt.Sprintf("Доступ запрещен"))
			c.Abort()
			return
		}

		c.Next()
	}
}
