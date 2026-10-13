// Copyright (c) 2026 Homoarea <homoarea114@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package gui

import (
	"path/filepath"

	"github.com/wailsapp/wails/v3/pkg/application"
)

const (
	AppName    = "pigeon"
	configFile = "config.toml"
	logFile    = "pigeon.log"
)

var Paths PathsConfig

type PathsConfig struct{}

func (p PathsConfig) Config() string {
	return filepath.Join(application.Path(application.PathConfigHome), AppName)
}
func (p PathsConfig) ConfigFile() string {
	return filepath.Join(p.Config(), configFile)
}
func (p PathsConfig) Log() string {
	return filepath.Join(application.Path(application.PathStateHome), AppName)
}
func (p PathsConfig) LogFile() string {
	return filepath.Join(p.Log(), logFile)
}
func (p PathsConfig) Cache() string {
	return filepath.Join(application.Path(application.PathCacheHome), AppName)
}
