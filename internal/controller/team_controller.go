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

type teamController struct {
	teamService service.ITeamService
}

func InitTeamController(teamService service.ITeamService) ITeamController {
	return &teamController{
		teamService: teamService,
	}
}

type ITeamController interface {
	GetAll(c *gin.Context)
	GetPaginated(c *gin.Context)
	GetByID(c *gin.Context)
	Create(c *gin.Context)
	Update(c *gin.Context)
	Delete(c *gin.Context)
	GetTemplateFile(c *gin.Context)
	Import(c *gin.Context)
  GetAllForSelect(c *gin.Context)
}

func (controller *teamController) GetAll(c *gin.Context) {
	projectID := c.GetUint("projectID")

	data, err := controller.teamService.GetAll(projectID)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Could not get Team data: %v", err))
		return
	}

	response.ResponseSuccess(c, data)
}

func (controller *teamController) GetPaginated(c *gin.Context) {
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

	data, err := controller.teamService.GetPaginated(page, limit, projectID)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Could not get the paginated data of Team: %v", err))
		return
	}

	dataCount, err := controller.teamService.Count(projectID)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Could not get the total amount of Team: %v", err))
		return
	}

	response.ResponsePaginatedData(c, data, dataCount)
}

func (controller *teamController) GetByID(c *gin.Context) {
	idRaw := c.Param("id")
	id, err := strconv.ParseUint(idRaw, 10, 64)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Incorrect parameter provided: %v", err))
		return
	}

	data, err := controller.teamService.GetByID(uint(id))
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Could not get the data with ID(%d): %v", id, err))
		return
	}

	response.ResponseSuccess(c, data)
}

func (controller *teamController) Create(c *gin.Context) {

	var createData dto.TeamMutation
	if err := c.ShouldBindJSON(&createData); err != nil {
		response.ResponseError(c, fmt.Sprintf("Invalid data recieved by server: %v", err))
		return
	}

	projectID := c.GetUint("projectID")
	createData.ProjectID = projectID

	exist, err := controller.teamService.DoesTeamNumberAlreadyExistForCreate(createData.Number, createData.ProjectID)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Could not perform team number check-up: %v", err))
		return
	}

	if exist {
		response.ResponseError(c, fmt.Sprintf("Бригада с таким номером уже существует"))
		return
	}

	data, err := controller.teamService.Create(createData)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Could perform the creation of Team: %v", err))
		return
	}

	response.ResponseSuccess(c, data)

}

func (controller *teamController) Update(c *gin.Context) {
	var updateData dto.TeamMutation
	if err := c.ShouldBindJSON(&updateData); err != nil {
		response.ResponseError(c, fmt.Sprintf("Invalid data recieved by server: %v", err))
		return
	}

	projectID := c.GetUint("projectID")
	updateData.ProjectID = projectID

	exist, err := controller.teamService.DoesTeamNumberAlreadyExistForUpdate(updateData.Number, updateData.ID, projectID)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Could not perform team number check-up: %v", err))
		return
	}

	if exist {
		response.ResponseError(c, fmt.Sprintf("Бригада с таким номером уже существует"))
		return
	}

	data, err := controller.teamService.Update(updateData)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Could not perform the updation of Team: %v", err))
		return
	}

	response.ResponseSuccess(c, data)
}

func (controller *teamController) Delete(c *gin.Context) {
	idRaw := c.Param("id")
	id, err := strconv.ParseUint(idRaw, 10, 64)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Incorrect parameter provided: %v", err))
		return
	}

	err = controller.teamService.Delete(uint(id))
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Could not perform the deletion of Team: %v", err))
		return
	}

	response.ResponseSuccess(c, "deleted")
}

func (controller *teamController) GetTemplateFile(c *gin.Context) {
	filepath := "./pkg/excels/templates/Шаблон для импорта Бригады.xlsx"
	projectID := c.GetUint("projectID")

	if err := controller.teamService.TemplateFile(projectID, filepath); err != nil {
		response.ResponseError(c, fmt.Sprintf("Внутренняя ошибка сервера: %v", err))
		return
	}

	c.FileAttachment(filepath, "Шаблон для импорта Бригады.xlsx")
}

func (controller *teamController) Import(c *gin.Context) {
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

	err = controller.teamService.Import(projectID, filePath)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Внутренняя ошибка сервера: %v", err))
		return
	}

	response.ResponseSuccess(c, true)
}

func (controller *teamController) GetAllForSelect(c *gin.Context) {
	projectID := c.GetUint("projectID")

	data, err := controller.teamService.GetAllForSelect(projectID)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Could not get Team data: %v", err))
		return
	}

	response.ResponseSuccess(c, data)
}
