package main

import (
	"ask/internal"
	"bufio"
	"fmt"
	"gopkg.in/yaml.v3"
	"os"
	"path/filepath"
	"strings"
)

type Config struct {
	AimDir   string   `yaml:"aimDir"`
	Output   string   `yaml:"output"`
	Exts     []string `yaml:"exts"`
	SkipDirs []string `yaml:"skipDirs"`
}

func main() {
	cfg, err := loadConfig("config.yaml")
	if err != nil {
		fmt.Println("读取配置失败:", err)
		os.Exit(1)
	}

	if err := NewWriter(cfg); err != nil {
		fmt.Println("执行失败:", err)
		os.Exit(1)
	}
}

func loadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

type Writer struct {
	writer   *bufio.Writer
	baseDir  string
	exts     []string
	fileList []string
	skipRule *internal.SkipMatcher
}

func (w *Writer) walkDir(basePath, relPath, prefix string) error {
	entries, err := os.ReadDir(basePath)
	if err != nil {
		return err
	}

	var filtered []os.DirEntry
	for _, e := range entries {
		if e.IsDir() && w.skipRule.Match(relPath, e.Name()) {
			continue
		}
		filtered = append(filtered, e)
	}

	for i, entry := range filtered {
		isLast := i == len(filtered)-1

		line := prefix
		if isLast {
			line += "└── "
		} else {
			line += "├── "
		}
		line += entry.Name()

		if entry.IsDir() {
			line += "/"
		}

		fmt.Fprintln(w.writer, line)

		if !entry.IsDir() {
			ext := filepath.Ext(entry.Name())
			if contains(w.exts, ext) {
				var rel string
				if relPath == "" {
					rel = entry.Name()
				} else {
					rel = filepath.Join(relPath, entry.Name())
				}
				w.fileList = append(w.fileList, filepath.ToSlash(rel))
			}
		}

		if entry.IsDir() {
			var subRel string
			if relPath == "" {
				subRel = entry.Name()
			} else {
				subRel = filepath.Join(relPath, entry.Name())
			}

			nextPrefix := prefix
			if isLast {
				nextPrefix += "    "
			} else {
				nextPrefix += "│   "
			}

			if err := w.walkDir(filepath.Join(basePath, entry.Name()), subRel, nextPrefix); err != nil {
				return err
			}
		}
	}

	return nil
}

func (w *Writer) writeFiles() {
	if len(w.fileList) == 0 {
		return
	}

	fmt.Fprintln(w.writer, "\nFile contents:")
	fmt.Fprintln(w.writer, strings.Repeat("=", 50))

	for _, rel := range w.fileList {
		fmt.Fprintf(w.writer, "\nFile: %s\n", rel)
		fmt.Fprintln(w.writer, strings.Repeat("-", 40))

		full := filepath.Join(w.baseDir, rel)
		data, err := os.ReadFile(full)
		if err != nil {
			fmt.Fprintf(w.writer, "[读取失败: %v]\n", err)
			continue
		}

		w.writer.Write(data)
		if len(data) > 0 && data[len(data)-1] != '\n' {
			w.writer.WriteString("\n")
		}
	}
}

func NewWriter(cfg *Config) error {
	f, err := os.Create(cfg.Output)
	if err != nil {
		return err
	}
	defer f.Close()

	writer := bufio.NewWriter(f)
	defer writer.Flush()

	w := &Writer{
		writer:   writer,
		baseDir:  cfg.AimDir,
		exts:     cfg.Exts,
		skipRule: internal.NewSkipMatcher(cfg.SkipDirs),
	}

	if err := w.walkDir(cfg.AimDir, "", ""); err != nil {
		return err
	}

	w.writeFiles()

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
