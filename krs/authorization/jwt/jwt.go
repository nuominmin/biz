package jwt

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/transport"
	"github.com/nuominmin/biz/krs/authorization"
	"github.com/spf13/cast"
)

const (
	contextKeyUserId = "user_id"
)

type Service interface {
	NewSecret() ([]byte, error)
	GenerateJWT(userId uint64, extra interface{}) (string, error)
	Middleware(ignoredPaths ...string) middleware.Middleware
	NewContextWithUserId(ctx context.Context, userId uint64) context.Context
	GetUserId(ctx context.Context) (uint64, error)
}
type service struct {
	secret []byte
}

func NewService(secret []byte) Service {
	return &service{
		secret: secret,
	}
}

func (s *service) NewSecret() ([]byte, error) {
	secret := make([]byte, 32)
	if _, err := rand.Read(secret); err != nil {
		return nil, fmt.Errorf("failed to generate jwt, error: %v", err)
	}
	return secret, nil
}

func (s *service) GenerateJWT(userId uint64, extra interface{}) (string, error) {
	now := time.Now()
	claims := jwt.MapClaims{
		contextKeyUserId: userId,
		"exp":            now.Add(time.Hour * 24 * 30).Unix(),
		"iat":            now.Unix(),
		"extra":          extra,
	}

	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(s.secret)
}

func (s *service) Middleware(ignoredPaths ...string) middleware.Middleware {
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

			token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, authorization.NewAuthorizationError(authorization.ErrInvalidToken)
				}
				return s.secret, nil
			})

			if err != nil || !token.Valid {
				return nil, authorization.NewAuthorizationError(authorization.ErrInvalidToken)
			}

			claims, ok := token.Claims.(jwt.MapClaims)
			if !ok {
				return nil, authorization.NewAuthorizationError(authorization.ErrInvalidToken)
			}

			return handler(s.NewContextWithUserId(ctx, cast.ToUint64(claims[contextKeyUserId])), req)
		}
	}
}

func (s *service) NewContextWithUserId(ctx context.Context, userId uint64) context.Context {
	return context.WithValue(ctx, contextKeyUserId, userId)
}

func (s *service) GetUserId(ctx context.Context) (uint64, error) {
	value := ctx.Value(contextKeyUserId)
	if userId, ok := value.(uint64); ok {
		return userId, nil
	}
	return 0, errors.New("failed to get user Id from context")
}
