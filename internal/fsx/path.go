package fsx

import (
	"errors"
	"path/filepath"
	"slices"
)

var ErrRootNotExist = errors.New("root tag directory is not existed.")

const RootTagDirName = ".pg"

// find the dir contains .pg dir in dirs from path to os' root dir
// (not include os' root dir and permission denied)
func RootDir(path string) (string, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}
	for {
		isExist, err := IsDirExist(filepath.Join(absPath, RootTagDirName))
		if err != nil {
			return "", err
		}
		if isExist {
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
	return filepath.Join(rootDir, RootTagDirName), nil
}

// return []string of paths form this path to RootDir
func PathsToRoot(path string) ([]string, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}
	res := make([]string, 0)
	for {
		isExist, err := IsDirExist(filepath.Join(absPath, RootTagDirName))
		if err != nil {
			return nil, err
		}

		res = append(res, absPath)
		if isExist {
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
