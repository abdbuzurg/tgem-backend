package controller

import (
	"backend-v2/internal/dto"
	"backend-v2/internal/service"
	"backend-v2/pkg/response"
	"fmt"
	"strconv"

	"github.com/gin-gonic/gin"
)

type invoiceCorrectionController struct {
	invoiceCorrectionService service.IInvoiceCorrectionService
}

func InitInvoiceCorrectionController(
	invoiceCorrectionService service.IInvoiceCorrectionService,
) IInvoiceCorrectionController {
	return &invoiceCorrectionController{
		invoiceCorrectionService: invoiceCorrectionService,
	}
}

type IInvoiceCorrectionController interface {
	GetAll(c *gin.Context)
	GetTotalMaterialInTeamByTeamNumber(c *gin.Context)
	GetInvoiceMaterialsByInvoiceObjectID(c *gin.Context)
	GetSerialNumbersOfMaterial(c *gin.Context)
	Create(c *gin.Context)
}

func (controller *invoiceCorrectionController) GetAll(c *gin.Context) {

	projectID := c.GetUint("projectID")

	data, err := controller.invoiceCorrectionService.GetAll(projectID)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Internal server error: %v", err))
		return
	}

	response.ResponseSuccess(c, data)
}

func (controller *invoiceCorrectionController) GetTotalMaterialInTeamByTeamNumber(c *gin.Context) {

	materialIDRaw := c.Param("materialID")
	materialID, err := strconv.ParseUint(materialIDRaw, 10, 64)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Incorrect parameter provided: %v", err))
		return
	}

	teamNumber := c.Param("teamNumber")
	projectID := c.GetUint("projectID")

	data, err := controller.invoiceCorrectionService.GetTotalAmounInLocationByTeamName(projectID, uint(materialID), teamNumber)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Internal server error: %v", err))
		return
	}

	response.ResponseSuccess(c, data)
}

func (controller *invoiceCorrectionController) GetInvoiceMaterialsByInvoiceObjectID(c *gin.Context) {

	idRaw := c.Param("id")
	id, err := strconv.ParseUint(idRaw, 10, 64)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Incorrect parameter provided: %v", err))
		return
	}

	data, err := controller.invoiceCorrectionService.GetInvoiceMaterialsByInvoiceObjectID(uint(id))
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Internal server error: %v", err))
		return
	}

	response.ResponseSuccess(c, data)

}

func (controller *invoiceCorrectionController) GetSerialNumbersOfMaterial(c *gin.Context) {
	materialIDRaw := c.Param("materialID")
	materialID, err := strconv.ParseUint(materialIDRaw, 10, 64)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Incorrect parameter provided: %v", err))
		return
	}

	teamIDRaw := c.Param("teamID")
	teamID, err := strconv.ParseUint(teamIDRaw, 10, 64)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Incorrect parameter provided: %v", err))
		return
	}

	projectID := c.GetUint("projectID")

	data, err := controller.invoiceCorrectionService.GetSerialNumberOfMaterialInTeam(projectID, uint(materialID), uint(teamID))
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Internal Server Error: %v", err))
		return
	}

	response.ResponseSuccess(c, data)

}

func (controller *invoiceCorrectionController) Create(c *gin.Context) {
	var createData dto.InvoiceCorrectionCreate
	if err := c.ShouldBindJSON(&createData); err != nil {
		response.ResponseError(c, fmt.Sprintf("Неверное тело запроса: %v", err))
		return
	}

	data, err := controller.invoiceCorrectionService.Create(createData)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Внутренняя ошибка сервера: %v", err))
		return
	}

  response.ResponseSuccess(c, data)
}
