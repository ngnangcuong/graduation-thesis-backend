package handler

import (
	"graduation-thesis/pkg/middleware"

	"github.com/gin-gonic/gin"
)

var router *gin.Engine

func NewRouter(assetHandler *AssetHandler) *gin.Engine {
	r := gin.Default()
	r.Use(middleware.Headers())
	r.Use(middleware.SetupCors())
	r.MaxMultipartMemory = 32 << 20 // Default

	absolutePath := r.Group("/v1")
	{
		absolutePath.POST("/", assetHandler.Upload)       // Upload
		absolutePath.GET("/:fid", assetHandler.Get)       // Get
		absolutePath.DELETE("/:fid", assetHandler.Delete) // Delete
	}
	return r
}

func GetRouter(assetHandler *AssetHandler) *gin.Engine {
	if router == nil {
		router = NewRouter(assetHandler)
	}

	return router
}
