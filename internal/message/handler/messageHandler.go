package handler

import (
	"graduation-thesis/internal/message/model"
	"graduation-thesis/internal/message/service"
	responseModel "graduation-thesis/pkg/model"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type MessageHandler struct {
	messageService   *service.MessageService
	authenticatorURL string
}

func NewMessageHandler(messageService *service.MessageService, authenticatorURL string) *MessageHandler {
	return &MessageHandler{
		messageService:   messageService,
		authenticatorURL: authenticatorURL,
	}
}

func (m *MessageHandler) SearchMessages(c *gin.Context) {
	var searchMessagesRequest model.SearchMessagesRequest
	if err := c.ShouldBindJSON(&searchMessagesRequest); err != nil {
		errorMessage := responseModel.ErrorResponse{
			Status:       http.StatusBadRequest,
			ErrorMessage: err.Error(),
		}
		c.JSON(errorMessage.Status, errorMessage)
		return
	}

	successResponse, errorResponse := m.messageService.SearchMessages(c, &searchMessagesRequest)
	if errorResponse != nil {
		c.JSON(errorResponse.Status, errorResponse)
		return
	}

	c.JSON(successResponse.Status, successResponse)
}

func (m *MessageHandler) SendMessages(c *gin.Context) {
	var sendMessageRequest model.SendMessageRequest
	if err := c.ShouldBindJSON(&sendMessageRequest); err != nil {
		errorMessage := responseModel.ErrorResponse{
			Status:       http.StatusBadRequest,
			ErrorMessage: err.Error(),
		}
		c.JSON(errorMessage.Status, errorMessage)
		return
	}

	successResponse, errorResponse := m.messageService.SendMessage(c, &sendMessageRequest)
	if errorResponse != nil {
		c.JSON(errorResponse.Status, errorResponse)
		return
	}

	c.JSON(successResponse.Status, successResponse)
}

func (m *MessageHandler) SearchConversation(c *gin.Context) {
	var searchConversionRequest model.SearchConversionRequest
	if err := c.ShouldBindJSON(&searchConversionRequest); err != nil {
		errorMessage := responseModel.ErrorResponse{
			Status:       http.StatusBadRequest,
			ErrorMessage: err.Error(),
		}
		c.JSON(errorMessage.Status, errorMessage)
		return
	}

	successResponse, errorResponse := m.messageService.SearchConversation(c, &searchConversionRequest)
	if errorResponse != nil {
		c.JSON(errorResponse.Status, errorResponse)
		return
	}

	c.JSON(successResponse.Status, successResponse)
}

func (m *MessageHandler) Inboxes(c *gin.Context) {
	userID := c.Param("user_id")
	if userID != c.Request.Header.Get("X-User-ID") {
		errorMessage := responseModel.ErrorResponse{
			Status:       http.StatusUnauthorized,
			ErrorMessage: "cannot query another user's inbox",
		}
		c.JSON(errorMessage.Status, errorMessage)
		return
	}
	limitQuery := c.DefaultQuery("limit", "1000")
	offsetQuery := c.DefaultQuery("offset", "1")
	lastInboxQuery := c.DefaultQuery("last_inbox", "0")
	limit, lErr := strconv.Atoi(limitQuery)
	offset, oErr := strconv.Atoi(offsetQuery)
	lastInbox, laErr := strconv.Atoi(lastInboxQuery)
	if lErr != nil || oErr != nil || laErr != nil {
		errorMessage := responseModel.ErrorResponse{
			Status:       http.StatusBadRequest,
			ErrorMessage: "invalid parameter",
		}
		c.JSON(errorMessage.Status, errorMessage)
		return
	}

	successReponse, errorResponse := m.messageService.UserInbox(c, userID, limit, offset, lastInbox)
	if errorResponse != nil {
		c.JSON(errorResponse.Status, errorResponse)
		return
	}

	c.JSON(successReponse.Status, successReponse)
}

func (m *MessageHandler) ConversationMessages(c *gin.Context) {
	userID := c.Request.Header.Get("X-User-ID")
	conversationID := c.Param("conv_id")
	limitQuery := c.DefaultQuery("limit", "20")
	beforeMsgQuery := c.DefaultQuery("before_msg", "0")
	limit, lErr := strconv.Atoi(limitQuery)
	beforeMsg, bErr := strconv.ParseInt(beforeMsgQuery, 10, 64)
	if lErr != nil || bErr != nil {
		errorResponse := responseModel.ErrorResponse{
			Status:       http.StatusBadRequest,
			ErrorMessage: "invalid parameters",
		}
		c.JSON(errorResponse.Status, errorResponse)
		return
	}

	successResponse, errorResponse := m.messageService.GetConversationMessages(c, userID, conversationID, limit, beforeMsg)
	if errorResponse != nil {
		c.JSON(errorResponse.Status, errorResponse)
		return
	}

	c.JSON(successResponse.Status, successResponse)
}

func (m *MessageHandler) ReadReceipts(c *gin.Context) {
	var readReceipt model.ReadReceiptRequest
	if err := c.ShouldBindJSON(&readReceipt); err != nil {
		errorMessage := responseModel.ErrorResponse{
			Status:       http.StatusBadRequest,
			ErrorMessage: err.Error(),
		}

		c.JSON(errorMessage.Status, errorMessage)
		return
	}

	successResponse, errorResponse := m.messageService.GetReadReceipts(c, &readReceipt)
	if errorResponse != nil {
		c.JSON(errorResponse.Status, errorResponse)
		return
	}

	c.JSON(successResponse.Status, successResponse)
}

func (m *MessageHandler) UpdateReadReceipts(c *gin.Context) {
	var updateReadReceiptRequest model.UpdateReadReceiptRequest
	if err := c.ShouldBindJSON(&updateReadReceiptRequest); err != nil {
		errorMessage := responseModel.ErrorResponse{
			Status:       http.StatusBadRequest,
			ErrorMessage: err.Error(),
		}

		c.JSON(errorMessage.Status, errorMessage)
		return
	}

	successResponse, errorResposne := m.messageService.UpdateReadReceipts(c, &updateReadReceiptRequest)
	if errorResposne != nil {
		c.JSON(errorResposne.Status, errorResposne)
		return
	}

	c.JSON(successResponse.Status, successResponse)
}

func (m *MessageHandler) SendMessage(c *gin.Context) {
	var sendMessageRequest model.SendMessageRequest
	if err := c.ShouldBindJSON(&sendMessageRequest); err != nil {
		errorMessage := responseModel.ErrorResponse{
			Status:       http.StatusBadRequest,
			ErrorMessage: err.Error(),
		}

		c.JSON(errorMessage.Status, errorMessage)
		return
	}

	successResponse, errorResponse := m.messageService.SendMessage(c, &sendMessageRequest)
	if errorResponse != nil {
		c.JSON(errorResponse.Status, errorResponse)
		return
	}

	c.JSON(successResponse.Status, successResponse)
}
