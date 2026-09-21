package service

import "github.com/wailsapp/wails/v3/pkg/application"

// A wrapper of [application.DialogManager] for using simplely
type PathSelectService struct {
	app *application.App
}

func NewPathSelectService(app *application.App) *PathSelectService {
	return &PathSelectService{app}
}
func (f *PathSelectService) ShowInFileManager(path string) {
	err := f.app.Env.OpenFileManager(path, true)
	if err != nil {
		f.app.Logger.Error("Could not open file manager", "path", path, "error", err)

		// Show error to user
		f.app.Dialog.Error().
			SetTitle("File Manager Error").
			SetMessage("Could not open file manager").
			Show()
	}
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
