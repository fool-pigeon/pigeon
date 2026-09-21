package service

import (
	"sync"

	"github.com/fool-pigeon/pigeon/internal/fsx"
)

type FileService struct {
	cache map[string][]byte
	mu    sync.RWMutex
}

func NewFileService() *FileService {
	return &FileService{}
}
func (f *FileService) findCache(path string) ([]byte, bool) {
	f.mu.RLock()
	defer f.mu.RUnlock()
	b, ok := f.cache[path]
	return b, ok
}
func (f *FileService) deleteCache(path string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	delete(f.cache, path)
}
func (f *FileService) setCache(path string, b []byte) {
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

	f.setCache(path, b)
	return string(b), nil
}
func (f *FileService) CloseFile(path string) error {
	f.deleteCache(path)
	return nil
}
func (f *FileService) WriteFile(path string, data string) error {
	b := []byte(data)
	if err := fsx.WriteFile(path, b); err != nil {
		return err
	}
	f.setCache(path, b)
	return nil
}
