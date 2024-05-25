package custom_error

import (
	"errors"
	"net/http"
)

var (
	ErrTimeout             = errors.New("time out")
	ErrNotFound            = errors.New("not found")
	ErrNoPermission        = errors.New("no permission")
	ErrConnectionErr       = errors.New("connection error")
	ErrInvalidParameter    = errors.New("invalid parameter")
	ErrInternalServerError = errors.New("internal server error")
	ErrEntityTooLarge      = errors.New("entity too large")
	ErrConflict            = errors.New("data conflict")
	ErrChannelHasClosed    = errors.New("channel has closed")
	ErrUnknown             = errors.New("unknown error")
)

func MappingError() map[error]int {
	result := make(map[error]int)
	result[ErrTimeout] = http.StatusGatewayTimeout
	result[ErrNotFound] = http.StatusNotFound
	result[ErrNoPermission] = http.StatusUnauthorized
	result[ErrConnectionErr] = http.StatusServiceUnavailable
	result[ErrInvalidParameter] = http.StatusUnprocessableEntity
	result[ErrInternalServerError] = http.StatusInternalServerError
	result[ErrConflict] = http.StatusConflict
	result[ErrEntityTooLarge] = http.StatusRequestEntityTooLarge
	result[ErrUnknown] = http.StatusInternalServerError
	return result
}

func MappingStatusError() map[int]error {
	result := make(map[int]error)
	result[http.StatusGatewayTimeout] = ErrTimeout
	result[http.StatusNotFound] = ErrNotFound
	result[http.StatusUnauthorized] = ErrNoPermission
	result[http.StatusServiceUnavailable] = ErrConnectionErr
	result[http.StatusUnprocessableEntity] = ErrInvalidParameter
	result[http.StatusInternalServerError] = ErrInternalServerError
	result[http.StatusConflict] = ErrConflict
	result[http.StatusRequestEntityTooLarge] = ErrEntityTooLarge
	return result
}
