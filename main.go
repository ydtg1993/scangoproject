package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	dir := flag.String("dir", ".", "扫描录进")
	flag.Parse()

	info, err := os.Stat(*dir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "无法访问目录 %s: %v\n", *dir, err)
		os.Exit(1)
	}
	if !info.IsDir() {
		fmt.Fprintf(os.Stderr, "%s 不是一个目录\n", *dir)
		os.Exit(1)
	}

	if err := NewWriter(*dir, "output.txt", []string{".go", ".yaml"}, []string{".idea", ".git", "test"}); err != nil {
		fmt.Fprintf(os.Stderr, "处理失败: %v\n", err)
		os.Exit(1)
	}
}

type Writer struct {
	writer   *bufio.Writer
	baseDir  string // 存储根目录
	exts     []string
	fileList []string
	skipDirs []string
}

func NewWriter(aimDir, output string, exts, skipDirs []string) error {
	f, err := os.Create(output)
	if err != nil {
		return fmt.Errorf("无法创建输出文件 %s: %w", output, err)
	}
	defer f.Close()
	writer := bufio.NewWriter(f)
	defer writer.Flush()

	w := &Writer{
		exts:     exts,
		writer:   writer,
		baseDir:  aimDir,
		skipDirs: skipDirs,
	}
	// 遍历目录并写入目录结构
	if err := w.walkDir(aimDir, "", ""); err != nil {
		return fmt.Errorf("遍历目录时出错: %w", err)
	}
	// 写入文件内容
	if len(w.fileList) > 0 {
		fmt.Fprintln(writer)
		fmt.Fprintln(writer, "File contents:")
		fmt.Fprintln(writer, strings.Repeat("=", 50))

		for _, relPath := range w.fileList {
			fmt.Fprintf(writer, "\nFile: %s\n", relPath)
			fmt.Fprintln(writer, strings.Repeat("-", 40))

			// 读取文件内容
			fullPath := filepath.Join(w.baseDir, relPath)
			content, err := os.ReadFile(fullPath)
			if err != nil {
				fmt.Fprintf(writer, "[读取文件失败: %v]\n", err)
				continue
			}
			// 写入内容，注意防止末尾换行重复
			writer.Write(content)
			if len(content) > 0 && content[len(content)-1] != '\n' {
				writer.WriteString("\n")
			}
		}
	}
	return nil
}

func (w *Writer) walkDir(basePath, relPath, prefix string) error {
	entries, err := os.ReadDir(basePath)
	if err != nil {
		return err
	}
	var filteredEntries []os.DirEntry
	for _, entry := range entries {
		if entry.IsDir() && contains(w.skipDirs, entry.Name()) {
			continue // 跳过该目录，不显示也不进入
		}
		filteredEntries = append(filteredEntries, entry)
	}

	for i, entry := range filteredEntries {
		isLast := i == len(filteredEntries)-1

		var line string
		if isLast {
			line = prefix + "└── "
		} else {
			line = prefix + "├── "
		}
		line += entry.Name()
		if entry.IsDir() {
			line += "/"
		}
		fmt.Fprintln(w.writer, line)

		if !entry.IsDir() {
			ext := filepath.Ext(entry.Name())
			if contains(w.exts, ext) {
				var fileRelPath string
				if relPath == "" {
					fileRelPath = entry.Name()
				} else {
					fileRelPath = filepath.Join(relPath, entry.Name())
				}
				fileRelPath = filepath.ToSlash(fileRelPath)
				w.fileList = append(w.fileList, fileRelPath)
			}
		}

		if entry.IsDir() {
			subBasePath := filepath.Join(basePath, entry.Name())
			var subRelPath string
			if relPath == "" {
				subRelPath = entry.Name()
			} else {
				subRelPath = filepath.Join(relPath, entry.Name())
			}
			nextPrefix := prefix
			if isLast {
				nextPrefix += "    "
			} else {
				nextPrefix += "│   "
			}
			if err := w.walkDir(subBasePath, subRelPath, nextPrefix); err != nil {
				return err
			}
		}
	}
	return nil
}

func contains(slice []string, item string) bool {
	for _, v := range slice {
		if v == item {
			return true
		}
	}
	return false
}
