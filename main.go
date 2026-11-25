package main

import (
	"embed"
	"log"

	"wreckscope/internal/app"

	"github.com/wailsapp/wails/v3/pkg/application"
)

//go:embed all:ui/dist
var assets embed.FS

func main() {
	service := app.New(assets)

	wailsApp := application.New(application.Options{
		Name:        "WreckScope",
		Description: "WreckScope dashboard",
		Services: []application.Service{
			application.NewService(service),
		},
		Assets: application.AssetOptions{
			Handler: application.AssetFileServerFS(assets),
		},
		Mac: application.MacOptions{
			ApplicationShouldTerminateAfterLastWindowClosed: true,
		},
	})

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
