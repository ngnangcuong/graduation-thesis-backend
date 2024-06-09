package service

import (
	"context"
	"fmt"
	"graduation-thesis/internal/message/model"
	"graduation-thesis/internal/message/repository"
	"graduation-thesis/pkg/logger"
	responseModel "graduation-thesis/pkg/model"
	request "graduation-thesis/pkg/requests"
	"math"
	"net/http"
	"time"
)

const MAXRETRY = 5

type MessageService struct {
	messageRepo     *repository.MessageRepo
	groupServiceUrl string
	logger          logger.Logger
}

func NewMessageService(messageRepo *repository.MessageRepo, groupServiceUrl string, logger logger.Logger) *MessageService {
	return &MessageService{
		messageRepo:     messageRepo,
		groupServiceUrl: groupServiceUrl,
		logger:          logger,
	}
}

func (m *MessageService) SearchMessages(ctx context.Context, request *model.SearchMessagesRequest) (*responseModel.SuccessResponse, *responseModel.ErrorResponse) {
	onlyUnread := request.OnlyUnread

	if request.Limit == 0 {
		request.Limit = 100
	}
	if onlyUnread {
		unreadMessages, err := m.messageRepo.GetUnRead(ctx, request.From, request.To, request.Limit)
		if err != nil {
			errorMessage := responseModel.ErrorResponse{
				Status:       http.StatusInternalServerError,
				ErrorMessage: err.Error(),
			}

			return nil, &errorMessage
		}

		successResponse := responseModel.SuccessResponse{
			Status: http.StatusOK,
			Result: unreadMessages,
		}
		return &successResponse, nil
	}

	messages, err := m.messageRepo.Get(ctx, request.From, request.To, request.Limit)
	if err != nil {
		errorMessage := responseModel.ErrorResponse{
			Status:       http.StatusInternalServerError,
			ErrorMessage: err.Error(),
		}

		return nil, &errorMessage
	}

	successResponse := responseModel.SuccessResponse{
		Status: http.StatusOK,
		Result: messages,
	}
	return &successResponse, nil
}

func (m *MessageService) SearchConversation(ctx context.Context, request *model.SearchConversionRequest) (*responseModel.SuccessResponse, *responseModel.ErrorResponse) {
	if len(request.Users) != 2 {
		return nil, &responseModel.ErrorResponse{
			Status:       http.StatusBadRequest,
			ErrorMessage: "invalid parameters",
		}
	}

	timeTo := request.TimeTo
	if timeTo.IsZero() {
		timeTo = time.Now()
	}

	timeFrom := request.TimeFrom
	if timeFrom.After(timeTo) || timeFrom.IsZero() {
		timeFrom = time.Now().Add(-24 * time.Hour)
	}

	messages, err := m.messageRepo.GetConversation(ctx, request.Users, timeFrom, timeTo, request.Limit)
	if err != nil {
		return nil, &responseModel.ErrorResponse{
			Status:       http.StatusInternalServerError,
			ErrorMessage: err.Error(),
		}
	}

	successResponse := responseModel.SuccessResponse{
		Result: messages,
		Status: http.StatusOK,
	}
	return &successResponse, nil
}

///////////////////////////////////////////////////

func (m *MessageService) UserInbox(ctx context.Context, userID string, limit, offset, lastInbox int) (*responseModel.SuccessResponse, *responseModel.ErrorResponse) {
	userInboxes, err := m.messageRepo.GetUserInbox(ctx, userID, limit, lastInbox)
	if err != nil {
		errorMessage := responseModel.ErrorResponse{
			Status:       http.StatusInternalServerError,
			ErrorMessage: err.Error(),
		}
		return nil, &errorMessage
	}

	successResponse := responseModel.SuccessResponse{
		Status: http.StatusOK,
		Result: userInboxes,
	}

	return &successResponse, nil
}

func (m *MessageService) getConversationMembers(ctx context.Context, userID, conversationID string) ([]string, error) {
	var (
		result interface{}
		err    error
	)
	for i := 1; i <= 5; i++ {
		result, err = request.HTTPRequestCall(
			fmt.Sprintf("%s/conversation/%s", m.groupServiceUrl, conversationID),
			http.MethodGet,
			userID,
			nil,
			5*time.Second,
		)
		if err != nil {
			time.Sleep(time.Second)
			continue
		}
	}
	if err != nil {
		return nil, err
	}

	type conversation struct {
		ID      string   `json:"id"`
		Members []string `json:"members"`
	}

	membersInterface := result.(map[string]interface{})["members"].([]interface{})
	members := make([]string, len(membersInterface))
	for i, value := range membersInterface {
		members[i] = fmt.Sprintf("%v", value)
	}
	return members, nil
}

func (m *MessageService) GetConversationMessages(ctx context.Context, userID, conversationID string, limit int, beforeMsg int64) (*responseModel.SuccessResponse, *responseModel.ErrorResponse) {
	conversationMembers, cErr := m.getConversationMembers(ctx, userID, conversationID)
	if cErr != nil {
		errorMessage := responseModel.ErrorResponse{
			Status:       http.StatusInternalServerError,
			ErrorMessage: cErr.Error(),
		}
		return nil, &errorMessage
	}

	isInConversation := false
	for _, member := range conversationMembers {
		if userID == member {
			isInConversation = true
			break
		}
	}
	if !isInConversation {
		errorMessage := responseModel.ErrorResponse{
			Status:       http.StatusUnauthorized,
			ErrorMessage: "only members can see conversation'messages",
		}
		return nil, &errorMessage
	}
	if beforeMsg <= 0 {
		beforeMsg = math.MaxInt64
	}
	conversationMessages, err := m.messageRepo.GetConversationMessages(ctx, conversationID, limit, beforeMsg)
	if err != nil {
		errorMessage := responseModel.ErrorResponse{
			Status:       http.StatusInternalServerError,
			ErrorMessage: err.Error(),
		}
		return nil, &errorMessage
	}

	successResponse := responseModel.SuccessResponse{
		Status: http.StatusOK,
		Result: conversationMessages,
	}

	return &successResponse, nil
}

func (m *MessageService) GetReadReceipts(ctx context.Context, readReceiptRequest *model.ReadReceiptRequest) (*responseModel.SuccessResponse, *responseModel.ErrorResponse) {
	if userID := readReceiptRequest.UserID; userID != "" {
		readReceipt, err := m.messageRepo.GetReadReceipt(ctx, readReceiptRequest.ConversationID, userID)
		if err != nil {
			errorResponse := responseModel.ErrorResponse{
				Status:       http.StatusInternalServerError,
				ErrorMessage: err.Error(),
			}

			return nil, &errorResponse
		}

		successResponse := responseModel.SuccessResponse{
			Status: http.StatusOK,
			Result: readReceipt,
		}
		return &successResponse, nil
	}

	readReceipts, err := m.messageRepo.GetReadReceipts(ctx, readReceiptRequest.ConversationID)
	if err != nil {
		errorResponse := responseModel.ErrorResponse{
			Status:       http.StatusInternalServerError,
			ErrorMessage: err.Error(),
		}

		return nil, &errorResponse
	}

	successResponse := responseModel.SuccessResponse{
		Status: http.StatusOK,
		Result: readReceipts,
	}
	return &successResponse, nil
}

func (m *MessageService) UpdateReadReceipts(ctx context.Context, updateReadReceiptRequest *model.UpdateReadReceiptRequest) (*responseModel.SuccessResponse, *responseModel.ErrorResponse) {
	err := m.messageRepo.UpdateReadReceipts(ctx, updateReadReceiptRequest.ConversationID, updateReadReceiptRequest.ReadReceiptUpdate)
	if err != nil {
		errorMessage := responseModel.ErrorResponse{
			Status:       http.StatusInternalServerError,
			ErrorMessage: err.Error(),
		}
		return nil, &errorMessage
	}

	go func(m *MessageService, ctx context.Context, updateReadReceiptRequest *model.UpdateReadReceiptRequest) {
		for _, readReceipt := range updateReadReceiptRequest.ReadReceiptUpdate {
			if err := m.DeleteUserInbox(ctx, readReceipt.UserID, updateReadReceiptRequest.ConversationID); err != nil {
				m.logger.Errorf("[DeleteUserInbox] Cannot clear user %v inbox: %v", readReceipt.UserID, err)
				time.Sleep(time.Second)
			}
		}
	}(m, ctx, updateReadReceiptRequest)

	successResponse := responseModel.SuccessResponse{
		Status: http.StatusNoContent,
	}
	return &successResponse, nil
}

func (m *MessageService) SendMessage(ctx context.Context, request *model.SendMessageRequest) (*responseModel.SuccessResponse, *responseModel.ErrorResponse) {
	var (
		convMsgID int64
		createErr error
	)
	if request.MessageTime == 0 {
		request.MessageTime = time.Now().Unix()
	}

	for i := 0; i < MAXRETRY; i++ {
		convMsgID, createErr = m.messageRepo.CreateConversationMessage(ctx, request.ConversationID, request.Sender, request.Content, request.MessageTime)
		if createErr == nil {
			break
		}
	}

	if createErr != nil {
		errorMessage := responseModel.ErrorResponse{
			Status:       http.StatusInternalServerError,
			ErrorMessage: createErr.Error(),
		}
		return nil, &errorMessage
	}
	go m.InsertUserInboxes(ctx, request.ConversationID, request.Sender, request.Content, convMsgID, request.MessageTime)
	go m.messageRepo.UpdateReadReceipts(ctx, request.ConversationID, []model.ReadReceiptUpdate{
		{
			UserID:    request.Sender,
			MessageID: convMsgID,
		},
	})

	successMessage := responseModel.SuccessResponse{
		Status: http.StatusCreated,
		Result: convMsgID,
	}
	return &successMessage, nil
}

func (m *MessageService) InsertUserInboxes(ctx context.Context, conversationID, sender, content string, convMsgID, messageTime int64) error {
	var (
		members []string
		err     error
	)
	for i := 0; i < MAXRETRY; i++ {
		members, err = m.getConversationMembers(ctx, sender, conversationID)
		if err == nil {
			break
		}
		m.logger.Errorf("[InsertUserInboxes] Cannot get conversation %s 'members: %v", conversationID, err)
		time.Sleep(time.Second)
	}
	if err != nil {
		return err
	}

	for _, member := range members {
		if member == sender {
			continue
		}

		go m.messageRepo.InsertUserInbox(ctx, member, conversationID, sender, content, convMsgID, messageTime)
	}

	return nil
}

func (m *MessageService) DeleteUserInbox(ctx context.Context, userID, conversationID string) error {
	if err := m.messageRepo.DeleteUserInbox(ctx, userID, conversationID); err != nil {
		return err
	}

	return nil
}
