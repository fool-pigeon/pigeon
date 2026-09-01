package fsx

import (
	"errors"
	"os"
)

func IsDirExist(name string) (bool, error) {
	f, err := os.Lstat(name)
	if errors.Is(err, ErrNotExist) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return f.IsDir(), nil
}
func IsDir(name string) bool {
	f, _ := os.Lstat(name)
	return f.IsDir()
}
func Mkdir(name string) error {
	return os.Mkdir(name, os.FileMode(0755))
}
