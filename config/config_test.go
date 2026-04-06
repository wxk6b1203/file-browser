package config

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func mustCanonicalPath(t *testing.T, path string) string {
	t.Helper()

	resolved, err := filepath.EvalSymlinks(path)
	if err != nil {
		t.Fatalf("EvalSymlinks(%q): %v", path, err)
	}
	return filepath.Clean(resolved)
}

func TestLoadAppConfigDefaultsWhenMissing(t *testing.T) {
	tmpDir := t.TempDir()
	prevWD, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "wails.json"), []byte("{}\n"), 0o644); err != nil {
		t.Fatalf("write wails marker: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(tmpDir, "frontend"), 0o755); err != nil {
		t.Fatalf("mkdir frontend marker: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "frontend", "package.json"), []byte("{}\n"), 0o644); err != nil {
		t.Fatalf("write frontend marker: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(prevWD)
	})

	loaded, err := LoadAppConfig("")
	if err != nil {
		t.Fatalf("LoadAppConfig: %v", err)
	}

	if loaded.Exists {
		t.Fatalf("expected missing config file to report Exists=false")
	}
	canonicalTmpDir := mustCanonicalPath(t, tmpDir)
	expectedPath := filepath.Join(canonicalTmpDir, DefaultAppConfigFileName)
	if loaded.Path != expectedPath {
		t.Fatalf("expected path %q, got %q", expectedPath, loaded.Path)
	}
	if loaded.Config.Log.Level != "info" {
		t.Fatalf("expected default log level info, got %q", loaded.Config.Log.Level)
	}
	if loaded.Config.UI.ExplorerFontSize != DefaultUIExplorerFontSize {
		t.Fatalf("expected default explorer font size %d, got %d", DefaultUIExplorerFontSize, loaded.Config.UI.ExplorerFontSize)
	}
	if loaded.Config.UI.FileListFontSize != DefaultUIFileListFontSize {
		t.Fatalf("expected default file list font size %d, got %d", DefaultUIFileListFontSize, loaded.Config.UI.FileListFontSize)
	}
	expectedLogPath := filepath.Join(canonicalTmpDir, DefaultLogFileName)
	if len(loaded.Config.Log.Outputs) != 2 || loaded.Config.Log.Outputs[1] != expectedLogPath {
		t.Fatalf("unexpected log outputs: %#v", loaded.Config.Log.Outputs)
	}
	if loaded.Config.Paths.ConnectionsFile != filepath.Join(canonicalTmpDir, DefaultConnectionsConfigFileName) {
		t.Fatalf("unexpected connections file path: %q", loaded.Config.Paths.ConnectionsFile)
	}
}

func TestLoadAppConfigDefaultsToUnixUserConfigDirOutsideProjectRoot(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("unix config directory default does not apply to windows")
	}

	tmpDir := t.TempDir()
	homeDir := filepath.Join(tmpDir, "home")
	wd := filepath.Join(tmpDir, "work")
	if err := os.MkdirAll(homeDir, 0o755); err != nil {
		t.Fatalf("mkdir home: %v", err)
	}
	if err := os.MkdirAll(wd, 0o755); err != nil {
		t.Fatalf("mkdir wd: %v", err)
	}

	prevWD, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	if err := os.Chdir(wd); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(prevWD)
	})

	t.Setenv("HOME", homeDir)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(tmpDir, "ignored-xdg-config"))

	loaded, err := LoadAppConfig("")
	if err != nil {
		t.Fatalf("LoadAppConfig: %v", err)
	}

	canonicalHome := mustCanonicalPath(t, homeDir)
	expectedPath := filepath.Join(canonicalHome, ".config", DefaultUnixConfigDirName, DefaultAppConfigFileName)
	if loaded.Path != expectedPath {
		t.Fatalf("expected path %q, got %q", expectedPath, loaded.Path)
	}
	if loaded.Config.Paths.ConnectionsFile != filepath.Join(canonicalHome, ".config", DefaultUnixConfigDirName, DefaultConnectionsConfigFileName) {
		t.Fatalf("unexpected connections path: %q", loaded.Config.Paths.ConnectionsFile)
	}
	if _, err := os.Stat(filepath.Dir(expectedPath)); err != nil {
		t.Fatalf("expected config directory to be created: %v", err)
	}
}

func TestLoadAppConfigUsesExistingUnixUserConfigYML(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("unix config directory default does not apply to windows")
	}

	tmpDir := t.TempDir()
	homeDir := filepath.Join(tmpDir, "home")
	wd := filepath.Join(tmpDir, "work")
	configDir := filepath.Join(homeDir, ".config", DefaultUnixConfigDirName)
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		t.Fatalf("mkdir config dir: %v", err)
	}
	if err := os.MkdirAll(wd, 0o755); err != nil {
		t.Fatalf("mkdir wd: %v", err)
	}
	configPath := filepath.Join(configDir, "config.yml")
	if err := os.WriteFile(configPath, []byte("log:\n  level: warn\n"), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	prevWD, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	if err := os.Chdir(wd); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(prevWD)
	})

	t.Setenv("HOME", homeDir)

	loaded, err := LoadAppConfig("")
	if err != nil {
		t.Fatalf("LoadAppConfig: %v", err)
	}

	canonicalConfigDir := mustCanonicalPath(t, configDir)
	if loaded.Path != filepath.Join(canonicalConfigDir, "config.yml") {
		t.Fatalf("expected existing config.yml, got %q", loaded.Path)
	}
	if loaded.Config.Log.Level != "warn" {
		t.Fatalf("expected loaded log level warn, got %q", loaded.Config.Log.Level)
	}
}

func TestLoadAppConfigYAMLNormalizesRelativePaths(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "app.yaml")
	body := []byte(`
app:
  locale: en
  theme: vscode-dark
  temp_dir: cache
log:
  level: debug
  outputs:
    - stdout
    - logs/custom.log
paths:
  connections_file: config/connections.yaml
  state_file: state/runtime.json
search:
  max_concurrency: 8
  result_limit: 1200
transfer:
  temp_dir: work/transfers
  download_dir: downloads
  overwrite_strategy: overwrite
ui:
  locale: en
  theme: islands-dark
  explorer_font_size: 10
  file_list_font_size: 20
`)
	if err := os.WriteFile(configPath, body, 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	loaded, err := LoadAppConfig(configPath)
	if err != nil {
		t.Fatalf("LoadAppConfig: %v", err)
	}

	canonicalTmpDir := mustCanonicalPath(t, tmpDir)

	if !loaded.Exists {
		t.Fatalf("expected config file to exist")
	}
	if loaded.Config.App.TempDir != filepath.Join(canonicalTmpDir, "cache") {
		t.Fatalf("unexpected app temp dir: %q", loaded.Config.App.TempDir)
	}
	if loaded.Config.Log.Outputs[1] != filepath.Join(canonicalTmpDir, "logs", "custom.log") {
		t.Fatalf("unexpected log file path: %q", loaded.Config.Log.Outputs[1])
	}
	if loaded.Config.Paths.ConnectionsFile != filepath.Join(canonicalTmpDir, "config", "connections.yaml") {
		t.Fatalf("unexpected connections path: %q", loaded.Config.Paths.ConnectionsFile)
	}
	if loaded.Config.Transfer.TempDir != filepath.Join(canonicalTmpDir, "work", "transfers") {
		t.Fatalf("unexpected transfer temp dir: %q", loaded.Config.Transfer.TempDir)
	}
	if loaded.Config.Transfer.DownloadDir != filepath.Join(canonicalTmpDir, "downloads") {
		t.Fatalf("unexpected download dir: %q", loaded.Config.Transfer.DownloadDir)
	}
	if loaded.Config.Search.MaxConcurrency != 8 || loaded.Config.Search.ResultLimit != 1200 {
		t.Fatalf("unexpected search config: %#v", loaded.Config.Search)
	}
	if loaded.Config.UI.ExplorerFontSize != MinUIFontSize {
		t.Fatalf("unexpected explorer font size clamp: %d", loaded.Config.UI.ExplorerFontSize)
	}
	if loaded.Config.UI.FileListFontSize != MaxUIFontSize {
		t.Fatalf("unexpected file list font size clamp: %d", loaded.Config.UI.FileListFontSize)
	}
}

func TestSaveAndLoadConnectionsConfigJSON(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "connections.json")

	cfg := &ConnectionsConfig{
		Connections: []ConnectionDefinition{
			{
				ID:      "local-default",
				Name:    "Local",
				Driver:  "Local",
				Enabled: true,
				Root:    "C:/",
				Config: map[string]any{
					"rootPath": "C:/",
				},
			},
		},
	}

	if err := SaveConnectionsConfig(configPath, cfg); err != nil {
		t.Fatalf("SaveConnectionsConfig: %v", err)
	}

	loaded, err := LoadConnectionsConfig(configPath)
	if err != nil {
		t.Fatalf("LoadConnectionsConfig: %v", err)
	}

	if len(loaded.Config.Connections) != 1 {
		t.Fatalf("expected one connection, got %d", len(loaded.Config.Connections))
	}
	conn := loaded.Config.Connections[0]
	if conn.ID != "local-default" || conn.Driver != "Local" {
		t.Fatalf("unexpected connection payload: %#v", conn)
	}
	if conn.Metadata == nil || conn.Config == nil {
		t.Fatalf("expected normalized maps, got metadata=%#v config=%#v", conn.Metadata, conn.Config)
	}
}

func TestLoadConnectionsConfigYAMLUsesYAMLV3(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "connections.yaml")

	body := []byte(`
connections:
  - id: local-docs
    name: Local Docs
    driver: Local
    enabled: true
    config:
      rootPath: /Users/wxk/Documents
`)
	if err := os.WriteFile(configPath, body, 0o644); err != nil {
		t.Fatalf("write connections yaml: %v", err)
	}

	loaded, err := LoadConnectionsConfig(configPath)
	if err != nil {
		t.Fatalf("LoadConnectionsConfig: %v", err)
	}

	if !loaded.Exists {
		t.Fatalf("expected yaml config file to exist")
	}
	if len(loaded.Config.Connections) != 1 {
		t.Fatalf("expected one connection, got %d", len(loaded.Config.Connections))
	}

	conn := loaded.Config.Connections[0]
	if conn.ID != "local-docs" || conn.Driver != "Local" {
		t.Fatalf("unexpected connection payload: %#v", conn)
	}
	if got := conn.Config["rootPath"]; got != "/Users/wxk/Documents" {
		t.Fatalf("unexpected rootPath: %#v", got)
	}
}
