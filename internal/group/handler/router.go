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
		groupPath.GET("", middleware.AuthMiddlewareV2(groupHandler.authenticatorURL), groupHandler.GetGroup)
		groupPath.PUT("/:group_id", middleware.AuthMiddlewareV2(groupHandler.authenticatorURL), groupHandler.UpdateGroup)
		groupPath.PUT("/:group_id/leave", middleware.AuthMiddlewareV2(groupHandler.authenticatorURL), groupHandler.LeaveGroup)
		groupPath.POST("", middleware.AuthMiddlewareV2(groupHandler.authenticatorURL), groupHandler.CreateGroup)
		groupPath.DELETE("/:group_id", middleware.AuthMiddlewareV2(groupHandler.authenticatorURL), groupHandler.DeleteGroup)
	}

	conversationPath := r.Group("/v1/conversation")
	{
		conversationPath.GET("/:conversation_id", conversationHandler.GetConversation)
		conversationPath.POST("", conversationHandler.CreateConversation)
		conversationPath.GET("/user/:user_id", middleware.AuthMiddlewareV2(conversationHandler.authenticatorURL), conversationHandler.GetConversationsContainUser)
		conversationPath.GET("/user", middleware.AuthMiddlewareV2(conversationHandler.authenticatorURL), conversationHandler.GetDirectedConversation)
	}

	return r
}

func GetRouter(groupHandler *GroupHandler, conversationHandler *ConversationHandler) *gin.Engine {
	if router == nil {
		router = NewRouter(groupHandler, conversationHandler)
	}

	return router
}
