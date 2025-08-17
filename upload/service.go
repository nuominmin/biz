package upload

import (
	"archive/zip"
	"bytes"
	"fmt"
	"github.com/nuominmin/biz/parser"
	"io"
	"mime"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
)

type Service interface {
	UploadFile(reader io.Reader, name string) (string, error)
	SaveFile(filePath string, name string) (string, error)
	ExtractAndSaveModel3D(zipPath string) (string, []parser.TextureMapping, error)
	DownloadFile(filename string) ([]byte, error)
	DeleteFile(filename string) error

	GetContentType(filename string) string
	GenerateUniqueFilename(originalFilename string) string
	RemoveDomainFromURL(host, fullURL string) string
	AddDomainToURL(host, relativePath string) string
}

type service struct {
	host string
	dir  string
	opts options
}

func newService(host, dir string, optFns ...Option) *service {
	return &service{
		host: strings.TrimRight(host, "/"),
		dir:  dir,
		opts: newOptions(optFns...),
	}
}

func NewService(host, dir string, optFns ...Option) Service {
	return newService(host, dir, optFns...)
}

// UploadFile 上传文件
func (s *service) UploadFile(reader io.Reader, name string) (string, error) {
	// 拼接文件路径
	filename := filepath.Join(DefaultUploadDir, s.dir, name)

	// 创建目录
	if err := os.MkdirAll(filepath.Dir(filename), 0755); err != nil {
		return "", fmt.Errorf("创建目录失败 (%s): %w", s.dir, err)
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

	// 确保 filename 里用的是 /
	filename = path.Clean(strings.ReplaceAll(filename, "\\", "/"))

	return fmt.Sprintf("%s/%s", s.host, filename), nil
}

// SaveFile 保存文件
func (s *service) SaveFile(filename string, name string) (string, error) {
	// 检查文件是否存在
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		return "", fmt.Errorf("文件不存在: %s", filename)
	}

	// 读取文件内容
	data, err := os.ReadFile(filename)
	if err != nil {
		return "", fmt.Errorf("打开文件失败: %w", err)
	}

	return s.UploadFile(bytes.NewReader(data), name)
}

// DownloadFile 下载文件
func (s *service) DownloadFile(filename string) ([]byte, error) {
	// 检查文件是否存在
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		return nil, fmt.Errorf("文件不存在: %s", filename)
	}

	// 读取文件内容
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("打开文件失败: %w", err)
	}

	return data, nil
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

// RemoveDomainFromURL 从URL中移除域名，只保留路径部分
func (s *service) RemoveDomainFromURL(host, fullURL string) string {
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

	// 检查是否是域名
	isOSSDomain := strings.Contains(host, parsedURL.Host)
	isLocalDomain := strings.HasPrefix(fullURL, "http://127.0.0.1") || strings.HasPrefix(fullURL, "http://localhost")

	if isOSSDomain || isLocalDomain {
		// 返回路径部分（包括查询参数和片段）
		parsedURLPath := parsedURL.Path
		if parsedURL.RawQuery != "" {
			parsedURLPath += "?" + parsedURL.RawQuery
		}
		if parsedURL.Fragment != "" {
			parsedURLPath += "#" + parsedURL.Fragment
		}
		return parsedURLPath
	}

	// 如果不是已知域名，保留原URL
	return fullURL
}

// AddDomainToURL 为相对路径添加域名
func (s *service) AddDomainToURL(host, relativePath string) string {
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

	baseURL := strings.TrimRight(host, "/")
	return baseURL + relativePath
}

// GetContentType 根据文件扩展名获取MIME类型
func (s *service) GetContentType(filename string) string {
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

// ExtractAndSaveModel3D 解压并保存模型文件
func (s *service) ExtractAndSaveModel3D(zipPath string) (string, []parser.TextureMapping, error) {
	// 创建FBX解析器实例
	fbxParser := parser.NewFBXParser()

	// 获取支持的文件格式
	modelExtensions := fbxParser.GetSupportedModelExtensions()
	textureExtensions := fbxParser.GetSupportedTextureExtensions()

	// 打开ZIP文件
	reader, err := zip.OpenReader(zipPath)
	if err != nil {
		return "", nil, fmt.Errorf("failed to open zip file: %v", err)
	}
	defer reader.Close()

	// 临时文件目录
	tempDir := "./temp"
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		return "", nil, fmt.Errorf("failed to create upload directory: %v", err)
	}

	var modelURL string
	var modelTextures []parser.TextureMapping
	var requiredTextures []string
	var fbxFilePath string

	// 第一遍：建立文件映射并找到FBX文件
	for _, file := range reader.File {
		if file.FileInfo().IsDir() {
			continue
		}

		// 检查是否为FBX文件
		ext := strings.ToLower(filepath.Ext(file.Name))
		if ext == ".fbx" && fbxFilePath == "" {
			// 临时保存FBX文件用于解析（总是保存到本地临时文件，因为FBX解析器需要本地文件）
			fbxTempPath := filepath.Join(tempDir, "temp_model.fbx")
			err := fbxParser.ExtractSingleFileFromZip(file, fbxTempPath)
			if err == nil {
				fbxFilePath = fbxTempPath
			}
		}
	}

	// 如果找到FBX文件，解析其贴图依赖
	if fbxFilePath != "" {
		requiredTextures, err = fbxParser.ParseTextureReferences(fbxFilePath)
		if err != nil {
			// 如果解析失败，回退到提取所有贴图文件
			fmt.Printf("Failed to parse FBX textures, falling back to extract all: %v\n", err)
		} else {
			fmt.Printf("Found %d texture references in FBX file\n", len(requiredTextures))
		}
		// 清理临时FBX文件
		os.Remove(fbxFilePath)
	}

	// 第二遍：解压文件
	for _, file := range reader.File {
		// 跳过目录
		if file.FileInfo().IsDir() {
			continue
		}

		// 获取文件扩展名
		ext := strings.ToLower(filepath.Ext(file.Name))
		fileName := filepath.Base(file.Name)

		// 打开压缩包中的文件
		srcFile, err := file.Open()
		if err != nil {
			fmt.Printf("Failed to open file in zip: %s, error: %v\n", file.Name, err)
			continue
		}

		// 生成唯一文件名
		uniqueFilename := s.GenerateUniqueFilename(fileName)

		// 上传
		var fileURL string
		if fileURL, err = s.UploadFile(srcFile, uniqueFilename); err != nil {
			fmt.Printf("Failed to upload file: %s, error: %v\n", fileName, err)
			continue
		}
		srcFile.Close()

		// 分类文件
		if modelExtensions[ext] && modelURL == "" {
			// 只保留第一个找到的模型文件
			modelURL = fileURL
		} else if textureExtensions[ext] {
			// 创建贴图映射对象
			modelTextures = append(modelTextures, parser.TextureMapping{
				Source: fileName,
				Target: fileURL,
			})
		}
	}

	// 验证是否找到了模型文件
	if modelURL == "" {
		return "", nil, fmt.Errorf("no supported 3D model file found in zip")
	}

	return modelURL, modelTextures, nil
}
