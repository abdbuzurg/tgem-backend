package controller

import (
	"backend-v2/internal/service"
	"backend-v2/model"
	"backend-v2/pkg/response"
	"fmt"
	"strconv"

	"github.com/gin-gonic/gin"
)

type permissionController struct {
	permissionService service.IPermissionService
}

func InitPermissionController(permissionService service.IPermissionService) IPermissionController {
  return &permissionController{
    permissionService: permissionService,
  }
}

type IPermissionController interface {
  GetAll(c *gin.Context)
  GetByRoleID(c *gin.Context)
  GetByRoleName(c *gin.Context)
  GetByResourceURL(c *gin.Context)
  Create(c *gin.Context)
  CreateBatch(c *gin.Context)
  Update(c *gin.Context)
  Delete(c *gin.Context)
}

func(controller *permissionController) GetAll(c *gin.Context) {
  data, err := controller.permissionService.GetAll()
  if err != nil {
    response.ResponseError(c, fmt.Sprintf("Internal server error: %v", err))
    return
  }

  response.ResponseSuccess(c, data)
}

func(controller *permissionController) GetByRoleID(c *gin.Context) {
  roleIDRaw := c.Param("roleID")
	roleID, err := strconv.ParseUint(roleIDRaw, 10, 64)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Incorrect parameter provided: %v", err))
		return
	}

  data, err := controller.permissionService.GetByRoleID(uint(roleID))
  if err != nil {
    response.ResponseError(c, fmt.Sprintf("Internal server error: %v", err))
    return
  }

  response.ResponseSuccess(c, data)
}

func(controller *permissionController) Create(c *gin.Context) {
  var requestData model.Permission
  if err := c.ShouldBindJSON(&requestData); err != nil {
    response.ResponseError(c, fmt.Sprintf("Invalid request body: %v", err))
    return
  }

  data, err := controller.permissionService.Create(requestData)
  if err != nil {
    response.ResponseError(c, fmt.Sprintf("Internal server error: %v", err))
    return
  }

  response.ResponseSuccess(c, data)
}

func(controller *permissionController) CreateBatch(c *gin.Context) {
  var requestData []model.Permission
  if err := c.ShouldBindJSON(&requestData); err != nil {
    response.ResponseError(c, fmt.Sprintf("Invalid request body: %v", err))
    return
  }

  err := controller.permissionService.CreateBatch(requestData)
  if err != nil {
    response.ResponseError(c, fmt.Sprintf("Internal server error: %v", err))
    return
  }

  response.ResponseSuccess(c, true)
}

func (controller *permissionController) Update(c *gin.Context) {
	var requestData model.Permission
	if err := c.ShouldBindJSON(&requestData); err != nil {
		response.ResponseError(c, fmt.Sprintf("Invalid body request: %v", err))
		return
	}

	data, err := controller.permissionService.Update(requestData)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Internal Server Error: %v", err))
		return
	}

	response.ResponseSuccess(c, data)
}

func (controller *permissionController) Delete(c *gin.Context) {
	idRaw := c.Param("id")
	id, err := strconv.ParseUint(idRaw, 10, 64)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Incorrect parameter provided: %v", err))
		return
	}

	err = controller.permissionService.Delete(uint(id))
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Internal Server Error: %v", err))
		return
	}

	response.ResponseSuccess(c, true)
}

func (controller *permissionController) GetByRoleName(c *gin.Context) {
  
  roleName := c.Param("roleName")

  permissions, err := controller.permissionService.GetByRoleName(roleName)
  if err != nil {
    response.ResponseError(c, fmt.Sprintf("Internal Server Error: %v", err))
		return
  }

  response.ResponseSuccess(c, permissions)

}

func(controller *permissionController) GetByResourceURL(c *gin.Context) {

  roleID := c.GetUint("roleID")

  resourceURL := c.Param("resourceURL")

  err := controller.permissionService.GetByResourceURL("/" + resourceURL, roleID)
  if err != nil {
    response.ResponseSuccess(c, false) 
    return
  }

  response.ResponseSuccess(c, true)
}
