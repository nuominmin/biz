package middleware

import (
	"log"
	"net/http"
	"time"

	"github.com/nuominmin/biz/krs/types"

	"github.com/go-kratos/kratos/v2/encoding"
	"github.com/spf13/cast"

	transporthttp "github.com/go-kratos/kratos/v2/transport/http"
)

// ResponseServerOption .
func ResponseServerOption() transporthttp.ServerOption {
	return transporthttp.ResponseEncoder(func(w http.ResponseWriter, r *http.Request, i interface{}) error {
		return handleJSONResponse(w, r, i)
	})
}

// handleJSONResponse 处理标准JSON响应
func handleJSONResponse(w http.ResponseWriter, r *http.Request, i interface{}) error {
	var nCode int
	if code := w.Header().Get("code"); code != "" {
		nCode = cast.ToInt(code)
	}

	reply := &types.Response{
		Code: nCode,
		Data: i,
		Ts:   time.Now().Format(time.RFC3339),
	}

	data, err := encoding.GetCodec("json").Marshal(reply)
	if err != nil {
		log.Printf("Failed to marshal JSON response: %v", err)
		errorResponse := types.NewErrorResponse(500, "Internal server error")
		errorData, _ := encoding.GetCodec("json").Marshal(errorResponse)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write(errorData)
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	// 根据业务错误码设置HTTP状态码
	if nCode != 0 {
		w.WriteHeader(http.StatusBadRequest)
	}
	_, _ = w.Write(data)
	return nil
}
