package main

import (
	"context"
	"fmt"

	wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"
	"github.com/wxk6b1203/file-util-manager/bootstrap"
	"github.com/wxk6b1203/file-util-manager/config"
	"github.com/wxk6b1203/file-util-manager/connection"
	"github.com/wxk6b1203/file-util-manager/fileops"
	"github.com/wxk6b1203/file-util-manager/folder"
	"github.com/wxk6b1203/file-util-manager/logging"
	"github.com/wxk6b1203/file-util-manager/render"
	"github.com/wxk6b1203/file-util-manager/search"
	"github.com/wxk6b1203/file-util-manager/shortcut"
	"github.com/wxk6b1203/file-util-manager/transfer"
	"go.uber.org/zap"
)

// App struct
type App struct {
	ctx      context.Context
	sc       *shortcut.Dispatcher
	rm       *render.Manager
	runtime  *bootstrap.Runtime
	conn     *connection.Service
	fileops  *fileops.Service
	search   *search.Service
	transfer *transfer.Service
}

// NewApp creates a new App application struct
func NewApp(rt *bootstrap.Runtime) *App {
	zap.S().Info("Creating new app")
	if rt != nil {
		zap.S().Infow("Runtime configuration loaded", "path", rt.AppConfigPath, "exists", rt.AppConfigExists)
	} else {
		rt = &bootstrap.Runtime{AppConfig: config.DefaultAppConfig()}
	}
	if rt.AppConfig == nil {
		rt.AppConfig = config.DefaultAppConfig()
	}

	sc := shortcut.NewDispatcher()
	// Register Go-side shortcut handlers here. Example:
	// sc.On("save", func() { zap.S().Info("save triggered from frontend") })

	connSvc := connection.NewService(connection.NewFileRepository(
		rt.AppConfig.Paths.ConnectionsFile,
	))

	return &App{
		sc:       sc,
		rm:       render.NewManager(),
		runtime:  rt,
		conn:     connSvc,
		fileops:  fileops.NewService(connSvc),
		search:   search.NewService(connSvc, rt.AppConfig.Search.MaxConcurrency, rt.AppConfig.Search.ResultLimit),
		transfer: transfer.NewService(connSvc, rt.AppConfig.Transfer.TempDir),
	}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	a.sc.Listen(ctx)
	render.Start(a.rm, ctx)
	a.transfer.SetObserver(func(event folder.TransferEvent) {
		wailsRuntime.EventsEmit(a.ctx, transfer.EventName, event)
	})
}

func (a *App) shutdown(ctx context.Context) {
	a.transfer.SetObserver(nil)
	a.sc.Stop()
	render.Stop(a.rm)
	zap.S().Info("App is shutting down")
}

func (a *App) applyRuntimeConfig(cfg *config.AppConfig) {
	if cfg == nil {
		return
	}

	logging.InitLogging(&logging.LogOptions{
		Level: cfg.Log.Level,
		Path:  append([]string(nil), cfg.Log.Outputs...),
	})

	a.search.UpdateDefaults(cfg.Search.MaxConcurrency, cfg.Search.ResultLimit)
	a.transfer.SetTempDir(cfg.Transfer.TempDir)
}

// ---------------------------------------------------------------------------
// Bound methods — exposed to the frontend via Wails bindings
// ---------------------------------------------------------------------------

// Greet returns a greeting for the given name
func (a *App) Greet(name string) string {
	return fmt.Sprintf("Hello %s, It's show time!", name)
}

func (a *App) GetAppConfig() *config.AppConfig {
	if a.runtime == nil {
		return config.DefaultAppConfig()
	}

	cloned := config.CloneAppConfig(a.runtime.AppConfig)
	if cloned == nil {
		return config.DefaultAppConfig()
	}
	return cloned
}

func (a *App) SaveAppConfig(next config.AppConfig) (*config.AppConfig, error) {
	cfg := config.CloneAppConfig(&next)
	if cfg == nil {
		cfg = config.DefaultAppConfig()
	}

	configPath := ""
	if a.runtime != nil {
		configPath = a.runtime.AppConfigPath
	}

	if err := config.SaveAppConfig(configPath, cfg); err != nil {
		return nil, err
	}

	loaded, err := config.LoadAppConfig(configPath)
	if err != nil {
		return nil, err
	}

	if a.runtime == nil {
		a.runtime = &bootstrap.Runtime{}
	}
	a.runtime.AppConfigPath = loaded.Path
	a.runtime.AppConfigExists = loaded.Exists
	a.runtime.AppConfig = loaded.Config

	a.applyRuntimeConfig(loaded.Config)
	return config.CloneAppConfig(loaded.Config), nil
}

func (a *App) ListDrivers() []folder.DriverInfo {
	return a.conn.ListDrivers()
}

func (a *App) ListConnections() ([]connection.Definition, error) {
	return a.conn.List(a.ctx)
}

func (a *App) SaveConnection(def connection.Definition) (connection.Definition, error) {
	return a.conn.Save(a.ctx, def)
}

func (a *App) DeleteConnection(id string) error {
	return a.conn.Delete(a.ctx, id)
}

func (a *App) OpenConnection(id string) (*connection.State, error) {
	return a.conn.Open(a.ctx, id)
}

func (a *App) CloseConnection(id string) error {
	return a.conn.Close(a.ctx, id)
}

func (a *App) ListConnectionStates() []connection.State {
	return a.conn.ListStates()
}

func (a *App) ListConnectionDirectory(connectionID string, dir string) ([]*folder.FileInfo, error) {
	return a.fileops.ListDirectory(a.ctx, connectionID, dir)
}

func (a *App) CreateConnectionDirectory(connectionID string, parentDir string, name string) (*folder.FileInfo, error) {
	return a.fileops.CreateDirectory(a.ctx, connectionID, parentDir, name)
}

func (a *App) RenameConnectionEntry(connectionID string, targetPath string, newName string) (*folder.FileInfo, error) {
	return a.fileops.RenameEntry(a.ctx, connectionID, targetPath, newName)
}

func (a *App) DeleteConnectionEntry(connectionID string, targetPath string) error {
	return a.fileops.DeleteEntry(a.ctx, connectionID, targetPath)
}

func (a *App) DownloadConnectionFileToTemp(connectionID string, remotePath string) ([]string, error) {
	return a.transfer.DownloadToTemp(a.ctx, connectionID, remotePath)
}

func (a *App) UploadConnectionLocalPath(connectionID string, remoteDir string, localPath string) ([]string, error) {
	return a.transfer.UploadLocalPath(a.ctx, connectionID, remoteDir, localPath)
}

func (a *App) TransferConnectionEntry(sourceConnectionID string, sourcePath string, targetConnectionID string, targetDir string) ([]string, error) {
	return a.transfer.TransferEntry(a.ctx, sourceConnectionID, sourcePath, targetConnectionID, targetDir)
}

func (a *App) ListTransferTasks() []*folder.TransferTask {
	return a.transfer.ListTasks()
}

func (a *App) CancelTransferTask(taskID string) error {
	return a.transfer.CancelTask(taskID)
}

func (a *App) RemoveTransferTask(taskID string) error {
	return a.transfer.RemoveTask(taskID)
}

func (a *App) ClearFinishedTransferTasks() {
	a.transfer.RemoveFinishedTasks()
}

func (a *App) StartSearch(req search.Request) (string, error) {
	return a.search.Start(a.ctx, req, func(event search.Event) {
		wailsRuntime.EventsEmit(a.ctx, search.EventName, event)
	})
}

func (a *App) CancelSearch(requestID string) error {
	return a.search.Cancel(requestID)
}
