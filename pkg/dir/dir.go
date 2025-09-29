package dir

import (
	"os"
	"path/filepath"
	"strings"
)

// List 返回给定目录的一层文件/目录名称（不含隐藏 .git 等）
func List(dir string) []string {
	fis, err := os.ReadDir(dir)
	if err != nil {
		return []string{}
	}
	var out []string
	for _, fi := range fis {
		// skip hidden start-with-dot entries (可根据需要修改)
		if strings.HasPrefix(fi.Name(), ".") {
			continue
		}
		out = append(out, fi.Name())
	}
	return out
}

// ListWithParent 返回给定目录的文件/目录名称，如果不在根目录则包含".."
func ListWithParent(dir string, workingDir string) []string {
	fis, err := os.ReadDir(dir)
	if err != nil {
		return []string{}
	}

	var out []string

	// 检查是否需要添加".."父目录条目
	cleanCurrentDir := filepath.Clean(dir)
	cleanWorkingDir := filepath.Clean(workingDir)

	// 如果当前目录不是工作目录，则添加".."
	if cleanCurrentDir != cleanWorkingDir {
		out = append(out, "..")
	}

	for _, fi := range fis {
		// // skip hidden start-with-dot entries (可根据需要修改)
		// if strings.HasPrefix(fi.Name(), ".") {
		// 	continue
		// }
		out = append(out, fi.Name())
	}
	return out
}
