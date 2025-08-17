package parser

import (
	"archive/zip"
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// FBXParser FBX文件解析器
type FBXParser interface {
	ParseTextureReferences(fbxPath string) ([]string, error)
	GetSupportedModelExtensions() map[string]bool
	GetSupportedTextureExtensions() map[string]bool
	ExtractSingleFileFromZip(file *zip.File, destPath string) error
}

type parser struct{}

// TextureMapping 贴图映射结构
type TextureMapping struct {
	Source string `json:"source"` // 原文件名
	Target string `json:"target"` // 目标URL
}

// NewFBXParser 创建FBX解析器实例
func NewFBXParser() FBXParser {
	return &parser{}
}

// ParseTextureReferences 解析FBX文件中的贴图引用
// 参数：fbxPath - FBX文件路径
// 返回：贴图文件名列表，错误信息
func (p *parser) ParseTextureReferences(fbxPath string) ([]string, error) {
	file, err := os.Open(fbxPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open FBX file: %v", err)
	}
	defer file.Close()

	// 读取文件前几个字节来判断是否为ASCII格式
	header := make([]byte, 23)
	_, err = file.Read(header)
	if err != nil {
		return nil, fmt.Errorf("failed to read FBX header: %v", err)
	}

	// 重置文件指针
	file.Seek(0, 0)

	// 检查是否为ASCII FBX格式
	if strings.HasPrefix(string(header), "Kaydara FBX Binary") {
		// 二进制格式FBX - 使用专门的解析方法
		return p.parseBinaryFBXTextures(file)
	} else {
		// ASCII格式FBX - 使用文本解析
		return p.parseASCIIFBXTextures(file)
	}
}

// parseASCIIFBXTextures 解析ASCII格式FBX文件中的贴图引用
func (p *parser) parseASCIIFBXTextures(file *os.File) ([]string, error) {
	var textures []string
	textureMap := make(map[string]bool) // 用于去重

	scanner := bufio.NewScanner(file)
	inTexture := false
	inMaterial := false

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// 检查是否进入Texture定义
		if strings.Contains(line, "Model:") && strings.Contains(line, "Texture") {
			inTexture = true
			continue
		}

		// 检查是否进入Material定义
		if strings.Contains(line, "Model:") && strings.Contains(line, "Material") {
			inMaterial = true
			continue
		}

		// 检查是否离开当前对象定义
		if strings.HasPrefix(line, "}") {
			inTexture = false
			inMaterial = false
			continue
		}

		// 在Texture或Material定义中查找文件路径
		if inTexture || inMaterial {
			// 查找RelativeFilename, Filename, 或类似的属性
			if strings.Contains(line, "RelativeFilename:") ||
				strings.Contains(line, "Filename:") ||
				strings.Contains(line, "FileName:") {

				// 提取引号中的文件路径
				start := strings.Index(line, "\"")
				if start != -1 {
					end := strings.LastIndex(line, "\"")
					if end > start {
						filePath := line[start+1 : end]
						// 清理路径并提取文件名
						fileName := filepath.Base(strings.ReplaceAll(filePath, "\\", "/"))
						if fileName != "" && !textureMap[fileName] {
							textures = append(textures, fileName)
							textureMap[fileName] = true
						}
					}
				}
			}
		}

		// 也检查Properties70中的贴图引用
		if strings.Contains(line, "DiffuseColor") ||
			strings.Contains(line, "BaseColor") ||
			strings.Contains(line, "NormalMap") ||
			strings.Contains(line, "SpecularColor") ||
			strings.Contains(line, "EmissiveColor") ||
			strings.Contains(line, "Bump") ||
			strings.Contains(line, "DisplacementColor") ||
			strings.Contains(line, "TransparencyFactor") ||
			strings.Contains(line, "ReflectionColor") {
			// 继续读取下一行寻找文件引用
			if scanner.Scan() {
				nextLine := strings.TrimSpace(scanner.Text())
				if strings.Contains(nextLine, "\"") {
					start := strings.Index(nextLine, "\"")
					if start != -1 {
						end := strings.LastIndex(nextLine, "\"")
						if end > start {
							filePath := nextLine[start+1 : end]
							fileName := filepath.Base(strings.ReplaceAll(filePath, "\\", "/"))
							if fileName != "" && strings.Contains(fileName, ".") && !textureMap[fileName] {
								textures = append(textures, fileName)
								textureMap[fileName] = true
							}
						}
					}
				}
			}
		}

		// 检查连接关系（Connections部分）
		if strings.Contains(line, "Connect:") && strings.Contains(line, "Texture") {
			// 在连接定义中可能包含贴图文件信息
			parts := strings.Split(line, ",")
			for _, part := range parts {
				part = strings.TrimSpace(part)
				if strings.Contains(part, "\"") {
					start := strings.Index(part, "\"")
					if start != -1 {
						end := strings.LastIndex(part, "\"")
						if end > start {
							potential := part[start+1 : end]
							if isTextureFile(potential) && !textureMap[potential] {
								textures = append(textures, potential)
								textureMap[potential] = true
							}
						}
					}
				}
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading FBX file: %v", err)
	}

	return textures, nil
}

// parseBinaryFBXTextures 解析二进制格式FBX文件中的贴图引用
func (p *parser) parseBinaryFBXTextures(file *os.File) ([]string, error) {
	var textures []string
	textureMap := make(map[string]bool)

	// 二进制FBX解析比较复杂，这里实现一个简化版本
	// 主要通过搜索文件中的常见贴图文件扩展名来找到贴图引用

	// 读取整个文件内容
	content, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read binary FBX file: %v", err)
	}

	// 搜索贴图文件扩展名的模式
	extensions := []string{".jpg", ".jpeg", ".png", ".bmp", ".tga", ".dds", ".exr", ".hdr", ".tif", ".tiff"}

	contentStr := string(content)
	for _, ext := range extensions {
		// 查找所有包含该扩展名的位置
		for i := 0; i < len(contentStr)-len(ext); i++ {
			if strings.ToLower(contentStr[i:i+len(ext)]) == ext {
				// 向前查找文件名的开始
				start := i
				for start > 0 && (contentStr[start-1] != 0 && contentStr[start-1] != '\\' &&
					contentStr[start-1] != '/' && contentStr[start-1] != '"' && contentStr[start-1] != ' ') {
					start--
				}

				// 提取文件名
				fileName := contentStr[start : i+len(ext)]
				// 清理无效字符
				fileName = strings.TrimFunc(fileName, func(r rune) bool {
					return r < 32 || r > 126
				})

				// 验证是否为有效的文件名
				if len(fileName) > len(ext) && isValidFileName(fileName) && !textureMap[fileName] {
					textures = append(textures, fileName)
					textureMap[fileName] = true
				}
			}
		}
	}

	return textures, nil
}

// ExtractSingleFileFromZip 从ZIP文件中提取单个文件
func (p *parser) ExtractSingleFileFromZip(file *zip.File, destPath string) error {
	// 打开ZIP中的文件
	rc, err := file.Open()
	if err != nil {
		return fmt.Errorf("failed to open file in zip: %v", err)
	}
	defer rc.Close()

	// 创建目标文件
	outFile, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("failed to create destination file: %v", err)
	}
	defer outFile.Close()

	// 复制文件内容
	_, err = io.Copy(outFile, rc)
	if err != nil {
		return fmt.Errorf("failed to copy file content: %v", err)
	}

	return nil
}

// FilterTexturesByFBX 根据FBX文件过滤贴图文件
// 参数：fbxPath - FBX文件路径，availableTextures - 可用的贴图文件列表
// 返回：过滤后的贴图文件列表，错误信息
func (p *parser) FilterTexturesByFBX(fbxPath string, availableTextures []string) ([]string, error) {
	requiredTextures, err := p.ParseTextureReferences(fbxPath)
	if err != nil {
		return availableTextures, fmt.Errorf("failed to parse FBX textures: %v", err)
	}

	if len(requiredTextures) == 0 {
		// 如果没有解析到贴图引用，返回所有可用贴图
		return availableTextures, nil
	}

	var filteredTextures []string
	for _, available := range availableTextures {
		for _, required := range requiredTextures {
			if strings.EqualFold(filepath.Base(available), required) ||
				strings.EqualFold(filepath.Base(available), filepath.Base(required)) {
				filteredTextures = append(filteredTextures, available)
				break
			}
		}
	}

	return filteredTextures, nil
}

// GetSupportedModelExtensions 获取支持的3D模型格式
func (p *parser) GetSupportedModelExtensions() map[string]bool {
	return map[string]bool{
		".fbx":  true,
		".glb":  true,
		".gltf": true,
		".obj":  true,
		".dae":  true,
		".3ds":  true,
		".ply":  true,
		".stl":  true,
	}
}

// GetSupportedTextureExtensions 获取支持的贴图格式
func (p *parser) GetSupportedTextureExtensions() map[string]bool {
	return map[string]bool{
		".jpg":  true,
		".jpeg": true,
		".png":  true,
		".bmp":  true,
		".tga":  true,
		".dds":  true,
		".exr":  true,
		".hdr":  true,
		".tif":  true,
		".tiff": true,
		".webp": true,
	}
}

// isTextureFile 检查是否为贴图文件
func isTextureFile(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	supportedExts := map[string]bool{
		".jpg": true, ".jpeg": true, ".png": true, ".bmp": true,
		".tga": true, ".dds": true, ".exr": true, ".hdr": true,
		".tif": true, ".tiff": true, ".webp": true,
	}
	return supportedExts[ext]
}

// isValidFileName 检查是否为有效的文件名
func isValidFileName(filename string) bool {
	// 检查文件名长度
	if len(filename) < 5 || len(filename) > 255 {
		return false
	}

	// 检查是否包含文件扩展名
	if !strings.Contains(filename, ".") {
		return false
	}

	// 检查是否只包含有效字符
	for _, r := range filename {
		if r < 32 || r > 126 {
			return false
		}
	}

	// 检查是否为贴图文件
	return isTextureFile(filename)
}

// ParseFBXInfo 解析FBX文件基本信息
type FBXInfo struct {
	Version     string
	Creator     string
	IsBinary    bool
	TextureRefs []string
}

// GetFBXInfo 获取FBX文件的基本信息
func (p *parser) GetFBXInfo(fbxPath string) (*FBXInfo, error) {
	file, err := os.Open(fbxPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open FBX file: %v", err)
	}
	defer file.Close()

	info := &FBXInfo{}

	// 读取文件头
	header := make([]byte, 27)
	_, err = file.Read(header)
	if err != nil {
		return nil, fmt.Errorf("failed to read FBX header: %v", err)
	}

	// 检查格式
	if strings.HasPrefix(string(header), "Kaydara FBX Binary") {
		info.IsBinary = true
	} else {
		info.IsBinary = false
	}

	// 重置文件指针
	file.Seek(0, 0)

	// 解析贴图引用
	info.TextureRefs, err = p.ParseTextureReferences(fbxPath)
	if err != nil {
		return nil, fmt.Errorf("failed to parse texture references: %v", err)
	}

	// 对于ASCII格式，尝试提取版本和创建者信息
	if !info.IsBinary {
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if strings.Contains(line, "FBXHeaderExtension:") {
				// 继续读取版本信息
				for scanner.Scan() {
					versionLine := strings.TrimSpace(scanner.Text())
					if strings.Contains(versionLine, "FBXVersion:") {
						parts := strings.Split(versionLine, ":")
						if len(parts) > 1 {
							info.Version = strings.TrimSpace(parts[1])
						}
					}
					if strings.Contains(versionLine, "Creator:") {
						start := strings.Index(versionLine, "\"")
						if start != -1 {
							end := strings.LastIndex(versionLine, "\"")
							if end > start {
								info.Creator = versionLine[start+1 : end]
							}
						}
					}
					if strings.HasPrefix(versionLine, "}") {
						break
					}
				}
				break
			}
		}
	}

	return info, nil
}
