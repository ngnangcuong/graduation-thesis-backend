package handler

import (
	"graduation-thesis/internal/websocket_handler/worker"
	responseModel "graduation-thesis/pkg/model"

	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

type Handler struct {
	upgrader         websocket.Upgrader
	worker           *worker.Worker
	authenticatorURL string
}

func NewHandler(worker *worker.Worker, authenticatorURL string) *Handler {
	upgrader := websocket.Upgrader{}
	upgrader.CheckOrigin = func(r *http.Request) bool { return true }
	return &Handler{
		upgrader:         upgrader,
		worker:           worker,
		authenticatorURL: authenticatorURL,
	}
}

func (h *Handler) EstablishConnetionWithPeer(c *gin.Context) {
	// websocketHandlerID := c.MustGet("websocket_handler_id").(string)
	websocketHandlerID := c.Query("websocket_id")
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
	userID := c.Query("user_id")
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
