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

	"github.com/fool-pigeon/pigeon/core"
	"github.com/fool-pigeon/pigeon/fsx"
	"github.com/wailsapp/wails/v3/pkg/application"
)

// batch script wrapper for [FileService] nad [FileExploreService]
type FileAccessService struct {
	fs           *FileService
	fes          *FileExploreService
	pathSelector *PathSelectService
	app          *application.App
}

func NewFileAccessService(fs *FileService, fes *FileExploreService, pathSelector *PathSelectService, app *application.App) *FileAccessService {
	return &FileAccessService{fs, fes, pathSelector, app}
}

// File Read
func (fa *FileAccessService) Read(path string) ([]byte, error) {
	return fa.fs.read(path)
}

// File ReadOriginal
func (fa *FileAccessService) ReadOriginal(path string) ([]byte, error) {
	return fa.fs.readOriginal(path)
}

// File Stage
func (fa *FileAccessService) Stage(path string, b []byte) (NodeMode, error) {
	if err := fa.fs.stage(path, b); err != nil {
		return 0, err
	}
	return getFileMode(b), nil
}

// File Write
func (fa *FileAccessService) Write(path string, b []byte) (NodeMode, error) {
	if err := fa.fs.write(path, b); err != nil {
		return 0, err
	}
	mode := getFileMode(b)
	go func() {
		if node := fa.fes.FindNode(path); node != nil {
			node.SetMode(getFileMode(b))
		}
	}()
	return mode, nil
}

// Rename Dir/File the path's item to name
// for example: ("/a/b","c") -> "/a/c"
func (fa *FileAccessService) Rename(path string, name string) error {
	path = filepath.Clean(path)
	node, err := fa.fes.GetNodeExact(path)
	if err != nil {
		return fmt.Errorf("node not found:%w", err)
	}
	if node.Pre().GetPath() == "" {
		return errors.New("can not rename collection")
	}

	parent := filepath.Dir(path)
	newPath := filepath.Join(parent, name)
	if _, err := os.Lstat(newPath); err == nil {
		return fmt.Errorf("%s:%w", newPath, os.ErrExist)
	} else if !errors.Is(err, os.ErrNotExist) {
		return err
	}
	if err := os.Rename(path, newPath); err != nil {
		return fmt.Errorf("os rename failed:%w", err)
	}
	fa.fes.ReloadPreCached(path, FesRename)
	fa.fs.renameKey(path, newPath)
	return nil
}

// Remove file or empty directory in collection
func (fa *FileAccessService) Delete(path string) error {
	path = filepath.Clean(path)
	if _, err := fsx.RootDir(path); err != nil {
		return fmt.Errorf("%s:%w", path, err)
	}
	if err := os.Remove(path); err != nil {
		return fmt.Errorf("os remove failed:%w", err)
	}
	go func() {
		if err := fa.fs.Close(path); err != nil {
			return
		}
		fa.app.Event.Emit(FsUpdatePaths, fa.fs.GetOpenedPaths())
	}()
	go func() {
		fa.fes.DeleteNode(path)
	}()
	return nil
}

// Remove all of path,danager operation
func (fa *FileAccessService) DeleteAll(path string) error {
	path = filepath.Clean(path)
	if _, err := fsx.RootDir(path); err != nil {
		return fmt.Errorf("%s:%w", path, err)
	}
	if err := os.RemoveAll(path); err != nil {
		return fmt.Errorf("os remove failed:%w", err)
	}
	go func() {
		if err := fa.fs.Close(path); err != nil {
			return
		}
		fa.app.Event.Emit(FsUpdatePaths, fa.fs.GetOpenedPaths())
	}()
	go func() {
		fa.fes.DeleteAllNode(path)
	}()
	return nil
}

// Unmount the collection from workspace
func (fa *FileAccessService) Unmount(path string) error {
	path = filepath.Clean(path)
	if _, err := fsx.RootDir(path); err != nil {
		return fmt.Errorf("%s:%w", path, err)
	}
	if !fa.fes.root.SubContain(path) {
		return fmt.Errorf("%s is not collection", path)
	}
	go func() {
		fa.fes.root.DelSub(path)
	}()
	return nil
}

// select a directory and init collection
func (fa *FileAccessService) CreateCollection() error {
	path, err := fa.pathSelector.SelectFolder()
	if err != nil {
		return fmt.Errorf("select path failed:%w", err)
	}
	path = filepath.Clean(path)
	info, err := os.Lstat(path)
	if err != nil {
		return fmt.Errorf("info get failed:%w", err)
	}
	if !info.IsDir() {
		return fmt.Errorf("%s:%w", path, fsx.ErrNotDir)
	}
	if err := core.InitCollection(path); err != nil {
		return fmt.Errorf("collection init failed:%w", err)
	}
	return fa.fes.AddRoot([]string{path})
}
