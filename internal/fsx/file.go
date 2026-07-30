package fsx

import (
	"errors"
	"os"

	"github.com/knadh/koanf/v2"
)

func IsFileExist(name string) (bool, error) {
	f, err := os.Lstat(name)
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	if f.Mode().IsRegular() {
		return true, nil
	}
	return false, nil
}

func StoreKoanf(k *koanf.Koanf, parser koanf.Parser, name string) error {
	b, err := k.Marshal(parser)
	if err != nil {
		return err
	}
	if err := os.WriteFile(name, b, os.FileMode(0666)); err != nil {
		return err
	}

	return nil
}
