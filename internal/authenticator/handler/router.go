package handler

import (
	"graduation-thesis/pkg/middleware"

	"github.com/gin-gonic/gin"
)

var router *gin.Engine

func NewRouter(authHandler *AuthHandler) *gin.Engine {
	r := gin.Default()
	r.Use(middleware.Headers())
	r.Use(middleware.SetupCors())

	absolutePath := r.Group("/v1/")
	{
		absolutePath.POST("/refresh", authHandler.Refresh)
		absolutePath.POST("/validate", authHandler.Validate)
		absolutePath.POST("/login", authHandler.Login)
		absolutePath.GET("/logout", authHandler.Logout)
		absolutePath.POST("/reset-password")
	}

	return r
}

func GetRouter(authHandler *AuthHandler) *gin.Engine {
	if router == nil {
		router = NewRouter(authHandler)
	}

	return router
}
