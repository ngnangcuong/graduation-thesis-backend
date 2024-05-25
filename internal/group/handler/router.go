package handler

import (
	"graduation-thesis/pkg/middleware"

	"github.com/gin-gonic/gin"
)

var router *gin.Engine

func NewRouter(groupHandler *GroupHandler, conversationHandler *ConversationHandler) *gin.Engine {
	r := gin.Default()
	r.Use(middleware.Headers())
	r.Use(middleware.SetupCors())

	groupPath := r.Group("/v1/group")
	{
		groupPath.GET("/:group_id", groupHandler.GetGroup)
		groupPath.GET("/:group_name", groupHandler.GetGroup)
		groupPath.PUT("/:group_id", groupHandler.UpdateGroup)
		groupPath.POST("/", groupHandler.CreateGroup)
		groupPath.DELETE("/:group_id", groupHandler.DeleteGroup)
	}

	conversationPath := r.Group("/v1/conversation")
	{
		conversationPath.GET("/:conversation_id", conversationHandler.GetConversation)
		conversationPath.POST("/", conversationHandler.CreateConversation)
		conversationPath.GET("/user/:user_id", conversationHandler.GetConversationsContainUser)
	}

	return r
}

func GetRouter(groupHandler *GroupHandler, conversationHandler *ConversationHandler) *gin.Engine {
	if router == nil {
		router = NewRouter(groupHandler, conversationHandler)
	}

	return router
}
