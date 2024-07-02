package controller

import (
	"backend-v2/internal/service"
	"backend-v2/model"
	"backend-v2/pkg/response"
	"fmt"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type workerController struct {
	workerService service.IWorkerService
}

func InitWorkerController(workerService service.IWorkerService) IWorkerController {
	return &workerController{
		workerService: workerService,
	}
}

type IWorkerController interface {
	GetAll(c *gin.Context)
	GetPaginated(c *gin.Context)
	GetByID(c *gin.Context)
	GetByJobTitleInProject(c *gin.Context)
	Create(c *gin.Context)
	Update(c *gin.Context)
	Delete(c *gin.Context)
	GetTemplateFile(c *gin.Context)
	Import(c *gin.Context)
}

func (controller *workerController) GetAll(c *gin.Context) {
	data, err := controller.workerService.GetAll()
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Could not get Worker data: %v", err))
		return
	}

	response.ResponseSuccess(c, data)
}

func (controller *workerController) GetPaginated(c *gin.Context) {
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

	filter := model.Worker{
    ProjectID: c.GetUint("projectID"),
	}

	data, err := controller.workerService.GetPaginated(page, limit, filter)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Could not get the paginated data of Worker: %v", err))
		return
	}

	dataCount, err := controller.workerService.Count()
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Could not get the total amount of Worker: %v", err))
		return
	}

	response.ResponsePaginatedData(c, data, dataCount)
}

func (controller *workerController) GetByID(c *gin.Context) {
	idRaw := c.Param("id")
	id, err := strconv.ParseUint(idRaw, 10, 64)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Incorrect parameter provided: %v", err))
		return
	}

	data, err := controller.workerService.GetByID(uint(id))
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Could not get the data with ID(%d): %v", id, err))
		return
	}

	response.ResponseSuccess(c, data)
}

func (controller *workerController) GetByJobTitleInProject(c *gin.Context) {
	jobTitleInProject := c.Param("jobTitleInProject")

	data, err := controller.workerService.GetByJobTitleInProject(jobTitleInProject)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Cannot get workers by the job title: %v", err))
		return
	}

	response.ResponseSuccess(c, data)
}

func (controller *workerController) Create(c *gin.Context) {
	var createData model.Worker
	if err := c.ShouldBindJSON(&createData); err != nil {
		response.ResponseError(c, fmt.Sprintf("Invalid data recieved by server: %v", err))
		return
	}

  createData.ProjectID = c.GetUint("projectID")

	data, err := controller.workerService.Create(createData)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Could perform the creation of Worker: %v", err))
		return
	}

	response.ResponseSuccess(c, data)
}

func (controller *workerController) Update(c *gin.Context) {
	var updateData model.Worker
	if err := c.ShouldBindJSON(&updateData); err != nil {
		response.ResponseError(c, fmt.Sprintf("Invalid data recieved by server: %v", err))
		return
	}

  updateData.ProjectID = c.GetUint("projectID")

	data, err := controller.workerService.Update(updateData)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Could not perform the updation of Worker: %v", err))
		return
	}

	response.ResponseSuccess(c, data)
}

func (controller *workerController) Delete(c *gin.Context) {
	idRaw := c.Param("id")
	id, err := strconv.ParseUint(idRaw, 10, 64)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Incorrect parameter provided: %v", err))
		return
	}

	err = controller.workerService.Delete(uint(id))
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Could not perform the deletion of Worker: %v", err))
		return
	}

	response.ResponseSuccess(c, "deleted")
}

func (controller *workerController) GetTemplateFile(c *gin.Context) {
	filepath := "./pkg/excels/templates/Шаблон для импорта Рабочего Персонала.xlsx"
	c.FileAttachment(filepath, "Шаблон для импорта Рабочего Персонала.xlsx")
}

func (controller *workerController) Import(c *gin.Context) {
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

	err = controller.workerService.Import(filePath, projectID)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Внутренняя ошибка сервера: %v", err))
		return
	}

	response.ResponseSuccess(c, true)
}
