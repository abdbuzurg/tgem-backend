package controller

import (
	"backend-v2/internal/dto"
	"backend-v2/internal/service"
	"backend-v2/model"
	"backend-v2/pkg/response"
	"fmt"
	"net/url"
	"strconv"

	"github.com/gin-gonic/gin"
)

type invoiceOutputController struct {
	invoiceOutputService service.IInvoiceOutputService
}

func InitInvoiceOutputController(invoiceOutputService service.IInvoiceOutputService) IInvoiceOutputController {
	return &invoiceOutputController{
		invoiceOutputService: invoiceOutputService,
	}
}

type IInvoiceOutputController interface {
	GetAll(c *gin.Context)
	GetPaginated(c *gin.Context)
	Create(c *gin.Context)
	// Update(c *gin.Context)
	Delete(c *gin.Context)
	GetDocument(c *gin.Context)
	GetInvoiceMaterialsWithoutSerialNumbers(c *gin.Context)
	GetInvoiceMaterialsWithSerialNumbers(c *gin.Context)
	Confirmation(c *gin.Context)
	UniqueCode(c *gin.Context)
	UniqueWarehouseManager(c *gin.Context)
	UniqueRecieved(c *gin.Context)
	UniqueDistrict(c *gin.Context)
	UniqueObject(c *gin.Context)
	UniqueTeam(c *gin.Context)
	Report(c *gin.Context)
	GetTotalAmountInWarehouse(c *gin.Context)
	GetCodesByMaterialID(c *gin.Context)
	GetAvailableMaterialsInWarehouse(c *gin.Context)
}

func (controller *invoiceOutputController) GetAll(c *gin.Context) {
	data, err := controller.invoiceOutputService.GetAll()
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Could not get Invoice Input data: %v", err))
		return
	}

	response.ResponseSuccess(c, data)
}

func (controller *invoiceOutputController) GetPaginated(c *gin.Context) {
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

	districtIDStr := c.DefaultQuery("districtID", "")
	districtID := 0
	if districtIDStr != "" {
		districtID, err = strconv.Atoi(districtIDStr)
		if err != nil {
			response.ResponseError(c, fmt.Sprintf("Cannot decode invoiceCo parameter: %v", err))
			return
		}
	}

	warehouseManagerWorkerIDStr := c.DefaultQuery("warehouseManagerWorkerID", "")
	warehouseManagerWorkerID := 0
	if warehouseManagerWorkerIDStr != "" {
		warehouseManagerWorkerID, err = strconv.Atoi(warehouseManagerWorkerIDStr)
		if err != nil {
			response.ResponseError(c, fmt.Sprintf("Cannot decode invoiceCo parameter: %v", err))
			return
		}
	}

	releasedWorkerIDStr := c.DefaultQuery("releasedWorkerID", "")
	releasedWorkerID := 0
	if releasedWorkerIDStr != "" {
		releasedWorkerID, err = strconv.Atoi(releasedWorkerIDStr)
		if err != nil {
			response.ResponseError(c, fmt.Sprintf("Cannot decode releasedWorkerID parameter: %v", err))
			return
		}
	}

	recipientWorkerIDStr := c.DefaultQuery("recipientWorkerID", "")
	recipientWorkerID := 0
	if recipientWorkerIDStr != "" {
		recipientWorkerID, err = strconv.Atoi(recipientWorkerIDStr)
		if err != nil {
			response.ResponseError(c, fmt.Sprintf("Cannot decode recipientWorkerID parameter: %v", err))
			return
		}
	}

	teamIDStr := c.DefaultQuery("teamID", "")
	teamID := 0
	if teamIDStr != "" {
		teamID, err = strconv.Atoi(teamIDStr)
		if err != nil {
			response.ResponseError(c, fmt.Sprintf("Cannot decode teamID parameter: %v", err))
			return
		}
	}

	objectIDStr := c.DefaultQuery("objectID", "")
	objectID := 0
	if objectIDStr != "" {
		objectID, err = strconv.Atoi(objectIDStr)
		if err != nil {
			response.ResponseError(c, fmt.Sprintf("Cannot decode objectID parameter: %v", err))
			return
		}
	}

	deliveryCode := c.DefaultQuery("deliveryCode", "")
	deliveryCode, err = url.QueryUnescape(deliveryCode)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Wrong query parameter provided for deliveryCode: %v", err))
		return
	}

	projectID := c.GetUint("projectID")
	filter := model.InvoiceOutput{
		ProjectID:                projectID,
		DistrictID:               uint(districtID),
		WarehouseManagerWorkerID: uint(warehouseManagerWorkerID),
		RecipientWorkerID:        uint(recipientWorkerID),
		ReleasedWorkerID:         uint(releasedWorkerID),
		TeamID:                   uint(teamID),
		ObjectID:                 uint(objectID),
		DeliveryCode:             deliveryCode,
	}

	data, err := controller.invoiceOutputService.GetPaginated(page, limit, filter)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Could not get the paginated data of Invoice: %v", err))
		return
	}

	dataCount, err := controller.invoiceOutputService.Count(projectID)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Could not get the total amount of Invoice: %v", err))
		return
	}

	response.ResponsePaginatedData(c, data, dataCount)
}

func (controller *invoiceOutputController) GetInvoiceMaterialsWithoutSerialNumbers(c *gin.Context) {

	idRaw := c.Param("id")
	id, err := strconv.ParseUint(idRaw, 10, 64)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Неверное тело запроса: %v", err))
		return
	}

	data, err := controller.invoiceOutputService.GetInvoiceMaterialsWithoutSerialNumbers(uint(id))
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Внутренняя ошибка сервера: %v", err))
		return
	}

	response.ResponseSuccess(c, data)
}

func (controller *invoiceOutputController) GetInvoiceMaterialsWithSerialNumbers(c *gin.Context) {

	idRaw := c.Param("id")
	id, err := strconv.ParseUint(idRaw, 10, 64)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Неверное тело запроса: %v", err))
		return
	}

	data, err := controller.invoiceOutputService.GetInvoiceMaterialsWithSerialNumbers(uint(id))
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Внутренняя ошибка сервера: %v", err))
		return
	}

	response.ResponseSuccess(c, data)
}

func (controller *invoiceOutputController) Create(c *gin.Context) {
	var createData dto.InvoiceOutput
	if err := c.ShouldBindJSON(&createData); err != nil {
		response.ResponseError(c, fmt.Sprintf("Invalid data recieved by server: %v", err))
		return
	}

	workerID := c.GetUint("workerID")
	createData.Details.ReleasedWorkerID = workerID

	projectID := c.GetUint("projectID")
	createData.Details.ProjectID = projectID

	data, err := controller.invoiceOutputService.Create(createData)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Could perform the creation of Invoice: %v", err))
		return
	}

	response.ResponseSuccess(c, data)
}

func (controller *invoiceOutputController) Delete(c *gin.Context) {
	idRaw := c.Param("id")
	id, err := strconv.ParseUint(idRaw, 10, 64)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Incorrect parameter provided: %v", err))
		return
	}

	err = controller.invoiceOutputService.Delete(uint(id))
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Could not perform the deletion of Invoice: %v", err))
		return
	}

	response.ResponseSuccess(c, "deleted")
}

func (controller *invoiceOutputController) Confirmation(c *gin.Context) {
	idRaw := c.Param("id")
	id, err := strconv.ParseUint(idRaw, 10, 64)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Incorrect parameter provided: %v", err))
		return
	}

	invoiceOutput, err := controller.invoiceOutputService.GetByID(uint(id))
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Cannot find invoice Output by id %v: %v", id, err))
	}

	file, err := c.FormFile("file")
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("cannot form file: %v", err))
		return
	}
	file.Filename = invoiceOutput.DeliveryCode

	filePath := "./pkg/excels/output/" + file.Filename
	err = c.SaveUploadedFile(file, filePath)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("cannot save file: %v", err))
		return
	}

	err = controller.invoiceOutputService.Confirmation(uint(id))
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("cannot confirm invoice input with id %v: %v", id, err))
		return
	}

	response.ResponseSuccess(c, true)
}

func (controller *invoiceOutputController) GetDocument(c *gin.Context) {
	deliveryCode := c.Param("deliveryCode")
	c.FileAttachment("./pkg/excels/output/"+deliveryCode+".xlsx", deliveryCode+".xlsx")
}

func (controller *invoiceOutputController) UniqueCode(c *gin.Context) {
	projectID := c.GetUint("projectID")
	data, err := controller.invoiceOutputService.UniqueCode(projectID)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Internal server error: %v", err))
		return
	}

	response.ResponseSuccess(c, data)
}

func (controller *invoiceOutputController) UniqueWarehouseManager(c *gin.Context) {
	projectID := c.GetUint("projectID")
	data, err := controller.invoiceOutputService.UniqueWarehouseManager(projectID)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Internal server error: %v", err))
		return
	}

	response.ResponseSuccess(c, data)
}

func (controller *invoiceOutputController) UniqueRecieved(c *gin.Context) {
	projectID := c.GetUint("projectID")
	data, err := controller.invoiceOutputService.UniqueRecieved(projectID)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Internal server error: %v", err))
		return
	}

	response.ResponseSuccess(c, data)
}

func (controller *invoiceOutputController) UniqueDistrict(c *gin.Context) {
	projectID := c.GetUint("projectID")
	data, err := controller.invoiceOutputService.UniqueDistrict(projectID)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Internal server error: %v", err))
		return
	}

	response.ResponseSuccess(c, data)
}

func (controller *invoiceOutputController) UniqueObject(c *gin.Context) {
	projectID := c.GetUint("projectID")
	data, err := controller.invoiceOutputService.UniqueObject(projectID)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Internal server error: %v", err))
		return
	}

	response.ResponseSuccess(c, data)
}

func (controller *invoiceOutputController) UniqueTeam(c *gin.Context) {
	projectID := c.GetUint("projectID")
	data, err := controller.invoiceOutputService.UniqueTeam(projectID)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Internal server error: %v", err))
		return
	}

	response.ResponseSuccess(c, data)
}

func (controller *invoiceOutputController) Report(c *gin.Context) {
	var filter dto.InvoiceOutputReportFilterRequest
	if err := c.ShouldBindJSON(&filter); err != nil {
		response.ResponseError(c, fmt.Sprintf("Invalid data recieved by server: %v", err))
		return
	}

	projectID := c.GetUint("projectID")
	filename, err := controller.invoiceOutputService.Report(filter, projectID)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Internal server error: %v", err))
		return
	}

	c.FileAttachment("./pkg/excels/report/"+filename, filename)
	// response.ResponseSuccess(c, true)
}

func (controller *invoiceOutputController) GetTotalAmountInWarehouse(c *gin.Context) {
	materialIDRaw := c.Param("materialID")
	materialID, err := strconv.ParseUint(materialIDRaw, 10, 64)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Incorrect parameter provided: %v", err))
		return
	}

	if materialID == 0 {
		response.ResponseSuccess(c, 0)
		return
	}

	projectID := c.GetUint("projectID")

	totalAmount, err := controller.invoiceOutputService.GetTotalMaterialAmount(projectID, uint(materialID))
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Internal server error: %v", err))
		return
	}

	response.ResponseSuccess(c, totalAmount)
}

func (controller *invoiceOutputController) GetCodesByMaterialID(c *gin.Context) {

	materialIDRaw := c.Param("materialID")
	materialID, err := strconv.ParseUint(materialIDRaw, 10, 64)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Incorrect parameter provided: %v", err))
		return
	}

	projectID := c.GetUint("projectID")

	data, err := controller.invoiceOutputService.GetSerialNumbersByMaterial(projectID, uint(materialID))
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Internal server error: %v", err))
		return
	}

	response.ResponseSuccess(c, data)

}

func (controller *invoiceOutputController) GetAvailableMaterialsInWarehouse(c *gin.Context) {
	projectID := c.GetUint("projectID")

	data, err := controller.invoiceOutputService.GetAvailableMaterialsInWarehouse(projectID)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Internal server error: %v", err))
		return
	}

	response.ResponseSuccess(c, data)

}
