package user

import (
	"context"
	"database/sql"
	"graduation-thesis/internal/user/model"
	"time"

	redis "github.com/redis/go-redis/v9"
)

type UserRepoRedis struct {
	redis *redis.Client
}

func NewUserRepoRedis(redisClient *redis.Client) *UserRepoRedis {
	return &UserRepoRedis{
		redis: redisClient,
	}
}

func (u *UserRepoRedis) WithTx(tx *sql.Tx) IUserRepo {
	return nil
}

func (u *UserRepoRedis) Get(ctx context.Context, userId string) (*model.User, error) {
	var user model.User
	err := u.redis.Get(ctx, userId).Scan(&user)

	return &user, err
}

func (u *UserRepoRedis) Delete(ctx context.Context, userId string) error {
	_, err := u.redis.Del(ctx, userId).Result()
	if err != nil {
		return err
	}

	return nil
}

func (u *UserRepoRedis) Create(ctx context.Context, user *model.User) error {
	err := u.redis.Set(ctx, user.ID, user, 10*time.Minute)
	return err.Err()
}
