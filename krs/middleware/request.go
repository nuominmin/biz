package middleware

import (
	"context"
	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/transport"
	"github.com/go-kratos/kratos/v2/transport/http"
	"net"
	nhttp "net/http"
	"strings"
)

// ClientIPFilter extracts the real client IP address from various headers
func ClientIPFilter() http.FilterFunc {
	return func(next nhttp.Handler) nhttp.Handler {
		return nhttp.HandlerFunc(func(w nhttp.ResponseWriter, req *nhttp.Request) {
			clientIP := extractClientIP(req)
			req.Header.Set("X-Real-IP", clientIP)
			next.ServeHTTP(w, req)
		})
	}
}

// ClientIPMiddleware adds client IP to context
func ClientIPMiddleware() middleware.Middleware {
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req interface{}) (reply interface{}, err error) {
			if tp, ok := transport.FromServerContext(ctx); ok {
				realIP := tp.RequestHeader().Get("X-Real-IP")
				if realIP != "" {
					ctx = context.WithValue(ctx, "client_ip", realIP)
				}
			}
			return handler(ctx, req)
		}
	}
}

// extractClientIP extracts the real client IP from request headers
func extractClientIP(req *nhttp.Request) string {
	// Check X-Forwarded-For header (most common)
	xff := req.Header.Get("X-Forwarded-For")
	if xff != "" {
		// X-Forwarded-For can contain multiple IPs, take the first one
		ips := strings.Split(xff, ",")
		clientIP := strings.TrimSpace(ips[0])
		if net.ParseIP(clientIP) != nil {
			return clientIP
		}
	}

	// Check X-Real-IP header
	xri := req.Header.Get("X-Real-IP")
	if xri != "" && net.ParseIP(xri) != nil {
		return xri
	}

	// Check CF-Connecting-IP header (Cloudflare)
	cfIP := req.Header.Get("CF-Connecting-IP")
	if cfIP != "" && net.ParseIP(cfIP) != nil {
		return cfIP
	}

	// Fall back to RemoteAddr
	host, _, err := net.SplitHostPort(req.RemoteAddr)
	if err != nil {
		return req.RemoteAddr
	}
	return host
}
