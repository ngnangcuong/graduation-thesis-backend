package custom_error

import (
	"errors"

	"github.com/redis/go-redis/v9"
)

func HandleRedisError(err error) error {
	if err == nil {
		return nil
	}

	if errors.Is(err, redis.Nil) {
		return ErrNotFound
	}

	return err
}
