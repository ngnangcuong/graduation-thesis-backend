package handler

import (
	"graduation-thesis/internal/user/model"
	"graduation-thesis/internal/user/service"
	responseModel "graduation-thesis/pkg/model"
	"net/http"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	userService *service.UserService
}

func NewUserHandler(userService *service.UserService) *UserHandler {
	return &UserHandler{
		userService: userService,
	}
}

func (u *UserHandler) GetUser(c *gin.Context) {
	id := c.Param("id")
	successResponse, errorResponse := u.userService.GetUser(c, id)
	if errorResponse != nil {
		c.JSON(errorResponse.Status, errorResponse)
		return
	}

	c.JSON(successResponse.Status, successResponse)
}

func (u *UserHandler) Register(c *gin.Context) {
	var createUserRequest model.CreateUserRequest
	if err := c.ShouldBindJSON(&createUserRequest); err != nil {
		errorResponse := responseModel.ErrorResponse{
			Status:       http.StatusBadRequest,
			ErrorMessage: model.ErrInvalidParameter.Error(),
		}
		c.JSON(errorResponse.Status, errorResponse)
		return
	}

	successResponse, errorResponse := u.userService.CreateUser(c, createUserRequest)
	if errorResponse != nil {
		c.JSON(errorResponse.Status, errorResponse)
		return
	}

	c.JSON(successResponse.Status, successResponse)
}

func (u *UserHandler) UpdateUser(c *gin.Context) {
	id := c.Param("id")
	if user_id := c.MustGet("user_id"); id != user_id {
		errorResponse := responseModel.ErrorResponse{
			Status:       http.StatusUnauthorized,
			ErrorMessage: model.ErrNoPermission.Error(),
		}
		c.JSON(errorResponse.Status, errorResponse)
		return
	}

	var updateUserRequest model.UpdateUserRequest
	if err := c.ShouldBindJSON(&updateUserRequest); err != nil {
		errorResponse := responseModel.ErrorResponse{
			Status:       http.StatusBadRequest,
			ErrorMessage: model.ErrInvalidParameter.Error(),
		}
		c.JSON(errorResponse.Status, errorResponse)
		return
	}

	successResponse, errorResponse := u.userService.UpdateUser(c, id, updateUserRequest)
	if errorResponse != nil {
		c.JSON(errorResponse.Status, errorResponse)
		return
	}

	c.JSON(successResponse.Status, successResponse)
}

func (u *UserHandler) Verify(c *gin.Context) {
	var loginRequest model.LoginRequest
	if err := c.ShouldBindJSON(&loginRequest); err != nil {
		errorResponse := responseModel.ErrorResponse{
			Status:       http.StatusBadRequest,
			ErrorMessage: model.ErrInvalidParameter.Error(),
		}
		c.JSON(errorResponse.Status, errorResponse)
		return
	}

	successResponse, errorResposne := u.userService.VerifyCredential(c, &loginRequest)
	if errorResposne != nil {
		c.JSON(errorResposne.Status, errorResposne)
		return
	}

	c.JSON(successResponse.Status, successResponse)
}
