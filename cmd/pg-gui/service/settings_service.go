// Copyright (c) 2026 Homoarea <homoarea114@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package service

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sync"

	"github.com/fool-pigeon/pigeon/cmd/pg-gui/gui"
	"github.com/fool-pigeon/pigeon/internal/fsx"
	"github.com/wailsapp/wails/v3/pkg/application"
)

type SettingsServiceEvent = string

const (
	SettingChanged SettingsServiceEvent = "settings-service:settings-changed"
)

type ThemeMode = string

const (
	ThemeAuto  ThemeMode = ""
	ThemeLight ThemeMode = "light"
	ThemeDark  ThemeMode = "dark"
)

type ThemeSettings struct {
	Mode ThemeMode `json:"mode" toml:"mode"`
}
type EditorSettings struct {
	PreventWrapLine bool `json:"preventWrapLine" toml:"prevent_wrap_line"`
}
type InterfaceSettings struct {
	// DispalyFontSize int

	// EditorFontSize int
	Editor EditorSettings `json:"editor" toml:"editor"`
}
type Settings struct {
	Theme     ThemeSettings     `json:"theme" toml:"theme"`
	Interface InterfaceSettings `json:"interface" toml:"interface"`
}
type SettingsService struct {
	settings Settings

	app *application.App
	mu  sync.RWMutex
}

func NewSettingsService(app *application.App) *SettingsService {
	return &SettingsService{app: app}
}

// ServiceStartup implements [application.ServiceStartup].
func (ss *SettingsService) ServiceStartup(ctx context.Context, options application.ServiceOptions) error {
	if err := ss.init(); err != nil {
		return err
	}
	return nil
}

// ServiceShutdown implements [application.ServiceShutdown].
func (ss *SettingsService) ServiceShutdown() error {
	return fsx.WriteToml(gui.Paths.ConfigFile(), ss.settings)
}

func (ss *SettingsService) init() error {
	if err := fsx.MkdirAll(gui.Paths.Config()); err != nil {
		return err
	}

	configFile := gui.Paths.ConfigFile()
	_, err := os.Lstat(configFile)
	if errors.Is(err, os.ErrNotExist) {
		if err := fsx.WriteToml(configFile, ss.settings); err != nil {
			return err
		}
		return nil
	} else if err != nil {
		return fmt.Errorf("[SettingsService] %w", err)
	}

	if err := fsx.ReadToml(configFile, &ss.settings); err != nil {
		return err
	}

	ss.app.Event.On(gui.SettingsWindowClosed, func(event *application.CustomEvent) {
		ss.app.Event.Emit(SettingChanged, ss.GetSettings())
	})

	return nil
}
func (ss *SettingsService) GetSettings() Settings {
	ss.mu.RLock()
	defer ss.mu.RUnlock()
	return ss.settings
}
func (ss *SettingsService) SetSettings(settings Settings) {
	ss.mu.Lock()
	defer ss.mu.Unlock()
	ss.settings = settings
	ss.app.Event.Emit(SettingChanged, ss.settings)
}

type PathsInfo struct {
	Config string `json:"config"`
	Log    string `json:"log"`
	Cache  string `json:"cache"`
}

func (ss *SettingsService) GetPathsInfo() PathsInfo {
	return PathsInfo{
		Config: gui.Paths.Config(),
		Log:    gui.Paths.Log(),
		Cache:  gui.Paths.Cache(),
	}
}
