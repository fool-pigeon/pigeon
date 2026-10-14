// Copyright (c) 2026 Homoarea <homoarea114@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package service

import (
	"fmt"
	"path/filepath"

	"github.com/fool-pigeon/pigeon/cmd/pg-gui/service/reveal"
	"github.com/wailsapp/wails/v3/pkg/application"
)

// A wrapper of [application.DialogManager] for using simplely
type PathSelectService struct {
	app *application.App
}

func NewPathSelectService(app *application.App) *PathSelectService {
	return &PathSelectService{app}
}
func (f *PathSelectService) ShowInFileManager(path string) error {
	if err := reveal.Open(path); err == nil {
		return nil
	}
	// back to wails defult open
	err := f.app.Env.OpenFileManager(path, true)
	if err != nil {
		// Fallback: try opening just the directory
		dir := filepath.Dir(path)
		err = f.app.Env.OpenFileManager(dir, false)
		if err != nil {
			f.app.Logger.Error("Could not open file manager", "path", path, "error", err)

			// Show error to user
			return fmt.Errorf("could not open file manager\npath:%s\nerr:%w", path, err)
		}
	}
	return nil
}
func (f *PathSelectService) SelectFolder() (string, error) {
	return f.app.Dialog.OpenFile().
		CanChooseDirectories(true).
		CanChooseFiles(false).
		CanCreateDirectories(true).
		PromptForSingleSelection()
}
func (f *PathSelectService) SelectFolders() ([]string, error) {
	return f.app.Dialog.OpenFile().
		CanChooseDirectories(true).
		CanChooseFiles(false).
		CanCreateDirectories(true).
		PromptForMultipleSelection()
}
func (f *PathSelectService) SelectFile() (string, error) {
	return f.app.Dialog.OpenFile().
		CanChooseDirectories(false).
		CanChooseFiles(true).
		CanCreateDirectories(true).
		PromptForSingleSelection()
}
func (f *PathSelectService) SelectFiles() ([]string, error) {
	return f.app.Dialog.OpenFile().
		CanChooseDirectories(false).
		CanChooseFiles(true).
		CanCreateDirectories(true).
		PromptForMultipleSelection()
}
func (f *PathSelectService) SelectTomlFile() (string, error) {
	return f.app.Dialog.OpenFile().
		CanChooseDirectories(false).
		CanChooseFiles(true).
		CanCreateDirectories(true).
		AddFilter("Toml File", "*.toml").
		PromptForSingleSelection()
}
func (f *PathSelectService) SelectTomlFiles() ([]string, error) {
	return f.app.Dialog.OpenFile().
		CanChooseDirectories(false).
		CanChooseFiles(true).
		CanCreateDirectories(true).
		AddFilter("Toml File", "*.toml").
		PromptForMultipleSelection()
}
