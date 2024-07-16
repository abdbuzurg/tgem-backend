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

type kl04kvObjectController struct {
	kl04kvObjectService service.IKL04KVObjectService
}

func InitKl04KVObjectController(kl04kvObjectService service.IKL04KVObjectService) IKL04KVObjectController {
	return &kl04kvObjectController{
		kl04kvObjectService: kl04kvObjectService,
	}
}

type IKL04KVObjectController interface {
	GetPaginated(c *gin.Context)
	Create(c *gin.Context)
	Delete(c *gin.Context)
	Update(c *gin.Context)
	GetTemplateFile(c *gin.Context)
	Import(c *gin.Context)
	Export(c *gin.Context)
	GetObjectNamesForSearch(c *gin.Context)
}

func (controller *kl04kvObjectController) GetPaginated(c *gin.Context) {
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

	tpObjectIDStr := c.DefaultQuery("tpObjectID", "0")
	tpObjectID, err := strconv.Atoi(tpObjectIDStr)
	if err != nil || tpObjectID < 0 {
		response.ResponseError(c, fmt.Sprintf("Неверное тело запроса tpObjectID: %v", err))
		return
	}

	filter := dto.KL04KVObjectSearchParameters{
		ProjectID:          c.GetUint("projectID"),
		TeamID:             uint(teamID),
		SupervisorWorkerID: uint(supervisorWorkerID),
		TPObjectID:         uint(tpObjectID),
		ObjectName:         c.DefaultQuery("objectName", ""),
	}

	data, err := controller.kl04kvObjectService.GetPaginated(page, limit, filter)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Внутренняя ошибка сервера: %v", err))
		return
	}

	dataCount, err := controller.kl04kvObjectService.Count(filter)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Внутренняя ошибка сервера: %v", err))
		return
	}

	response.ResponsePaginatedData(c, data, dataCount)
}

func (controller *kl04kvObjectController) Create(c *gin.Context) {
	var data dto.KL04KVObjectCreate
	if err := c.ShouldBindJSON(&data); err != nil {
		response.ResponseError(c, fmt.Sprintf("Неправильно тело запроса: %v", err))
		return
	}

	projectID := c.GetUint("projectID")
	data.BaseInfo.ProjectID = projectID

	_, err := controller.kl04kvObjectService.Create(data)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Внутренняя ошибка сервера: %v", err))
		return
	}

	response.ResponseSuccess(c, true)
}

func (controller *kl04kvObjectController) Delete(c *gin.Context) {
	idRaw := c.Param("id")
	id, err := strconv.ParseUint(idRaw, 10, 64)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Неправильный параметр в запросе: %v", err))
		return
	}

	projectID := c.GetUint("projectID")

	err = controller.kl04kvObjectService.Delete(projectID, uint(id))
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Внутренняя ошибка сервера: %v", err))
		return
	}

	response.ResponseSuccess(c, true)
}

func (controller *kl04kvObjectController) Update(c *gin.Context) {
	var data dto.KL04KVObjectCreate
	if err := c.ShouldBindJSON(&data); err != nil {
		response.ResponseError(c, fmt.Sprintf("Неправильно тело запроса: %v", err))
		return
	}

	projectID := c.GetUint("projectID")
	data.BaseInfo.ProjectID = projectID

	_, err := controller.kl04kvObjectService.Update(data)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Внутренняя ошибка сервера: %v", err))
		return
	}

	response.ResponseSuccess(c, true)
}

func (controller *kl04kvObjectController) GetTemplateFile(c *gin.Context) {
	templateFilePath := filepath.Join("./pkg/excels/templates/Шаблон для импорта КЛ 04 КВ.xlsx")

	if err := controller.kl04kvObjectService.TemplateFile(templateFilePath, c.GetUint("projectID")); err != nil {
		response.ResponseError(c, fmt.Sprintf("Внутренняя ошибка сервера: %v", err))
		return
	}

	c.FileAttachment(templateFilePath, "Шаблон для импорта КЛ 04 КВ.xlsx")
}

func (controller *kl04kvObjectController) Import(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Файл не может быть сформирован, проверьте файл: %v", err))
		return
	}

	date := time.Now()
	importFilePath := filepath.Join("./pkg/excels/temp/", date.Format("2006-01-02 15-04-05")+file.Filename)
	err = c.SaveUploadedFile(file, importFilePath)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Файл не может быть сохранен на сервере: %v", err))
		return
	}

	projectID := c.GetUint("projectID")
	err = controller.kl04kvObjectService.Import(projectID, importFilePath)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Внутренняя ошибка сервера: %v", err))
		return
	}

	response.ResponseSuccess(c, true)
}

func (controller *kl04kvObjectController) Export(c *gin.Context) {
	projectID := c.GetUint("projectID")

	exportFileName, err := controller.kl04kvObjectService.Export(projectID)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Внутренняя ошибка сервера: %v", err))
		return
	}

	exportFilePath := filepath.Join("./pkg/excels/temp/", exportFileName)
	c.FileAttachment(exportFilePath, exportFileName)
	os.Remove(exportFileName)
}

func (controller *kl04kvObjectController) GetObjectNamesForSearch(c *gin.Context) {
	data, err := controller.kl04kvObjectService.GetObjectNamesForSearch(c.GetUint("projectID"))
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Внутренняя ошибка сервера: %v", err))
		return
	}

  response.ResponseSuccess(c, data)
}
