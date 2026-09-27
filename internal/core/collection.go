package core

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/fool-pigeon/pigeon/internal/config"
	"github.com/fool-pigeon/pigeon/internal/fsx"
)

// create a .pg/ of collection in dir
func InitCollection(dir string) error {
	info, err := os.Lstat(dir)
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return fmt.Errorf("%s:%w", dir, fsx.ErrNotDir)
	}
	// create root tag dir
	tagDir := filepath.Join(dir, fsx.RootTagDirName)
	if err := fsx.Mkdir(tagDir); err != nil {
		return err
	}
	// create envs and runs dir
	for _, d := range []string{filepath.Join(tagDir, fsx.EnvDirName), filepath.Join(tagDir, fsx.RunDirName)} {
		fsx.MkdirAll(d)
	}
	// create empty config
	if err := fsx.WriteToml(filepath.Join(tagDir, fsx.ConfigDefaultName), config.New()); err != nil {
		return err
	}

	return nil
}
