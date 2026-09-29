package service

import (
	"errors"
	"maps"
	"slices"
	"sync"

	"github.com/fool-pigeon/pigeon/internal/fsx"
	"github.com/fool-pigeon/pigeon/internal/request/httpx"
	"github.com/pelletier/go-toml/v2"
)

var (
	ErrOpened    = errors.New("the file is already opened.")
	ErrNotOpened = errors.New("the file is not opened yet.")
)

type OpenedState struct {
	Path  string `json:"path"`
	Saved bool   `json:"saved"`
}

type FileCache struct {
	mu    sync.RWMutex
	cache map[string][]byte
}

func (f *FileCache) find(path string) ([]byte, bool) {
	f.mu.RLock()
	defer f.mu.RUnlock()
	b, ok := f.cache[path]
	return b, ok
}
func (f *FileCache) hasKey(path string) bool {
	f.mu.RLock()
	defer f.mu.RUnlock()
	_, ok := f.cache[path]
	return ok
}
func (f *FileCache) set(path string, b []byte) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.cache == nil {
		f.cache = make(map[string][]byte)
	}
	f.cache[path] = b
}
func (f *FileCache) delete(path string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	delete(f.cache, path)
}
func (f *FileCache) copy() map[string][]byte {
	f.mu.RLock()
	defer f.mu.RUnlock()
	res := make(map[string][]byte)
	maps.Copy(res, f.cache)
	return res
}

// func (f *FileCache) length() int {
// 	f.mu.RLock()
// 	defer f.mu.RUnlock()
// 	return len(f.cache)
// }

type FileService struct {
	mu sync.RWMutex
	// Cache of original file content
	original FileCache
	// Cache of current temp file content that not saved
	temp FileCache
	// Store open file paths in open order
	opendPaths []string
}

func NewFileService() *FileService {
	return &FileService{}
}
func (f *FileService) Original() map[string][]byte {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.original.copy()
}
func (f *FileService) Temp() map[string][]byte {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.temp.copy()
}
func (f *FileService) GetOpenedState() ([]OpenedState, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()
	if len(f.opendPaths) <= 0 {
		return []OpenedState{}, nil
	}
	states := make([]OpenedState, 0, len(f.opendPaths))
	for _, path := range f.opendPaths {
		states = append(states, OpenedState{path, !f.temp.hasKey(path)})
	}
	return states, nil
}
func (f *FileService) Open(path string) error {
	_, err := f.open(path)
	return err
}
func (f *FileService) Close(path string) error {
	return f.close(path)
}
func (f *FileService) CloseAll() {
	for _, path := range f.opendPaths {
		go func(path string) {
			f.close(path)
		}(path)
	}
}
func (f *FileService) Save(path string) error {
	return f.save(path)
}
func (f *FileService) SaveAll() error {
	f.mu.RLock()
	paths := append([]string{}, f.opendPaths...)
	f.mu.RUnlock()
	var wg sync.WaitGroup
	for _, p := range paths {
		wg.Go(func() { f.save(p) })
	}
	wg.Wait()
	return nil
}
func (f *FileService) IsSaved(path string) bool {
	return !f.temp.hasKey(path)
}

// first read into cache and insert path to [FileService.openedPaths]
func (f *FileService) open(path string) ([]byte, error) {
	b, err := fsx.ReadFile(path)
	if err != nil {
		return nil, err
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	if b, ok := f.original.find(path); ok {
		return b, nil
	}
	f.original.set(path, b)
	if f.opendPaths == nil {
		f.opendPaths = make([]string, 0)
	}
	f.opendPaths = append(f.opendPaths, path)
	return b, nil
}

// read newest file content,the order is [FileService.temp]->[FileService.original]->disk
func (f *FileService) read(path string) ([]byte, error) {
	if b, ok := f.temp.find(path); ok {
		return b, nil
	}
	if b, ok := f.original.find(path); ok {
		return b, nil
	}
	return nil, ErrNotOpened
}

// read original file content,the order is [FileService.original]->disk
func (f *FileService) readOriginal(path string) ([]byte, error) {
	if b, ok := f.original.find(path); ok {
		return b, nil
	}
	return nil, ErrNotOpened
}

// write file content into [FileService.temp] if is opened
func (f *FileService) stage(path string, b []byte) error {
	if f.original.hasKey(path) {
		f.temp.set(path, b)
		return nil
	}
	return errors.Join(errors.New("stage failed"), ErrNotOpened)
}

// write file content to disk
func (f *FileService) write(path string, b []byte) error {
	if err := fsx.WriteFile(path, b); err != nil {
		return err
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.original.hasKey(path) {
		f.original.set(path, b)
	}
	f.temp.delete(path)
	return nil
}

// write file content of temp into disk
func (f *FileService) save(path string) error {
	if !f.temp.hasKey(path) {
		return errors.New("not in the temp cache.")
	}
	if !f.original.hasKey(path) {
		return ErrNotOpened
	}

	if b, ok := f.temp.find(path); ok {
		return f.write(path, b)
	}
	return nil
}

// remove all cache whose path equals this and remove from [FileService.openedPaths]
func (f *FileService) close(path string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if !f.original.hasKey(path) {
		return ErrNotOpened
	}
	f.temp.delete(path)
	f.original.delete(path)
	f.opendPaths = slices.DeleteFunc(f.opendPaths, func(p string) bool { return p == path })
	return nil
}

// the wrapper of FileService
type FileServiceWrapper[T any] interface {
	Read(path string) (T, error)
	ReadOriginal(path string) (T, error)
	Write(path string, data T) (NodeMode, error)
	Stage(path string, data T) (NodeMode, error)
}

// FileServiceWrapper[string]
type TextFileService struct {
	fs *FileService
}

func NewTextFileService(fs *FileService) *TextFileService {
	return &TextFileService{fs}
}
func (t *TextFileService) Read(path string) (string, error) {
	b, err := t.fs.read(path)
	return string(b), err
}

func (t *TextFileService) ReadOriginal(path string) (string, error) {
	b, err := t.fs.readOriginal(path)
	return string(b), err
}

func (t *TextFileService) Write(path string, data string) (NodeMode, error) {
	if err := t.fs.write(path, []byte(data)); err != nil {
		return 0, err
	}
	return getFileMode([]byte(data)), nil
}

func (t *TextFileService) Stage(path string, data string) (NodeMode, error) {
	if err := t.fs.stage(path, []byte(data)); err != nil {
		return 0, err
	}
	return getFileMode([]byte(data)), nil
}

// FileServiceWrapper[http.ReqData]
type HttpReqFileService struct {
	fs *FileService
}

func NewHttpReqFileService(fs *FileService) *HttpReqFileService {
	return &HttpReqFileService{fs}
}
func (r *HttpReqFileService) Read(path string) (*httpx.ReqData, error) {
	b, err := r.fs.read(path)
	if err != nil {
		return nil, err
	}
	var req httpx.ReqData
	if err := toml.Unmarshal(b, &req); err != nil {
		return nil, err
	}
	return &req, nil
}

func (r *HttpReqFileService) ReadOriginal(path string) (*httpx.ReqData, error) {
	b, err := r.fs.readOriginal(path)
	if err != nil {
		return nil, err
	}
	var req httpx.ReqData
	if err := toml.Unmarshal(b, &req); err != nil {
		return nil, err
	}
	return &req, nil
}

func (r *HttpReqFileService) Write(path string, data *httpx.ReqData) (NodeMode, error) {
	b, err := toml.Marshal(data)
	if err != nil {
		return 0, err
	}
	r.fs.write(path, b)
	return FileMode | Request | parseModeFromHttpReq(data), nil
}

func (r *HttpReqFileService) Stage(path string, data *httpx.ReqData) (NodeMode, error) {
	b, err := toml.Marshal(data)
	if err != nil {
		return 0, err
	}
	r.fs.stage(path, b)
	// temp := r.fs.Temp()
	// pterm.Warning.Println("TEMP:")
	// for k, v := range temp {
	// 	pterm.Info.Println(k, string(v))
	// }
	return FileMode | Request | parseModeFromHttpReq(data), nil
}
