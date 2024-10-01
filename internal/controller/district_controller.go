package controller

import (
	"backend-v2/internal/service"
	"backend-v2/model"
	"backend-v2/pkg/response"
	"fmt"
	"strconv"

	"github.com/gin-gonic/gin"
)

type districtController struct {
	districtService   service.IDistrictService
	userActionService service.IUserActionService
}

func InitDistrictController(
	districtService service.IDistrictService,
	userActionService service.IUserActionService,
) IDistictController {
	return &districtController{
		districtService:   districtService,
		userActionService: userActionService,
	}
}

type IDistictController interface {
	GetAll(c *gin.Context)
	GetPaginated(c *gin.Context)
	GetByID(c *gin.Context)
	Create(c *gin.Context)
	Update(c *gin.Context)
	Delete(c *gin.Context)
}

func (controller *districtController) GetAll(c *gin.Context) {

	projectID := c.GetUint("projectID")

	data, err := controller.districtService.GetAll(projectID)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Внутренняя ошибка сервера: %v", err))
		return

	}

	response.ResponseSuccess(c, data)
}

func (controller *districtController) GetPaginated(c *gin.Context) {

	projectID := c.GetUint("projectID")

	pageStr := c.DefaultQuery("page", "1")
	page, err := strconv.Atoi(pageStr)
	if err != nil || page <= 0 {

		response.ResponseError(c, fmt.Sprintf("Сервер получил неправильные данные: %v", err))
		return

	}

	limitStr := c.DefaultQuery("limit", "10")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {

		response.ResponseError(c, fmt.Sprintf("Сервер получил неправильные данные: %v", err))
		return

	}

	data, err := controller.districtService.GetPaginated(page, limit, projectID)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Внутренняя ошибка сервера: %v", err))
		return
	}

	dataCount, err := controller.districtService.Count(projectID)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Внутренняя ошибка сервера: %v", err))
		return
	}

	response.ResponsePaginatedData(c, data, dataCount)
}

func (controller *districtController) GetByID(c *gin.Context) {
	idRaw := c.Param("id")
	id, err := strconv.ParseUint(idRaw, 10, 64)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Сервер получил неправильные данные: %v", err))
		return
	}

	data, err := controller.districtService.GetByID(uint(id))
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Внутренняя ошибка сервера: %v", err))
		return
	}

	response.ResponseSuccess(c, data)
}

func (controller *districtController) Create(c *gin.Context) {

	projectID := c.GetUint("projectID")

	var createData model.District
	if err := c.ShouldBindJSON(&createData); err != nil {
		response.ResponseError(c, fmt.Sprintf("Неверное тело запроса: %v", err))
		return
	}

  createData.ProjectID = projectID

	data, err := controller.districtService.Create(createData)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Внутрення ошибка сервера: %v", err))
		return
	}

	response.ResponseSuccess(c, data)
}

func (controller *districtController) Update(c *gin.Context) {

	projectID := c.GetUint("projectID")

	var updateData model.District
	if err := c.ShouldBindJSON(&updateData); err != nil {

		response.ResponseError(c, fmt.Sprintf("Неверное тело запроса: %v", err))
		return
	}

  updateData.ProjectID = projectID

	data, err := controller.districtService.Update(updateData)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Внутренняя ошибка сервера: %v", err))
		return
	}

	response.ResponseSuccess(c, data)
}

func (controller *districtController) Delete(c *gin.Context) {

	idRaw := c.Param("id")
	id, err := strconv.ParseUint(idRaw, 10, 64)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Неправильный параметер запроса: %v", err))
		return
	}

	err = controller.districtService.Delete(uint(id))
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Внутренняя ошибка сервера: %v", err))
		return
	}

	response.ResponseSuccess(c, "deleted")
}
