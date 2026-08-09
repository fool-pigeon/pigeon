package fsx

import (
	"errors"
	"path/filepath"
	"slices"
)

var (
	ErrRootNotExist = errors.New("Not found root tag file")
)

const RootTagName = ".pg"

// find the dir contains .pg file in dirs from path to os' root dir
// (not include os' root dir and permission denied)
func RootDir(path string) (string, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}
	for {
		isFileExist, err := IsFileExist(filepath.Join(absPath, RootTagName))
		if err != nil {
			return "", err
		}
		if isFileExist {
			return absPath, nil
		}

		if filepath.Dir(absPath) == absPath {
			return "", ErrRootNotExist
		}

		absPath = filepath.Dir(absPath)
	}
}

// return []string of paths form this path to RootDir
func PathsToRoot(path string) ([]string, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}
	res := make([]string, 0)
	for {
		isFileExist, err := IsFileExist(filepath.Join(absPath, RootTagName))
		if err != nil {
			return nil, err
		}

		res = append(res, absPath)
		if isFileExist {
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
	if err == nil {
		slices.Reverse(paths)
	}
	return paths, err
}
