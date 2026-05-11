package main

import (
	"embed"
	"io/fs"
	"log"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:build/frontend
var assets embed.FS

func main() {
	// Sub-FS rooted at build/frontend/browser — served by the display HTTP server
	// so remote operators can open http://<host>:8080 in a browser.
	browserFS, err := fs.Sub(assets, "build/frontend/browser")
	if err != nil {
		log.Fatal("could not create browser sub-fs:", err)
	}

	app := NewApp(browserFS)

	err = wails.Run(&options.App{
		Title:            "Judo App",
		Width:            1280,
		Height:           800,
		AssetServer:      &assetserver.Options{Assets: browserFS},
		BackgroundColour: &options.RGBA{R: 13, G: 13, B: 26, A: 1},
		OnStartup:        app.startup,
		OnShutdown:       app.shutdown,
		Bind:             []interface{}{app},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}
