package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
)

var wsServer *WebSocketServer
var watchPath string

func main() {
	filePath := flag.String("f", "", "文件路径")
	dirPath := flag.String("d", "", "目录路径")

	// 同时支持 -f/--file 和 -d/--dir 的格式
	flag.StringVar(filePath, "file", "", "文件路径")
	flag.StringVar(dirPath, "dir", "", "目录路径")

	flag.Parse()

	// 检查是否至少提供了一个参数
	if *filePath == "" && *dirPath == "" {
		fmt.Println("错误：必须指定 --file 或 --dir 参数")
		flag.Usage()
		os.Exit(1)
	}

	// 不能同时指定 -f/--file 和 -d/--dir 参数
	if *filePath != "" && *dirPath != "" {
		fmt.Println("错误：不能同时指定 -f/--file 和 -d/--dir 参数")
		flag.Usage()
		os.Exit(1)
	}

	// 启动 websocket 服务
	wsServer = NewWebSocketServer(onSocketMessage)
	go wsServer.Start("localhost:8081")

	// 如果指定了 -f/--file 参数，则监控文件
	if *filePath != "" {
		watchPath = *filePath
		watcher, err := NewFileWatcher()
		if err != nil {
			log.Fatalf("NewFileWatcher() error: %v", err)
		}
		go watcher.WatchFile(*filePath, onFileChange)
	}

	// // 如果指定了 -d/--dir 参数，则监控目录
	if *dirPath != "" {
		watchPath = *dirPath
		go WatchDirectory(*dirPath, onFileChange)
	}

	// 启动 http 服务
	httpServer := NewHTTPServer(":8081")
	err := httpServer.Start(true)
	if err != nil {
		log.Fatalf("httpServer.Start(true) error: %v", err)
	}
}

type SocketMessage struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}

func onSocketMessage(message []byte) ([]byte, error) {
	msg := SocketMessage{}
	err := json.Unmarshal(message, &msg)
	if err != nil {
		log.Fatalf("json.Unmarshal(message) error: %v", err)
		return nil, err
	}
	var response interface{}
	switch msg.Type {
	case "get_info":
		response = interface{}(map[string]interface{}{
			"type": "info",
			"data": map[string]interface{}{
				"type": "file",
				"path": watchPath,
			},
		})
	case "get_file_content":
		content, err := os.ReadFile(msg.Data.(string))
		if err != nil {
			log.Fatalf("os.ReadFile(msg.Data.(string)) error: %v", err)
			return nil, err
		}
		response = interface{}(map[string]interface{}{
			"type": "file_content",
			"data": string(content),
		})
	case "ping":
		response = interface{}(map[string]interface{}{
			"type": "pong",
		})
	default:
		return nil, fmt.Errorf("未知的消息类型: %s", msg.Type)
	}

	msgStr, err := json.Marshal(response)
	if err != nil {
		log.Fatalf("json.Marshal(msg) error: %v", err)
		return nil, err
	}
	return msgStr, nil
}

func onFileChange(change FileChange) {
	fmt.Printf("文件变化:\n")
	fmt.Printf("路径: %s\n", change.Path)
	fmt.Printf("操作: %s\n", change.Op)
	if change.Op == "modified" {
		for _, lineChange := range change.LineChanges {
			fmt.Printf("修改 (行 %d -> %d):\n  原内容: %s\n  新内容: %s\n",
				lineChange.OldLine,
				lineChange.NewLine,
				lineChange.OldText,
				lineChange.NewText)
		}
	}
	fmt.Println("------------------------")
	msg := interface{}(map[string]interface{}{
		"type": "log",
		"data": change,
	})
	msgStr, err := json.Marshal(msg)
	if err != nil {
		log.Fatalf("json.Marshal(msg) error: %v", err)
	}
	wsServer.broadcast <- msgStr
}
