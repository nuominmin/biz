package jwt

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/nuominmin/biz/krs/middleware/constant"
	"github.com/nuominmin/biz/krs/middleware/errresp"

	"github.com/golang-jwt/jwt/v5"

	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/transport"
	"github.com/spf13/cast"
)

// Config holds the configuration for JWT service
type Config struct {
	ContextKey   string        // Context key for storing user ID
	ExpiryTime   time.Duration // Token expiry time
	RefreshTime  time.Duration // Token refresh time (optional)
	Issuer       string        // Token issuer
	SecretLength int           // Secret key length for generation
}

// DefaultConfig returns default configuration
func DefaultConfig() *Config {
	return &Config{
		ContextKey:   constant.DefaultJwtContextKey,
		ExpiryTime:   time.Hour * 24 * 30, // 30 days
		RefreshTime:  time.Hour * 24 * 7,  // 7 days for refresh
		Issuer:       "biz-service",
		SecretLength: 32,
	}
}

type Service interface {
	NewSecret() ([]byte, error)
	GenerateJWT(userId uint64, extra any) (string, error)
	GenerateJWTWithExpiry(userId uint64, extra any, expiry time.Duration) (string, error)
	Middleware(ignoredPaths ...string) middleware.Middleware
	NewContextWithUserId(ctx context.Context, userId uint64) context.Context
	GetUserId(ctx context.Context) (uint64, error)
	GetConfig() *Config
}

type service struct {
	config *Config
	secret []byte
}

func NewService(secret []byte) Service {
	return &service{
		config: DefaultConfig(),
		secret: secret,
	}
}

func NewServiceWithConfig(secret []byte, config *Config) Service {
	if config == nil {
		config = DefaultConfig()
	}
	return &service{
		config: config,
		secret: secret,
	}
}

// NewServiceWithContextKey creates a new JWT service with a custom context key.
// Deprecated: Use NewServiceWithConfig instead
func NewServiceWithContextKey(secret []byte, contextKey string) Service {
	config := DefaultConfig()
	config.ContextKey = contextKey
	return &service{
		config: config,
		secret: secret,
	}
}

func (s *service) GetConfig() *Config {
	return s.config
}

func (s *service) NewSecret() ([]byte, error) {
	secret := make([]byte, s.config.SecretLength)
	if _, err := rand.Read(secret); err != nil {
		return nil, fmt.Errorf("failed to generate jwt secret, error: %v", err)
	}
	return secret, nil
}

func (s *service) GenerateJWT(userId uint64, extra any) (string, error) {
	return s.GenerateJWTWithExpiry(userId, extra, s.config.ExpiryTime)
}

func (s *service) GenerateJWTWithExpiry(userId uint64, extra any, expiry time.Duration) (string, error) {
	now := time.Now()
	claims := jwt.MapClaims{
		s.config.ContextKey: userId,
		"exp":               now.Add(expiry).Unix(),
		"iat":               now.Unix(),
		"iss":               s.config.Issuer,
		"extra":             extra,
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

			token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, errresp.NewAuthorizationError(constant.ErrInvalidToken)
				}
				return s.secret, nil
			})

			if err != nil || !token.Valid {
				return nil, errresp.NewAuthorizationError(constant.ErrInvalidToken)
			}

			claims, ok := token.Claims.(jwt.MapClaims)
			if !ok {
				return nil, errresp.NewAuthorizationError(constant.ErrInvalidToken)
			}

			return handler(s.NewContextWithUserId(ctx, cast.ToUint64(claims[s.config.ContextKey])), req)
		}
	}
}

func (s *service) NewContextWithUserId(ctx context.Context, userId uint64) context.Context {
	return context.WithValue(ctx, s.config.ContextKey, userId)
}

func (s *service) GetUserId(ctx context.Context) (uint64, error) {
	value := ctx.Value(s.config.ContextKey)
	if userId, ok := value.(uint64); ok {
		return userId, nil
	}
	return 0, errors.New("failed to get user Id from context")
}
