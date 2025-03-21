package main

import (
	"embed"
	"fmt"
	"io/fs"
	"path"
)

//go:embed all:ui/dist/ui/browser
var uiFiles embed.FS

// GetEmbedFile 从嵌入的文件系统中获取文件内容
func GetEmbedFile(filename string) (string, bool) {
	// 构建完整的文件路径
	filepath := path.Join("ui/dist/ui/browser", filename)

	// 读取文件内容
	content, err := fs.ReadFile(uiFiles, filepath)
	if err != nil {
		return "", false
	}

	return string(content), true
}

// ListEmbedFiles 列出所有嵌入的文件
func ListEmbedFiles() []string {
	var files []string
	fs.WalkDir(uiFiles, "ui/dist/ui/browser", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			// 移除前缀路径，只保留文件名
			filename := path[len("ui/dist/ui/browser/"):]
			files = append(files, filename)
		}
		return nil
	})
	return files
}

// PrintEmbedFiles 打印所有嵌入的文件
func PrintEmbedFiles() {
	files := ListEmbedFiles()
	fmt.Println("嵌入的文件列表：")
	for _, file := range files {
		fmt.Printf("- %s\n", file)
	}
}
