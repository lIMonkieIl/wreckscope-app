package main

import (
	"embed"
	"log"

	"wreckscope/internal/app"
	"wreckscope/internal/settings"
	"wreckscope/internal/update"

	"github.com/wailsapp/wails/v3/pkg/application"
)

//go:embed all:ui/dist
var assets embed.FS

func main() {
	service := app.New(assets)

	windowsOpts := application.WindowsOptions{}
	if dataDir, err := settings.WebviewDataDir(); err != nil {
		log.Printf("WebView data dir: %v", err)
	} else {
		// Stable profile — bare WreckScope-windows-amd64.exe must not use %AppData%\<exe-name>.
		windowsOpts.WebviewUserDataPath = dataDir
	}

	var wailsApp *application.App
	wailsApp = application.New(application.Options{
		Name:        "WreckScope",
		Description: "WreckScope dashboard",
		Services: []application.Service{
			application.NewService(service),
		},
		Assets: application.AssetOptions{
			Handler: application.AssetFileServerFS(assets),
		},
		SingleInstance: &application.SingleInstanceOptions{
			UniqueID: "com.wreckscope.app",
			OnSecondInstanceLaunch: func(_ application.SecondInstanceData) {
				focusMainWindow(wailsApp)
			},
		},
		Windows: windowsOpts,
		Mac: application.MacOptions{
			ApplicationShouldTerminateAfterLastWindowClosed: true,
		},
	})

	if err := update.Configure(wailsApp); err != nil {
		log.Printf("Updater: %v", err)
	}

	wailsApp.Window.NewWithOptions(application.WebviewWindowOptions{
		Title:            "WreckScope",
		Width:            1280,
		Height:           800,
		BackgroundColour: application.NewRGB(27, 38, 54),
		URL:              "/",
	})

	if err := wailsApp.Run(); err != nil {
		log.Fatal(err)
	}
}

func focusMainWindow(wailsApp *application.App) {
	if wailsApp == nil {
		return
	}
	win := wailsApp.Window.Current()
	if win == nil {
		for _, w := range wailsApp.Window.GetAll() {
			win = w
			break
		}
	}
	if win == nil {
		return
	}
	win.Restore()
	win.Show()
	win.Focus()
}
