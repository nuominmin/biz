package upload

import (
	"fmt"
	"io"
	"mime"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
)

type Service interface {
	UploadFile(reader io.Reader, dir, name string) (string, error)
	DownloadFile(filename string) (io.ReadCloser, error)
	DeleteFile(filename string) error
	GenerateUniqueFilename(originalFilename string) string
}

type service struct {
	opts options
}

func newService(optFns ...Option) *service {
	opts := newOptions(optFns...)
	return &service{
		opts: opts,
	}
}

func NewService(optFns ...Option) Service {
	return newService(optFns...)
}

// UploadFile 上传文件
func (s *service) UploadFile(reader io.Reader, dir, name string) (string, error) {
	// 拼接文件路径
	filename := filepath.Join(dir, name)

	// 创建目录
	if err := os.MkdirAll(filepath.Dir(filename), 0755); err != nil {
		return "", fmt.Errorf("创建目录失败 (%s): %w", dir, err)
	}

	// 创建目标文件
	destFile, err := os.Create(filename)
	if err != nil {
		return "", fmt.Errorf("创建目标文件失败: %v", err)
	}

	// 复制文件内容
	_, err = io.Copy(destFile, reader)

	// 关闭目标文件
	destFile.Close()
	if err != nil {
		os.Remove(filename)
		return "", fmt.Errorf("复制文件内容失败: %w", err)
	}

	return filename, nil
}

// DownloadFile 下载文件
func (s *service) DownloadFile(filename string) (io.ReadCloser, error) {
	// 检查文件是否存在
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		return nil, fmt.Errorf("文件不存在: %s", filename)
	}

	// 打开文件
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("打开文件失败: %w", err)
	}

	// 返回 io.ReadCloser，调用方用完后需关闭
	return file, nil
}

// DeleteFile 删除文件
func (s *service) DeleteFile(filename string) error {
	// 检查文件是否存在
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		return fmt.Errorf("文件不存在: %s", filename)
	}

	// 删除文件
	if err := os.Remove(filename); err != nil {
		return fmt.Errorf("删除文件失败: %w", err)
	}

	return nil
}

// joinPath 拼接目录和文件名，返回统一格式的相对路径。
// 1. 将目录分隔符统一为 "/"（跨平台一致性）。
// 2. 使用 path.Join 拼接目录和文件名，自动去掉多余的 "/"。
// 3. 去掉结果开头的 "/"，保证返回的是相对路径。
func (s *service) joinPath(dir, name string) string {
	// 统一分隔符
	cleanDir := filepath.ToSlash(dir)
	// 使用 path.Join 拼接（它会自动去掉多余的 "/")
	key := path.Join(cleanDir, name)
	// 去掉开头的 "/"
	return strings.TrimPrefix(key, "/")
}

// GenerateUniqueFilename 生成唯一文件名
// originalFilename: 原文件名，用于提取扩展名
func (s *service) GenerateUniqueFilename(originalFilename string) string {
	ext := filepath.Ext(originalFilename)
	name := strings.ReplaceAll(uuid.New().String(), "-", "")
	return fmt.Sprintf("%s%s", name, ext)
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
