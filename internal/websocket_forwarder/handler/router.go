package handler

import (
	"graduation-thesis/pkg/middleware"

	"github.com/gin-gonic/gin"
)

var router *gin.Engine

func NewRouter(websocketForwarder *WebsocketForwarder) *gin.Engine {
	r := gin.Default()
	r.Use(middleware.Headers())
	r.Use(middleware.SetupCors())
	r.POST("/ws", websocketForwarder.HandleRequest)
	return r
}

func GetRouter(websocketForwarder *WebsocketForwarder) *gin.Engine {
	if router == nil {
		router = NewRouter(websocketForwarder)
	}
	return router
}
