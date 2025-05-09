package apperrors

import (
	"fmt"
	"net/http"
)

type Type string

const (
	BadRequest         Type = "BAD_REQUEST"
	NotFound           Type = "NOT_FOUND"
	Conflict           Type = "CONFLICT"
	Unauthorized       Type = "UNAUTHORIZED"
	Forbidden          Type = "FORBIDDEN"
	InternalServer     Type = "INTERNAL_SERVER"
	ServiceUnavailable Type = "SERVICE_UNAVAILABLE"
)

type Error struct {
	Type    Type   `json:"type"`
	Message string `json:"message"`
}

func (e Error) Error() string {
	return e.Message
}

func (e Error) StatusCode() int {
	switch e.Type {
	case BadRequest:
		return http.StatusBadRequest
	case NotFound:
		return http.StatusNotFound
	case Conflict:
		return http.StatusConflict
	case Unauthorized:
		return http.StatusUnauthorized
	case Forbidden:
		return http.StatusForbidden
	case ServiceUnavailable:
		return http.StatusServiceUnavailable
	default:
		return http.StatusInternalServerError
	}
}

func New(errType Type, message string) Error {
	return Error{
		Type:    errType,
		Message: message,
	}
}

func NewWithFormat(errType Type, format string, args ...interface{}) Error {
	return Error{
		Type:    errType,
		Message: fmt.Sprintf(format, args...),
	}
}

func Wrap(err error, errType Type) Error {
	return Error{
		Type:    errType,
		Message: err.Error(),
	}
}
