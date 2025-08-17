package server

import (
	"github.com/go-kratos/kratos/v2/transport/http"
	"github.com/nuominmin/biz/krs/types"
	"github.com/nuominmin/biz/upload"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"path/filepath"
	"strings"
)

// Upload 上传
func (s *service) Upload(uploadSvc upload.Service) func(http.Context) error {
	return func(ctx http.Context) error {
		// 获取文件
		file, handler, err := ctx.Request().FormFile("file")
		if err != nil {
			return status.Errorf(codes.Internal, "Failed to read file: %v", err)
		}
		defer file.Close()

		// 检查文件大小是否超过最大限制
		if handler.Size > defaultMaxFileSize {
			return status.Errorf(codes.InvalidArgument, "File size exceeds maximum limit of %d MB", defaultMaxFileSize/(1024*1024))
		}

		// 获取文件扩展名
		ext := strings.ToLower(filepath.Ext(handler.Filename))

		// 检查是否为允许的文件类型，如果未配置，则允许所有类型
		if s.opts.allowedTypes != nil && len(s.opts.allowedTypes) > 0 {
			if _, ok := s.opts.allowedTypes[ext]; !ok {
				return status.Errorf(codes.InvalidArgument, "File type %s is not allowed", ext)
			}
		}

		// 生成唯一文件名
		filename := uploadSvc.GenerateUniqueFilename(handler.Filename)

		// 上传
		var fileURL string
		if fileURL, err = uploadSvc.UploadFile(file, filename); err != nil {
			return status.Errorf(codes.Internal, "Failed to upload: %v, dir: %s, filename: %s", err, filename)
		}

		data := types.Upload{
			Url:      fileURL,
			Filename: filepath.Base(fileURL),
			Size:     handler.Size,
		}

		return ctx.JSON(200, types.NewSuccessResponse(data))
	}
}
