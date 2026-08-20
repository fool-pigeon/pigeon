package fsx

import (
	"os"

	"github.com/knadh/koanf/v2"
	"github.com/pelletier/go-toml/v2"
)

func IsFileExist(name string) (bool, error) {
	f, err := os.Lstat(name)
	if err != nil {
		return false, err
	}
	return f.Mode().IsRegular(), nil
}

func IsDirExist(name string) (bool, error) {
	f, err := os.Lstat(name)
	if err != nil {
		return false, err
	}

	return f.IsDir(), nil
}

func IsFile(name string) bool {
	f, _ := os.Lstat(name)
	return f.Mode().IsRegular()
}

func ReadFile(name string) ([]byte, error) {
	return os.ReadFile(name)
}
func ReadFileAsString(name string) (string, error) {
	b, err := os.ReadFile(name)
	return string(b), err
}
func WriteFile(name string, data []byte) error {
	return os.WriteFile(name, data, os.FileMode(0666))
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
