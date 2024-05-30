package handler

import (
	"graduation-thesis/internal/websocket_handler/worker"
	responseModel "graduation-thesis/pkg/model"

	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

type Handler struct {
	upgrader websocket.Upgrader
	worker   *worker.Worker
}

func NewHandler(worker *worker.Worker) *Handler {
	return &Handler{
		upgrader: websocket.Upgrader{},
		worker:   worker,
	}
}

func (h *Handler) EstablishConnetionWithPeer(c *gin.Context) {
	// websocketHandlerID := c.MustGet("websocket_handler_id").(string)
	websocketHandlerID := c.Request.Header.Get("X-Websocket-ID")
	conn, err := h.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		errorResponse := responseModel.ErrorResponse{
			Status:       http.StatusBadRequest,
			ErrorMessage: err.Error(),
		}
		c.JSON(errorResponse.Status, errorResponse)
		return
	}

	if err := h.worker.KeepPeersConnection(conn, websocketHandlerID); err != nil {
		errorResponse := responseModel.ErrorResponse{
			Status:       http.StatusBadRequest,
			ErrorMessage: err.Error(),
		}
		c.JSON(errorResponse.Status, errorResponse)
		return
	}
}

func (h *Handler) EstablishConnetionWithUser(c *gin.Context) {
	// userID := c.MustGet("user_id").(string)
	userID := c.Request.Header.Get("X-User-ID")
	conn, err := h.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		errorResponse := responseModel.ErrorResponse{
			Status:       http.StatusBadRequest,
			ErrorMessage: err.Error(),
		}
		c.JSON(errorResponse.Status, errorResponse)
		return
	}

	if err := h.worker.KeepUsersConnection(conn, userID); err != nil {
		errorResponse := responseModel.ErrorResponse{
			Status:       http.StatusBadRequest,
			ErrorMessage: err.Error(),
		}
		c.JSON(errorResponse.Status, errorResponse)
		return
	}
}
