package controller

import (
	"backend-v2/internal/dto"
	"backend-v2/internal/service"
	"backend-v2/pkg/response"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

type invoiceOutputOutOfProjectController struct {
	invoiceOutputOutOfProjectService service.IInvoiceOutputOutOfProjectService
}

func InitInvoiceOutputOutOfProjectController(
	invoiceOutputOutOfProjectService service.IInvoiceOutputOutOfProjectService,
) IInvoiceOutputOutOfProjectController {
	return &invoiceOutputOutOfProjectController{
		invoiceOutputOutOfProjectService: invoiceOutputOutOfProjectService,
	}
}

type IInvoiceOutputOutOfProjectController interface {
	GetPaginated(c *gin.Context)
	Create(c *gin.Context)
	GetInvoiceMaterialsWithSerialNumbers(c *gin.Context)
	GetInvoiceMaterialsWithoutSerialNumbers(c *gin.Context)
	Confirmation(c *gin.Context)
	Update(c *gin.Context)
	GetMaterialsForEdit(c *gin.Context)
	UniqueNameOfProjects(c *gin.Context)
	Report(c *gin.Context)
}

func (controller *invoiceOutputOutOfProjectController) GetPaginated(c *gin.Context) {
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

	nameOfProject := c.DefaultQuery("nameOfProject", "")

	releasedWorkerIDStr := c.DefaultQuery("releasedWorkerID", "0")
	releasedWorkerID, err := strconv.Atoi(releasedWorkerIDStr)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Wrong query parameter provided for releasedWorkerID: %v", err))
		return
	}

	filter := dto.InvoiceOutputOutOfProjectSearchParameters{
		ProjectID:        c.GetUint("projectID"),
		NameOfProject:    nameOfProject,
		ReleasedWorkerID: uint(releasedWorkerID),
	}

	data, err := controller.invoiceOutputOutOfProjectService.GetPaginated(page, limit, filter)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Could not get the paginated data of Invoice: %v", err))
		return
	}

	dataCount, err := controller.invoiceOutputOutOfProjectService.Count(filter)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Could not get the total amount of Invoice: %v", err))
		return
	}

	response.ResponsePaginatedData(c, data, dataCount)
}

func (controller *invoiceOutputOutOfProjectController) Create(c *gin.Context) {
	createData := dto.InvoiceOutputOutOfProject{}
	if err := c.ShouldBindJSON(&createData); err != nil {
		response.ResponseError(c, fmt.Sprintf("Invalid data recieved by server: %v", err))
		return
	}

	workerID := c.GetUint("workerID")
	createData.Details.ReleasedWorkerID = workerID

	projectID := c.GetUint("projectID")
	createData.Details.ProjectID = projectID

	data, err := controller.invoiceOutputOutOfProjectService.Create(createData)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Could perform the creation of Invoice: %v", err))
		return
	}

	response.ResponseSuccess(c, data)
}

func (controller *invoiceOutputOutOfProjectController) Delete(c *gin.Context) {
	idRaw := c.Param("id")
	id, err := strconv.ParseUint(idRaw, 10, 64)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Incorrect parameter provided: %v", err))
		return
	}

	err = controller.invoiceOutputOutOfProjectService.Delete(uint(id))
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Could not perform the deletion of Invoice: %v", err))
		return
	}

	response.ResponseSuccess(c, "deleted")
}

func (controller *invoiceOutputOutOfProjectController) GetInvoiceMaterialsWithoutSerialNumbers(c *gin.Context) {

	idRaw := c.Param("id")
	id, err := strconv.ParseUint(idRaw, 10, 64)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Неверное тело запроса: %v", err))
		return
	}

	data, err := controller.invoiceOutputOutOfProjectService.GetInvoiceMaterialsWithoutSerialNumbers(uint(id))
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Внутренняя ошибка сервера: %v", err))
		return
	}

	response.ResponseSuccess(c, data)
}

func (controller *invoiceOutputOutOfProjectController) GetInvoiceMaterialsWithSerialNumbers(c *gin.Context) {

	idRaw := c.Param("id")
	id, err := strconv.ParseUint(idRaw, 10, 64)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Неверное тело запроса: %v", err))
		return
	}

	data, err := controller.invoiceOutputOutOfProjectService.GetInvoiceMaterialsWithSerialNumbers(uint(id))
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Внутренняя ошибка сервера: %v", err))
		return
	}

	response.ResponseSuccess(c, data)
}

func (controller *invoiceOutputOutOfProjectController) Update(c *gin.Context) {
	updateData := dto.InvoiceOutputOutOfProject{}
	if err := c.ShouldBindJSON(&updateData); err != nil {
		response.ResponseError(c, fmt.Sprintf("Invalid data recieved by server: %v", err))
		return
	}

	workerID := c.GetUint("workerID")
	updateData.Details.ReleasedWorkerID = workerID

	projectID := c.GetUint("projectID")
	updateData.Details.ProjectID = projectID

	data, err := controller.invoiceOutputOutOfProjectService.Update(updateData)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Could perform the creation of Invoice: %v", err))
		return
	}

	response.ResponseSuccess(c, data)
}

func (controller *invoiceOutputOutOfProjectController) Confirmation(c *gin.Context) {
	idRaw := c.Param("id")
	id, err := strconv.ParseUint(idRaw, 10, 64)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Incorrect parameter provided: %v", err))
		return
	}

	invoiceOutputOutOfProject, err := controller.invoiceOutputOutOfProjectService.GetByID(uint(id))
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Cannot find invoice Output by id %v: %v", id, err))
	}

	file, err := c.FormFile("file")
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("cannot form file: %v", err))
		return
	}

	fileNameAndExtension := strings.Split(file.Filename, ".")
	fileExtension := fileNameAndExtension[1]
	file.Filename = invoiceOutputOutOfProject.DeliveryCode + "." + fileExtension
	filePath := filepath.Join("./pkg/excels/output/", file.Filename)

	err = c.SaveUploadedFile(file, filePath)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("cannot save file: %v", err))
		return
	}

	excelFilePath := filepath.Join("./pkg/excels/output/", invoiceOutputOutOfProject.DeliveryCode+".xlsx")
	os.Remove(excelFilePath)

	err = controller.invoiceOutputOutOfProjectService.Confirmation(uint(id))
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("cannot confirm invoice input with id %v: %v", id, err))
		return
	}

	response.ResponseSuccess(c, true)
}

func (controller *invoiceOutputOutOfProjectController) GetMaterialsForEdit(c *gin.Context) {
	idRaw := c.Param("id")
	id, err := strconv.ParseUint(idRaw, 10, 64)

	result, err := controller.invoiceOutputOutOfProjectService.GetMaterialsForEdit(uint(id))
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Внутренняя ошибка сервера: %v", err))
		return
	}

	response.ResponseSuccess(c, result)
}

func (controller *invoiceOutputOutOfProjectController) UniqueNameOfProjects(c *gin.Context) {
	result, err := controller.invoiceOutputOutOfProjectService.GetUniqueNameOfProjects(c.GetUint("projectID"))
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Внутренняя ошибка сервера: %v", err))
		return
	}

	response.ResponseSuccess(c, result)
}

func (controller *invoiceOutputOutOfProjectController) Report(c *gin.Context) {
	var filter dto.InvoiceOutputOutOfProjectReportFilter
	if err := c.ShouldBindJSON(&filter); err != nil {
		response.ResponseError(c, fmt.Sprintf("Invalid data recieved by server: %v", err))
		return
	}

	filter.ProjectID = c.GetUint("projectID")
	filename, err := controller.invoiceOutputOutOfProjectService.Report(filter)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Внутренняя ошибка сервера: %v", err))
		return
	}

	filePath := filepath.Join("./pkg/excels/temp/", filename)
	c.FileAttachment(filePath, filename)
	os.Remove(filePath)

}
