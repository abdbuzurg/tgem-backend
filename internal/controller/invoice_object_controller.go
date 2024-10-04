package controller

import (
	"backend-v2/internal/dto"
	"backend-v2/internal/service"
	"backend-v2/pkg/response"
	"fmt"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type invoiceObjectController struct {
	invoiceObjectService service.IInvoiceObjectService
}

func InitInvoiceObjectController(
	invoiceObjectService service.IInvoiceObjectService,
) IInvoiceObjectController {
	return &invoiceObjectController{
		invoiceObjectService: invoiceObjectService,
	}
}

type IInvoiceObjectController interface {
	GetInvoiceObjectDescriptiveDataByID(c *gin.Context)
	GetTeamsMaterials(c *gin.Context)
	GetSerialNumbersOfMaterial(c *gin.Context)
	Create(c *gin.Context)
	GetMaterialAmountInTeam(c *gin.Context)
	GetPaginated(c *gin.Context)
	GetTeamsFromObjectID(c *gin.Context)
	GetOperationsBasedOnMaterialsInTeamID(c *gin.Context)
}

func (controller *invoiceObjectController) GetInvoiceObjectDescriptiveDataByID(c *gin.Context) {
	idRaw := c.Param("id")
	id, err := strconv.ParseUint(idRaw, 10, 64)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Incorrect parameter provided: %v", err))
		return
	}

	data, err := controller.invoiceObjectService.GetInvoiceObjectDescriptiveDataByID(uint(id))
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Internal Server Error: %v", err))
		return
	}

	response.ResponseSuccess(c, data)

}

func (controller *invoiceObjectController) GetPaginated(c *gin.Context) {
	pageStr := c.DefaultQuery("page", "1")
	page, err := strconv.Atoi(pageStr)
	if err != nil || page <= 0 {
		response.ResponseError(c, fmt.Sprintf("Wrong query parameter provided for page: %v", err))
		return
	}

	limitStr := c.DefaultQuery("limit", "10")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		response.ResponseError(c, fmt.Sprintf("Wrong query parameter provided for limit: %v", err))
		return
	}

	projectID := c.GetUint("projectID")

	data, err := controller.invoiceObjectService.GetPaginated(limit, page, projectID)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Internal Server Error: %v", err))
		return
	}

	dataCount, err := controller.invoiceObjectService.Count(projectID)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Internal Server Error: %v", err))
		return
	}

	response.ResponsePaginatedData(c, data, dataCount)
}

func (controller *invoiceObjectController) Delete(c *gin.Context) {
	idRaw := c.Param("id")
	id, err := strconv.ParseUint(idRaw, 10, 64)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Incorrect parameter provided: %v", err))
		return
	}

	err = controller.invoiceObjectService.Delete(uint(id))
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Internal Server Error: %v", err))
		return
	}

	response.ResponseSuccess(c, true)
}

func (controller *invoiceObjectController) GetTeamsMaterials(c *gin.Context) {
	teamIDRaw := c.Param("teamID")
	teamID, err := strconv.ParseUint(teamIDRaw, 10, 64)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Incorrect parameter provided: %v", err))
		return
	}

	projectID := c.GetUint("projectID")

	data, err := controller.invoiceObjectService.GetTeamsMaterials(projectID, uint(teamID))
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Internal Server Error: %v", err))
		return
	}

	response.ResponseSuccess(c, data)

}

func (controller *invoiceObjectController) GetSerialNumbersOfMaterial(c *gin.Context) {
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

	data, err := controller.invoiceObjectService.GetSerialNumberOfMaterial(projectID, uint(materialID), uint(teamID))
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Internal Server Error: %v", err))
		return
	}

	response.ResponseSuccess(c, data)

}

func (controller *invoiceObjectController) Create(c *gin.Context) {

	var data dto.InvoiceObjectCreate
	if err := c.ShouldBindJSON(&data); err != nil {
		response.ResponseError(c, fmt.Sprintf("Неверное тело запроса: %v", err))
		return
	}

	workerID := c.GetUint("workerID")
	data.Details.SupervisorWorkerID = workerID

	projectID := c.GetUint("projectID")
	data.Details.ProjectID = projectID

	location, err := time.LoadLocation("Asia/Dushanbe")
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Неверное тело запроса: %v", err))
		return
	}
	date := time.Now().In(location)
	data.Details.DateOfInvoice = date

	_, err = controller.invoiceObjectService.Create(data)
	if err != nil {

		response.ResponseError(c, fmt.Sprintf("Внутренняя ошибка сервера: %v", err))
		return

	}

	response.ResponseSuccess(c, true)
}

func (controller *invoiceObjectController) GetMaterialAmountInTeam(c *gin.Context) {

	projectID := c.GetUint("projectID")

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

	data, err := controller.invoiceObjectService.GetAvailableMaterialAmount(projectID, uint(materialID), uint(teamID))
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Внутренняя ошибка сервера: %v", err))
		return
	}

	response.ResponseSuccess(c, data)
}

func (controller *invoiceObjectController) GetTeamsFromObjectID(c *gin.Context) {

	objectIDRaw := c.Param("objectID")
	objectID, err := strconv.ParseUint(objectIDRaw, 10, 64)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Incorrect parameter provided: %v", err))
		return
	}

	data, err := controller.invoiceObjectService.GetTeamsFromObjectID(uint(objectID))
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Внутренняя ошибка сервера: %v", err))
		return
	}

	response.ResponseSuccess(c, data)
}

func (controller *invoiceObjectController) GetOperationsBasedOnMaterialsInTeamID(c *gin.Context) {

	teamIDRaw := c.Param("teamID")
	teamID, err := strconv.ParseUint(teamIDRaw, 10, 64)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Incorrect parameter provided: %v", err))
		return
	}

	data, err := controller.invoiceObjectService.GetOperationsBasedOnMaterialsInTeamID(c.GetUint("projectID"), uint(teamID))
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Внутренняя ошибка сервера: %v", err))
		return
	}

	response.ResponseSuccess(c, data)
}
