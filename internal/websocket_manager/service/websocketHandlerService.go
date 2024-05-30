package service

import (
	"context"
	"errors"
	"graduation-thesis/internal/websocket_manager/model"
	"graduation-thesis/internal/websocket_manager/repository"
	"graduation-thesis/pkg/custom_error"
	"graduation-thesis/pkg/logger"
	responseModel "graduation-thesis/pkg/model"
	"hash/fnv"
	"net/http"
	"sync"
	"time"
)

type WebsocketManagerService struct {
	websocketManagerRepo *repository.WebsocketManagerRepo
	userRepo             *repository.UserRepo
	errorMap             map[error]int
	mu                   []*sync.Mutex
	numMu                int
	websocketHandlers    *model.MapWebsocketHandlerMonitoring
	wg                   sync.WaitGroup
	heartbeatInterval    time.Duration
	maxRetries           int
	retryInterval        time.Duration
	logger               logger.Logger
	close                chan struct{}
}

func NewWebsocketManagerService(
	websocketManagerRepo *repository.WebsocketManagerRepo,
	userRepo *repository.UserRepo,
	errorMap map[error]int,
	numMu int,
	heartbeatInterval time.Duration,
	maxRetries int,
	retryInterval time.Duration,
	logger logger.Logger) *WebsocketManagerService {
	w := WebsocketManagerService{
		websocketManagerRepo: websocketManagerRepo,
		userRepo:             userRepo,
		errorMap:             errorMap,
		mu:                   make([]*sync.Mutex, numMu),
		numMu:                numMu,
		websocketHandlers:    model.NewMapWebsocketHandlerMonitoring(),
		heartbeatInterval:    heartbeatInterval,
		maxRetries:           maxRetries,
		retryInterval:        retryInterval,
		logger:               logger,
		close:                make(chan struct{}),
	}
	for i := 0; i < numMu; i++ {
		w.mu[i] = &sync.Mutex{}
	}

	return &w
}

func (w *WebsocketManagerService) MonitorWebsocketHandler() {
	w.logger.Info("[Monitoring] Start Monitoring Websocket Handler")
	defer w.logger.Info("[Monitoring] Shutting down Monitoring Websocket Handler")
	ctx := context.Background()
	websocketHandlers, err := w.websocketManagerRepo.GetWebsocketHandlers(ctx)
	if err != nil {
		w.logger.Errorf("[Monitoring] Failed to get list of websocket handlers: %v", err)
		return
	}

	w.logger.Info("[Monitoring] List of Websocket Handlers: %v", websocketHandlers)
	for ID, IPAddress := range websocketHandlers {
		w.wg.Add(1)
		websocketHandlerMonitoring := model.WebsocketHandlerMonitoring{
			ID:        ID,
			IPAddress: IPAddress,
			Hearbeat:  make(chan struct{}, 1),
		}
		w.websocketHandlers.Set(ID, websocketHandlerMonitoring)
		go w.checkHearbeat(&websocketHandlerMonitoring)
	}
}

func (w *WebsocketManagerService) checkHearbeat(monitoring *model.WebsocketHandlerMonitoring) {
	w.logger.Infof("[%s] Start monitoring websocket handler %s", monitoring.ID, monitoring.ID)
	defer w.wg.Done()
	timer := time.NewTimer(w.heartbeatInterval)

	for {
		select {
		case <-timer.C:
			w.logger.Infof("[%s] Lost heartbeat from websocket handler %s", monitoring.ID, monitoring.ID)
			timer.Stop()
			// Comment this line temporarily to avoid race condition
			// close(monitoring.Hearbeat)
			w.leaveGroup(monitoring.ID)
			w.logger.Infof("[%s] Canceling monitoring websocket handler %s", monitoring.ID, monitoring.ID)
			return
		case _, ok := <-monitoring.Hearbeat:
			if !ok {
				return
			}
			timer.Reset(w.heartbeatInterval)
		case <-w.close:
			w.logger.Infof("[%s] Shutting down monitoring websocket handler %s", monitoring.ID, monitoring.ID)
			timer.Stop()
			return
		}
	}
}

func (w *WebsocketManagerService) leaveGroup(websocketHandlerID string) {
	ctx := context.Background()
	var (
		users []string
		err   error
	)
	for i := 1; i <= w.maxRetries; i++ {
		users, err = w.websocketManagerRepo.Get(ctx, websocketHandlerID)
		if err != nil {
			w.logger.Errorf("[%s] Failed to get information's websocket handler %s for %dth time: %v", websocketHandlerID, websocketHandlerID, i, err)
			time.Sleep(w.retryInterval)
			continue
		}
		break
	}

	for _, user := range users {
		for i := 1; i <= w.maxRetries; i++ {
			removeUserRequest := model.AddNewUserRequest{
				WebsocketID: websocketHandlerID,
				UserID:      user,
			}
			_, errResp := w.RemoveUser(ctx, &removeUserRequest)
			if errResp != nil {
				w.logger.Errorf("[%s] Failed to remove user %s from websocket handler %s's list", websocketHandlerID, user, websocketHandlerID)
				time.Sleep(w.retryInterval)
				continue
			}
			break
		}
	}

	for i := 1; i <= w.maxRetries; i++ {
		if err := w.websocketManagerRepo.RemoveWebSocketHandler(ctx, websocketHandlerID); err != nil {
			w.logger.Errorf("[%s] Failed to remove websocket handler %s from redis for %dth time: %v", websocketHandlerID, websocketHandlerID, i, err)
			time.Sleep(w.retryInterval)
			continue
		}
		break
	}

	w.websocketHandlers.Del(websocketHandlerID)
}

func (w *WebsocketManagerService) Pong(ctx context.Context, request *model.PingRequest) (*responseModel.SuccessResponse, *responseModel.ErrorResponse) {
	monitoring := w.websocketHandlers.Get(request.ID)
	if monitoring == nil {
		errorResponse := responseModel.ErrorResponse{
			Status:       w.errorMap[custom_error.ErrNotFound],
			ErrorMessage: custom_error.ErrNotFound.Error(),
		}
		return nil, &errorResponse
	}
	go func(monitoring *model.WebsocketHandlerMonitoring) {
		monitoring.Hearbeat <- struct{}{}
	}(monitoring)

	successResponse := responseModel.SuccessResponse{
		Status: http.StatusOK,
	}
	return &successResponse, nil
}

func (w *WebsocketManagerService) GetUsers(ctx context.Context, websocketHandlerID string) (*responseModel.SuccessResponse, *responseModel.ErrorResponse) {
	users, err := w.websocketManagerRepo.Get(ctx, websocketHandlerID)
	if err != nil {
		errorResponse := responseModel.ErrorResponse{
			Status:       w.errorMap[err],
			ErrorMessage: err.Error(),
		}
		return nil, &errorResponse
	}

	successResponse := responseModel.SuccessResponse{
		Status: http.StatusOK,
		Result: users,
	}

	return &successResponse, nil
}

func (w *WebsocketManagerService) GetWebsocketHandlers(ctx context.Context) (*responseModel.SuccessResponse, *responseModel.ErrorResponse) {
	mapIDToIP, err := w.websocketManagerRepo.GetWebsocketHandlers(ctx)
	if err != nil {
		errorResponse := responseModel.ErrorResponse{
			Status:       w.errorMap[err],
			ErrorMessage: err.Error(),
		}
		return nil, &errorResponse
	}

	websocketHandlers := make([]model.WebsocketHandlerClient, len(mapIDToIP))
	for ID, IPAddress := range mapIDToIP {
		numberClient, err := w.websocketManagerRepo.GetNumberClient(ctx, ID)
		if err != nil {
			errorResponse := responseModel.ErrorResponse{
				Status:       w.errorMap[err],
				ErrorMessage: err.Error(),
			}
			return nil, &errorResponse
		}
		websocketHandlers = append(websocketHandlers, model.WebsocketHandlerClient{
			ID:           ID,
			IPAddress:    IPAddress,
			NumberClient: numberClient,
		})
	}

	successResponse := responseModel.SuccessResponse{
		Status: http.StatusOK,
		Result: websocketHandlers,
	}
	return &successResponse, nil
}

func (w *WebsocketManagerService) AddNewWebsocketHandler(ctx context.Context, request *model.AddNewWebsocketHandlerRequest) (*responseModel.SuccessResponse, *responseModel.ErrorResponse) {
	websocketHandlerClient := model.WebsocketHandlerClient{
		ID:        request.ID,
		IPAddress: request.IPAddress,
	}

	err := w.websocketManagerRepo.AddWebsocketHandler(ctx, websocketHandlerClient)
	if err != nil {
		errorResponse := responseModel.ErrorResponse{
			Status:       w.errorMap[err],
			ErrorMessage: err.Error(),
		}
		return nil, &errorResponse
	}

	oldWebsocketHandlerMonitoring := w.websocketHandlers.Get(websocketHandlerClient.ID)
	if oldWebsocketHandlerMonitoring != nil {
		close(oldWebsocketHandlerMonitoring.Hearbeat)
	}

	websocketHandlerMonitoring := model.WebsocketHandlerMonitoring{
		ID:        websocketHandlerClient.ID,
		IPAddress: websocketHandlerClient.IPAddress,
		Hearbeat:  make(chan struct{}, 1),
	}
	w.websocketHandlers.Set(websocketHandlerClient.ID, websocketHandlerMonitoring)
	w.wg.Add(1)
	go w.checkHearbeat(&websocketHandlerMonitoring)

	successResponse := responseModel.SuccessResponse{
		Status: http.StatusCreated,
		Result: websocketHandlerClient,
	}
	return &successResponse, nil
}

func (w *WebsocketManagerService) AddNewUser(ctx context.Context, request *model.AddNewUserRequest) (*responseModel.SuccessResponse, *responseModel.ErrorResponse) {
	w.mu[hash(request.UserID)%uint32(w.numMu)].Lock()
	defer w.mu[hash(request.UserID)%uint32(w.numMu)].Unlock()

	websocketHandlerRequest, err := w.websocketManagerRepo.GetAWebsocketHandler(ctx, request.WebsocketID)
	if err != nil {
		errorResponse := responseModel.ErrorResponse{
			Status:       w.errorMap[err],
			ErrorMessage: err.Error(),
		}
		return nil, &errorResponse
	}

	currentWebsocketHandler, cErr := w.userRepo.Get(ctx, request.UserID)
	if cErr != nil && !errors.Is(cErr, custom_error.ErrNotFound) { // TODO: Handle more things when error is ErrNotFound
		errorResponse := responseModel.ErrorResponse{
			Status:       w.errorMap[cErr],
			ErrorMessage: cErr.Error(),
		}
		return nil, &errorResponse
	}

	if currentWebsocketHandler != nil && currentWebsocketHandler.ID != websocketHandlerRequest.ID {
		if err := w.websocketManagerRepo.Remove(ctx, currentWebsocketHandler.ID, request.UserID); err != nil {
			errorResponse := responseModel.ErrorResponse{
				Status:       w.errorMap[cErr],
				ErrorMessage: cErr.Error(),
			}
			return nil, &errorResponse
		}
	}

	if err := w.userRepo.Set(ctx, request.UserID, *websocketHandlerRequest); err != nil {
		errorResponse := responseModel.ErrorResponse{
			Status:       w.errorMap[err],
			ErrorMessage: err.Error(),
		}
		return nil, &errorResponse
	}

	if err := w.websocketManagerRepo.Add(ctx, request.WebsocketID, request.UserID); err != nil {
		_ = w.userRepo.Del(ctx, []string{request.UserID})
		errorResponse := responseModel.ErrorResponse{
			Status:       w.errorMap[err],
			ErrorMessage: err.Error(),
		}
		return nil, &errorResponse
	}

	successResponse := responseModel.SuccessResponse{
		Status: http.StatusNoContent,
	}
	return &successResponse, nil
}

func (w *WebsocketManagerService) RemoveUser(ctx context.Context, request *model.AddNewUserRequest) (*responseModel.SuccessResponse, *responseModel.ErrorResponse) {
	w.mu[hash(request.UserID)%uint32(w.numMu)].Lock()
	defer w.mu[hash(request.UserID)%uint32(w.numMu)].Unlock()

	websocketHandler, err := w.userRepo.Get(ctx, request.UserID)
	if err != nil && !errors.Is(err, custom_error.ErrNotFound) { // TODO: Handle more things when error is ErrNotFound
		errorResponse := responseModel.ErrorResponse{
			Status:       w.errorMap[err],
			ErrorMessage: err.Error(),
		}
		return nil, &errorResponse
	}

	if err := w.websocketManagerRepo.Remove(ctx, request.WebsocketID, request.UserID); err != nil {
		errorResponse := responseModel.ErrorResponse{
			Status:       w.errorMap[err],
			ErrorMessage: err.Error(),
		}
		return nil, &errorResponse
	}

	if websocketHandler != nil && websocketHandler.ID == request.WebsocketID {
		if err := w.userRepo.Del(ctx, []string{request.UserID}); err != nil {
			errorResponse := responseModel.ErrorResponse{
				Status:       w.errorMap[err],
				ErrorMessage: err.Error(),
			}
			return nil, &errorResponse
		}
	}

	successResponse := responseModel.SuccessResponse{
		Status: http.StatusNoContent,
	}
	return &successResponse, nil
}

func hash(s string) uint32 {
	h := fnv.New32a()
	h.Write([]byte(s))
	return h.Sum32()
}
