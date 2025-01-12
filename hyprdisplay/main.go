package main

import (
	"embed"
	"flag"
	"hyprdisplay/backend"
	"log"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	daemonize := flag.Bool("daemon", false, "--daemon to start daemon")
	verbose := flag.Bool("v", false, "-v verbose")

	flag.Parse()

	backend.Verbose = *verbose

	var daemon *backend.Daemon = nil

	if *daemonize {
		daemon = backend.Daemonize()
	}

	_ = startApp()

	if daemon != nil {
		log.Printf("quitting daemon")
		daemon.Close()
	}
}

func startApp() *App {
	// Create an instance of the app structure
	app := NewApp()

	// Create application with options
	err := wails.Run(&options.App{
		Title:  "hyprdisplay",
		Width:  1024,
		Height: 768,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 1},
		OnStartup:        app.startup,
		Bind: []interface{}{
			app,
		},
	})

	if err != nil {
		println("Error:", err.Error())
	}

	return app
}
