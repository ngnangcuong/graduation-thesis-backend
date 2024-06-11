package handler

import (
	"errors"
	"graduation-thesis/internal/websocket_handler/worker"
	"graduation-thesis/pkg/custom_error"
	responseModel "graduation-thesis/pkg/model"
	request "graduation-thesis/pkg/requests"
	"time"

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
	authToken := c.Query("user_id")
	userID, vErr := h.validateToken(authToken)
	if vErr != nil {
		errorResponse := responseModel.ErrorResponse{
			Status:       http.StatusBadRequest,
			ErrorMessage: vErr.Error(),
		}
		c.JSON(errorResponse.Status, errorResponse)
		return
	}
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

func (h *Handler) validateToken(authToken string) (string, error) {
	var (
		result interface{}
		err    error
	)
	for i := 1; i <= 5; i++ {
		result, err = request.HTTPRequestCall(
			h.authenticatorURL,
			http.MethodPost,
			authToken,
			nil,
			5*time.Second,
		)
		if err != nil && !errors.Is(err, custom_error.ErrNoPermission) {
			continue
		}
		break
	}

	if err != nil {
		return "", err
	}

	userID := result.(string)
	return userID, nil
}
