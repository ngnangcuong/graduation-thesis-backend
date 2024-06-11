package handler

import (
	"graduation-thesis/internal/group/model"
	"graduation-thesis/internal/group/service"
	responseModel "graduation-thesis/pkg/model"
	"net/http"

	"github.com/gin-gonic/gin"
)

type ConversationHandler struct {
	conversationService *service.ConversationService
	authenticatorURL    string
}

func NewConversationHandler(conversationService *service.ConversationService, authenticatorURL string) *ConversationHandler {
	return &ConversationHandler{
		conversationService: conversationService,
		authenticatorURL:    authenticatorURL,
	}
}

func (con *ConversationHandler) GetConversation(c *gin.Context) {
	conversationID := c.Param("conversation_id")
	successResponse, errorResponse := con.conversationService.GetConversation(c, conversationID)
	if errorResponse != nil {
		c.JSON(errorResponse.Status, errorResponse)
		return
	}

	c.JSON(successResponse.Status, successResponse)

}

func (con *ConversationHandler) CreateConversation(c *gin.Context) {
	var createConversationRequest model.CreateConversationRequest
	if err := c.ShouldBindJSON(&createConversationRequest); err != nil {
		errorResponse := responseModel.ErrorResponse{
			Status:       http.StatusBadRequest,
			ErrorMessage: err.Error(),
		}
		c.JSON(errorResponse.Status, errorResponse)
		return
	}

	successResponse, errorResponse := con.conversationService.CreateConversation(c, &createConversationRequest)
	if errorResponse != nil {
		c.JSON(errorResponse.Status, errorResponse)
		return
	}

	c.JSON(successResponse.Status, successResponse)
}

func (con *ConversationHandler) GetConversationsContainUser(c *gin.Context) {
	userID := c.Param("user_id")
	if userID != c.Request.Header.Get("X-User-ID") {
		errorResponse := responseModel.ErrorResponse{
			Status:       http.StatusForbidden,
			ErrorMessage: "no permission",
		}
		c.JSON(errorResponse.Status, errorResponse)
		return
	}
	successResponse, errorResponse := con.conversationService.GetConversationsContainUser(c, userID)
	if errorResponse != nil {
		c.JSON(errorResponse.Status, errorResponse)
		return
	}

	c.JSON(successResponse.Status, successResponse)
}

func (con *ConversationHandler) GetDirectedConversation(c *gin.Context) {
	otherUser := c.Query("with")
	userID := c.Request.Header.Get("X-User-ID")

	successResponse, errorResponse := con.conversationService.GetDirectedConversation(c, userID, otherUser)
	if errorResponse != nil {
		c.JSON(errorResponse.Status, errorResponse)
		return
	}

	c.JSON(successResponse.Status, successResponse)
}
