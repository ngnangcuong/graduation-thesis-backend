package service

import (
	"context"
	"graduation-thesis/internal/message/model"
	"graduation-thesis/internal/message/repository"
	responseModel "graduation-thesis/pkg/model"
	"math"
	"net/http"
	"time"
)

const MAXRETRY = 5

type MessageService struct {
	messageRepo *repository.MessageRepo
}

func NewMessageService(messageRepo *repository.MessageRepo) *MessageService {
	return &MessageService{
		messageRepo: messageRepo,
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

	var conversationList = make(map[string][]*model.MessageResponse)
	userInboxResponse := model.UserInboxResponse{
		UserID: userID,
	}
	for _, userInbox := range userInboxes {
		if _, ok := conversationList[userInbox.ConversationID]; !ok {
			conversationList[userInbox.ConversationID] = make([]*model.MessageResponse, 0)
		}

		conversationList[userInbox.ConversationID] = append(conversationList[userInbox.ConversationID], &model.MessageResponse{
			ConversationMessageID: userInbox.ConversationMessageID,
			MessageTime:           userInbox.MessageTime,
			Sender:                userInbox.Sender,
			Content:               userInbox.Content,
		})
	}

	userInboxResponse.Inboxes = make([]model.Inbox, 0, len(conversationList))
	for conversationID, messageResponses := range conversationList {
		inboxEntry := model.Inbox{
			Count:          len(messageResponses),
			ConversationID: conversationID,
			Messages:       messageResponses,
		}
		userInboxResponse.Inboxes = append(userInboxResponse.Inboxes, inboxEntry)
	}
	successResponse := responseModel.SuccessResponse{
		Status: http.StatusOK,
		Result: userInboxResponse,
	}

	return &successResponse, nil
}

func (m *MessageService) GetConversationMessages(ctx context.Context, conversationID string, limit int, beforeMsg int64) (*responseModel.SuccessResponse, *responseModel.ErrorResponse) {
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

func (m *MessageService) InsertUserInboxes(ctx context.Context, conversationID, sender string, content []byte, convMsgID, messageTime int64) error {
	var (
		readReceipts []*model.ReadReceipt
		err          error
	)
	for i := 0; i < MAXRETRY; i++ {
		readReceipts, err = m.messageRepo.GetReadReceipts(ctx, conversationID) // TODO: Review???
		if err == nil {
			break
		}
	}
	if err != nil {
		return err
	}

	for _, readReceipt := range readReceipts {
		if readReceipt.UserID == sender {
			continue
		}

		go m.messageRepo.InsertUserInbox(ctx, readReceipt.UserID, conversationID, sender, content, convMsgID, messageTime)
	}

	return nil
}
