package upload

import (
	"errors"
	"fmt"
	"io"
	"net/url"
	"strings"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
)

// OssService 是OSS服务的接口
type OssService interface {
	Service
	SetBucketCORS(rules ...oss.CORSRule) error
}

// 默认CORS规则
var defaultCorsRule = oss.CORSRule{
	AllowedOrigin: []string{"*"},
	AllowedMethod: []string{"GET", "HEAD"},
	AllowedHeader: []string{"*"},
	ExposeHeader:  []string{"ETag", "Content-Length", "Content-Type"},
	MaxAgeSeconds: 86400,
}

type ossService struct {
	*service
	client *oss.Client
	bucket *oss.Bucket
}

func NewOssService(optFns ...Option) (OssService, error) {
	ossSvc := &ossService{
		service: newService(optFns...),
	}

	// 验证OSS配置的完整性
	err := ossSvc.validateOSSConfig(ossSvc.opts)
	if err != nil {
		return nil, err
	}

	// 创建OSS客户端
	client, err := oss.New(ossSvc.opts.endpoint, ossSvc.opts.accessKeyId, ossSvc.opts.accessKeySecret)
	if err != nil {
		return nil, fmt.Errorf("failed to create OSS client: %v", err)
	}
	ossSvc.client = client

	// 获取存储桶
	bucket, err := client.Bucket(ossSvc.opts.bucketName)
	if err != nil {
		return nil, fmt.Errorf("failed to get OSS bucket '%s': %v", ossSvc.opts.bucketName, err)
	}
	ossSvc.bucket = bucket

	return ossSvc, nil
}

// 验证OSS配置的完整性
func (s *ossService) validateOSSConfig(opts options) error {
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
func (s *ossService) UploadFile(reader io.Reader, dir, name string) (string, error) {
	// 拼接文件路径
	filename := s.joinPath(dir, name)

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
func (s *ossService) DownloadFile(filename string) (io.ReadCloser, error) {
	// 从OSS获取文件
	reader, err := s.bucket.GetObject(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to download file from OSS: %v", err)
	}
	return reader, nil
}

// DeleteFile 从OSS删除文件
func (s *ossService) DeleteFile(filename string) error {
	err := s.bucket.DeleteObject(filename)
	if err != nil {
		return fmt.Errorf("failed to delete file from OSS: %v", err)
	}
	return nil
}

/*
SetBucketCORS 设置CORS配置

例子：

	var rule1 = oos.CORSRule{
		AllowedOrigin: []string{"*"},
		AllowedMethod: []string{"GET", "HEAD"},
		AllowedHeader: []string{"*"},
		ExposeHeader:  []string{"ETag", "Content-Length", "Content-Type"},
		MaxAgeSeconds: 86400,
	}

	var rule2 = oos.CORSRule{
		AllowedOrigin: []string{"http://www.a.com", "http://www.b.com"},
		AllowedMethod: []string{"GET"},
		AllowedHeader: []string{"Authorization"},
		ExposeHeader:  []string{"x-oss-test", "x-oss-test1"},
		MaxAgeSeconds: 200,
	}
*/
func (s *ossService) SetBucketCORS(rules ...oss.CORSRule) error {
	// 为空使用默认的
	if len(rules) == 0 {
		rules = append(rules, defaultCorsRule)
	}

	return s.client.SetBucketCORS(s.bucket.BucketName, rules)
}

// RemoveDomainFromURL 从URL中移除域名，只保留路径部分
func (s *ossService) RemoveDomainFromURL(fullURL string) string {
	if fullURL == "" {
		return ""
	}

	// 如果已经是相对路径，直接返回
	if !strings.HasPrefix(fullURL, "http://") && !strings.HasPrefix(fullURL, "https://") {
		return fullURL
	}

	// 解析URL
	parsedURL, err := url.Parse(fullURL)
	if err != nil {
		// 如果解析失败，返回原URL
		return fullURL
	}

	// 检查是否是OSS域名或本地域名
	host := parsedURL.Host
	isOSSDomain := strings.Contains(s.opts.baseUrl, host)
	isLocalDomain := strings.HasPrefix(fullURL, "http://127.0.0.1") || strings.HasPrefix(fullURL, "http://localhost")

	if isOSSDomain || isLocalDomain {
		// 返回路径部分（包括查询参数和片段）
		path := parsedURL.Path
		if parsedURL.RawQuery != "" {
			path += "?" + parsedURL.RawQuery
		}
		if parsedURL.Fragment != "" {
			path += "#" + parsedURL.Fragment
		}
		return path
	}

	// 如果不是已知域名，保留原URL
	return fullURL
}

// AddDomainToURL 为相对路径添加域名
func (s *ossService) AddDomainToURL(relativePath string) string {
	if relativePath == "" {
		return ""
	}

	// 如果已经是完整URL，直接返回
	if strings.HasPrefix(relativePath, "http://") || strings.HasPrefix(relativePath, "https://") {
		return relativePath
	}

	// 确保路径以 / 开头
	if !strings.HasPrefix(relativePath, "/") {
		relativePath = "/" + relativePath
	}

	// 如果启用了OSS，使用OSS的base_url
	if s.opts.baseUrl != "" {
		baseURL := strings.TrimRight(s.opts.baseUrl, "/")
		return baseURL + relativePath
	}

	// 否则使用本地路径（由前端处理域名拼接）
	return relativePath
}
