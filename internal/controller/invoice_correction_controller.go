package controller

import (
	"backend-v2/internal/dto"
	"backend-v2/internal/service"
	"backend-v2/pkg/response"
	"fmt"
	"os"
	"path/filepath"
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
	UniqueObject(c *gin.Context)
	UniqueTeam(c *gin.Context)
	Report(c *gin.Context)
	GetPaginated(c *gin.Context)
}

func (controller *invoiceCorrectionController) GetPaginated(c *gin.Context) {
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

	data, err := controller.invoiceCorrectionService.GetPaginated(page, limit, projectID)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Внутренняя ошибка сервера: %v", err))
		return
	}

	dataCount, err := controller.invoiceCorrectionService.Count(projectID)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Внутренняя ошибка сервера: %v", err))
		return
	}

	response.ResponsePaginatedData(c, data, dataCount)
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

	createData.Details.OperatorWorkerID = c.GetUint("workerID")

	data, err := controller.invoiceCorrectionService.Create(createData)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Внутренняя ошибка сервера: %v", err))
		return
	}

	response.ResponseSuccess(c, data)
}

func (controller *invoiceCorrectionController) UniqueObject(c *gin.Context) {
	projectID := c.GetUint("projectID")
	data, err := controller.invoiceCorrectionService.UniqueObject(projectID)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Internal Server Error: %v", err))
		return
	}

	response.ResponseSuccess(c, data)
}

func (controller *invoiceCorrectionController) UniqueTeam(c *gin.Context) {
	projectID := c.GetUint("projectID")
	data, err := controller.invoiceCorrectionService.UniqueTeam(projectID)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Internal Server Error: %v", err))
		return
	}

	response.ResponseSuccess(c, data)
}

func (controller *invoiceCorrectionController) Report(c *gin.Context) {
	var filter dto.InvoiceCorrectionReportFilter
	if err := c.ShouldBindJSON(&filter); err != nil {
		response.ResponseError(c, fmt.Sprintf("Неверное тело запроса: %v", err))
		return
	}

	filter.ProjectID = c.GetUint("projectID")

	reportFileName, err := controller.invoiceCorrectionService.Report(filter)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Internal Server Error: %v", err))
		return
	}

	reportFilePath := filepath.Join("./pkg/excels/temp/", reportFileName)
	c.FileAttachment(reportFilePath, reportFileName)
	os.Remove(reportFilePath)
}
