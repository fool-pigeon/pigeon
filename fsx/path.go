// Copyright (c) 2026 Homoarea <homoarea114@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package fsx

import (
	"errors"
	"os"
	"path/filepath"
	"slices"
)

var ErrRootNotExist = errors.New("root tag directory is not existed")

const (
	RootTag             = ".pg"
	RootRun             = "run"
	RootConf            = "conf"
	DefaultRootConfFile = "conf.toml"

	Conf = ".conf"
)

// find the dir contains .pg dir in dirs from path to os' root dir
// (not include os' root dir and permission denied)
func RootDir(path string) (string, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}
	for {
		isExist, err := IsDirExist(filepath.Join(absPath, RootTag))

		if err == nil && isExist {
			return absPath, nil
		}

		// arrived the root dir of OS
		if filepath.Dir(absPath) == absPath {
			return "", ErrRootNotExist
		}

		absPath = filepath.Dir(absPath)
		absPath = filepath.Clean(absPath)
	}
}

// find the tag dir
func RootTagDir(path string) (string, error) {
	rootDir, err := RootDir(path)
	if err != nil {
		return "", err
	}
	return filepath.Join(rootDir, RootTag), nil
}

// return []string of paths form this path to RootDir
func PathsToRoot(path string) ([]string, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}
	res := make([]string, 0)
	for {
		isExist, err := IsDirExist(filepath.Join(absPath, RootTag))
		res = append(res, absPath)
		if err == nil && isExist {
			return res, nil
		}

		if filepath.Dir(absPath) == absPath {
			return nil, ErrRootNotExist
		}

		absPath = filepath.Dir(absPath)
	}
}

// return []string of paths form rootDir to this path
func PathsFromRoot(path string) ([]string, error) {
	paths, err := PathsToRoot(path)
	if err != nil {
		return nil, err
	}
	slices.Reverse(paths)
	return paths, nil
}

// get RootConf directory from current path
func RootConfDir(path string) (string, error) {
	return rootSubDir(path, Conf)
}
func RootRunDir(path string) (string, error) {
	return rootSubDir(path, RootRun)
}

// get sub directory of root from current path
func rootSubDir(path string, subDir string) (string, error) {
	tagDir, err := RootTagDir(path)
	if err != nil {
		return "", err
	}
	subPath := filepath.Join(tagDir, subDir)
	info, err := os.Lstat(subPath)
	if err != nil {
		return "", err
	}
	if !info.IsDir() {
		return "", ErrNotDir
	}
	return subPath, nil
}
