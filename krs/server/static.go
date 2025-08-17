package server

import (
	"crypto/tls"
	"encoding/json"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/transport/http"
	"mime"
	"os"
	"path/filepath"
	"strings"
)

// 静态服务配置文件名
const staticServerConfigFilename = "config.json"

// StaticServerWithConfig 带配置的静态服务器
func StaticServerWithConfig[T any](root, addr string, tlsCertFile string, tlsKeyFile string, configData T) *http.Server {
	httpServer := StaticServer(root, addr, tlsCertFile, tlsKeyFile)

	// 序列化为 JSON
	jsonData, err := json.Marshal(configData)
	if err != nil {
		log.Errorf("failed to marshal config: %v, configData: %+v", err, configData)
		return httpServer
	}

	// 写入文件
	configPath := filepath.Join(root, staticServerConfigFilename)
	if err = os.WriteFile(configPath, jsonData, 0644); err != nil {
		log.Errorf("failed to write config.json: %v, configPath: %s", err, configPath)
	}

	return httpServer
}

// StaticServer 静态服务器
func StaticServer(root, addr string, tlsCertFile string, tlsKeyFile string) *http.Server {
	// 使用 Kratos HTTP 服务器创建静态文件服务
	opts := []http.ServerOption{
		http.Address(addr),
	}

	// 添加 TLS 支持
	if tlsCertFile != "" && tlsKeyFile != "" {
		cert, err := tls.LoadX509KeyPair(tlsCertFile, tlsKeyFile)
		if err != nil {
			log.Errorf("加载%s静态服务器 TLS 证书失败: %v", root, err)
		} else {
			opts = append(opts, http.TLSConfig(&tls.Config{
				Certificates: []tls.Certificate{cert},
			}))
			log.Infof("%s静态服务器已启用 HTTPS", root)
		}
	}

	server := http.NewServer(opts...)

	// 使用通配符路由统一处理所有静态资源请求
	// 支持路径: /assets/*, /, /index.html 等
	server.Route("/").GET("/", func(ctx http.Context) error {
		serveStaticFileUnified(ctx.Response(), ctx.Request(), root)
		return nil
	})
	server.Route("/").GET("{path:.*}", func(ctx http.Context) error {
		serveStaticFileUnified(ctx.Response(), ctx.Request(), root)
		return nil
	})
	return server
}

// serveStaticFileUnified 统一处理静态文件请求
func serveStaticFileUnified(w http.ResponseWriter, r *http.Request, root string) {
	// 获取请求路径，移除前导斜杠
	path := strings.TrimPrefix(r.URL.Path, "/")

	// 如果路径为空，默认返回 index.html
	if path == "" {
		path = "index.html"
	}

	log.Infof("静态文件请求: path='%s', root='%s'", path, root)

	// 构建完整的文件路径
	filePath := filepath.Join(root, path)

	// 安全检查，防止路径遍历攻击
	cleanRoot := filepath.Clean(root)
	cleanFilePath := filepath.Clean(filePath)

	if !filepath.HasPrefix(cleanFilePath, cleanRoot) {
		log.Errorf("路径遍历攻击尝试: path='%s', filePath='%s', root='%s'", path, cleanFilePath, cleanRoot)
		w.WriteHeader(403)
		w.Write([]byte("403 Forbidden"))
		return
	}

	log.Infof("尝试访问文件: '%s'", filePath)

	// 检查文件是否存在
	info, err := os.Stat(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			log.Errorf("文件不存在: '%s'", filePath)
			// 对于 SPA 应用，静态资源文件不存在时不应该返回 index.html
			// 只有在访问路由页面时才返回 index.html
			if shouldFallbackToIndex(path) {
				indexPath := filepath.Join(root, "index.html")
				if _, indexErr := os.Stat(indexPath); indexErr == nil {
					log.Infof("返回 index.html 用于 SPA 路由: %s", indexPath)
					filePath = indexPath
				} else {
					log.Errorf("index.html 也不存在: %s", indexPath)
					w.WriteHeader(404)
					w.Write([]byte("404 Not Found"))
					return
				}
			} else {
				w.WriteHeader(404)
				w.Write([]byte("404 Not Found"))
				return
			}
		}
		log.Errorf("访问文件时出错: %v", err)
		w.WriteHeader(500)
		w.Write([]byte("500 Internal Server Error"))
		return
	}

	if info.IsDir() {
		// 如果是目录，尝试返回目录下的 index.html
		indexPath := filepath.Join(filePath, "index.html")
		if _, indexErr := os.Stat(indexPath); indexErr == nil {
			filePath = indexPath
		} else {
			w.WriteHeader(404)
			w.Write([]byte("404 Not Found"))
			return
		}
	}

	// 读取文件内容
	data, err := os.ReadFile(filePath)
	if err != nil {
		log.Errorf("读取静态文件出错: %v", err)
		w.WriteHeader(500)
		w.Write([]byte("500 Internal Server Error"))
		return
	}

	// 设置 Content-Type
	ext := filepath.Ext(filePath)
	mimeType := mime.TypeByExtension(ext)
	if mimeType == "" {
		mimeType = "application/octet-stream"
	}

	// 设置响应头
	w.Header().Set("Content-Type", mimeType)
	if strings.Contains(filePath, "index.html") {
		w.Header().Set("Cache-Control", "no-cache")
	} else {
		w.Header().Set("Cache-Control", "public, max-age=3600") // 其他文件缓存1小时
	}

	log.Infof("成功返回文件: '%s', 大小: %d bytes, Content-Type: %s", filePath, len(data), mimeType)

	// 返回文件内容
	w.Write(data)
}

// shouldFallbackToIndex 判断是否应该回退到 index.html
// 对于静态资源文件(js, css, 图片等)，不应该回退到 index.html
func shouldFallbackToIndex(path string) bool {
	// 如果是静态资源文件，不回退到 index.html
	staticExtensions := []string{".js", ".css", ".png", ".jpg", ".jpeg", ".gif", ".svg", ".ico", ".woff", ".woff2", ".ttf", ".eot"}
	ext := strings.ToLower(filepath.Ext(path))

	for _, staticExt := range staticExtensions {
		if ext == staticExt {
			return false
		}
	}

	// 如果路径以 /assets/ 开头，也不回退
	if strings.HasPrefix(path, "assets/") {
		return false
	}

	return true
}
