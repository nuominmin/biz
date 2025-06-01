package middleware

import (
	"context"
	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/transport"
	"github.com/go-kratos/kratos/v2/transport/http"
	nhttp "net/http"
	"strings"
)

func LocalHttpRequestFilter() http.FilterFunc {
	return func(next nhttp.Handler) nhttp.Handler {
		return nhttp.HandlerFunc(func(w nhttp.ResponseWriter, req *nhttp.Request) {
			req.Header.Add("X-RemoteAddr", strings.Split(req.RemoteAddr, ":")[0])
			next.ServeHTTP(w, req)
		})
	}
}
func LocalHttpRequestMiddleware() middleware.Middleware {
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req interface{}) (reply interface{}, err error) {
			if tp, ok := transport.FromServerContext(ctx); ok {
				ctx = context.WithValue(ctx, "X-RemoteAddr", tp.RequestHeader().Get("X-RemoteAddr"))
			}
			return handler(ctx, req)
		}
	}
}
