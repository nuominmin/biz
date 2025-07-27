package middleware

import (
	"github.com/nuominmin/biz/krs/middleware/errresp"
	"github.com/spf13/cast"
	"log"
	"net/http"
	"net/url"

	"github.com/go-kratos/kratos/v2/errors"
	transporthttp "github.com/go-kratos/kratos/v2/transport/http"
)

// ErrorEncoderOption returns a server option that configures error encoding
func ErrorEncoderOption() transporthttp.ServerOption {
	return transporthttp.ErrorEncoder(errorEncoder)
}

func errorEncoder(w http.ResponseWriter, r *http.Request, err error) {
	if err == nil {
		return
	}

	// Log the error for debugging
	log.Printf("Error handling request %s %s: %v", r.Method, r.URL.Path, err)

	var file *errresp.File
	if errors.As(err, &file) {
		handleFileResponse(w, file)
		return
	}

	var redirect *errresp.Redirect
	if errors.As(err, &redirect) {
		handleRedirectResponse(w, redirect)
		return
	}

	var authErr *errresp.Error
	if errors.As(err, &authErr) {
		handleErrorResponse(w, authErr)
		return
	}

	transporthttp.DefaultErrorEncoder(w, r, err)
}

func handleFileResponse(w http.ResponseWriter, file *errresp.File) {
	// 设置Content-Type，默认为 application/octet-stream
	contentType := file.ContentType
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	w.Header().Set("Content-Type", contentType)

	// 设置Content-Disposition
	disposition := "attachment"
	if file.Inline {
		disposition = "inline"
	}
	if file.Filename != "" {
		disposition += "; filename=\"" + url.QueryEscape(file.Filename) + "\""
	}
	w.Header().Set("Content-Disposition", disposition)

	// 设置Content-Length
	w.Header().Set("Content-Length", cast.ToString(len(file.Content)))

	// 写入文件内容
	_, _ = w.Write(file.Content)
}

func handleRedirectResponse(w http.ResponseWriter, redirect *errresp.Redirect) {
	w.Header().Set("Location", redirect.URL)
	if redirect.Permanent {
		w.WriteHeader(http.StatusMovedPermanently)
	} else {
		w.WriteHeader(http.StatusFound)
	}
}

func handleErrorResponse(w http.ResponseWriter, authErr *errresp.Error) {
	w.Header().Set("Content-Type", "application/json")
	// Note: Consider setting appropriate HTTP status code based on error code
	// w.WriteHeader(authErr.Code)
	_, _ = w.Write([]byte(authErr.Error()))
}
