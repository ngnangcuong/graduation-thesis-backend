package handler

import (
	"net/http"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"

	"graduation-thesis/internal/user/model"
	"graduation-thesis/internal/user/service"
	responseModel "graduation-thesis/pkg/model"
)

type AuthHandler struct {
	authService  *service.AuthService
	tokenService *service.TokenService
	userService  *service.UserService
}

func NewAuthHandler(authService *service.AuthService, tokenService *service.TokenService, userService *service.UserService) *AuthHandler {
	return &AuthHandler{
		authService:  authService,
		tokenService: tokenService,
		userService:  userService,
	}
}

func (a *AuthHandler) Refresh(c *gin.Context) {
	var refreshRequest model.RefreshRequest
	if err := c.ShouldBindJSON(&refreshRequest); err != nil {
		errorResponse := responseModel.ErrorResponse{
			Status:       http.StatusBadRequest,
			ErrorMessage: model.ErrNoPermission.Error(),
		}
		c.JSON(errorResponse.Status, errorResponse)
		return
	}

	tokenDetails, err := a.tokenService.Refresh(refreshRequest.RefreshToken)
	if err != nil {
		errorResponse := responseModel.ErrorResponse{
			Status:       http.StatusBadRequest,
			ErrorMessage: model.ErrInvalidParameter.Error(),
		}
		c.JSON(errorResponse.Status, errorResponse)
		return
	}

	successResponse := responseModel.SuccessResponse{
		Status: http.StatusCreated,
		Result: tokenDetails,
	}

	c.JSON(successResponse.Status, successResponse)
}

func (a *AuthHandler) Logout(c *gin.Context) {
	tokenUuid := c.GetString("access_uuid")
	_, err := a.tokenService.DeleteToken(tokenUuid)
	if err != nil {
		errorResponse := responseModel.ErrorResponse{
			Status:       http.StatusInternalServerError,
			ErrorMessage: model.ErrInternalServerError.Error(),
		}
		c.JSON(errorResponse.Status, errorResponse)
		return
	}
	successResponse := responseModel.SuccessResponse{
		Status: http.StatusOK,
	}

	c.JSON(successResponse.Status, successResponse)
}

func (a *AuthHandler) Login(c *gin.Context) {
	var loginRequest model.LoginRequest
	if err := c.ShouldBindJSON(&loginRequest); err != nil {
		errorResponse := responseModel.ErrorResponse{
			Status:       http.StatusBadRequest,
			ErrorMessage: model.ErrInvalidParameter.Error(),
		}
		c.JSON(errorResponse.Status, errorResponse)
		return
	}

	successResponse, errorResposne := a.authService.Login(c, &loginRequest)
	if errorResposne != nil {
		c.JSON(errorResposne.Status, errorResposne)
		return
	}

	c.JSON(successResponse.Status, successResponse)
}

func (a *AuthHandler) AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := a.tokenService.ExtractTokenFromRequest(c.Request)
		token, err := a.tokenService.ValidateToken(tokenString)
		if err != nil {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		claims, _ := token.Claims.(jwt.MapClaims)
		accessUuid := claims["access_uuid"]
		userId := claims["user_id"]
		id, err := a.tokenService.FetchUser(accessUuid.(string))
		if err != nil || userId != id {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		c.Set("access_uuid", accessUuid)
		c.Set("user_id", userId)
		c.Next()
	}
}
