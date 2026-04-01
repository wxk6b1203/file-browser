package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadAppConfigDefaultsWhenMissing(t *testing.T) {
	tmpDir := t.TempDir()
	prevWD, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("chdir: %v", err)
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
	expectedPath := filepath.Join(tmpDir, DefaultAppConfigFileName)
	if loaded.Path != expectedPath {
		t.Fatalf("expected path %q, got %q", expectedPath, loaded.Path)
	}
	if loaded.Config.Log.Level != "info" {
		t.Fatalf("expected default log level info, got %q", loaded.Config.Log.Level)
	}
	expectedLogPath := filepath.Join(tmpDir, DefaultLogFileName)
	if len(loaded.Config.Log.Outputs) != 2 || loaded.Config.Log.Outputs[1] != expectedLogPath {
		t.Fatalf("unexpected log outputs: %#v", loaded.Config.Log.Outputs)
	}
	if loaded.Config.Paths.ConnectionsFile != filepath.Join(tmpDir, DefaultConnectionsConfigFileName) {
		t.Fatalf("unexpected connections file path: %q", loaded.Config.Paths.ConnectionsFile)
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
  overwrite_strategy: overwrite
ui:
  locale: en
  theme: islands-dark
`)
	if err := os.WriteFile(configPath, body, 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	loaded, err := LoadAppConfig(configPath)
	if err != nil {
		t.Fatalf("LoadAppConfig: %v", err)
	}

	if !loaded.Exists {
		t.Fatalf("expected config file to exist")
	}
	if loaded.Config.App.TempDir != filepath.Join(tmpDir, "cache") {
		t.Fatalf("unexpected app temp dir: %q", loaded.Config.App.TempDir)
	}
	if loaded.Config.Log.Outputs[1] != filepath.Join(tmpDir, "logs", "custom.log") {
		t.Fatalf("unexpected log file path: %q", loaded.Config.Log.Outputs[1])
	}
	if loaded.Config.Paths.ConnectionsFile != filepath.Join(tmpDir, "config", "connections.yaml") {
		t.Fatalf("unexpected connections path: %q", loaded.Config.Paths.ConnectionsFile)
	}
	if loaded.Config.Transfer.TempDir != filepath.Join(tmpDir, "work", "transfers") {
		t.Fatalf("unexpected transfer temp dir: %q", loaded.Config.Transfer.TempDir)
	}
	if loaded.Config.Search.MaxConcurrency != 8 || loaded.Config.Search.ResultLimit != 1200 {
		t.Fatalf("unexpected search config: %#v", loaded.Config.Search)
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
