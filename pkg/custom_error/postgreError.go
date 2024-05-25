package custom_error

import (
	"context"
	"database/sql"
	"errors"
	"os"

	"github.com/lib/pq"
)

func HandlePostgreError(err error) error {
	if err == nil {
		return nil
	}

	pgErr, ok := err.(*pq.Error)
	if ok {
		switch pgErr.Code.Class() {
		case "XX":
			return ErrInternalServerError
		case "23":
			return ErrConflict
		case "22":
			return ErrInvalidParameter
		case "08":
			return ErrConnectionErr
		case "38":
			return ErrNoPermission
		default:
			return ErrUnknown
		}
	}

	if errors.Is(err, sql.ErrNoRows) {
		return ErrNotFound
	}

	if errors.Is(err, os.ErrDeadlineExceeded) || errors.Is(err, context.DeadlineExceeded) {
		return ErrTimeout
	}

	return ErrUnknown
}
