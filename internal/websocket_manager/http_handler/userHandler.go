package http_handler

import (
	"graduation-thesis/internal/websocket_manager/service"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	userService *service.UserService
}

func NewUserHandler(userService *service.UserService) *UserHandler {
	return &UserHandler{
		userService: userService,
	}
}

func (u *UserHandler) GetWebsocketHandler(c *gin.Context) {
	userID := c.Param("user_id")
	successResponse, errorResponse := u.userService.GetWebsocketHandler(c, userID)
	if errorResponse != nil {
		c.JSON(errorResponse.Status, errorResponse)
		return
	}

	c.JSON(successResponse.Status, successResponse)
}
