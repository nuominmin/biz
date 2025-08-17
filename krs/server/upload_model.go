package server

import (
	"github.com/go-kratos/kratos/v2/transport/http"
	"github.com/nuominmin/biz/krs/types"
	"github.com/nuominmin/biz/upload"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func (s *service) UploadModel3D(uploadSvc upload.Service) func(http.Context) error {
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

		// 检查文件类型必须是ZIP
		ext := strings.ToLower(filepath.Ext(handler.Filename))
		if ext != ".zip" {
			return status.Errorf(codes.InvalidArgument, "File type must be ZIP")
		}

		// 创建临时目录用于解压
		tempDir, err := os.MkdirTemp("", "model_upload_*")
		if err != nil {
			return status.Errorf(codes.Internal, "Failed to create temp directory: %v", err)
		}
		defer os.RemoveAll(tempDir)

		// 保存上传的ZIP文件到临时位置
		tempZipPath := filepath.Join(tempDir, "model.zip")
		tempZipFile, err := os.Create(tempZipPath)
		if err != nil {
			return status.Errorf(codes.Internal, "Failed to create temp zip file: %v", err)
		}
		defer tempZipFile.Close()
		// 复制上传的文件内容到临时ZIP文件
		_, err = io.Copy(tempZipFile, file)
		if err != nil {
			return status.Errorf(codes.Internal, "Failed to save temp zip file: %v", err)
		}
		tempZipFile.Close()

		// 解压ZIP文件
		modelURL, modelTextures, err := uploadSvc.ExtractAndSaveModel3D(tempZipPath)
		if err != nil {
			return status.Errorf(codes.Internal, "Failed to extract model: %v", err)
		}

		// 返回模型URL和贴图映射列表
		return ctx.JSON(200, types.NewSuccessResponse(map[string]interface{}{
			"model_url":          modelURL,
			"model_texture_urls": modelTextures, // 确保字段名为 model_texture_urls
			"model_textures":     modelTextures, // 兼容前端当前使用的字段名
		}))
	}
}
