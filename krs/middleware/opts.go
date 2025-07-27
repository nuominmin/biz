package middleware

import (
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/transport/http"
	"github.com/gorilla/handlers"
	"github.com/nuominmin/biz/krs/middleware/jwt"
	"time"
)

func DefaultServerOption() []http.ServerOption {
	return []http.ServerOption{
		http.Timeout(time.Second * 30),
		http.Filter(ProductionCORS(), ClientIPFilter()),
		ErrorEncoderOption(),
		ResponseServerOption(),
		http.Middleware(
			recovery.Recovery(),
			ClientIPMiddleware(),
		),
	}
}

func DefaultServerOptionWithJwt(jwtSvc jwt.Service, ignoredPaths ...string) []http.ServerOption {
	return []http.ServerOption{
		http.Timeout(time.Second * 30),
		http.Filter(ProductionCORS(), ClientIPFilter()),
		ErrorEncoderOption(),
		ResponseServerOption(),
		http.Middleware(
			recovery.Recovery(),
			jwtSvc.Middleware(ignoredPaths...),
			ClientIPMiddleware(),
		),
	}
}

func ProductionCORS(allowedOrigins ...string) http.FilterFunc {
	if len(allowedOrigins) == 0 {
		allowedOrigins = append(allowedOrigins, "*")
	}
	return handlers.CORS(
		handlers.AllowedHeaders([]string{"X-Requested-With", "Content-Type", "Authorization", "User-Agent", "Content-Length", "Access-Control-Allow-Credentials"}),
		handlers.AllowedMethods([]string{"GET", "POST", "PUT", "DELETE", "HEAD", "OPTIONS"}),
		handlers.AllowedOrigins([]string{"*"}),
		handlers.AllowCredentials(),
	)
}
