package controller

import (
	"backend-v2/internal/service"
	"backend-v2/model"
	"backend-v2/pkg/response"
	"fmt"
	"strconv"

	"github.com/gin-gonic/gin"
)

type serialNumberController struct {
	serialNumberService service.ISerialNumberService
}

func InitSerialNumberController(serialNumberService service.ISerialNumberService) ISerialNumberController {
	return &serialNumberController{
		serialNumberService: serialNumberService,
	}
}

type ISerialNumberController interface {
	GetAll(c *gin.Context)
	GetCodes(c *gin.Context)
	Create(c *gin.Context)
	Update(c *gin.Context)
	Delete(c *gin.Context)
}

func (controller *serialNumberController) GetAll(c *gin.Context) {
	data, err := controller.serialNumberService.GetAll()
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Internal Server Error: %v", err))
		return
	}

	response.ResponseSuccess(c, data)
}

func (controller *serialNumberController) GetCodes(c *gin.Context) {
	materialIDRaw := c.Param("materialID")
	materialID, err := strconv.ParseUint(materialIDRaw, 10, 64)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Incorrect request parameter: %v", err))
		return
	}

	codes, err := controller.serialNumberService.GetCodesByMaterialID(uint(materialID))
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Internal server error: %v", err))
		return
	}

	response.ResponseSuccess(c, codes)
}

func (controller *serialNumberController) Create(c *gin.Context) {
	var data model.SerialNumber
	if err := c.ShouldBindJSON(&data); err != nil {
		response.ResponseError(c, fmt.Sprintf("Incorrect request body: %v", err))
		return
	}

	data, err := controller.serialNumberService.Create(data)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Internal Server Error: %v", err))
		return
	}

	response.ResponseSuccess(c, data)
}

func (controller *serialNumberController) Update(c *gin.Context) {
	var data model.SerialNumber
	if err := c.ShouldBindJSON(&data); err != nil {
		response.ResponseError(c, fmt.Sprintf("Incorrect request body: %v", err))
		return
	}

	data, err := controller.serialNumberService.Update(data)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Internal Server Errror: %v", err))
		return
	}

	response.ResponseSuccess(c, data)
}

func (controller *serialNumberController) Delete(c *gin.Context) {
	idRaw := c.Param("id")
	id, err := strconv.ParseUint(idRaw, 10, 64)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Incorrect request parameter: %v", err))
		return
	}

	err = controller.serialNumberService.Delete(uint(id))
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Internal Server error: %v", err))
		return
	}

	response.ResponseSuccess(c, true)
}
