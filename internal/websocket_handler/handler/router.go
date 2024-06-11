package handler

import (
	"graduation-thesis/pkg/middleware"

	"github.com/gin-gonic/gin"
)

var router *gin.Engine

func NewRouter(handler *Handler) *gin.Engine {
	r := gin.Default()
	r.Use(middleware.Headers())
	r.Use(middleware.SetupCors())

	r.GET("/user/ws", handler.EstablishConnetionWithUser)
	r.GET("/peer/ws", handler.EstablishConnetionWithPeer)
	return r
}

func GetRouter(handler *Handler) *gin.Engine {
	if router == nil {
		router = NewRouter(handler)
	}
	return router
}
