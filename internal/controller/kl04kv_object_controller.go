package controller

import (
	"backend-v2/internal/dto"
	"backend-v2/internal/service"
	"backend-v2/pkg/response"
	"fmt"
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

	projectID := c.GetUint("projectID")

	data, err := controller.kl04kvObjectService.GetPaginated(page, limit, projectID)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Внутренняя ошибка сервера: %v", err))
		return
	}

	dataCount, err := controller.kl04kvObjectService.Count(projectID)
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
	filepath := "./pkg/excels/templates/Шаблон для импорта КЛ 04 КВ.xlsx"

	if err := controller.kl04kvObjectService.TemplateFile(filepath); err != nil {
		response.ResponseError(c, fmt.Sprintf("Внутренняя ошибка сервера: %v", err))
		return
	}

	c.FileAttachment(filepath, "Шаблон для импорта КЛ 04 КВ.xlsx")
}

func (controller *kl04kvObjectController) Import(c *gin.Context) {
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
  err = controller.kl04kvObjectService.Import(projectID, filePath)
  if err != nil {
    response.ResponseError(c, fmt.Sprintf("Внутренняя ошибка сервера: %v", err))
    return
  }

  response.ResponseSuccess(c, true)
}
