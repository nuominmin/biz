package oss

import (
	"errors"
	"fmt"
	"io"
	"mime"
	"path"
	"path/filepath"
	"strings"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/google/uuid"
)

type Service interface {
	GenerateUniqueFilepath(dir, originalFilename string) string
	UploadFile(reader io.Reader, originalFilename string) (string, error)
	DownloadFile(filename string) (io.ReadCloser, error)
	DeleteFile(filename string) error
}

type service struct {
	opts   options
	client *oss.Client
	bucket *oss.Bucket
}

func NewOss(optFns ...Option) (Service, error) {
	opts := newOptions(optFns...)

	// 验证OSS配置的完整性
	err := validateOSSConfig(opts)
	if err != nil {
		return nil, err
	}

	// 创建OSS客户端
	client, err := oss.New(opts.endpoint, opts.accessKeyId, opts.accessKeySecret)
	if err != nil {
		return nil, fmt.Errorf("failed to create OSS client: %v", err)
	}

	// 获取存储桶
	bucket, err := client.Bucket(opts.bucketName)
	if err != nil {
		return nil, fmt.Errorf("failed to get OSS bucket '%s': %v", opts.bucketName, err)
	}

	return &service{
		opts:   opts,
		client: client,
		bucket: bucket,
	}, nil
}

// 验证OSS配置的完整性
func validateOSSConfig(opts options) error {
	if opts.endpoint == "" {
		return fmt.Errorf("endpoint is required")
	}
	if opts.accessKeyId == "" {
		return fmt.Errorf("access_key_id is required")
	}
	if opts.accessKeySecret == "" {
		return fmt.Errorf("access_key_secret is required")
	}
	if opts.bucketName == "" {
		return fmt.Errorf("bucket_name is required")
	}
	if opts.baseUrl == "" {
		return fmt.Errorf("base_url is required")
	}
	// 检查 endpoint 和 base_url 是否匹配
	if !strings.Contains(opts.baseUrl, opts.endpoint) {
		return fmt.Errorf("endpoint '%s' and base_url '%s' region mismatch - they should use the same region",
			opts.endpoint, opts.baseUrl)
	}

	return nil
}

// UploadFile 上传文件到OSS或本地存储
func (s *service) UploadFile(reader io.Reader, filename string) (string, error) {
	// 设置上传选项
	opts := []oss.Option{
		oss.ContentType(s.getContentType(filename)),
		oss.ObjectACL(oss.ACLPublicRead), // 设置为公共读
		// 添加缓存控制，允许浏览器缓存
		oss.CacheControl("public, max-age=31536000"), // 1年缓存
		// 添加 CORS 相关头部（虽然主要的 CORS 配置需要在控制台设置）
		oss.ContentDisposition("inline"), // 浏览器内联显示而不是下载
	}

	// 上传文件到OSS
	err := s.bucket.PutObject(filename, reader, opts...)
	if err != nil {
		// 解析OSS错误，提供更详细的错误信息
		var ossErr oss.ServiceError
		if errors.As(err, &ossErr) {
			switch ossErr.Code {
			case "AccessDenied":
				if strings.Contains(ossErr.Message, "endpoint") {
					return "", fmt.Errorf("region mismatch: bucket is in different region than configured endpoint. Please check your OSS configuration. Error: %s", ossErr.Message)
				}
				return "", fmt.Errorf("access denied: insufficient permissions to upload file. Error: %s", ossErr.Message)
			case "NoSuchBucket":
				return "", fmt.Errorf("bucket '%s' does not exist. Error: %s", s.opts.bucketName, ossErr.Message)
			case "InvalidAccessKeyId":
				return "", fmt.Errorf("invalid AccessKeyId in configuration. Error: %s", ossErr.Message)
			case "SignatureDoesNotMatch":
				return "", fmt.Errorf("invalid AccessKeySecret in configuration. Error: %s", ossErr.Message)
			case "RequestTimeTooSkewed":
				return "", fmt.Errorf("system time is incorrect. Please sync your system time. Error: %s", ossErr.Message)
			default:
				return "", fmt.Errorf("OSS upload error [%s]: %s", ossErr.Code, ossErr.Message)
			}
		}
		return "", fmt.Errorf("failed to upload file to OSS: %v", err)
	}

	// 返回文件的公网访问URL
	fileURL := fmt.Sprintf("%s/%s", strings.TrimRight(s.opts.baseUrl, "/"), filename)
	return fileURL, nil
}

// DownloadFile 从OSS下载文件
func (s *service) DownloadFile(filename string) (io.ReadCloser, error) {
	// 从OSS获取文件
	reader, err := s.bucket.GetObject(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to download file from OSS: %v", err)
	}
	return reader, nil
}

// DeleteFile 从OSS删除文件
func (s *service) DeleteFile(filename string) error {
	err := s.bucket.DeleteObject(filename)
	if err != nil {
		return fmt.Errorf("failed to delete file from OSS: %v", err)
	}
	return nil
}

// GenerateUniqueFilepath 生成唯一文件路径（带目录）
// dir: 目标目录（可以为空）
// originalFilename: 原文件名，用于提取扩展名
func (s *service) GenerateUniqueFilepath(dir, originalFilename string) string {
	ext := filepath.Ext(originalFilename)
	name := strings.ReplaceAll(uuid.New().String(), "-", "")
	filename := fmt.Sprintf("%s%s", name, ext)

	if dir != "" {
		// 归一化目录分隔符并使用 URL 风格路径拼接，避免 Windows 下反斜杠导致的对象 Key/URL 异常
		cleanDir := strings.ReplaceAll(dir, "\\", "/")
		key := path.Join(cleanDir, filename)
		return strings.TrimLeft(key, "/")
	}
	return filename
}

// getContentType 根据文件扩展名获取MIME类型
func (s *service) getContentType(filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))

	// 首先尝试标准MIME类型
	if mimeType := mime.TypeByExtension(ext); mimeType != "" {
		return mimeType
	}

	// 为3D模型文件设置特定的MIME类型
	switch ext {
	case ".fbx":
		return "model/fbx" // 更具体的 MIME 类型，有助于浏览器识别
	case ".obj":
		return "model/obj"
	case ".dae":
		return "model/vnd.collada+xml"
	case ".gltf":
		return "model/gltf+json"
	case ".glb":
		return "model/gltf-binary"
	case ".3ds":
		return "model/3ds"
	case ".blend":
		return "application/x-blender"
	case ".max":
		return "application/x-3dsmax"
	case ".ma", ".mb":
		return "application/x-maya"
	// 视频文件
	case ".mp4":
		return "video/mp4"
	case ".avi":
		return "video/x-msvideo"
	case ".mov":
		return "video/quicktime"
	default:
		return "application/octet-stream"
	}
}
