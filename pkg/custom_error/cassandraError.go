package custom_error

import (
	"context"
	"errors"
	"os"

	"github.com/gocql/gocql"
)

func HandleCassandraError(err error) error {
	if errors.Is(err, gocql.ErrNotFound) {
		return ErrNotFound
	}

	if errors.Is(err, os.ErrDeadlineExceeded) ||
		errors.Is(err, context.DeadlineExceeded) ||
		errors.Is(err, gocql.ErrTimeoutNoResponse) {
		return ErrTimeout
	}

	if errors.Is(err, gocql.ErrFrameTooBig) {
		return ErrEntityTooLarge
	}

	if errors.Is(err, gocql.ErrKeyspaceDoesNotExist) || errors.Is(err, gocql.ErrNoKeyspace) || errors.Is(err, gocql.ErrNoHosts) {
		return ErrConflict
	}

	if errors.Is(err, gocql.ErrConnectionClosed) ||
		errors.Is(err, gocql.ErrCannotFindHost) ||
		errors.Is(err, gocql.ErrNoConnections) ||
		errors.Is(err, gocql.ErrUnavailable) {
		return ErrConnectionErr
	}
	return err
}
