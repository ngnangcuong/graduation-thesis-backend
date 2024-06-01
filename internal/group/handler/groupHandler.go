package handler

import (
	"graduation-thesis/internal/group/model"
	"graduation-thesis/internal/group/service"
	responseModel "graduation-thesis/pkg/model"
	"net/http"

	"github.com/gin-gonic/gin"
)

type GroupHandler struct {
	groupService *service.GroupService
}

func NewGroupHandler(groupService *service.GroupService) *GroupHandler {
	return &GroupHandler{
		groupService: groupService,
	}
}

func (g *GroupHandler) GetGroup(c *gin.Context) {
	// userID := c.MustGet("user_id").(string)
	userID := c.Request.Header.Get("X-User-ID")
	groupID := c.Query("group_id")
	groupName := c.Query("group_name")
	conversationID := c.Query("conv_id")

	successResponse, errorResponse := g.groupService.GetGroup(c, userID, groupID, groupName, conversationID)
	if errorResponse != nil {
		c.JSON(errorResponse.Status, errorResponse)
		return
	}

	c.JSON(successResponse.Status, successResponse)
}

func (g *GroupHandler) CreateGroup(c *gin.Context) {
	// userID := c.MustGet("user_id").(string)
	userID := c.Request.Header.Get("X-User-ID")

	var createGroupRequest model.CreateGroupRequest
	if err := c.ShouldBindJSON(&createGroupRequest); err != nil {
		errorResponse := responseModel.ErrorResponse{
			Status:       http.StatusBadRequest,
			ErrorMessage: err.Error(),
		}
		c.JSON(errorResponse.Status, errorResponse)
		return
	}

	successResponse, errorResponse := g.groupService.CreateGroup(c, &createGroupRequest, userID)
	if errorResponse != nil {
		c.JSON(errorResponse.Status, errorResponse)
		return
	}

	c.JSON(successResponse.Status, successResponse)

}

func (g *GroupHandler) UpdateGroup(c *gin.Context) {
	groupID := c.Param("group_id")
	userID := c.Request.Header.Get("X-User-ID")
	// userID := c.MustGet("user_id").(string)
	var updateGroupRequest model.UpdateGroupRequest
	if err := c.ShouldBindJSON(&updateGroupRequest); err != nil {
		errorResponse := responseModel.ErrorResponse{
			Status:       http.StatusBadRequest,
			ErrorMessage: err.Error(),
		}
		c.JSON(errorResponse.Status, errorResponse)
		return
	}

	successResponse, errorResponse := g.groupService.UpdateGroup(c, &updateGroupRequest, groupID, userID)
	if errorResponse != nil {
		c.JSON(errorResponse.Status, errorResponse)
		return
	}

	c.JSON(successResponse.Status, successResponse)
}

func (g *GroupHandler) DeleteGroup(c *gin.Context) {
	groupID := c.Param("group_id")
	userID := c.Request.Header.Get("X-User-ID")
	// userID := c.MustGet("user_id").(string)
	successResponse, errorResponse := g.groupService.DeleteGroup(c, groupID, userID)
	if errorResponse != nil {
		c.JSON(errorResponse.Status, errorResponse)
		return
	}

	c.JSON(successResponse.Status, successResponse)
}
