// Copyright (c) 2026 Homoarea <homoarea114@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package main

import (
	"embed"
	"errors"
	"sync/atomic"

	"log"

	"github.com/fool-pigeon/pigeon/cmd/pg-gui/gui"
	"github.com/fool-pigeon/pigeon/cmd/pg-gui/service"
	"github.com/wailsapp/wails/v3/pkg/application"
	"github.com/wailsapp/wails/v3/pkg/events"
)

// Wails uses Go's `embed` package to embed the frontend files into the binary.
// Any files in the frontend/dist folder will be embedded into the binary and
// made available to the frontend.
// See https://pkg.go.dev/embed for more information.

//go:embed all:frontend/dist
var assets embed.FS

type AppWrapper struct {
	app            *application.App
	mainWindow     *application.WebviewWindow
	settingsWindow *application.WebviewWindow

	shouldQuit atomic.Bool
}

// Create AppWrapper,set the app basic settings and windows' menu
func NewAppWrapper() *AppWrapper {
	aw := &AppWrapper{}

	// Create a new Wails application by providing the necessary options.
	// Variables 'Name' and 'Description' are for application metadata.
	// 'Assets' configures the asset server with the 'FS' variable pointing to the frontend files.
	// 'Bind' is a list of Go struct instances. The frontend has access to the methods of these instances.
	// 'Mac' options tailor the application when running an macOS.
	aw.app = application.New(application.Options{
		Name:   "pigeon",
		Logger: gui.NewLogger(),

		Description: "Version: 0.1\ngo: 1.26.5\nwails: v3/beta\n",
		Services: []application.Service{
			application.NewService(&service.EventConstExposeService{}),
		},
		Assets: application.AssetOptions{
			Handler: application.AssetFileServerFS(assets),
		},
		Mac: application.MacOptions{
			ApplicationShouldTerminateAfterLastWindowClosed: true,
		},
	})
	app := aw.app

	// Create a new window with the necessary options.
	// 'Title' is the title of the window.
	// 'Mac' options tailor the window when running on macOS.
	// 'BackgroundColour' is the background colour of the window.
	// 'URL' is the URL that will be loaded into the webview.
	aw.mainWindow = app.Window.NewWithOptions(application.WebviewWindowOptions{
		URL:   "/#/",
		Title: "Pigeon",
		// Window sized to the golden ratio (1000 / 618 ≈ 1.618).
		Width: 1000, Height: 618,
		Mac: application.MacWindow{
			Backdrop: application.MacBackdropTranslucent,
			TitleBar: application.MacTitleBarHiddenInset,

			InvisibleTitleBarHeight: 50,
		},
		BackgroundColour: application.NewRGB(6, 7, 15),
		// use localized menu for app
		UseApplicationMenu: true,
	})
	aw.mainWindow.SetMenu(gui.MainMenu(app))
	aw.mainWindow.Center()

	aw.settingsWindow = app.Window.NewWithOptions(application.WebviewWindowOptions{
		URL:  "/#/settings",
		Name: "settings", Title: "Settings",
		Width: 600, Height: 400,
		Hidden: true,
	})
	return aw
}

// Register the hooks,listeners and events betweeen windows
func (aw *AppWrapper) init() error {
	if aw.app == nil || aw.mainWindow == nil || aw.settingsWindow == nil {
		return errors.New("the app wrapper's attr is nil")
	}
	type WrapperEvent = string

	// app will quit when close mainWindow
	aw.mainWindow.OnWindowEvent(events.Common.WindowClosing, func(event *application.WindowEvent) {
		aw.shouldQuit.Store(true)
		aw.app.Quit()
	})

	// settingsWindow will hide when close except the mainWindow closed
	aw.settingsWindow.RegisterHook(events.Common.WindowClosing, func(e *application.WindowEvent) {
		if aw.shouldQuit.Load() {
			return
		}
		aw.settingsWindow.Hide()
		e.Cancel()
		aw.app.Event.Emit(gui.SettingsWindowClosed)
	})
	// register event listener to show settingsWindow
	aw.app.Event.On(gui.Preferences, func(event *application.CustomEvent) {
		aw.showSettingsWindow()
	})
	// register event to close settingsWindow
	aw.app.Event.On(gui.CloseSettingsWindow, func(event *application.CustomEvent) {
		aw.settingsWindow.Close()
	})

	return nil
}
func (aw *AppWrapper) showSettingsWindow() {
	if aw.settingsWindow != nil {
		aw.settingsWindow.Show()
		aw.settingsWindow.Focus()
	}
}
func RegisterServices(app *application.App) {

	app.RegisterService(application.NewService(service.NewSettingsService(app)))

	db := service.NewDatabase()
	app.RegisterService(application.NewService(db))

	alertService := service.NewAlertService(app)

	pathSelectService := service.NewPathSelectService(app)
	app.RegisterService(application.NewService(pathSelectService))

	fileService := service.NewFileService()
	app.RegisterService(application.NewService(fileService))

	fileExploreService := service.NewFileExploreService(alertService, app)
	app.RegisterService(application.NewService(fileExploreService))

	fileAccessService := service.NewFileAccessService(fileService, fileExploreService, pathSelectService, app)
	app.RegisterService(application.NewService(fileAccessService))
	app.RegisterService(application.NewService(service.NewTextFileService(fileAccessService)))
	app.RegisterService(application.NewService(service.NewReqFileService(fileAccessService)))

	reqCtxService := service.NewReqCtxService(app)
	app.RegisterService(application.NewService(reqCtxService))

	taskRunService := service.NewTaskRunService(alertService, app)
	app.RegisterService(application.NewService(taskRunService))
}
func init() {
	// Register a custom event whose associated data type is string.
	// This is not required, but the binding generator will pick up registered events
	// and provide a strongly typed JS/TS API for them.
	// For exmaple:
	// 		application.RegisterEvent[string]("time")
}

// main function serves as the application's entry point. It initializes the application, creates a window,
// and starts a goroutine that emits a time-based event every second. It subsequently runs the application and
// logs any error that might occur.
func main() {

	if err := gui.InitLogFile(); err != nil {
		log.Fatal(err)
	}

	aw := NewAppWrapper()
	aw.init()
	// Register Services
	RegisterServices(aw.app)
	// Run the application. This blocks until the application has been exited.
	err := aw.app.Run()
	// If an error occurred while running the application, log it and exit.
	if err != nil {
		log.Fatal(err)
	}
}
