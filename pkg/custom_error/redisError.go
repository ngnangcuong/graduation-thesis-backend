package custom_error

import (
	"context"
	"errors"
	"os"

	"github.com/redis/go-redis/v9"
)

func HandleRedisError(err error) error {
	if err == nil {
		return nil
	}

	if errors.Is(err, redis.Nil) {
		return ErrNotFound
	}

	if errors.Is(err, os.ErrDeadlineExceeded) || errors.Is(err, context.DeadlineExceeded) {
		return ErrTimeout
	}

	return err
}
