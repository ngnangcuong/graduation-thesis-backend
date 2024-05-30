package repository

import (
	"context"
	"graduation-thesis/internal/websocket_manager/model"
	"graduation-thesis/pkg/custom_error"

	"github.com/redis/go-redis/v9"
)

const LISTWEBSOCKETKEY = "list_websocket"

type WebsocketManagerRepo struct {
	redis *redis.Client
}

func NewWebsocketManagerRepo(redis *redis.Client) *WebsocketManagerRepo {
	return &WebsocketManagerRepo{
		redis: redis,
	}
}

func (w *WebsocketManagerRepo) Get(ctx context.Context, websocketHandlerID string) ([]string, error) {
	users, err := w.redis.SMembers(ctx, websocketHandlerID).Result()
	if err != nil {
		return nil, custom_error.HandleRedisError(err)
	}

	return users, nil
}

func (w *WebsocketManagerRepo) GetWebsocketHandlers(ctx context.Context) (map[string]string, error) {
	result := w.redis.HGetAll(ctx, LISTWEBSOCKETKEY)
	if err := result.Err(); err != nil {
		return nil, custom_error.HandleRedisError(err)
	}
	return result.Val(), nil
}

func (w *WebsocketManagerRepo) GetAWebsocketHandler(ctx context.Context, websocketHandlerID string) (*model.WebsocketHandlerClient, error) {
	ipAddress, err := w.redis.HGet(ctx, LISTWEBSOCKETKEY, websocketHandlerID).Result()
	if err != nil {
		return nil, custom_error.HandleRedisError(err)
	}

	websocketHandler := model.WebsocketHandlerClient{
		ID:        websocketHandlerID,
		IPAddress: ipAddress,
	}
	return &websocketHandler, nil
}

func (w *WebsocketManagerRepo) GetNumberClient(ctx context.Context, websocketHandlerID string) (int, error) {
	result, err := w.redis.SCard(ctx, websocketHandlerID).Result()
	if err != nil {
		return 0, custom_error.HandleRedisError(err)
	}

	return int(result), nil
}

func (w *WebsocketManagerRepo) AddWebsocketHandler(ctx context.Context, websocketHandler model.WebsocketHandlerClient) error {
	mapIDToIP := make(map[string]string, 1)
	mapIDToIP[websocketHandler.ID] = websocketHandler.IPAddress
	err := w.redis.HSet(ctx, LISTWEBSOCKETKEY, mapIDToIP).Err()
	return custom_error.HandleRedisError(err)
}

func (w *WebsocketManagerRepo) RemoveWebSocketHandler(ctx context.Context, websocketHandlerID string) error {
	err := w.redis.HDel(ctx, LISTWEBSOCKETKEY, websocketHandlerID).Err()
	return custom_error.HandleRedisError(err)
}

func (w *WebsocketManagerRepo) Add(ctx context.Context, websocketHandlerID, userID string) error {
	err := w.redis.SAdd(ctx, websocketHandlerID, userID).Err()
	return custom_error.HandleRedisError(err)
}

func (w *WebsocketManagerRepo) Remove(ctx context.Context, websocketHandlerID, userID string) error {
	err := w.redis.SRem(ctx, websocketHandlerID, userID).Err()
	return custom_error.HandleRedisError(err)
}

func (w *WebsocketManagerRepo) Watch(ctx context.Context, f func(*redis.Tx) error, keys ...string) error {
	w.redis.Watch(ctx, f, keys...)
	return nil
}
