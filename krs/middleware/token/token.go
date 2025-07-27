package token

import (
	"context"
	"errors"
	"github.com/nuominmin/biz/krs/middleware/constant"
	"github.com/nuominmin/biz/krs/middleware/errresp"
	"strings"

	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/transport"
	"github.com/google/uuid"
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
		contextKey: constant.DefaultTokenContextKey,
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

				authHeader := tr.RequestHeader().Get(constant.HeaderAuthorizationKey)
				if authHeader == "" {
					return nil, errresp.NewAuthorizationError(constant.ErrMissingToken)
				}

				parts := strings.SplitN(authHeader, " ", 2)
				if len(parts) != 2 || parts[0] != constant.AuthorizationValueBearer {
					return nil, errresp.NewAuthorizationError(constant.ErrInvalidToken)
				}

				tokenString = parts[1]
			} else {
				return nil, errresp.NewAuthorizationError(constant.ErrMissingToken)
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
	return "", errors.New("failed to get token from context")
}
