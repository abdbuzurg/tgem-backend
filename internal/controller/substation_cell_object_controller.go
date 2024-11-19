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

type substationCellObjectController struct {
	substationCellObjectService service.ISubstationCellObjectService
}

func InitSubstationCellObjectController(
	substationCellObjectService service.ISubstationCellObjectService,
) ISubstationCellObjectController {
	return &substationCellObjectController{
		substationCellObjectService: substationCellObjectService,
	}
}

type ISubstationCellObjectController interface {
	GetPaginated(c *gin.Context)
	Create(c *gin.Context)
	Update(c *gin.Context)
	Delete(c *gin.Context)
	GetTemplateFile(c *gin.Context)
	Import(c *gin.Context)
	Export(c *gin.Context)
	GetObjectNamesForSearch(c *gin.Context)
}

func (controller *substationCellObjectController) GetPaginated(c *gin.Context) {
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

	filter := dto.SubstationCellObjectSearchParameters{
		ProjectID:          c.GetUint("projectID"),
		TeamID:             uint(teamID),
		SupervisorWorkerID: uint(supervisorWorkerID),
		ObjectName:         c.DefaultQuery("objectName", ""),
	}

	data, err := controller.substationCellObjectService.GetPaginated(page, limit, filter)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Внутренняя ошибка сервера: %v", err))
		return
	}

	dataCount, err := controller.substationCellObjectService.Count(filter)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Внутренняя ошибка сервера: %v", err))
		return
	}

	response.ResponsePaginatedData(c, data, dataCount)
}

func (controller *substationCellObjectController) Create(c *gin.Context) {
	var createData dto.SubstationCellObjectCreate
	if err := c.ShouldBindJSON(&createData); err != nil {
		response.ResponseError(c, fmt.Sprintf("Неверное тело запроса: %v", err))
		return
	}

	projectID := c.GetUint("projectID")
	createData.BaseInfo.ProjectID = projectID

	data, err := controller.substationCellObjectService.Create(createData)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Внутренняя ошибка сервера: %v", err))
		return
	}

	response.ResponseSuccess(c, data)

}

func (controller *substationCellObjectController) Update(c *gin.Context) {
	var updateData dto.SubstationCellObjectCreate
	if err := c.ShouldBindJSON(&updateData); err != nil {
		response.ResponseError(c, fmt.Sprintf("Неверное тело запроса: %v", err))
		return
	}

	projectID := c.GetUint("projectID")
	updateData.BaseInfo.ProjectID = projectID

	data, err := controller.substationCellObjectService.Update(updateData)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Внутренняя ошибка сервера: %v", err))
		return
	}

	response.ResponseSuccess(c, data)
}

func (controller *substationCellObjectController) Delete(c *gin.Context) {
	idRaw := c.Param("id")
	id, err := strconv.ParseUint(idRaw, 10, 64)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Неправильный параметр в запросе: %v", err))
		return
	}

	projectID := c.GetUint("projectID")

	err = controller.substationCellObjectService.Delete(uint(id), projectID)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Внутренняя ошибка сервера: %v", err))
		return
	}

	response.ResponseSuccess(c, true)
}

func (controller *substationCellObjectController) GetTemplateFile(c *gin.Context) {
	filepath := "./pkg/excels/templates/Шаблон для импорт Ячеек Подстанции.xlsx"

	tmpFilePath, err := controller.substationCellObjectService.TemplateFile(filepath, c.GetUint("projectID"))
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Внутренняя ошибка сервера: %v", err))
		return
	}

	c.FileAttachment(tmpFilePath, "Шаблон для импорт Ячеек Подстанции.xlsx")
	if err := os.Remove(tmpFilePath); err != nil {
		response.ResponseError(c, fmt.Sprintf("Внутренняя ошибка сервера: %v", err))
		return
	}
}

func (controller *substationCellObjectController) Import(c *gin.Context) {
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
	err = controller.substationCellObjectService.Import(projectID, filePath)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Внутренняя ошибка сервера: %v", err))
		return
	}

	response.ResponseSuccess(c, true)
}

func (controller *substationCellObjectController) Export(c *gin.Context) {
	projectID := c.GetUint("projectID")

	exportFileName, err := controller.substationCellObjectService.Export(projectID)
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

func (controller *substationCellObjectController) GetObjectNamesForSearch(c *gin.Context) {
	data, err := controller.substationCellObjectService.GetObjectNamesForSearch(c.GetUint("projectID"))
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Внутренняя ошибка сервера: %v", err))
		return
	}

	response.ResponseSuccess(c, data)
}
