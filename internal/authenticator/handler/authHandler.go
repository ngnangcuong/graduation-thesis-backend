package handler

import (
	"net/http"

	"graduation-thesis/internal/authenticator/model"
	"graduation-thesis/internal/authenticator/service"
	"graduation-thesis/pkg/custom_error"
	responseModel "graduation-thesis/pkg/model"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	authService  *service.AuthService
	tokenService *service.TokenService
}

func NewAuthHandler(authService *service.AuthService, tokenService *service.TokenService) *AuthHandler {
	return &AuthHandler{
		authService:  authService,
		tokenService: tokenService,
	}
}

func (a *AuthHandler) Refresh(c *gin.Context) {
	var refreshRequest model.RefreshRequest
	if err := c.ShouldBindJSON(&refreshRequest); err != nil {
		errorResponse := responseModel.ErrorResponse{
			Status:       http.StatusUnauthorized,
			ErrorMessage: custom_error.ErrNoPermission.Error(),
		}
		c.JSON(errorResponse.Status, errorResponse)
		return
	}

	successResponse, errorResponse := a.tokenService.Refresh(refreshRequest.RefreshToken)
	if errorResponse != nil {
		c.JSON(errorResponse.Status, errorResponse)
		return
	}

	c.JSON(successResponse.Status, successResponse)
}

func (a *AuthHandler) Validate(c *gin.Context) {
	// token := a.tokenService.ExtractTokenFromRequest(c.Request)
	// successResponse, errorResponse := a.tokenService.ValidateToken(token)
	// if errorResponse != nil {
	// 	c.JSON(errorResponse.Status, errorResponse)
	// 	return
	// }

	// c.JSON(successResponse.Status, successResponse)
	c.JSON(200, gin.H{
		"user_id": 1,
	})
}

func (a *AuthHandler) Login(c *gin.Context) {
	var loginRequest model.LoginRequest
	if err := c.ShouldBindJSON(&loginRequest); err != nil {
		errorResponse := responseModel.ErrorResponse{
			Status:       http.StatusBadRequest,
			ErrorMessage: err.Error(),
		}
		c.JSON(errorResponse.Status, errorResponse)
		return
	}

	successResponse, errorResponse := a.authService.Login(c, &loginRequest)
	if errorResponse != nil {
		c.JSON(errorResponse.Status, errorResponse)
		return
	}

	c.JSON(successResponse.Status, successResponse)
}

func (a *AuthHandler) Logout(c *gin.Context) {

}
