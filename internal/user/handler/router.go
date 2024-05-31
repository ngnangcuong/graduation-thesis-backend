package handler

import (
	"graduation-thesis/pkg/middleware"

	"github.com/gin-gonic/gin"
)

var router *gin.Engine

func InitRouter(authHandler *AuthHandler, userHandler *UserHandler) {
	router = gin.Default()

	router.GET("/health", func(ctx *gin.Context) {
		ctx.JSON(200, "OK")
	})
	router.Use(middleware.Headers())
	router.Use(middleware.SetupCors())

	authPath := router.Group("/v1/auth")
	{
		authPath.POST("/login", authHandler.Login)
		authPath.POST("/refresh", authHandler.Refresh)
		authPath.GET("/logout", authHandler.AuthMiddleware(), authHandler.Logout)
	}

	userPath := router.Group("/v1/user")
	{
		userPath.POST("/verify", userHandler.Verify)
		userPath.GET("/:id", userHandler.GetUser)
		userPath.POST("/", userHandler.Register)
		userPath.PUT("/:id", userHandler.UpdateUser)
	}
}

func GetRouter(authHandler *AuthHandler, userHandler *UserHandler) *gin.Engine {
	if router == nil {
		InitRouter(authHandler, userHandler)
	}
	return router
}
