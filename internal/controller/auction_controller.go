package controller

import (
	"backend-v2/internal/dto"
	"backend-v2/internal/service"
	"backend-v2/pkg/response"
	"fmt"
	"strconv"

	"github.com/gin-gonic/gin"
)

type auctionController struct {
	auctionService service.IAuctionService
}

func InitAuctionController(auctionService service.IAuctionService) IAuctionController {
	return &auctionController{
		auctionService: auctionService,
	}
}

type IAuctionController interface {
	GetAuctionDataForPublic(c *gin.Context)
	GetAuctionDataForPrivate(c *gin.Context)
	SaveParticipantChanges(c *gin.Context)
}

func (controller *auctionController) GetAuctionDataForPublic(c *gin.Context) {
	auctionIDRaw := c.Param("auctionID")
	auctionID, err := strconv.ParseUint(auctionIDRaw, 10, 64)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Неправильный параметер запроса: %v", err))
		return
	}

	result, err := controller.auctionService.GetAuctionDataForPublic(uint(auctionID))
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Внутренняя ошибка сервера: %v", err))
		return
	}

	response.ResponseSuccess(c, result)
}

func (controller *auctionController) GetAuctionDataForPrivate(c *gin.Context) {
	auctionIDRaw := c.Param("auctionID")
	auctionID, err := strconv.ParseUint(auctionIDRaw, 10, 64)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Неправильный параметер запроса: %v", err))
		return
	}

	result, err := controller.auctionService.GetAuctionDataForPrivate(uint(auctionID), c.GetUint("userID"))
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Внутренняя ошибка сервера: %v", err))
		return
	}

	response.ResponseSuccess(c, result)
}

func (controller *auctionController) SaveParticipantChanges(c *gin.Context) {
	var participantChanges []dto.ParticipantDataForSave
	if err := c.ShouldBindJSON(&participantChanges); err != nil {
		response.ResponseError(c, fmt.Sprintf("Request Error: %v", err))
		return
	}

	err := controller.auctionService.SaveParticipantChanges(c.GetUint("userID"), participantChanges)
	if err != nil {
		response.ResponseError(c, fmt.Sprintf("Internal Server Error: %v", err))
		return
	}

	response.ResponseSuccess(c, true)
}
