package server

import (
	"fmt"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/transport/http"
	"github.com/nuominmin/biz/upload"
	"path/filepath"
	"strings"
)

// StaticFileRead 静态文件读取
func (s *service) StaticFileRead(uploadSvc upload.Service) func(http.Context) error {
	return func(ctx http.Context) error {
		filename := ctx.Vars().Get("filename")

		// 拼接文件路径
		filename = filepath.Join(upload.DefaultUploadDir, filename)

		data, err := uploadSvc.DownloadFile(filename)
		if err != nil {
			log.Errorf("read file error (%+v), filename: %s", err, filename)
			ctx.Response().WriteHeader(500)
			_, _ = ctx.Response().Write([]byte("500 Internal Server Error"))
			return nil
		}

		contentType := uploadSvc.GetContentType(filename)

		// 设置正确的Content-Type和缓存头
		ctx.Response().Header().Set("Content-Type", contentType)
		ctx.Response().Header().Set("Cache-Control", "public, max-age=31536000") // 缓存1年
		ctx.Response().Header().Set("Content-Length", fmt.Sprintf("%d", len(data)))

		// 对于视频文件，添加支持范围请求的头部
		if strings.HasPrefix(contentType, "video/") {
			ctx.Response().Header().Set("Accept-Ranges", "bytes")
		}

		_, _ = ctx.Response().Write(data)
		return nil
	}
}
