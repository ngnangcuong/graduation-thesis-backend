package service

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"time"

	"graduation-thesis/internal/authenticator/model"
	"graduation-thesis/pkg/custom_error"
	responseModel "graduation-thesis/pkg/model"
	request "graduation-thesis/pkg/requests"
)

type AuthService struct {
	tokenService   *TokenService
	mapError       map[error]int
	userServiceUrl string
}

func NewAuthService(tokenService *TokenService, mapError map[error]int, userServiceUrl string) *AuthService {
	return &AuthService{
		tokenService:   tokenService,
		mapError:       mapError,
		userServiceUrl: userServiceUrl,
	}
}

func (a *AuthService) Login(ctx context.Context, loginRequest *model.LoginRequest) (*responseModel.SuccessResponse, *responseModel.ErrorResponse) {
	var successResponse responseModel.SuccessResponse
	var errorResponse responseModel.ErrorResponse

	userID, err := a.verifyCredential(ctx, loginRequest)
	if err != nil {
		errorResponse.Status = a.mapError[err]
		errorResponse.ErrorMessage = err.Error()
		return nil, &errorResponse
	}

	tokenDetails, tokenErr := a.tokenService.CreateToken(userID)
	if tokenErr != nil {
		errorResponse.Status = a.mapError[err]
		errorResponse.ErrorMessage = err.Error()
		return nil, &errorResponse
	}

	successResponse.Result = tokenDetails
	successResponse.Status = http.StatusOK
	return &successResponse, nil
}

func (a *AuthService) verifyCredential(ctx context.Context, loginRequest *model.LoginRequest) (string, error) {
	body := new(bytes.Buffer)
	if err := json.NewEncoder(body).Encode(&loginRequest); err != nil {
		return "", custom_error.ErrInvalidParameter
	}

	result, err := request.HTTPRequestCall(
		a.userServiceUrl,
		http.MethodPost,
		"",
		body,
		5*time.Second,
	)
	if err != nil {
		return "", err
	}

	return result.(string), nil
}

func (a *AuthService) Logout(ctx context.Context) {

}
