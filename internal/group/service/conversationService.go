package service

import (
	"context"
	"database/sql"
	"graduation-thesis/internal/group/model"
	"graduation-thesis/internal/group/repository"
	"graduation-thesis/pkg/custom_error"
	responseModel "graduation-thesis/pkg/model"
	"net/http"
	"time"

	"github.com/twinj/uuid"
)

type ConversationService struct {
	db               *sql.DB
	conversationRepo *repository.ConversationRepo
	errorMap         map[error]int
}

func NewConversationService(db *sql.DB, conversationRepo *repository.ConversationRepo, errorMap map[error]int) *ConversationService {
	return &ConversationService{
		db:               db,
		conversationRepo: conversationRepo,
		errorMap:         errorMap,
	}
}

func (c *ConversationService) execTx(ctx context.Context, fn func(c *repository.ConversationRepo) error) error {
	tx, err := c.db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return err
	}

	convRepoWithTx := c.conversationRepo.WithTx(tx)
	fErr := fn(convRepoWithTx)
	if fErr != nil {
		_ = tx.Rollback()
		return fErr
	}

	return tx.Commit()
}

func (c *ConversationService) GetConversation(ctx context.Context, conversationID string) (*responseModel.SuccessResponse, *responseModel.ErrorResponse) {
	queryContext, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	conversation, err := c.conversationRepo.GetMembers(queryContext, conversationID)
	if err != nil {
		errorResponse := responseModel.ErrorResponse{
			Status:       c.errorMap[err],
			ErrorMessage: err.Error(),
		}
		return nil, &errorResponse
	}

	successResponse := responseModel.SuccessResponse{
		Status: http.StatusOK,
		Result: conversation,
	}
	return &successResponse, nil
}

func (c *ConversationService) CreateConversation(ctx context.Context, request *model.CreateConversationRequest) (*responseModel.SuccessResponse, *responseModel.ErrorResponse) {
	if len(request.Members) < 2 {
		errorResponse := responseModel.ErrorResponse{
			Status:       c.errorMap[custom_error.ErrInvalidParameter],
			ErrorMessage: custom_error.ErrInvalidParameter.Error(),
		}
		return nil, &errorResponse
	}
	queryContext, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	conversationID := uuid.NewV4().String()
	err := c.execTx(ctx, func(c *repository.ConversationRepo) error {
		if err := c.Create(queryContext, conversationID); err != nil {
			return err
		}

		addErr := c.AddMembers(queryContext, conversationID, request.Members)
		if addErr != nil {
			return addErr
		}

		return nil
	})

	if err != nil {
		errorResponse := responseModel.ErrorResponse{
			Status:       c.errorMap[err],
			ErrorMessage: err.Error(),
		}
		return nil, &errorResponse
	}

	successResponse := responseModel.SuccessResponse{
		Status: http.StatusCreated,
		Result: model.Conversation{
			ID:      conversationID,
			Members: request.Members,
		},
	}
	return &successResponse, nil
}

func (c *ConversationService) GetConversationsContainUser(ctx context.Context, userID string) (*responseModel.SuccessResponse, *responseModel.ErrorResponse) {
	conversations, err := c.conversationRepo.GetConversations(ctx, userID)
	if err != nil {
		errorResponse := responseModel.ErrorResponse{
			Status:       c.errorMap[err],
			ErrorMessage: err.Error(),
		}
		return nil, &errorResponse
	}

	successResponse := responseModel.SuccessResponse{
		Status: http.StatusOK,
		Result: conversations,
	}
	return &successResponse, nil
}
