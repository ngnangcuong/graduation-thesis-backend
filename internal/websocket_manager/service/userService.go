package service

import (
	"context"
	"graduation-thesis/internal/websocket_manager/repository"
	responseModel "graduation-thesis/pkg/model"
	"net/http"
)

type UserService struct {
	userRepo *repository.UserRepo
	errorMap map[error]int
}

func NewUserService(userRepo *repository.UserRepo, errorMap map[error]int) *UserService {
	return &UserService{
		userRepo: userRepo,
		errorMap: errorMap,
	}
}

func (u *UserService) GetWebsocketHandler(ctx context.Context, userID string) (*responseModel.SuccessResponse, *responseModel.ErrorResponse) {
	websocketHandler, err := u.userRepo.Get(ctx, userID)
	if err != nil {
		errorResponse := responseModel.ErrorResponse{
			Status:       u.errorMap[err],
			ErrorMessage: err.Error(),
		}
		return nil, &errorResponse
	}

	successResponse := responseModel.SuccessResponse{
		Status: http.StatusOK,
		Result: websocketHandler,
	}
	return &successResponse, nil
}
