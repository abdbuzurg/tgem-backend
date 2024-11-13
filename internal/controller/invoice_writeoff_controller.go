package controller

import (
	// "backend-v2/internal/dto"
	"backend-v2/internal/dto"
	"backend-v2/internal/service"
	"backend-v2/pkg/response"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

type invoiceWriteOffController struct {
	invoiceWriteOffService service.IInvoiceWriteOffService
}

func InitInvoiceWriteOffController(invoiceWriteOffService service.IInvoiceWriteOffService) IInvoiceWriteOffController {
	return &invoiceWriteOffController{
		invoiceWriteOffService: invoiceWriteOffService,
	}
}

type IInvoiceWriteOffController interface {
	GetPaginated(c *gin.Context)
	GetInvoiceMaterialsWithoutSerialNumber(c *gin.Context)
	Create(c *gin.Context)
	Update(c *gin.Context)
	Delete(c *gin.Context)
	GetMaterialsForEdit(c *gin.Context)
	GetRawDocument(c *gin.Context)
	Confirmation(c *gin.Context)
	GetDocument(c *gin.Context)
	Report(c *gin.Context)
	GetMaterialsInLocation(c *gin.Context)
}

func (controller *invoiceWriteOffController) GetPaginated(c *gin.Context) {
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

	writeOffType := c.DefaultQuery("writeOffType", "")
	writeOffType, err = url.QueryUnescape(writeOffType)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Wrong query parameter provided for returnerType: %v", err))
		return
	}

	projectID := c.GetUint("projectID")
	filter := dto.InvoiceWriteOffSearchParameters{
		ProjectID:    projectID,
		WriteOffType: writeOffType,
	}

	data, err := controller.invoiceWriteOffService.GetPaginated(page, limit, filter)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Could not get the paginated data of Invoice: %v", err))
		return
	}

	dataCount, err := controller.invoiceWriteOffService.Count(filter)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Could not get the total amount of Invoice: %v", err))
		return
	}

	response.ResponsePaginatedData(c, data, dataCount)
}

func (controller *invoiceWriteOffController) Create(c *gin.Context) {
	var createData dto.InvoiceWriteOff
	if err := c.ShouldBindJSON(&createData); err != nil {
		response.ResponseError(c, fmt.Sprintf("Invalid data recieved by server: %v", err))
		return
	}

	createData.Details.ProjectID = c.GetUint("projectID")
	createData.Details.ReleasedWorkerID = c.GetUint("workerID")

	data, err := controller.invoiceWriteOffService.Create(createData)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Could perform the creation of Invoice: %v", err))
		return
	}

	response.ResponseSuccess(c, data)
}

func (controller *invoiceWriteOffController) Update(c *gin.Context) {
	var updateData dto.InvoiceWriteOff
	if err := c.ShouldBindJSON(&updateData); err != nil {
		response.ResponseError(c, fmt.Sprintf("Invalid data recieved by server: %v", err))
		return
	}

	updateData.Details.ProjectID = c.GetUint("projectID")
	updateData.Details.ReleasedWorkerID = c.GetUint("workerID")

	data, err := controller.invoiceWriteOffService.Update(updateData)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Could not perform the updation of Invoice: %v", err))
		return
	}

	response.ResponseSuccess(c, data)
}

func (controller *invoiceWriteOffController) Delete(c *gin.Context) {
	idRaw := c.Param("id")
	id, err := strconv.ParseUint(idRaw, 10, 64)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Incorrect parameter provided: %v", err))
		return
	}

	err = controller.invoiceWriteOffService.Delete(uint(id))
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Could not perform the deletion of Invoice: %v", err))
		return
	}

	response.ResponseSuccess(c, "deleted")
}

func (controller *invoiceWriteOffController) GetRawDocument(c *gin.Context) {
	deliveryCode := c.Param("deliveryCode")
	c.FileAttachment("./pkg/excels/writeoff/"+deliveryCode+".xlsx", deliveryCode+".xlsx")
}

func (controller *invoiceWriteOffController) GetInvoiceMaterialsWithoutSerialNumber(c *gin.Context) {

	idRaw := c.Param("id")
	id, err := strconv.ParseUint(idRaw, 10, 64)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Неверное тело запроса: %v", err))
		return
	}

	data, err := controller.invoiceWriteOffService.GetInvoiceMaterialsWithoutSerialNumbers(uint(id))
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Внутренняя ошибка сервера: %v", err))
		return
	}

	response.ResponseSuccess(c, data)

}

func (controller *invoiceWriteOffController) GetMaterialsForEdit(c *gin.Context) {
	idRaw := c.Param("id")
	id, err := strconv.ParseUint(idRaw, 10, 64)

	locationIDRaw := c.Param("locationID")
	locationID, err := strconv.Atoi(locationIDRaw)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Invalid parameters in request: %v", err))
		return
	}

	locationType := c.Param("locationType")

	result, err := controller.invoiceWriteOffService.GetMaterialsForEdit(uint(id), locationType, uint(locationID))
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Внутренняя ошибка сервера: %v", err))
		return
	}

	response.ResponseSuccess(c, result)
}

func (controller *invoiceWriteOffController) Confirmation(c *gin.Context) {

	projectID := c.GetUint("projectID")

	idRaw := c.Param("id")
	id, err := strconv.ParseUint(idRaw, 10, 64)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Неверное тело запроса: %v", err))
		return
	}

	invoiceWriteOff, err := controller.invoiceWriteOffService.GetByID(uint(id))
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Внутренняя ошибка сервера: %v", err))
		return
	}

	file, err := c.FormFile("file")
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Внутренняя ошибка сервера: %v", err))
		return
	}

	fileNameAndExtension := strings.Split(file.Filename, ".")
	fileExtension := fileNameAndExtension[len(fileNameAndExtension) - 1]
	if fileExtension != "pdf" {
		response.ResponseError(c, fmt.Sprintf("Файл должен быть формата PDF"))
		return
	}
	file.Filename = invoiceWriteOff.DeliveryCode + "." + fileExtension
	filePath := filepath.Join("./pkg/excels/writeoff/", file.Filename)

	err = c.SaveUploadedFile(file, filePath)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Внутренняя ошибка сервера: %v", err))
		return
	}

	err = controller.invoiceWriteOffService.Confirmation(uint(id), projectID)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Внутренняя ошибка сервера: %v", err))
		return
	}

	response.ResponseSuccess(c, true)
}

func (controller *invoiceWriteOffController) GetDocument(c *gin.Context) {

	deliveryCode := c.Param("deliveryCode")

	filePath := filepath.Join("./pkg/excels/writeoff/", deliveryCode)
	fileGlob, err := filepath.Glob(filePath + ".*")
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Внутренняя ошибка сервера: %v", err))
		return
	}

	filePath = fileGlob[0]
	pathSeparated := strings.Split(filePath, ".")
	deliveryCodeExtension := pathSeparated[len(pathSeparated)-1]

	c.FileAttachment(filePath, deliveryCode+"."+deliveryCodeExtension)
}

func (controller *invoiceWriteOffController) Report(c *gin.Context) {
	var reportParameters dto.InvoiceWriteOffReportParameters
	if err := c.ShouldBindJSON(&reportParameters); err != nil {
		response.ResponseError(c, fmt.Sprintf("Invalid data recieved by server: %v", err))
		return
	}

	reportParameters.ProjectID = c.GetUint("projectID")

	filename, err := controller.invoiceWriteOffService.Report(reportParameters)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Внутренняя ошибка сервера: %v", err))
		return
	}

	filePath := filepath.Join("./pkg/excels/temp/", filename)
	c.FileAttachment(filePath, filename)
	os.Remove(filePath)
}

func (controller *invoiceWriteOffController) GetMaterialsInLocation(c *gin.Context) {

	projectID := c.GetUint("projectID")

	locationIDRaw := c.Param("locationID")
	locationID, err := strconv.Atoi(locationIDRaw)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Invalid parameters in request: %v", err))
		return
	}

	locationType := c.Param("locationType")

	data, err := controller.invoiceWriteOffService.GetMaterialsInLocation(projectID, uint(locationID), locationType)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Internal server error: %v", err))
		return
	}

	response.ResponseSuccess(c, data)
}
