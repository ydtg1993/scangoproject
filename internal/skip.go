package internal

import (
	"path/filepath"
	"strings"
)

type SkipMatcher struct {
	globalNames map[string]struct{} // ba
	rootPaths   map[string]struct{} // /ba
	fullPaths   map[string]struct{} // /b1/ba
	prefixMatch []string            // aa*
}

func NewSkipMatcher(rules []string) *SkipMatcher {
	m := &SkipMatcher{
		globalNames: make(map[string]struct{}),
		rootPaths:   make(map[string]struct{}),
		fullPaths:   make(map[string]struct{}),
	}

	for _, r := range rules {
		r = strings.TrimSpace(r)

		// 模糊匹配
		if strings.HasSuffix(r, "*") {
			m.prefixMatch = append(m.prefixMatch, strings.TrimSuffix(r, "*"))
			continue
		}

		// 绝对路径
		if strings.HasPrefix(r, "/") {
			parts := strings.Split(strings.TrimPrefix(r, "/"), "/")
			if len(parts) == 1 {
				m.rootPaths[parts[0]] = struct{}{}
			} else {
				m.fullPaths[strings.Join(parts, "/")] = struct{}{}
			}
			continue
		}

		// 普通名称
		m.globalNames[r] = struct{}{}
	}

	return m
}

func (m *SkipMatcher) Match(relPath string, name string) bool {
	// 1️⃣ 全局名称匹配
	if _, ok := m.globalNames[name]; ok {
		return true
	}

	// 2️⃣ 根目录匹配
	if relPath == "" {
		if _, ok := m.rootPaths[name]; ok {
			return true
		}
	}

	// 3️⃣ 完整路径匹配
	full := filepath.ToSlash(filepath.Join(relPath, name))
	if _, ok := m.fullPaths[full]; ok {
		return true
	}

	// 4️⃣ 前缀匹配
	for _, prefix := range m.prefixMatch {
		if strings.HasPrefix(name, prefix) {
			return true
		}
	}

	return false
}
