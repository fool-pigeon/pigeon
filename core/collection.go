// Copyright (c) 2026 Homoarea <homoarea114@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package core

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/fool-pigeon/pigeon/fsx"
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
	tagDir := filepath.Join(dir, fsx.RootTag)
	if err := fsx.Mkdir(tagDir); err != nil {
		return err
	}
	// create conf directory and run directory
	for _, d := range []string{filepath.Join(tagDir, fsx.RootConf), filepath.Join(tagDir, fsx.RootRun)} {
		fsx.MkdirAll(d)
	}
	// create empty config
	if err := fsx.WriteToml(filepath.Join(tagDir, fsx.RootConf, fsx.DefaultRootConfFile), RootConf{}); err != nil {
		return err
	}

	return nil
}
