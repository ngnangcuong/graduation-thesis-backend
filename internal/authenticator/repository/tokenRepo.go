package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type TokenRepo struct {
	redis *redis.Client
}

func NewTokenRepo(redisClient *redis.Client) *TokenRepo {
	return &TokenRepo{
		redis: redisClient,
	}
}

func (t *TokenRepo) StoreToken(userId string, tokenUuid string, expired time.Time) error {
	err := t.redis.Set(context.Background(), tokenUuid, userId, time.Until(expired))
	if err.Err() != nil {
		return err.Err()
	}

	return nil
}

func (t *TokenRepo) FetchUser(tokenUuid string) (string, error) {
	userId, err := t.redis.Get(context.Background(), tokenUuid).Result()
	if err != nil {
		return "", err
	}

	return userId, nil
}

func (t *TokenRepo) DeleteToken(tokenUuid string) (int64, error) {
	deleted, err := t.redis.Del(context.Background(), tokenUuid).Result()
	if err != nil {
		return 0, err
	}

	return deleted, nil
}

func (t *TokenRepo) DeleteAllToken(userId string) error {
	length, err := t.redis.SCard(context.Background(), userId).Result()
	if err != nil {
		return err
	}

	for i := 0; i < int(length); i++ {
		if err := t.redis.SPop(context.Background(), userId).Err(); err != nil {
			return err
		}
	}

	return nil
}

func (t *TokenRepo) AddToken(userId string, tokenString string) error {
	added, err := t.redis.SAdd(context.Background(), userId, tokenString).Result()
	if err != nil {
		return err
	}

	if added == 0 {
		return fmt.Errorf("add token error")
	}

	return nil
}

func (t *TokenRepo) IsForgotTokenOf(forgotToken string, userId string) (bool, error) {
	return t.redis.SIsMember(context.Background(), userId, forgotToken).Result()
}
