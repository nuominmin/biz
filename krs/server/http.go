package server

import (
	"github.com/go-kratos/kratos/v2/transport/http"
	"github.com/nuominmin/biz/captcha"
	"github.com/nuominmin/biz/upload"
)

type Service interface {
	Upload(uploadSvc upload.Service) func(http.Context) error
	UploadModel3D(uploadSvc upload.Service) func(http.Context) error
	StaticFileRead(uploadSvc upload.Service) func(http.Context) error
	Captcha(captchaSvc captcha.Service) func(http.Context) error
}

type service struct {
	opts options
}

func NewService(optFns ...Option) Service {
	return &service{
		opts: newOptions(optFns...),
	}
}

const (
	// 默认最大文件大小
	defaultMaxFileSize = 500 * 1024 * 1024
	// 路由路径
	RouteFilePath = "{filename:.*}"
)
