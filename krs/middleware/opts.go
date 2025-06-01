package middleware

import (
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/transport/http"
	"github.com/gorilla/handlers"
	"github.com/nuominmin/biz/krs/authorization/jwt"
	"time"
)

func DefaultServerOption() []http.ServerOption {
	return []http.ServerOption{
		http.Timeout(time.Second * 30),
		http.Filter(DefaultCORS(), LocalHttpRequestFilter()),
		ErrorServerOption(),    // error handler
		ResponseServerOption(), // response handler
		http.Middleware(
			recovery.Recovery(),
			LocalHttpRequestMiddleware(),
		),
	}
}

func DefaultServerOptionWithJwt(jwtSvc jwt.Service, ignoredPaths ...string) []http.ServerOption {
	return []http.ServerOption{
		http.Timeout(time.Second * 30),
		http.Filter(DefaultCORS(), LocalHttpRequestFilter()),
		ErrorServerOption(),    // error handler
		ResponseServerOption(), // response handler
		http.Middleware(
			recovery.Recovery(),
			jwtSvc.Middleware(ignoredPaths...),
			LocalHttpRequestMiddleware(),
		),
	}
}

func DefaultCORS() http.FilterFunc {
	return handlers.CORS(
		handlers.AllowedHeaders([]string{"X-Requested-With", "Content-Type", "Authorization", "User-Agent", "Content-Length", "Access-Control-Allow-Credentials"}),
		handlers.AllowedMethods([]string{"GET", "POST", "PUT", "HEAD", "OPTIONS"}),
		handlers.AllowedOrigins([]string{"*"}), handlers.AllowedOrigins([]string{"*"}),
		handlers.AllowCredentials(),
	)
}
