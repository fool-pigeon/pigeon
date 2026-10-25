// Copyright (c) 2026 Homoarea <homoarea114@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package service

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/fool-pigeon/pigeon/fsx"
	"github.com/wailsapp/wails/v3/pkg/application"
)

var (
	ErrNilValue = errors.New("the value is nil.")
)

type FileMeta struct {
	Name string `json:"name"`
	Path string `json:"path"`
}
type ReqCtxService struct {
	//current collection root path
	root string
	// root -> RootConf File path
	rootConfMap map[string]FileMeta

	app *application.App
	mu  sync.RWMutex
}

func NewReqCtxService(app *application.App) *ReqCtxService {
	return &ReqCtxService{app: app, rootConfMap: map[string]FileMeta{}}
}

// Switch root from path and returns root
func (r *ReqCtxService) SetRootFromPath(path string) (string, error) {
	root, err := fsx.RootDir(path)
	if err != nil {
		return "", err
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	r.root = root
	return root, nil
}
func (r *ReqCtxService) SetRootConfMeta(f FileMeta) error {
	root, err := fsx.RootDir(f.Path)
	if err != nil {
		return err
	}
	info, err := os.Lstat(f.Path)
	if err != nil {
		return err
	}
	if info.IsDir() {
		return fsx.ErrNotFile
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.rootConfMap[root] = f
	return nil
}
func (r *ReqCtxService) GetRootConfFiles(path string) ([]FileMeta, error) {
	root, err := fsx.RootDir(path)
	if err != nil {
		return nil, err
	}
	rootConf := filepath.Join(root, fsx.RootTag, fsx.RootConf)
	entries, err := os.ReadDir(rootConf)
	if err != nil {
		return nil, fmt.Errorf("%s:%w", rootConf, err)
	}
	res := make([]FileMeta, 0)
	for _, entry := range entries {
		if entry.Type().IsRegular() {
			res = append(res, FileMeta{entry.Name(), filepath.Join(rootConf, entry.Name())})
		}
	}
	return res, nil
}
