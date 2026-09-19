package service

import (
	"sync"

	"github.com/fool-pigeon/pigeon/internal/fsx"
)

type FileService struct {
	cache map[string][]byte
	mu    sync.RWMutex
}

func (f *FileService) findCache(path string) ([]byte, bool) {
	f.mu.RLock()
	defer f.mu.RUnlock()
	b, ok := f.cache[path]
	return b, ok
}
func (f *FileService) DeleteCache(path string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	delete(f.cache, path)
}
func (f *FileService) SetCache(path string, b []byte) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.cache == nil {
		f.cache = make(map[string][]byte)
	}
	f.cache[path] = b
}
func (f *FileService) ReadFile(path string) (string, error) {
	if b, ok := f.findCache(path); ok {
		return string(b), nil
	}

	// read file to memory
	b, err := fsx.ReadFile(path)
	if err != nil {
		return "", err
	}

	f.SetCache(path, b)
	return string(b), nil
}
