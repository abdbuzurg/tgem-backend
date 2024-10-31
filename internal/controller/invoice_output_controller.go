package controller

import (
	"backend-v2/internal/dto"
	"backend-v2/internal/service"
	"backend-v2/model"
	"backend-v2/pkg/response"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

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
	Update(c *gin.Context)
	Delete(c *gin.Context)
	GetDocument(c *gin.Context)
	GetInvoiceMaterialsWithoutSerialNumbers(c *gin.Context)
	GetInvoiceMaterialsWithSerialNumbers(c *gin.Context)
	Confirmation(c *gin.Context)
	UniqueCode(c *gin.Context)
	UniqueWarehouseManager(c *gin.Context)
	UniqueRecieved(c *gin.Context)
	UniqueDistrict(c *gin.Context)
	UniqueTeam(c *gin.Context)
	Report(c *gin.Context)
	GetTotalAmountInWarehouse(c *gin.Context)
	GetCodesByMaterialID(c *gin.Context)
	GetAvailableMaterialsInWarehouse(c *gin.Context)
	GetMaterialsForEdit(c *gin.Context)
	Import(c *gin.Context)
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

	projectID := c.GetUint("projectID")
	filter := model.InvoiceOutput{
		ProjectID: projectID,
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

	fileNameAndExtension := strings.Split(file.Filename, ".")
	fileExtension := fileNameAndExtension[len(fileNameAndExtension)-1]
	if fileExtension != "pdf" {
		response.ResponseError(c, fmt.Sprintf("Файл должен быть формата PDF"))
		return
	}
	file.Filename = invoiceOutput.DeliveryCode + "." + fileExtension
	filePath := filepath.Join("./pkg/excels/output/", file.Filename)

	err = c.SaveUploadedFile(file, filePath)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("cannot save file: %v", err))
		return
	}

	excelFilePath := filepath.Join("./pkg/excels/output/", invoiceOutput.DeliveryCode+".xlsx")
	os.Remove(excelFilePath)

	err = controller.invoiceOutputService.Confirmation(uint(id))
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("cannot confirm invoice input with id %v: %v", id, err))
		return
	}

	response.ResponseSuccess(c, true)
}

func (controller *invoiceOutputController) GetDocument(c *gin.Context) {
	deliveryCode := c.Param("deliveryCode")
	filePath := filepath.Join("./pkg/excels/output/", deliveryCode+".xlsx")
	c.FileAttachment(filePath, deliveryCode+".xlsx")
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

	filter.ProjectID = c.GetUint("projectID")
	filename, err := controller.invoiceOutputService.Report(filter)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Internal server error: %v", err))
		return
	}

	filePath := filepath.Join("./pkg/excels/temp/", filename)
	c.FileAttachment(filePath, filename)
	os.Remove(filePath)
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

func (controller *invoiceOutputController) GetMaterialsForEdit(c *gin.Context) {
	idRaw := c.Param("id")
	id, err := strconv.ParseUint(idRaw, 10, 64)

	result, err := controller.invoiceOutputService.GetMaterialsForEdit(uint(id))
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Внутренняя ошибка сервера: %v", err))
		return
	}

	response.ResponseSuccess(c, result)
}

func (controller *invoiceOutputController) Update(c *gin.Context) {
	var updateData dto.InvoiceOutput
	if err := c.ShouldBindJSON(&updateData); err != nil {
		response.ResponseError(c, fmt.Sprintf("Invalid data recieved by server: %v", err))
		return
	}

	workerID := c.GetUint("workerID")
	updateData.Details.ReleasedWorkerID = workerID

	projectID := c.GetUint("projectID")
	updateData.Details.ProjectID = projectID

	data, err := controller.invoiceOutputService.Update(updateData)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Could perform the creation of Invoice: %v", err))
		return
	}

	response.ResponseSuccess(c, data)
}

func (controller *invoiceOutputController) Import(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Файл не может быть сформирован, проверьте файл: %v", err))
		return
	}

	date := time.Now()
	importFileName := date.Format("2006-01-02 15-04-05") + file.Filename
	importFilePath := filepath.Join("./pkg/excels/temp/", importFileName)
	err = c.SaveUploadedFile(file, importFilePath)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Файл не может быть сохранен на сервере: %v", err))
		return
	}

	projectID := c.GetUint("projectID")
	workerID := c.GetUint("workerID")
	err = controller.invoiceOutputService.Import(importFilePath, projectID, workerID)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Внутренняя ошибка сервера: %v", err))
		return
	}

	response.ResponseSuccess(c, true)

}
