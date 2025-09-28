package dir

import (
	"os"
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
