package repository

import (
	"context"
	"graduation-thesis/internal/websocket_manager/model"
	"graduation-thesis/pkg/custom_error"

	"github.com/redis/go-redis/v9"
)

type UserRepo struct {
	redis *redis.Client
}

func NewUserRepo(redis *redis.Client) *UserRepo {
	return &UserRepo{
		redis: redis,
	}
}

func (u *UserRepo) Get(ctx context.Context, userID string) (*model.WebsocketHandlerClient, error) {
	var websocketHandlerClient model.WebsocketHandlerClient
	err := u.redis.HGetAll(ctx, userID).Scan(&websocketHandlerClient)
	if err != nil {
		return nil, custom_error.HandleRedisError(err)
	}

	return &websocketHandlerClient, nil
}

func (u *UserRepo) Set(ctx context.Context, userID string, websocketHandler model.WebsocketHandlerClient) error {
	err := u.redis.HSet(ctx, userID, websocketHandler).Err()
	return custom_error.HandleRedisError(err)
}

func (u *UserRepo) Del(ctx context.Context, usersID []string) error {
	err := u.redis.Del(ctx, usersID...).Err()
	return custom_error.HandleRedisError(err)
}
