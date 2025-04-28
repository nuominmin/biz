package token

import (
	"context"
	"errors"
	"strings"

	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/transport"
	"github.com/google/uuid"
	"github.com/nuominmin/biz/krs/authorization"
)

const (
	// default context key for token
	defaultContextKey string = "token"
)

type Service interface {
	GenerateToken() string
	Middleware(ignoredPaths []string, m ...middleware.Middleware) middleware.Middleware
	GetToken(ctx context.Context) (string, error)
}

type service struct {
	contextKey string
}

func NewService() Service {
	return &service{
		contextKey: defaultContextKey,
	}
}

// NewServiceWithContextKey creates a new token service with a custom context key.
func NewServiceWithContextKey(contextKey string) Service {
	return &service{
		contextKey: contextKey,
	}
}

func (s *service) GenerateToken() string {
	return strings.ReplaceAll(uuid.New().String(), "-", "")
}

func (s *service) Middleware(ignoredPaths []string, m ...middleware.Middleware) middleware.Middleware {
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req interface{}) (interface{}, error) {
			var tokenString string

			if tr, ok := transport.FromServerContext(ctx); ok {
				operation := tr.Operation()

				// check if the request path is in the ignore list
				for i := 0; i < len(ignoredPaths); i++ {
					if operation == ignoredPaths[i] {
						// ignore this path, call the next handler
						return handler(ctx, req)
					}
				}

				authHeader := tr.RequestHeader().Get(authorization.HeaderAuthorizationKey)
				if authHeader == "" {
					return nil, authorization.NewAuthorizationError(authorization.ErrMissingToken)
				}

				parts := strings.SplitN(authHeader, " ", 2)
				if len(parts) != 2 || parts[0] != authorization.AuthorizationValueBearer {
					return nil, authorization.NewAuthorizationError(authorization.ErrInvalidToken)
				}

				tokenString = parts[1]
			} else {
				return nil, authorization.NewAuthorizationError(authorization.ErrMissingToken)
			}

			// 将 token 信息传递给 handler
			ctx = s.newContextWithToken(ctx, tokenString)

			for i := 0; i < len(m); i++ {
				handler = m[i](handler) // 链式调用中间件
			}

			return handler(ctx, req)
		}
	}
}

func (s *service) newContextWithToken(ctx context.Context, token string) context.Context {
	return context.WithValue(ctx, s.contextKey, token)
}

func (s *service) GetToken(ctx context.Context) (string, error) {
	if token, ok := ctx.Value(s.contextKey).(string); ok {
		return token, nil
	}
	return "", errors.New("failed to token from context")
}
