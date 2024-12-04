package controller

import (
	"backend-v2/internal/dto"
	"backend-v2/internal/service"
	"backend-v2/pkg/response"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type operationController struct {
	operationService service.IOperationService
}

func InitOperationController(operationService service.IOperationService) IOperationController {
	return &operationController{
		operationService: operationService,
	}
}

type IOperationController interface {
	GetAll(c *gin.Context)
	GetPaginated(c *gin.Context)
	GetByID(c *gin.Context)
	Create(c *gin.Context)
	Update(c *gin.Context)
	Delete(c *gin.Context)
	Import(c *gin.Context)
	GetTemplateFile(c *gin.Context)
}

func (controller *operationController) GetAll(c *gin.Context) {
	data, err := controller.operationService.GetAll(c.GetUint("projectID"))
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Could not get Operation data: %v", err))
		return
	}

	response.ResponseSuccess(c, data)
}

func (controller *operationController) GetPaginated(c *gin.Context) {
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

	name := c.DefaultQuery("name", "")
	name, err = url.QueryUnescape(name)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Wrong query parameter provided for name: %v", err))
		return
	}

	code := c.DefaultQuery("code", "")
	code, err = url.QueryUnescape(code)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Wrong query parameter provided for code: %v", err))
		return
	}

	materialIDStr := c.DefaultQuery("materialID", "0")
	materialID, err := strconv.Atoi(materialIDStr)
	if err != nil || materialID < 0 {
		response.ResponseError(c, fmt.Sprintf("Wrong query parameter provided for materialID: %v", err))
		return
	}

	filter := dto.OperationSearchParameters{
		Name:       name,
		Code:       code,
		ProjectID:  c.GetUint("projectID"),
		MaterialID: uint(materialID),
	}

	data, err := controller.operationService.GetPaginated(page, limit, filter)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Could not get the paginated data of Operation: %v", err))
		return
	}

	dataCount, err := controller.operationService.Count(filter)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Could not get the total amount of Operation: %v", err))
		return
	}

	response.ResponsePaginatedData(c, data, dataCount)
}

func (controller *operationController) GetByID(c *gin.Context) {
	idRaw := c.Param("id")
	id, err := strconv.ParseUint(idRaw, 10, 64)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Incorrect parameter provided: %v", err))
		return
	}

	data, err := controller.operationService.GetByID(uint(id))
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Could not get the data with ID(%d): %v", id, err))
		return
	}

	response.ResponseSuccess(c, data)
}

func (controller *operationController) Create(c *gin.Context) {
	var createData dto.Operation
	if err := c.ShouldBindJSON(&createData); err != nil {
		response.ResponseError(c, fmt.Sprintf("Invalid data recieved by server: %v", err))
		return
	}

	createData.ProjectID = c.GetUint("projectID")

	operation, err := controller.operationService.GetByName(createData.Name, createData.ProjectID)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Ошибка проверки имени услуги: %v", err))
		return
	}

	if operation.Name == createData.Name {
		response.ResponseError(c, fmt.Sprint("Услуга с таким именем уже существует"))
		return
	}

	data, err := controller.operationService.Create(createData)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Could perform the creation of Operation: %v", err))
		return
	}

	response.ResponseSuccess(c, data)
}

func (controller *operationController) Update(c *gin.Context) {
	var updateData dto.Operation
	if err := c.ShouldBindJSON(&updateData); err != nil {
		response.ResponseError(c, fmt.Sprintf("Invalid data recieved by server: %v", err))
		return
	}

	updateData.ProjectID = c.GetUint("projectID")

	operation, err := controller.operationService.GetByName(updateData.Name, updateData.ProjectID)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Ошибка проверки имени услуги: %v", err))
		return
	}

	if operation.Name == updateData.Name && operation.ID != updateData.ID {
		response.ResponseError(c, fmt.Sprint("Услуга с таким именем уже существует"))
		return
	}

	data, err := controller.operationService.Update(updateData)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Could not perform the updation of Operation: %v", err))
		return
	}

	response.ResponseSuccess(c, data)
}

func (controller *operationController) Delete(c *gin.Context) {
	idRaw := c.Param("id")
	id, err := strconv.ParseUint(idRaw, 10, 64)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Incorrect parameter provided: %v", err))
		return
	}

	err = controller.operationService.Delete(uint(id))
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Could not perform the deletion of Operation: %v", err))
		return
	}

	response.ResponseSuccess(c, "deleted")
}

func (controller *operationController) Import(c *gin.Context) {
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
	err = controller.operationService.Import(projectID, importFilePath)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Внутренняя ошибка сервера: %v", err))
		return
	}

	response.ResponseSuccess(c, true)
}

func (controller *operationController) GetTemplateFile(c *gin.Context) {
	templateFilePath := filepath.Join("./pkg/excels/templates/Шаблон для импорта Услуг.xlsx")

	tmpFilePath, err := controller.operationService.TemplateFile(templateFilePath, c.GetUint("projectID"))
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Внутренняя ошибка сервера: %v", err))
		return
	}

	c.FileAttachment(tmpFilePath, "Шаблон для импорта КЛ 04 КВ.xlsx")
	os.Remove(tmpFilePath)
}
