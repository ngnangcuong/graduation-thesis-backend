package http_handler

import "github.com/gin-gonic/gin"

var router *gin.Engine

func NewRouter(userHandler *UserHandler, websocketHandler *WebsocketHandler) *gin.Engine {
	r := gin.Default()
	userPath := r.Group("/v1/user")
	{
		userPath.GET("/:user_id", userHandler.GetWebsocketHandler)
	}

	websocketHandlerPath := r.Group("/v1/websocket_handler")
	{
		websocketHandlerPath.POST("/ping", websocketHandler.Ping)
		websocketHandlerPath.GET("/:id", websocketHandler.GetUserList)
		websocketHandlerPath.POST("/register", websocketHandler.AddNewWebsocketHandler)
		websocketHandlerPath.GET("", websocketHandler.GetWebsocketHandlerList)
		websocketHandlerPath.POST("/user", websocketHandler.AddNewUser)
		websocketHandlerPath.DELETE("/user", websocketHandler.DisconnectUser)
	}
	return r
}

func GetRouter(userHandler *UserHandler, websocketHandler *WebsocketHandler) *gin.Engine {
	if router == nil {
		router = NewRouter(userHandler, websocketHandler)
	}
	return router
}
