package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/fsnotify/fsnotify"
)

type FileEventCallback func(FileChange)

// FileChange 文件变化信息结构
type FileChange struct {
	Path        string       `json:"path"`         // 文件路径
	Op          string       `json:"op"`           // 操作类型
	Content     string       `json:"-"`            // 当前文件内容
	PrevContent string       `json:"-"`            // 之前的文件内容
	LineChanges []LineChange `json:"line_changes"` // 行级别的变化
	FileInfo    os.FileInfo  `json:"file_info"`    // 文件信息
}

// LineChange 行变化信息
type LineChange struct {
	Type    string `json:"type"`     // added, removed, modified
	OldLine int    `json:"old_line"` // 原行号
	NewLine int    `json:"new_line"` // 新行号
	OldText string `json:"old_text"` // 原内容
	NewText string `json:"new_text"` // 新内容
}

// FileWatcher 文件监视器
type FileWatcher struct {
	watcher    *fsnotify.Watcher
	fileCache  map[string]string // 缓存文件内容
	cacheMutex sync.RWMutex
}

// NewFileWatcher 创建新的文件监视器
func NewFileWatcher() (*FileWatcher, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	return &FileWatcher{
		watcher:   watcher,
		fileCache: make(map[string]string),
	}, nil
}

// compareLines 比较两个文本的差异
func (fw *FileWatcher) compareLines(oldContent, newContent string) []LineChange {
	oldLines := strings.Split(oldContent, "\n")
	newLines := strings.Split(newContent, "\n")
	changes := make([]LineChange, 0)

	// 使用最长公共子序列（LCS）算法的简化版本
	m := len(oldLines)
	n := len(newLines)

	// 创建动态规划表
	dp := make([][]int, m+1)
	for i := range dp {
		dp[i] = make([]int, n+1)
	}

	// 填充动态规划表
	for i := 1; i <= m; i++ {
		for j := 1; j <= n; j++ {
			if oldLines[i-1] == newLines[j-1] {
				dp[i][j] = dp[i-1][j-1] + 1
			} else {
				dp[i][j] = max(dp[i-1][j], dp[i][j-1])
			}
		}
	}

	// 回溯找出差异
	i, j := m, n
	for i > 0 || j > 0 {
		switch {
		case i > 0 && j > 0 && oldLines[i-1] == newLines[j-1]:
			i--
			j--
		case j > 0 && (i == 0 || dp[i][j-1] >= dp[i-1][j]):
			// 新增行
			changes = append([]LineChange{{
				Type:    "added",
				NewLine: j,
				NewText: newLines[j-1],
			}}, changes...)
			j--
		case i > 0 && (j == 0 || dp[i][j-1] < dp[i-1][j]):
			// 删除行
			changes = append([]LineChange{{
				Type:    "removed",
				OldLine: i,
				OldText: oldLines[i-1],
			}}, changes...)
			i--
		}
	}

	return changes
}

// WatchFile 监听单个文件的变化
func (fw *FileWatcher) WatchFile(filePath string, callback func(FileChange)) error {
	// 首次读取文件内容并缓存
	content, _, err := fw.readFileContent(filePath)
	if err != nil {
		return err
	}

	fw.cacheMutex.Lock()
	fw.fileCache[filePath] = content
	fw.cacheMutex.Unlock()

	err = fw.watcher.Add(filePath)
	if err != nil {
		return fmt.Errorf("添加文件监听失败: %v", err)
	}

	go func() {
		for {
			select {
			case event, ok := <-fw.watcher.Events:
				if !ok {
					return
				}
				log.Printf("文件变化: %v\n", event.Name)
				if event.Has(fsnotify.Write) {
					fw.handleFileChange(event.Name, callback)
				}
			case err, ok := <-fw.watcher.Errors:
				if !ok {
					return
				}
				log.Printf("监听错误: %v\n", err)
			}
		}
	}()

	return nil
}

// handleFileChange 处理文件变化
func (fw *FileWatcher) handleFileChange(filePath string, callback func(FileChange)) {
	// 读取新内容
	newContent, fileInfo, err := fw.readFileContent(filePath)
	if err != nil {
		log.Printf("读取文件失败 %s: %v", filePath, err)
		return
	}

	// 获取旧内容
	fw.cacheMutex.RLock()
	oldContent := fw.fileCache[filePath]
	fw.cacheMutex.RUnlock()

	// 比较差异
	changes := fw.compareLines(oldContent, newContent)

	// 更新缓存
	fw.cacheMutex.Lock()
	fw.fileCache[filePath] = newContent
	fw.cacheMutex.Unlock()

	// 创建变化信息
	fileChange := FileChange{
		Path:        filePath,
		Op:          "modified",
		Content:     newContent,
		PrevContent: oldContent,
		LineChanges: changes,
		FileInfo:    fileInfo,
	}

	// 调用回调函数
	if callback != nil {
		callback(fileChange)
	}
}

// readFileContent 读取文件内容
func (fw *FileWatcher) readFileContent(filePath string) (string, os.FileInfo, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", nil, err
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		return "", nil, err
	}

	maxSize := int64(10 * 1024 * 1024) // 10MB
	if fileInfo.Size() > maxSize {
		return "", fileInfo, fmt.Errorf("文件太大: %d > %d", fileInfo.Size(), maxSize)
	}

	var buffer bytes.Buffer
	reader := bufio.NewReader(file)

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				buffer.WriteString(line)
				break
			}
			return "", nil, err
		}
		buffer.WriteString(line)
	}

	return buffer.String(), fileInfo, nil
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// Close 关闭监视器
func (fw *FileWatcher) Close() {
	fw.watcher.Close()
}

// WatchDirectory 监听目录下所有文件的变化
func WatchDirectory(dirPath string, callback FileEventCallback) error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("创建监听器失败: %v", err)
	}
	defer watcher.Close()

	done := make(chan bool)
	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}

				// 获取操作类型
				var op string
				switch {
				case event.Has(fsnotify.Write):
					op = "modified"
				case event.Has(fsnotify.Create):
					op = "created"
				case event.Has(fsnotify.Remove):
					op = "removed"
				case event.Has(fsnotify.Rename):
					op = "renamed"
				default:
					continue
				}

				// 对于删除和重命名操作，无法读取文件内容
				if op == "removed" || op == "renamed" {
					change := FileChange{
						Path: event.Name,
						Op:   op,
					}
					if callback != nil {
						callback(change)
					}
					continue
				}

				// 读取文件内容
				content, fileInfo, err := readFileContent(event.Name)
				if err != nil {
					log.Printf("读取文件内容失败 %s: %v", event.Name, err)
					continue
				}

				change := FileChange{
					Path:     event.Name,
					Op:       op,
					Content:  content,
					FileInfo: fileInfo,
				}

				if callback != nil {
					callback(change)
				}

			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Printf("监听错误: %v\n", err)
			}
		}
	}()

	// 递归添加目录下的所有子目录
	err = filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			err = watcher.Add(path)
			if err != nil {
				return fmt.Errorf("添加目录监听失败 %s: %v", path, err)
			}
			log.Printf("正在监听目录: %s\n", path)
		}
		return nil
	})

	if err != nil {
		return fmt.Errorf("遍历目录失败: %v", err)
	}

	<-done
	return nil
}

// readFileContent 读取文件内容
func readFileContent(filePath string) (string, os.FileInfo, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", nil, err
	}
	defer file.Close()

	// 获取文件信息
	fileInfo, err := file.Stat()
	if err != nil {
		return "", nil, err
	}

	// 如果文件太大，可能需要限制读取大小
	maxSize := int64(10 * 1024 * 1024) // 10MB
	if fileInfo.Size() > maxSize {
		return "", fileInfo, fmt.Errorf("文件太大: %d > %d", fileInfo.Size(), maxSize)
	}

	// 读取文件内容
	content, err := io.ReadAll(file)
	if err != nil {
		return "", fileInfo, err
	}

	return string(content), fileInfo, nil
}
