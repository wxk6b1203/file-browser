package main

import (
	"embed"
	"fmt"
	"os"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wxk6b1203/file-util-manager/bootstrap"
	_ "github.com/wxk6b1203/file-util-manager/folder/alibaba-oss"
	_ "github.com/wxk6b1203/file-util-manager/folder/local"
	_ "github.com/wxk6b1203/file-util-manager/folder/s3"
	_ "github.com/wxk6b1203/file-util-manager/folder/sftp"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	rt, err := bootstrap.Initialize(os.Args[1:])
	if err != nil {
		fmt.Fprintf(os.Stderr, "startup failed: %v\n", err)
		os.Exit(1)
	}

	app := NewApp(rt)

	err = wails.Run(&options.App{
		Title:  "file-browser",
		Width:  1024,
		Height: 768,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 255, G: 255, B: 255, A: 1},
		OnStartup:        app.startup,
		OnShutdown:       app.shutdown,
		// DragAndDrop enables OS-level file drag-and-drop into the window.
		// The frontend registers handlers via runtime.OnFileDrop().
		DragAndDrop: &options.DragAndDrop{
			EnableFileDrop: true,
		},
		Bind: []interface{}{
			app,
			app.rm,
		},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}
