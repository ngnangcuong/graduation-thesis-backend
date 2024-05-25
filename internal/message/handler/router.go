package handler

import (
	"graduation-thesis/pkg/middleware"

	"github.com/gin-gonic/gin"
)

var router *gin.Engine

func initRouter(messageHandler *MessageHandler) *gin.Engine {
	router := gin.Default()
	router.Use(middleware.Headers())
	router.Use(middleware.SetupCors())

	messagePath := router.Group("/v1/message")
	{
		// Previous version
		messagePath.POST("/_search", messageHandler.SearchMessages)
		messagePath.POST("/", messageHandler.SendMessages)
		messagePath.PUT("/")
		messagePath.DELETE("/")
		messagePath.POST("/_search/conversation", messageHandler.SearchConversation)

		// New version
		messagePath.GET("/inbox/:user_id", messageHandler.Inboxes)
		messagePath.GET("/conversation/:conv_id", messageHandler.ConversationMessages)
		messagePath.POST("/read_receipt", messageHandler.ReadReceipts)
		messagePath.PUT("/read_receipt", messageHandler.UpdateReadReceipts)
		messagePath.POST("/message", messageHandler.SendMessage)

	}

	return router
}

func GetRouter(messageHandler *MessageHandler) *gin.Engine {
	if router == nil {
		router = initRouter(messageHandler)
	}

	return router
}
