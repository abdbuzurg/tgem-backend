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

	"github.com/gin-gonic/gin"
)

type invoiceInputController struct {
	invoiceInputService service.IInvoiceInputService
	userActionService   service.IUserActionService
}

func InitInvoiceInputController(
	invoiceInputService service.IInvoiceInputService,
	userActionService service.IUserActionService,
) IInvoiceInputController {
	return &invoiceInputController{
		invoiceInputService: invoiceInputService,
		userActionService:   userActionService,
	}
}

type IInvoiceInputController interface {
	GetAll(c *gin.Context)
	GetPaginated(c *gin.Context)
	GetInvoiceMaterialsWithoutSerialNumbers(c *gin.Context)
	GetInvoiceMaterialsWithSerialNumbers(c *gin.Context)
	Create(c *gin.Context)
	Update(c *gin.Context)
	Delete(c *gin.Context)
	Confirmation(c *gin.Context)
	GetDocument(c *gin.Context)
	UniqueCode(c *gin.Context)
	UniqueReleased(c *gin.Context)
	UniqueWarehouseManager(c *gin.Context)
	Report(c *gin.Context)
	NewMaterial(c *gin.Context)
	NewMaterialCost(c *gin.Context)
	GetMaterialsForEdit(c *gin.Context)
}

func (controller *invoiceInputController) GetAll(c *gin.Context) {

	data, err := controller.invoiceInputService.GetAll()
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Could not get Invoice Input data: %v", err))
		return
	}

	response.ResponseSuccess(c, data)

}

func (controller *invoiceInputController) GetPaginated(c *gin.Context) {

	projectID := c.GetUint("projectID")

	pageStr := c.DefaultQuery("page", "1")
	page, err := strconv.Atoi(pageStr)
	if err != nil || page <= 0 {

		response.ResponseError(c, fmt.Sprintf("Неверное тело запроса: %v", err))
		return

	}

	limitStr := c.DefaultQuery("limit", "10")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {

		response.ResponseError(c, fmt.Sprintf("Неверное тело запроса: %v", err))
		return

	}

	filter := model.InvoiceInput{
		ProjectID: projectID,
	}

	data, err := controller.invoiceInputService.GetPaginated(page, limit, filter)
	if err != nil {

		response.ResponseError(c, fmt.Sprintf("Внутренняя ошибка сервера: %v", err))
		return

	}

	dataCount, err := controller.invoiceInputService.Count(projectID)
	if err != nil {

		response.ResponseError(c, fmt.Sprintf("Внутренняя ошибка сервера: %v", err))
		return

	}

	response.ResponsePaginatedData(c, data, dataCount)

}

func (controller *invoiceInputController) GetInvoiceMaterialsWithoutSerialNumbers(c *gin.Context) {

	idRaw := c.Param("id")
	id, err := strconv.ParseUint(idRaw, 10, 64)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Неверное тело запроса: %v", err))
		return
	}

	data, err := controller.invoiceInputService.GetInvoiceMaterialsWithoutSerialNumbers(uint(id))
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Внутренняя ошибка сервера: %v", err))
		return
	}

	response.ResponseSuccess(c, data)
}

func (controller *invoiceInputController) GetInvoiceMaterialsWithSerialNumbers(c *gin.Context) {
	idRaw := c.Param("id")
	id, err := strconv.ParseUint(idRaw, 10, 64)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Неверное тело запроса: %v", err))
		return
	}

	data, err := controller.invoiceInputService.GetInvoiceMaterialsWithSerialNumbers(uint(id))
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Внутренняя ошибка сервера: %v", err))
		return
	}

	response.ResponseSuccess(c, data)
}

func (controller *invoiceInputController) Create(c *gin.Context) {

	workerID := c.GetUint("workerID")
	projectID := c.GetUint("projectID")

	var createData dto.InvoiceInput
	if err := c.ShouldBindJSON(&createData); err != nil {
		response.ResponseError(c, fmt.Sprintf("Неверное тело запроса: %v", err))
		return
	}

	createData.Details.ProjectID = projectID
	createData.Details.ReleasedWorkerID = workerID

	data, err := controller.invoiceInputService.Create(createData)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Внутренняя ошибка сервера: %v", err))
		return
	}

	response.ResponseSuccess(c, data)
}

func (controller *invoiceInputController) Update(c *gin.Context) {
	var updateData dto.InvoiceInput
	if err := c.ShouldBindJSON(&updateData); err != nil {
		response.ResponseError(c, fmt.Sprintf("Invalid data recieved by server: %v", err))
		return
	}

	workerID := c.GetUint("workerID")
	projectID := c.GetUint("projectID")

	updateData.Details.ProjectID = projectID
	updateData.Details.ReleasedWorkerID = workerID

	data, err := controller.invoiceInputService.Update(updateData)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Could not perform the updation of Invoice: %v", err))
		return
	}

	response.ResponseSuccess(c, data)
}

func (controller *invoiceInputController) Delete(c *gin.Context) {
	idRaw := c.Param("id")
	id, err := strconv.ParseUint(idRaw, 10, 64)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Неверное тело запроса: %v", err))
		return
	}

	err = controller.invoiceInputService.Delete(uint(id))
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Внутренняя ошибка сервера: %v", err))
		return
	}

	response.ResponseSuccess(c, "deleted")
}

func (controller *invoiceInputController) Confirmation(c *gin.Context) {

	projectID := c.GetUint("projectID")

	idRaw := c.Param("id")
	id, err := strconv.ParseUint(idRaw, 10, 64)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Неверное тело запроса: %v", err))
		return
	}

	invoiceInput, err := controller.invoiceInputService.GetByID(uint(id))
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
	fileExtension := fileNameAndExtension[1]
	file.Filename = invoiceInput.DeliveryCode + "." + fileExtension
	filePath := filepath.Join("./pkg/excels/input/", file.Filename)

	err = c.SaveUploadedFile(file, filePath)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Внутренняя ошибка сервера: %v", err))
		return
	}

	err = controller.invoiceInputService.Confirmation(uint(id), projectID)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Внутренняя ошибка сервера: %v", err))
		return
	}

	response.ResponseSuccess(c, true)
}

func (controller *invoiceInputController) GetDocument(c *gin.Context) {

	deliveryCode := c.Param("deliveryCode")

	filePath := filepath.Join("./pkg/excels/input/", deliveryCode)
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

func (controller *invoiceInputController) UniqueCode(c *gin.Context) {

	projectID := c.GetUint("projectID")

	data, err := controller.invoiceInputService.UniqueCode(projectID)
	if err != nil {

		response.ResponseError(c, fmt.Sprintf("Внутренняя ошибка сервера: %v", err))
		return

	}

	response.ResponseSuccess(c, data)
}

func (controller *invoiceInputController) UniqueWarehouseManager(c *gin.Context) {

	projectID := c.GetUint("projectID")

	data, err := controller.invoiceInputService.UniqueWarehouseManager(projectID)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Внутренняя ошибка сервера: %v", err))
		return
	}

	response.ResponseSuccess(c, data)
}

func (controller *invoiceInputController) UniqueReleased(c *gin.Context) {

	projectID := c.GetUint("projectID")

	data, err := controller.invoiceInputService.UniqueReleased(projectID)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Внутренняя ошибка сервера: %v", err))
		return
	}

	response.ResponseSuccess(c, data)
}

func (controller *invoiceInputController) Report(c *gin.Context) {

	projectID := c.GetUint("projectID")

	var filter dto.InvoiceInputReportFilterRequest
	if err := c.ShouldBindJSON(&filter); err != nil {
		response.ResponseError(c, fmt.Sprintf("Неверное тело запроса: %v", err))
		return
	}

	filter.ProjectID = projectID
	filename, err := controller.invoiceInputService.Report(filter, projectID)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Внутренняя ошибка сервера: %v", err))
		return
	}

	filePath := filepath.Join("./pkg/excels/temp/", filename)
	c.FileAttachment(filePath, filename)
	os.Remove(filePath)
	// response.ResponseSuccess(c, true)
}

func (controller *invoiceInputController) NewMaterial(c *gin.Context) {
	var data dto.NewMaterialDataFromInvoiceInput
	if err := c.ShouldBindJSON(&data); err != nil {
		response.ResponseError(c, fmt.Sprintf("Неверное тело запроса %v", err))
		return
	}

	projectID := c.GetUint("projectID")
	data.ProjectID = projectID

	err := controller.invoiceInputService.NewMaterialAndItsCost(data)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Внутренняя ошибка сервера: %v", err))
		return
	}

	response.ResponseSuccess(c, true)
}

func (controller *invoiceInputController) NewMaterialCost(c *gin.Context) {
	var data model.MaterialCost
	if err := c.ShouldBindJSON(&data); err != nil {
		response.ResponseError(c, fmt.Sprintf("Неверное тело запроса %v", err))
		return
	}

	err := controller.invoiceInputService.NewMaterialCost(data)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Внутренняя ошибка сервера: %v", err))
		return
	}

	response.ResponseSuccess(c, true)
}

func (controller *invoiceInputController) GetMaterialsForEdit(c *gin.Context) {
	idRaw := c.Param("id")
	id, err := strconv.ParseUint(idRaw, 10, 64)

	result, err := controller.invoiceInputService.GetMaterialsForEdit(uint(id))
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Внутренняя ошибка сервера: %v", err))
		return
	}

	response.ResponseSuccess(c, result)
}
