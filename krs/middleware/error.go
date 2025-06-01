package middleware

import (
	auth "github.com/nuominmin/biz/krs/authorization"
	"net/http"

	"github.com/go-kratos/kratos/v2/errors"
	transporthttp "github.com/go-kratos/kratos/v2/transport/http"
)

// redirect error
type RedirectError struct {
	URL        string
	Permanent  bool
	underlying error
}

func (e *RedirectError) Error() string {
	return e.underlying.Error()
}

func (e *RedirectError) Unwrap() error {
	return e.underlying
}

func NewRedirectError(url string, permanent bool) *RedirectError {
	return &RedirectError{
		URL:       url,
		Permanent: permanent,
	}
}

func ErrorServerOption() transporthttp.ServerOption {
	return transporthttp.ErrorEncoder(func(w http.ResponseWriter, r *http.Request, err error) {

		if err == nil {
			return
		}

		var redErr *RedirectError
		if errors.As(err, &redErr) {
			if redErr.Permanent {
				// set permanent redirect status code and Location header
				w.Header().Set("Location", redErr.URL)
				w.WriteHeader(http.StatusMovedPermanently)
				return
			}

			// set temporary redirect status code and Location header
			w.Header().Set("Location", redErr.URL)
			w.WriteHeader(http.StatusFound)
			return
		}

		var authErr *auth.Error
		if errors.As(err, &authErr) {
			w.Header().Set("Content-Type", "application/json")

			//w.WriteHeader(http.StatusUnauthorized) // 401 Unauthorized
			_, _ = w.Write([]byte(authErr.Error()))
			return
		}

		transporthttp.DefaultErrorEncoder(w, r, err)
		return
	})
}
