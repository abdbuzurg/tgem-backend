package controller

import (
	"backend-v2/internal/service"
	"backend-v2/pkg/response"
	"fmt"
	"strconv"

	"github.com/gin-gonic/gin"
)

type userActionController struct {
	userActionService service.IUserActionService
}

func InitUserActionController(userActionService service.IUserActionService) IUserActionController {
	return &userActionController{
		userActionService: userActionService,
	}
}

type IUserActionController interface {
	GetAllByUserID(c *gin.Context)
}

func (controller *userActionController) GetAllByUserID(c *gin.Context) {
	userIDRaw := c.Param("userID")
	userID, err := strconv.ParseUint(userIDRaw, 10, 64)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Incorrect parameter provided: %v", err))
		return
	}

	data, err := controller.userActionService.GetAllByUserID(uint(userID))
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Internal Server Error: %v", err))
		return
	}

	response.ResponseSuccess(c, data)

}
