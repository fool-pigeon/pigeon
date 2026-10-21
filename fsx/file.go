// Copyright (c) 2026 Homoarea <homoarea114@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package fsx

import (
	"errors"
	"os"

	"github.com/knadh/koanf/v2"
	"github.com/pelletier/go-toml/v2"
)

var (
	ErrInvalid    = os.ErrInvalid    // "invalid argument"
	ErrPermission = os.ErrPermission // "permission denied"
	ErrExist      = os.ErrExist      // "file already exists"
	ErrNotExist   = os.ErrNotExist   // "file does not exist"
	ErrClosed     = os.ErrClosed     // "file already closed"

	ErrNotDir  = errors.New("not dir type")
	ErrNotFile = errors.New("not file type")
)

func IsFileExist(name string) (bool, error) {
	f, err := os.Lstat(name)
	if errors.Is(err, ErrNotExist) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return f.Mode().IsRegular(), nil
}

func IsFile(name string) bool {
	f, _ := os.Lstat(name)
	return f.Mode().IsRegular()
}

func WriteFile(name string, data []byte) error {
	return os.WriteFile(name, data, os.FileMode(0644))
}

func CreateFileIfNotExists(path string) error {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0644)
	if err != nil {
		if errors.Is(err, os.ErrExist) {
			return nil
		}
		return err
	}
	return f.Close()
}

func ReadFile(name string) ([]byte, error) {
	return os.ReadFile(name)
}
func ReadFileAsString(name string) (string, error) {
	b, err := os.ReadFile(name)
	return string(b), err
}

func StoreKoanf(k *koanf.Koanf, parser koanf.Parser, name string) error {
	b, err := k.Marshal(parser)
	if err != nil {
		return err
	}
	if err := WriteFile(name, b); err != nil {
		return err
	}

	return nil
}

func ReadToml(name string, v interface{}) error {
	b, err := ReadFile(name)
	if err != nil {
		return err
	}
	if err := toml.Unmarshal(b, v); err != nil {
		return err
	}
	return nil
}
func WriteToml(name string, v interface{}) error {
	b, err := toml.Marshal(v)
	if err != nil {
		return err
	}
	if err := WriteFile(name, b); err != nil {
		return err
	}
	return nil
}

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
func MkdirAll(name string) error {
	return os.MkdirAll(name, os.FileMode(0755))
}
