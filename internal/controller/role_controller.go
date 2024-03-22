package controller

import (
	"backend-v2/internal/service"
	"backend-v2/model"
	"backend-v2/pkg/response"
	"fmt"
	"strconv"

	"github.com/gin-gonic/gin"
)

type roleController struct {
	roleService service.IRoleService
}

func InitRoleController(roleService service.IRoleService) IRoleController {
	return &roleController{
		roleService: roleService,
	}
}

type IRoleController interface{
  GetAll(c *gin.Context)
  Create(c *gin.Context)
  Update(c *gin.Context)
  Delete(c *gin.Context)
}

func (controller *roleController) GetAll(c *gin.Context) {
	data, err := controller.roleService.GetAll()
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Internal Server Error: %v", err))
		return
	}

	response.ResponseSuccess(c, data)
}

func (controller *roleController) Create(c *gin.Context) {
	var requestData model.Role
	if err := c.ShouldBindJSON(&requestData); err != nil {
		response.ResponseError(c, fmt.Sprintf("Invalid body request: %v", err))
		return
	}

	data, err := controller.roleService.Create(requestData)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Internal Server Error: %v", err))
		return
	}

	response.ResponseSuccess(c, data)
}

func (controller *roleController) Update(c *gin.Context) {
	var requestData model.Role
	if err := c.ShouldBindJSON(&requestData); err != nil {
		response.ResponseError(c, fmt.Sprintf("Invalid body request: %v", err))
		return
	}

	data, err := controller.roleService.Update(requestData)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Internal Server Error: %v", err))
		return
	}

	response.ResponseSuccess(c, data)
}

func (controller *roleController) Delete(c *gin.Context) {
	idRaw := c.Param("id")
	id, err := strconv.ParseUint(idRaw, 10, 64)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Incorrect parameter provided: %v", err))
		return
	}

	err = controller.roleService.Delete(uint(id))
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Internal Server Error: %v", err))
		return
	}

	response.ResponseSuccess(c, true)
}
