package service

import (
	"context"
	"database/sql"
	"errors"
	"graduation-thesis/internal/user/helper/argon2"
	"graduation-thesis/internal/user/model"
	"graduation-thesis/internal/user/repository/user"
	responseModel "graduation-thesis/pkg/model"
	"net/http"
)

type AuthService struct {
	userRepo     *user.UserRepoPostgres
	tokenService *TokenService
}

func NewAuthService(userRepo *user.UserRepoPostgres, tokenService *TokenService) *AuthService {
	return &AuthService{
		userRepo:     userRepo,
		tokenService: tokenService,
	}
}

func (a *AuthService) Login(ctx context.Context, loginRequest *model.LoginRequest) (*responseModel.SuccessResponse, *responseModel.ErrorResponse) {
	var successResponse responseModel.SuccessResponse
	var errorResponse responseModel.ErrorResponse

	user, err := a.userRepo.GetByUsername(ctx, loginRequest.Username)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			errorResponse.Status = http.StatusBadRequest
			errorResponse.ErrorMessage = model.ErrNoPermission.Error()
		} else {
			errorResponse.Status = http.StatusInternalServerError
			errorResponse.ErrorMessage = model.ErrInternalServerError.Error()
		}
		return nil, &errorResponse
	}

	check, checkErr := argon2.Compare(user.Password, []byte(loginRequest.Password))
	if !check || checkErr != nil {
		errorResponse.Status = http.StatusBadRequest
		errorResponse.ErrorMessage = model.ErrNoPermission.Error()
		return nil, &errorResponse
	}

	tokenDetails, tokenErr := a.tokenService.CreateToken(user.ID)
	if tokenErr != nil {
		errorResponse.Status = http.StatusInternalServerError
		errorResponse.ErrorMessage = model.ErrInternalServerError.Error()
		return nil, &errorResponse
	}

	successResponse.Result = tokenDetails
	successResponse.Status = http.StatusOK
	return &successResponse, nil
}
