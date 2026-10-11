package main

import (
	"embed"

	"log"

	"github.com/fool-pigeon/pigeon/cmd/pg-gui/gui"
	"github.com/fool-pigeon/pigeon/cmd/pg-gui/service"
	"github.com/wailsapp/wails/v3/pkg/application"
)

// Wails uses Go's `embed` package to embed the frontend files into the binary.
// Any files in the frontend/dist folder will be embedded into the binary and
// made available to the frontend.
// See https://pkg.go.dev/embed for more information.

//go:embed all:frontend/dist
var assets embed.FS

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

	// Create a new Wails application by providing the necessary options.
	// Variables 'Name' and 'Description' are for application metadata.
	// 'Assets' configures the asset server with the 'FS' variable pointing to the frontend files.
	// 'Bind' is a list of Go struct instances. The frontend has access to the methods of these instances.
	// 'Mac' options tailor the application when running an macOS.
	app := application.New(application.Options{
		Name:        "Pigeon",
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
	app.RegisterService(application.NewService(service.NewHttpReqFileService(fileAccessService)))

	reqCtxService := service.NewReqCtxService(app)
	app.RegisterService(application.NewService(reqCtxService))

	taskRunService := service.NewTaskRunService(reqCtxService, alertService, app)
	app.RegisterService(application.NewService(taskRunService))

	// Create a new window with the necessary options.
	// 'Title' is the title of the window.
	// 'Mac' options tailor the window when running on macOS.
	// 'BackgroundColour' is the background colour of the window.
	// 'URL' is the URL that will be loaded into the webview.
	mainWindow := app.Window.NewWithOptions(application.WebviewWindowOptions{
		Title: "Pigeon",
		// Window sized to the golden ratio (1000 / 618 ≈ 1.618).
		Width:  1000,
		Height: 618,
		Mac: application.MacWindow{
			InvisibleTitleBarHeight: 50,
			Backdrop:                application.MacBackdropTranslucent,
			TitleBar:                application.MacTitleBarHiddenInset,
		},
		BackgroundColour: application.NewRGB(6, 7, 15),
		URL:              "/",
		// use localized menu for app
		UseApplicationMenu: true,
	})

	mainWindow.Center()
	gui.InitMenu(app)

	// Run the application. This blocks until the application has been exited.
	err := app.Run()

	// If an error occurred while running the application, log it and exit.
	if err != nil {
		log.Fatal(err)
	}
}
