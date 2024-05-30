package custom_error

import (
	"errors"
	"net/http"
)

type ErrorTimeout struct {
	Description string
}

func (e *ErrorTimeout) Error() string {
	return e.Description
}

type ErrorNotFound struct {
	Description string
}

func (e *ErrorNotFound) Error() string {
	return e.Description
}

type ErrorNoPermission struct {
	Description string
}

func (e *ErrorNoPermission) Error() string {
	return e.Description
}

type ErrorConnection struct {
	Description string
}

func (e *ErrorConnection) Error() string {
	return e.Description
}

type ErrorInvalidParameter struct {
	Description string
}

func (e *ErrorInvalidParameter) Error() string {
	return e.Description
}

type ErrorInternalServerError struct {
	Description string
}

func (e *ErrorInternalServerError) Error() string {
	return e.Description
}

type ErrorConflict struct {
	Description string
}

func (e *ErrorConflict) Error() string {
	return e.Description
}

type ErrorUnknown struct {
	Description string
}

func (e *ErrorUnknown) Error() string {
	return e.Description
}

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
