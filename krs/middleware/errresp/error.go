package errresp

import (
	"fmt"
	"github.com/nuominmin/biz/krs/middleware/constant"
)

// File 文件
type File struct {
	Content     []byte // 文件内容
	Filename    string // 文件名
	ContentType string // Content-Type，默认为 application/octet-stream
	Inline      bool   // 是否内联显示，false为下载
}

func NewFile(f *File) *File {
	return f
}

func (e *File) Error() string {
	return ""
}

// Redirect 重定向
type Redirect struct {
	URL       string
	Permanent bool // 是否设置永久重定向状态码和位置标头
}

func (e *Redirect) Error() string {
	return ""
}

func NewRedirect(url string, permanent bool) *Redirect {
	return &Redirect{
		URL:       url,
		Permanent: permanent,
	}
}

// Error 错误
type Error struct {
	Code    int
	Message string
}

func (e *Error) Error() string {
	return fmt.Sprintf(`{"code": %d, "message": "%s"}`, e.Code, e.Message)
}

func NewAuthorizationError(format string, a ...any) *Error {
	if format == "" {
		format = constant.ErrMessageUnauthorized
	}
	return &Error{
		Code:    401,
		Message: fmt.Sprintf(format, a...),
	}
}
