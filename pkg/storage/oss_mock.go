package storage

import (
	"os"
	"path/filepath"
	"strings"
	"sync"
)

var _ Client = (*ossMock)(nil)

type ossMock struct {
	base string
	mu   sync.Mutex
}

func NewOSSMock(base string) Client {
	if err := os.MkdirAll(base, 0o755); err != nil {
		panic(err)
	}
	return &ossMock{base: base}
}

func (o *ossMock) List(prefix string) ([]string, error) {
	o.mu.Lock()
	defer o.mu.Unlock()
	var out []string
	err := filepath.Walk(o.base, func(p string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(o.base, p)
		if err != nil {
			return err
		}
		rel = filepath.ToSlash(rel)
		if prefix == "" || strings.HasPrefix(rel, prefix) {
			out = append(out, rel)
		}
		return nil
	})
	return out, err
}

func (o *ossMock) keyPath(key string) string {
	return filepath.Join(o.base, filepath.FromSlash(key))
}

func (o *ossMock) Upload(key string, data []byte) error {
	o.mu.Lock()
	defer o.mu.Unlock()
	p := o.keyPath(key)
	dir := filepath.Dir(p)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	return os.WriteFile(p, data, 0o644)
}

func (o *ossMock) Download(key string) ([]byte, error) {
	o.mu.Lock()
	defer o.mu.Unlock()
	p := o.keyPath(key)
	return os.ReadFile(p)
}

func (o *ossMock) Delete(key string) error {
	return nil
}
