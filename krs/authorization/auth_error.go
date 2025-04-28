package authorization

import (
	"fmt"
)

const (
	errFormat = `{"code": %d, "message": "%s"}`

	errMessageUnauthorized = "Unauthorized"
)

type AuthorizationError struct {
	Code    int
	Message string
}

func (e *AuthorizationError) Error() string {
	return fmt.Sprintf(errFormat, e.Code, e.Message)
}

func NewAuthorizationError(format string, a ...any) *AuthorizationError {
	if format == "" {
		format = errMessageUnauthorized
	}
	return &AuthorizationError{
		Code:    401,
		Message: fmt.Sprintf(format, a...),
	}
}
