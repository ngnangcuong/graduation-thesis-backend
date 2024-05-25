package http_handler

import (
	"graduation-thesis/internal/websocket_manager/model"
	"graduation-thesis/internal/websocket_manager/service"
	responseModel "graduation-thesis/pkg/model"
	"net/http"

	"github.com/gin-gonic/gin"
)

type WebsocketHandler struct {
	websocketManagerService *service.WebsocketManagerService
}

func NewWebSocketHandler(websocketManagerService *service.WebsocketManagerService) *WebsocketHandler {
	return &WebsocketHandler{
		websocketManagerService: websocketManagerService,
	}
}

func (w *WebsocketHandler) GetUserList(c *gin.Context) {
	websocketHandlerID := c.Param("id")
	successResponse, errorResponse := w.websocketManagerService.GetUsers(c, websocketHandlerID)
	if errorResponse != nil {
		c.JSON(errorResponse.Status, errorResponse)
		return
	}

	c.JSON(successResponse.Status, successResponse)
}

func (w *WebsocketHandler) GetWebsocketHandlerList(c *gin.Context) {
	successResponse, errorResponse := w.websocketManagerService.GetWebsocketHandlers(c)
	if errorResponse != nil {
		c.JSON(errorResponse.Status, errorResponse)
		return
	}

	c.JSON(successResponse.Status, errorResponse)
}

func (w *WebsocketHandler) AddNewWebsocketHandler(c *gin.Context) {
	var addNewWebsocketHandlerRequest model.AddNewWebsocketHandlerRequest
	if err := c.ShouldBindJSON(&addNewWebsocketHandlerRequest); err != nil {
		errorResponse := responseModel.ErrorResponse{
			Status:       http.StatusBadRequest,
			ErrorMessage: err.Error(),
		}
		c.JSON(errorResponse.Status, errorResponse)
	}

	successResponse, errorResponse := w.websocketManagerService.AddNewWebsocketHandler(c, &addNewWebsocketHandlerRequest)
	if errorResponse != nil {
		c.JSON(errorResponse.Status, errorResponse)
		return
	}

	c.JSON(successResponse.Status, successResponse)
}

func (w *WebsocketHandler) AddNewUser(c *gin.Context) {
	var addNewUserRequest model.AddNewUserRequest
	if err := c.ShouldBindJSON(&addNewUserRequest); err != nil {
		errorResponse := responseModel.ErrorResponse{
			Status:       http.StatusBadRequest,
			ErrorMessage: err.Error(),
		}
		c.JSON(errorResponse.Status, errorResponse)
	}

	successResponse, errorResponse := w.websocketManagerService.AddNewUser(c, &addNewUserRequest)
	if errorResponse != nil {
		c.JSON(errorResponse.Status, errorResponse)
		return
	}

	c.JSON(successResponse.Status, successResponse)
}

func (w *WebsocketHandler) DisconnectUser(c *gin.Context) {
	var disconnectUserRequest model.AddNewUserRequest
	if err := c.ShouldBindJSON(&disconnectUserRequest); err != nil {
		errorResponse := responseModel.ErrorResponse{
			Status:       http.StatusBadRequest,
			ErrorMessage: err.Error(),
		}
		c.JSON(errorResponse.Status, errorResponse)
		return
	}

	successResponse, errorResponse := w.websocketManagerService.RemoveUser(c, &disconnectUserRequest)
	if errorResponse != nil {
		c.JSON(errorResponse.Status, errorResponse)
		return
	}

	c.JSON(successResponse.Status, successResponse)
}

func (w *WebsocketHandler) Ping(c *gin.Context) {
	var pingRequest model.PingRequest
	if err := c.ShouldBindJSON(&pingRequest); err != nil {
		errorResponse := responseModel.ErrorResponse{
			Status:       http.StatusBadRequest,
			ErrorMessage: err.Error(),
		}
		c.JSON(errorResponse.Status, errorResponse)
		return
	}

	successResponse, errorResponse := w.websocketManagerService.Pong(c, &pingRequest)
	if errorResponse != nil {
		c.JSON(errorResponse.Status, errorResponse)
		return
	}

	c.JSON(successResponse.Status, successResponse)
}
