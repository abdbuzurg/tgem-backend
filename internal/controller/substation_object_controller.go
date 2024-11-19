package controller

import (
	"backend-v2/internal/dto"
	"backend-v2/internal/service"
	"backend-v2/pkg/response"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type substationObjectController struct {
	substationObjectService service.ISubstationObjectService
}

func InitSubstationObjectController(
	substationObjectService service.ISubstationObjectService,
) ISubstationObjectController {
	return &substationObjectController{
		substationObjectService: substationObjectService,
	}
}

type ISubstationObjectController interface {
	GetPaginated(c *gin.Context)
	Create(c *gin.Context)
	Update(c *gin.Context)
	Delete(c *gin.Context)
	GetTemplateFile(c *gin.Context)
	Import(c *gin.Context)
	GetObjectNamesForSearch(c *gin.Context)
	Export(c *gin.Context)
	GetAll(c *gin.Context)
}

func (controller *substationObjectController) GetAll(c *gin.Context) {
	data, err := controller.substationObjectService.GetAll(c.GetUint("projectID"))
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Внутренняя ошибка сервера: %v", err))
		return
	}

  response.ResponseSuccess(c, data)
}

func (controller *substationObjectController) GetPaginated(c *gin.Context) {

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

	teamIDStr := c.DefaultQuery("teamID", "0")
	teamID, err := strconv.Atoi(teamIDStr)
	if err != nil || teamID < 0 {
		response.ResponseError(c, fmt.Sprintf("Неверное тело запроса teamID: %v", err))
		return
	}

	supervisorWorkerIDStr := c.DefaultQuery("supervisorWorkerID", "0")
	supervisorWorkerID, err := strconv.Atoi(supervisorWorkerIDStr)
	if err != nil || supervisorWorkerID < 0 {
		response.ResponseError(c, fmt.Sprintf("Неверное тело запроса supervisorWorkerID: %v", err))
		return
	}

	filter := dto.SubstationObjectSearchParameters{
		ProjectID:          c.GetUint("projectID"),
		TeamID:             uint(teamID),
		SupervisorWorkerID: uint(supervisorWorkerID),
		ObjectName:         c.DefaultQuery("objectName", ""),
	}

	data, err := controller.substationObjectService.GetPaginated(page, limit, filter)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Внутренняя ошибка сервера: %v", err))
		return
	}

	dataCount, err := controller.substationObjectService.Count(filter)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Внутренняя ошибка сервера: %v", err))
		return
	}

	response.ResponsePaginatedData(c, data, dataCount)
}

func (controller *substationObjectController) Create(c *gin.Context) {
	var createData dto.SubstationObjectCreate
	if err := c.ShouldBindJSON(&createData); err != nil {
		response.ResponseError(c, fmt.Sprintf("Неверное тело запроса: %v", err))
		return
	}

	projectID := c.GetUint("projectID")
	createData.BaseInfo.ProjectID = projectID

	data, err := controller.substationObjectService.Create(createData)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Внутренняя ошибка сервера: %v", err))
		return
	}

	response.ResponseSuccess(c, data)
}

func (controller *substationObjectController) Update(c *gin.Context) {
	var updateData dto.SubstationObjectCreate
	if err := c.ShouldBindJSON(&updateData); err != nil {
		response.ResponseError(c, fmt.Sprintf("Неверное тело запроса: %v", err))
		return
	}

	projectID := c.GetUint("projectID")
	updateData.BaseInfo.ProjectID = projectID

	data, err := controller.substationObjectService.Update(updateData)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Внутренняя ошибка сервера: %v", err))
		return
	}

	response.ResponseSuccess(c, data)
}

func (controller *substationObjectController) Delete(c *gin.Context) {
	idRaw := c.Param("id")
	id, err := strconv.ParseUint(idRaw, 10, 64)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Неправильный параметр в запросе: %v", err))
		return
	}

	projectID := c.GetUint("projectID")

	err = controller.substationObjectService.Delete(uint(id), projectID)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Внутренняя ошибка сервера: %v", err))
		return
	}

	response.ResponseSuccess(c, true)
}

func (controller *substationObjectController) GetTemplateFile(c *gin.Context) {
	templateFilePath := filepath.Join("./pkg/excels/templates/", "Шаблон для импорта Подстанции.xlsx")

	tmpFilePath, err := controller.substationObjectService.TemplateFile(templateFilePath, c.GetUint("projectID"))
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Внутренняя ошибка сервера: %v", err))
		return
	}

	c.FileAttachment(tmpFilePath, "Шаблон для импорта Подстанции.xlsx")
	if err := os.Remove(tmpFilePath); err != nil {
		response.ResponseError(c, fmt.Sprintf("Внутренняя ошибка сервера: %v", err))
		return
	}
}

func (controller *substationObjectController) Import(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Файл не может быть сформирован, проверьте файл: %v", err))
		return
	}

	date := time.Now()
	filePath := "./pkg/excels/temp/" + date.Format("2006-01-02 15-04-05") + file.Filename
	err = c.SaveUploadedFile(file, filePath)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Файл не может быть сохранен на сервере: %v", err))
		return
	}

	projectID := c.GetUint("projectID")
	err = controller.substationObjectService.Import(projectID, filePath)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Внутренняя ошибка сервера: %v", err))
		return
	}

	response.ResponseSuccess(c, true)
}

func (controller *substationObjectController) GetObjectNamesForSearch(c *gin.Context) {
	data, err := controller.substationObjectService.GetObjectNamesForSearch(c.GetUint("projectID"))
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Внутренняя ошибка сервера: %v", err))
		return
	}

	response.ResponseSuccess(c, data)
}

func (controller *substationObjectController) Export(c *gin.Context) {
	projectID := c.GetUint("projectID")

	exportFileName, err := controller.substationObjectService.Export(projectID)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Внутренняя ошибка сервера: %v", err))
		return
	}

	exportFilePath := filepath.Join("./pkg/excels/temp/", exportFileName)
	c.FileAttachment(exportFilePath, exportFileName)
	if err := os.Remove(exportFilePath); err != nil {
		response.ResponseError(c, fmt.Sprintf("Внутренняя ошибка сервера: %v", err))
		return
	}
}
