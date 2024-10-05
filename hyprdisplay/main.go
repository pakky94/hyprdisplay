package main

import (
	"embed"
	"hyprdisplay/backend"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	dbPath, err := backend.DefaultDbPath()
	if err != nil {
		panic(err)
	}

	db, err := backend.InitDb(dbPath, backend.DB_NAME)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	currentSetup, err := backend.ReadHyprMonitors()
	if err != nil {
		panic(err)
	}

	dbSetup, err := backend.FindSetup(db, backend.ToKey(currentSetup))
	if err != nil {
		panic(err)
	}

	ctl, err := backend.OpenConn()
	if err != nil {
		panic(err)
	}
	defer ctl.Close()

	if len(dbSetup) == 0 {
		println("no setup found in db")

		err := backend.SaveSetup(db, backend.ToKey(currentSetup), currentSetup)
		if err != nil {
			panic(err)
		}
	} else {
		cmds := backend.Diff(currentSetup, dbSetup)
		err := backend.Apply(cmds)
		if err != nil {
			panic(err)
		}
	}

	/*
		ctl, err := backend.OpenConn()
		if err != nil {
			panic(err)
		}
		defer ctl.Close()

		cmdChan := make(chan string)
		go func() {
			for true {
				cmd := <-cmdChan
				if strings.HasPrefix(cmd, "monitoradded>>") || strings.HasPrefix(cmd, "monitorremoved>>") {
					println(cmd)
					monitors, err := backend.ReadHyprMonitors()

					if err != nil {
						panic(err)
					}

					println(fmt.Sprintf("%+v", monitors))
				}
			}
		}()

		ctl.Loop(cmdChan)
	*/

	// ctl.SendRaw([]byte("/keyword monitor eDP-1, disable"))
	// ctl.SendRaw([]byte("/keyword monitor eDP-1,preferred,0x0,1,transform,3"))
}

func startApp() {
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
}
