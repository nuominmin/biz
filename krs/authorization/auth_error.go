package authorization

import (
	"fmt"
)

const (
	errFormat = `{"code": %d, "message": "%s"}`

	errMessageUnauthorized = "Unauthorized"
)

type Error struct {
	Code    int
	Message string
}

func (e *Error) Error() string {
	return fmt.Sprintf(errFormat, e.Code, e.Message)
}

func NewAuthorizationError(format string, a ...any) *Error {
	if format == "" {
		format = errMessageUnauthorized
	}
	return &Error{
		Code:    401,
		Message: fmt.Sprintf(format, a...),
	}
}
