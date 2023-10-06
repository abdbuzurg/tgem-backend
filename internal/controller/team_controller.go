package controller

import (
	"backend-v2/internal/service"
	"backend-v2/model"
	"backend-v2/pkg/response"
	"fmt"
	"net/url"
	"strconv"

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
}

func (controller *teamController) GetAll(c *gin.Context) {
	data, err := controller.teamService.GetAll()
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

	leaderWorkerIDStr := c.DefaultQuery("leaderWorkerID", "")
	leaderWorkerID := 0
	if leaderWorkerIDStr != "" {
		leaderWorkerID, err = strconv.Atoi(leaderWorkerIDStr)
		if err != nil {
			response.ResponseError(c, fmt.Sprintf("Cannot decode leaderWorkerID parameter: %v", err))
			return
		}
	}

	number := c.DefaultQuery("number", "")
	number, err = url.QueryUnescape(number)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Wrong query parameter provided for number: %v", err))
		return
	}

	mobileNumber := c.DefaultQuery("mobileNumber", "")
	mobileNumber, err = url.QueryUnescape(mobileNumber)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Wrong query parameter provided for mobileNumber: %v", err))
		return
	}

	company := c.DefaultQuery("company", "")
	company, err = url.QueryUnescape(company)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Wrong query parameter provided for company: %v", err))
		return
	}

	filter := model.Team{
		LeaderWorkerID: uint(leaderWorkerID),
		Number:         number,
		MobileNumber:   mobileNumber,
		Company:        company,
	}

	data, err := controller.teamService.GetPaginated(page, limit, filter)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Could not get the paginated data of Team: %v", err))
		return
	}

	dataCount, err := controller.teamService.Count()
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
	var createData model.Team
	if err := c.ShouldBindJSON(&createData); err != nil {
		response.ResponseError(c, fmt.Sprintf("Invalid data recieved by server: %v", err))
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
	var updateData model.Team
	if err := c.ShouldBindJSON(&updateData); err != nil {
		response.ResponseError(c, fmt.Sprintf("Invalid data recieved by server: %v", err))
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
