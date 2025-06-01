package middleware

import (
	"net/http"
	"time"

	"github.com/go-kratos/kratos/v2/encoding"
	"github.com/spf13/cast"

	transporthttp "github.com/go-kratos/kratos/v2/transport/http"
)

// Response define standard response format
type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Ts      string      `json:"ts"`
}

// ResponseServerOption .
func ResponseServerOption() transporthttp.ServerOption {
	return transporthttp.ResponseEncoder(func(w http.ResponseWriter, r *http.Request, i interface{}) error {
		var nCode int
		if code := w.Header().Get("code"); code != "" {
			nCode = cast.ToInt(code)
		}

		reply := &Response{
			Code: nCode,
			Data: i,
			Ts:   time.Now().Format(time.RFC3339),
		}
		data, _ := encoding.GetCodec("json").Marshal(reply)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(data)
		return nil
	})
}
