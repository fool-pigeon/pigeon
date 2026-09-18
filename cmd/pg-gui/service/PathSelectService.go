package service

import "github.com/wailsapp/wails/v3/pkg/application"

// A wrapper of [application.DialogManager] for using simplely
type PathSelectService struct {
	dialog *application.DialogManager
}

func NewPathSelectService() *PathSelectService {
	return &PathSelectService{}
}

func (f *PathSelectService) SelectFolder() (string, error) {
	return f.dialog.OpenFile().
		CanChooseDirectories(true).
		CanChooseFiles(false).
		CanCreateDirectories(true).
		PromptForSingleSelection()
}
func (f *PathSelectService) SelectFolders() ([]string, error) {
	return f.dialog.OpenFile().
		CanChooseDirectories(true).
		CanChooseFiles(false).
		CanCreateDirectories(true).
		PromptForMultipleSelection()
}
func (f *PathSelectService) SelectFile() (string, error) {
	return f.dialog.OpenFile().
		CanChooseDirectories(false).
		CanChooseFiles(true).
		CanCreateDirectories(true).
		PromptForSingleSelection()
}
func (f *PathSelectService) SelectFiles() ([]string, error) {
	return f.dialog.OpenFile().
		CanChooseDirectories(false).
		CanChooseFiles(true).
		CanCreateDirectories(true).
		PromptForMultipleSelection()
}
func (f *PathSelectService) SelectTomlFile() (string, error) {
	return f.dialog.OpenFile().
		CanChooseDirectories(false).
		CanChooseFiles(true).
		CanCreateDirectories(true).
		AddFilter("Toml File", "*.toml").
		PromptForSingleSelection()
}
func (f *PathSelectService) SelectTomlFiles() ([]string, error) {
	return f.dialog.OpenFile().
		CanChooseDirectories(false).
		CanChooseFiles(true).
		CanCreateDirectories(true).
		AddFilter("Toml File", "*.toml").
		PromptForMultipleSelection()
}
