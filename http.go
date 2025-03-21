package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"
)

// HTTPServer HTTP服务器结构
type HTTPServer struct {
	addr string
}

// NewHTTPServer 创建新的HTTP服务器
func NewHTTPServer(addr string) *HTTPServer {
	return &HTTPServer{
		addr: addr,
	}
}

// Start 启动HTTP服务器
func (s *HTTPServer) Start(devMode bool) error {
	if devMode {
		s.enableDevMode()
	} else {
		http.HandleFunc("/", s.handleRequest)
	}

	log.Printf("HTTP服务器启动在 %s，开发模式：%v", s.addr, devMode)
	return http.ListenAndServe(s.addr, nil)
}

// handleRequest 处理请求
func (s *HTTPServer) handleRequest(w http.ResponseWriter, r *http.Request) {
	// 从路径中获取文件名
	filename := strings.TrimPrefix(r.URL.Path, "/")
	if filename == "" {
		filename = "index.html"
	}

	// 获取文件内容
	content, exists := GetEmbedFile(filename)
	if !exists {
		http.NotFound(w, r)
		return
	}

	// 设置 Content-Type
	switch {
	case strings.HasSuffix(filename, ".html"):
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
	case strings.HasSuffix(filename, ".js"):
		w.Header().Set("Content-Type", "application/javascript; charset=utf-8")
	case strings.HasSuffix(filename, ".css"):
		w.Header().Set("Content-Type", "text/css; charset=utf-8")
	case strings.HasSuffix(filename, ".png"):
		w.Header().Set("Content-Type", "image/png")
	case strings.HasSuffix(filename, ".jpg"), strings.HasSuffix(filename, ".jpeg"):
		w.Header().Set("Content-Type", "image/jpeg")
	case strings.HasSuffix(filename, ".gif"):
		w.Header().Set("Content-Type", "image/gif")
	case strings.HasSuffix(filename, ".svg"):
		w.Header().Set("Content-Type", "image/svg+xml")
	case strings.HasSuffix(filename, ".ico"):
		w.Header().Set("Content-Type", "image/x-icon")
	}

	// 设置缓存控制
	if !strings.HasSuffix(filename, ".html") {
		w.Header().Set("Cache-Control", "public, max-age=31536000")
	}

	fmt.Fprint(w, content)
}

// 开发模式配置
func (s *HTTPServer) enableDevMode() {
	// 添加CORS支持
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// 设置CORS头
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		// 处理OPTIONS请求
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		// 所有请求都使用 handleRequest 处理
		s.handleRequest(w, r)
	})
}
