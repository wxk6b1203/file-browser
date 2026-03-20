package main

import (
	"context"
	"fmt"

	"github.com/wxk6b1203/file-util-manager/render"
	"github.com/wxk6b1203/file-util-manager/shortcut"
	"go.uber.org/zap"
)

// App struct
type App struct {
	ctx context.Context
	sc  *shortcut.Dispatcher
	rm  *render.Manager
}

// NewApp creates a new App application struct
func NewApp() *App {
	zap.S().Info("Creating new app")

	sc := shortcut.NewDispatcher()
	// Register Go-side shortcut handlers here. Example:
	// sc.On("save", func() { zap.S().Info("save triggered from frontend") })

	return &App{
		sc: sc,
		rm: render.NewManager(),
	}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	a.sc.Listen(ctx)
	a.rm.Startup(ctx)
}

func (a *App) shutdown(ctx context.Context) {
	a.sc.Stop()
	a.rm.Shutdown()
	zap.S().Info("App is shutting down")
}

// ---------------------------------------------------------------------------
// Bound methods — exposed to the frontend via Wails bindings
// ---------------------------------------------------------------------------

// Greet returns a greeting for the given name
func (a *App) Greet(name string) string {
	return fmt.Sprintf("Hello %s, It's show time!", name)
}
