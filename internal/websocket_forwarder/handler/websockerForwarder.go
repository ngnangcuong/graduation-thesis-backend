package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"graduation-thesis/internal/websocket_forwarder/model"
	"graduation-thesis/pkg/logger"
	responseModel "graduation-thesis/pkg/model"
	request "graduation-thesis/pkg/requests"
)

type WebsocketForwarder struct {
	websocketManagerUrl string
	errorMap            map[error]int
	timeout             time.Duration
	maxRetries          int
	retryInterval       time.Duration
	logger              logger.Logger
}

func NewWebsocketForwarder(
	websocketManagerUrl string,
	errorMap map[error]int,
	timeout time.Duration,
	maxRetries int,
	retryInterval time.Duration,
	logger logger.Logger) *WebsocketForwarder {
	return &WebsocketForwarder{
		websocketManagerUrl: websocketManagerUrl,
		errorMap:            errorMap,
		timeout:             timeout,
		maxRetries:          maxRetries,
		retryInterval:       retryInterval,
		logger:              logger,
	}
}

func (w *WebsocketForwarder) getListWebsocketHandlers() ([]model.WebsocketHandler, error) {
	var (
		result interface{}
		err    error
	)
	for i := 1; i <= w.maxRetries; i++ {
		result, err = request.HTTPRequestCall(
			fmt.Sprintf("%s/websocket_handler", w.websocketManagerUrl),
			http.MethodGet,
			"",
			nil,
			w.timeout,
		)
		if err != nil {
			time.Sleep(w.retryInterval)
			continue
		}
		break
	}

	if err != nil {
		return nil, err
	}

	listWebsocketHandlersJSON, _ := json.Marshal(result)
	var listWebsocketHandlers []model.WebsocketHandler
	json.Unmarshal(listWebsocketHandlersJSON, &listWebsocketHandlers)
	return listWebsocketHandlers, nil
}

func (w *WebsocketForwarder) HandleRequest(c *gin.Context) {
	websocketHandlers, err := w.getListWebsocketHandlers()
	if err != nil {
		errorMessage := responseModel.ErrorResponse{
			Status:       w.errorMap[err],
			ErrorMessage: err.Error(),
		}
		c.JSON(errorMessage.Status, errorMessage)
		return
	}

	websocketHandler := w.selectAWebsocketHandler(websocketHandlers)
	successResponse := responseModel.SuccessResponse{
		Status: http.StatusOK,
		Result: model.HandleRequestResponse{
			IPAddress: websocketHandler.IPAddress,
		},
	}
	c.JSON(successResponse.Status, successResponse)
}

func (w *WebsocketForwarder) selectAWebsocketHandler(websocketHandlers []model.WebsocketHandler) *model.WebsocketHandler {
	result := websocketHandlers[0]
	for _, websocketHandler := range websocketHandlers {
		if result.NumberClient > websocketHandler.NumberClient {
			result = websocketHandler
		}
	}
	return &result
}
