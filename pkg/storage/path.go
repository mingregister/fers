package storage

import (
	"path/filepath"
	"strings"
)

func NormalizePath(path string) string {
	if path == "" || path == "." {
		return ""
	}

	// 使用filepath.Clean清理路径（兼容不同操作系统）
	path = filepath.Clean(path)

	// 转换为Unix风格的斜杠（OSS使用/作为路径分隔符）
	path = filepath.ToSlash(path)

	// 移除开头和结尾的斜杠，但保留中间路径
	path = strings.Trim(path, "/")

	// 如果清理后为空，返回空字符串
	if path == "." || path == "" {
		return ""
	}

	return path
}
