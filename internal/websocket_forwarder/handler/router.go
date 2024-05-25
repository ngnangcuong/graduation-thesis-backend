package handler

import "github.com/gin-gonic/gin"

var router *gin.Engine

func NewRouter(websocketForwarder *WebsocketForwarder) *gin.Engine {
	r := gin.Default()
	r.POST("/ws", websocketForwarder.HandleRequest)
	return r
}

func GetRouter(websocketForwarder *WebsocketForwarder) *gin.Engine {
	if router == nil {
		router = NewRouter(websocketForwarder)
	}
	return router
}
