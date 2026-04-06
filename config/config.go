package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/spf13/viper"
	"go.yaml.in/yaml/v3"
)

const (
	DefaultAppConfigFileName         = "config.yaml"
	DefaultConnectionsConfigFileName = "connections.yaml"
	DefaultStateFileName             = "state.yaml"
	DefaultLogFileName               = "logs/app.log"
	DefaultTempDir                   = "tmp"
	DefaultTransferTempDirName       = "transfers"
	DefaultUnixConfigDirName         = "file-browser"
	DefaultUIExplorerFontSize        = 13
	DefaultUIFileListFontSize        = 13
	MinUIFontSize                    = 11
	MaxUIFontSize                    = 18
)

type AppConfig struct {
	App      AppSection      `json:"app" yaml:"app" mapstructure:"app"`
	Log      LogSection      `json:"log" yaml:"log" mapstructure:"log"`
	Paths    PathSection     `json:"paths" yaml:"paths" mapstructure:"paths"`
	Search   SearchSection   `json:"search" yaml:"search" mapstructure:"search"`
	Transfer TransferSection `json:"transfer" yaml:"transfer" mapstructure:"transfer"`
	UI       UISection       `json:"ui" yaml:"ui" mapstructure:"ui"`
}

type AppSection struct {
	Locale  string `json:"locale" yaml:"locale" mapstructure:"locale"`
	Theme   string `json:"theme" yaml:"theme" mapstructure:"theme"`
	TempDir string `json:"tempDir" yaml:"temp_dir" mapstructure:"temp_dir"`
}

type LogSection struct {
	Level   string   `json:"level" yaml:"level" mapstructure:"level"`
	Outputs []string `json:"outputs" yaml:"outputs" mapstructure:"outputs"`
}

type PathSection struct {
	ConnectionsFile string `json:"connectionsFile" yaml:"connections_file" mapstructure:"connections_file"`
	StateFile       string `json:"stateFile" yaml:"state_file" mapstructure:"state_file"`
}

type SearchSection struct {
	MaxConcurrency int `json:"maxConcurrency" yaml:"max_concurrency" mapstructure:"max_concurrency"`
	ResultLimit    int `json:"resultLimit" yaml:"result_limit" mapstructure:"result_limit"`
}

type TransferSection struct {
	TempDir           string `json:"tempDir" yaml:"temp_dir" mapstructure:"temp_dir"`
	DownloadDir       string `json:"downloadDir" yaml:"download_dir" mapstructure:"download_dir"`
	OverwriteStrategy string `json:"overwriteStrategy" yaml:"overwrite_strategy" mapstructure:"overwrite_strategy"`
}

type UISection struct {
	Locale           string `json:"locale" yaml:"locale" mapstructure:"locale"`
	Theme            string `json:"theme" yaml:"theme" mapstructure:"theme"`
	ExplorerFontSize int    `json:"explorerFontSize" yaml:"explorer_font_size" mapstructure:"explorer_font_size"`
	FileListFontSize int    `json:"fileListFontSize" yaml:"file_list_font_size" mapstructure:"file_list_font_size"`
}

type ConnectionDefinition struct {
	ID          string            `json:"id" yaml:"id" mapstructure:"id"`
	Name        string            `json:"name" yaml:"name" mapstructure:"name"`
	Driver      string            `json:"driver" yaml:"driver" mapstructure:"driver"`
	Description string            `json:"description,omitempty" yaml:"description,omitempty" mapstructure:"description"`
	Enabled     bool              `json:"enabled" yaml:"enabled" mapstructure:"enabled"`
	ReadOnly    bool              `json:"readOnly,omitempty" yaml:"read_only,omitempty" mapstructure:"read_only"`
	Root        string            `json:"root,omitempty" yaml:"root,omitempty" mapstructure:"root"`
	Tags        []string          `json:"tags,omitempty" yaml:"tags,omitempty" mapstructure:"tags"`
	Metadata    map[string]string `json:"metadata,omitempty" yaml:"metadata,omitempty" mapstructure:"metadata"`
	Config      map[string]any    `json:"config,omitempty" yaml:"config,omitempty" mapstructure:"config"`
}

type ConnectionsConfig struct {
	Connections []ConnectionDefinition `json:"connections" yaml:"connections" mapstructure:"connections"`
}

type LoadedAppConfig struct {
	Path   string
	Exists bool
	Config *AppConfig
}

type LoadedConnectionsConfig struct {
	Path   string
	Exists bool
	Config *ConnectionsConfig
}

func CloneAppConfig(cfg *AppConfig) *AppConfig {
	if cfg == nil {
		return nil
	}

	body, err := json.Marshal(cfg)
	if err != nil {
		clone := *cfg
		return &clone
	}

	var out AppConfig
	if err := json.Unmarshal(body, &out); err != nil {
		clone := *cfg
		return &clone
	}

	return &out
}

func DefaultAppConfig() *AppConfig {
	return &AppConfig{
		App: AppSection{
			Locale:  "zh",
			Theme:   "system",
			TempDir: DefaultTempDir,
		},
		Log: LogSection{
			Level:   "info",
			Outputs: []string{"stdout", DefaultLogFileName},
		},
		Paths: PathSection{
			ConnectionsFile: DefaultConnectionsConfigFileName,
			StateFile:       DefaultStateFileName,
		},
		Search: SearchSection{
			MaxConcurrency: 4,
			ResultLimit:    500,
		},
		Transfer: TransferSection{
			TempDir:           DefaultTransferTempDirPath(),
			DownloadDir:       "",
			OverwriteStrategy: "rename",
		},
		UI: UISection{
			Locale:           "zh",
			Theme:            "system",
			ExplorerFontSize: DefaultUIExplorerFontSize,
			FileListFontSize: DefaultUIFileListFontSize,
		},
	}
}

func DefaultConnectionsConfig() *ConnectionsConfig {
	return &ConnectionsConfig{
		Connections: make([]ConnectionDefinition, 0),
	}
}

func DefaultTransferTempDirPath() string {
	if runtime.GOOS == "windows" {
		base := strings.TrimSpace(os.Getenv("USERPROFILE"))
		if base == "" {
			base = os.TempDir()
		}
		return filepath.Join(base, "AppData", "Local", "Temp", DefaultUnixConfigDirName, DefaultTransferTempDirName)
	}
	return filepath.Join(string(os.PathSeparator), "tmp", DefaultUnixConfigDirName, DefaultTransferTempDirName)
}

func ResolveAppConfigPath(input string) (string, error) {
	if strings.TrimSpace(input) != "" {
		return canonicalPath(strings.TrimSpace(input))
	}

	wd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("resolve app config path: %w", err)
	}

	if workingDirLooksLikeProjectRoot(wd) {
		return resolveAppConfigPathInDir(wd)
	}

	if dir, ok := defaultUnixConfigDir(); ok {
		return resolveAppConfigPathInDir(dir)
	}

	return canonicalPath(filepath.Join(wd, DefaultAppConfigFileName))
}

func LoadAppConfig(path string) (*LoadedAppConfig, error) {
	resolvedPath, err := ResolveAppConfigPath(path)
	if err != nil {
		return nil, err
	}

	if err := ensureParentDir(resolvedPath); err != nil {
		return nil, fmt.Errorf("prepare app config directory: %w", err)
	}

	cfg := DefaultAppConfig()
	exists, err := fileExists(resolvedPath)
	if err != nil {
		return nil, fmt.Errorf("load app config: %w", err)
	}
	if exists {
		v, configType, err := newViperForFile(resolvedPath)
		if err != nil {
			return nil, err
		}
		v.SetConfigType(configType)
		if err := v.ReadInConfig(); err != nil {
			return nil, fmt.Errorf("load app config %q: %w", resolvedPath, err)
		}
		if err := v.Unmarshal(cfg); err != nil {
			return nil, fmt.Errorf("decode app config %q: %w", resolvedPath, err)
		}
	}

	cfg.Normalize(filepath.Dir(resolvedPath))

	return &LoadedAppConfig{
		Path:   resolvedPath,
		Exists: exists,
		Config: cfg,
	}, nil
}

func SaveAppConfig(path string, cfg *AppConfig) error {
	if cfg == nil {
		cfg = DefaultAppConfig()
	}

	resolvedPath, err := ResolveAppConfigPath(path)
	if err != nil {
		return err
	}

	if err := ensureParentDir(resolvedPath); err != nil {
		return err
	}

	return writeConfigFile(resolvedPath, cfg)
}

func appConfigCandidates(baseDir string) []string {
	return []string{
		filepath.Join(baseDir, DefaultAppConfigFileName),
		filepath.Join(baseDir, "config.yml"),
		filepath.Join(baseDir, "config.json"),
	}
}

func resolveAppConfigPathInDir(dir string) (string, error) {
	for _, candidate := range appConfigCandidates(dir) {
		if _, statErr := os.Stat(candidate); statErr == nil {
			return canonicalPath(candidate)
		}
	}
	return canonicalPath(filepath.Join(dir, DefaultAppConfigFileName))
}

func workingDirLooksLikeProjectRoot(dir string) bool {
	if strings.TrimSpace(dir) == "" {
		return false
	}

	if _, err := os.Stat(filepath.Join(dir, "wails.json")); err != nil {
		return false
	}
	if _, err := os.Stat(filepath.Join(dir, "frontend", "package.json")); err != nil {
		return false
	}
	return true
}

func defaultUnixConfigDir() (string, bool) {
	if runtime.GOOS == "windows" {
		return "", false
	}

	home, err := os.UserHomeDir()
	if err != nil || strings.TrimSpace(home) == "" {
		return "", false
	}

	return filepath.Join(home, ".config", DefaultUnixConfigDirName), true
}

func LoadConnectionsConfig(path string) (*LoadedConnectionsConfig, error) {
	if strings.TrimSpace(path) == "" {
		return nil, fmt.Errorf("load connections config: path is required")
	}

	resolvedPath, err := filepath.Abs(filepath.Clean(path))
	if err != nil {
		return nil, fmt.Errorf("load connections config: %w", err)
	}
	resolvedPath, err = canonicalPath(resolvedPath)
	if err != nil {
		return nil, fmt.Errorf("load connections config: %w", err)
	}

	cfg := DefaultConnectionsConfig()
	exists, err := fileExists(resolvedPath)
	if err != nil {
		return nil, fmt.Errorf("load connections config: %w", err)
	}
	if exists {
		if err := decodeConnectionsConfigFile(resolvedPath, cfg); err != nil {
			return nil, fmt.Errorf("decode connections config %q: %w", resolvedPath, err)
		}
	}

	cfg.Normalize()

	return &LoadedConnectionsConfig{
		Path:   resolvedPath,
		Exists: exists,
		Config: cfg,
	}, nil
}

func SaveConnectionsConfig(path string, cfg *ConnectionsConfig) error {
	if strings.TrimSpace(path) == "" {
		return fmt.Errorf("save connections config: path is required")
	}
	if cfg == nil {
		cfg = DefaultConnectionsConfig()
	}
	cfg.Normalize()

	resolvedPath, err := canonicalPath(path)
	if err != nil {
		return fmt.Errorf("save connections config: %w", err)
	}

	if err := ensureParentDir(resolvedPath); err != nil {
		return err
	}

	return writeConfigFile(resolvedPath, cfg)
}

func decodeConnectionsConfigFile(path string, cfg *ConnectionsConfig) error {
	configType, err := detectConfigType(path)
	if err != nil {
		return err
	}

	body, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	switch configType {
	case "json":
		if len(body) == 0 {
			return nil
		}
		if err := json.Unmarshal(body, cfg); err != nil {
			return err
		}
		return nil
	case "yaml":
		if len(body) == 0 {
			return nil
		}
		if err := yaml.Unmarshal(body, cfg); err != nil {
			return err
		}
		return nil
	default:
		return fmt.Errorf("unsupported config type %q", configType)
	}
}

func (c *AppConfig) Normalize(baseDir string) {
	if c == nil {
		return
	}

	if strings.TrimSpace(baseDir) == "" {
		baseDir = "."
	}

	c.App.Locale = defaultString(c.App.Locale, "zh")
	c.App.Theme = defaultString(c.App.Theme, "system")
	c.App.TempDir = resolvePath(baseDir, defaultString(c.App.TempDir, DefaultTempDir))

	c.Log.Level = defaultString(c.Log.Level, "info")
	c.Log.Outputs = normalizeOutputPaths(baseDir, c.Log.Outputs)

	c.Paths.ConnectionsFile = resolvePath(baseDir, defaultString(c.Paths.ConnectionsFile, DefaultConnectionsConfigFileName))
	c.Paths.StateFile = resolvePath(baseDir, defaultString(c.Paths.StateFile, DefaultStateFileName))

	if c.Search.MaxConcurrency <= 0 {
		c.Search.MaxConcurrency = 4
	}
	if c.Search.ResultLimit <= 0 {
		c.Search.ResultLimit = 500
	}

	c.Transfer.TempDir = defaultPath(c.Transfer.TempDir, DefaultTransferTempDirPath(), baseDir)
	c.Transfer.DownloadDir = resolvePath(baseDir, c.Transfer.DownloadDir)
	c.Transfer.OverwriteStrategy = defaultString(c.Transfer.OverwriteStrategy, "rename")

	c.UI.Locale = defaultString(c.UI.Locale, c.App.Locale)
	c.UI.Theme = defaultString(c.UI.Theme, c.App.Theme)
	c.UI.ExplorerFontSize = normalizeUIFontSize(c.UI.ExplorerFontSize, DefaultUIExplorerFontSize)
	c.UI.FileListFontSize = normalizeUIFontSize(c.UI.FileListFontSize, DefaultUIFileListFontSize)
}

func normalizeUIFontSize(value int, fallback int) int {
	if fallback < MinUIFontSize || fallback > MaxUIFontSize {
		fallback = DefaultUIExplorerFontSize
	}
	if value <= 0 {
		return fallback
	}
	if value < MinUIFontSize {
		return MinUIFontSize
	}
	if value > MaxUIFontSize {
		return MaxUIFontSize
	}
	return value
}

func (c *ConnectionsConfig) Normalize() {
	if c == nil {
		return
	}

	if c.Connections == nil {
		c.Connections = make([]ConnectionDefinition, 0)
	}

	for i := range c.Connections {
		conn := &c.Connections[i]
		conn.ID = strings.TrimSpace(conn.ID)
		conn.Name = strings.TrimSpace(conn.Name)
		conn.Driver = strings.TrimSpace(conn.Driver)
		conn.Description = strings.TrimSpace(conn.Description)
		conn.Root = strings.TrimSpace(conn.Root)
		if conn.Metadata == nil {
			conn.Metadata = map[string]string{}
		}
		if conn.Config == nil {
			conn.Config = map[string]any{}
		}
	}
}

func newViperForFile(path string) (*viper.Viper, string, error) {
	configType, err := detectConfigType(path)
	if err != nil {
		return nil, "", err
	}

	v := viper.New()
	v.SetConfigFile(path)
	return v, configType, nil
}

func detectConfigType(path string) (string, error) {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".yaml", ".yml":
		return "yaml", nil
	case ".json":
		return "json", nil
	default:
		return "", fmt.Errorf("unsupported config extension for %q", path)
	}
}

func fileExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func ensureParentDir(path string) error {
	dir := filepath.Dir(path)
	if dir == "" || dir == "." {
		return nil
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("ensure config directory %q: %w", dir, err)
	}
	return nil
}

func writeConfigFile(path string, cfg any) error {
	configType, err := detectConfigType(path)
	if err != nil {
		return err
	}

	var body []byte
	switch configType {
	case "json":
		body, err = json.MarshalIndent(cfg, "", "  ")
		if err == nil {
			body = append(body, '\n')
		}
	case "yaml":
		body, err = yaml.Marshal(cfg)
		if err == nil && len(body) > 0 && body[len(body)-1] != '\n' {
			body = append(body, '\n')
		}
	default:
		return fmt.Errorf("unsupported config type %q", configType)
	}
	if err != nil {
		return fmt.Errorf("encode config %q: %w", path, err)
	}

	if err := os.WriteFile(path, body, 0o644); err != nil {
		return fmt.Errorf("write config %q: %w", path, err)
	}
	return nil
}

func normalizeOutputPaths(baseDir string, outputs []string) []string {
	if len(outputs) == 0 {
		return []string{"stdout", resolvePath(baseDir, DefaultLogFileName)}
	}

	normalized := make([]string, 0, len(outputs))
	for _, output := range outputs {
		trimmed := strings.TrimSpace(output)
		switch strings.ToLower(trimmed) {
		case "", "stdout", "stderr":
			if trimmed != "" {
				normalized = append(normalized, strings.ToLower(trimmed))
			}
		default:
			normalized = append(normalized, resolvePath(baseDir, trimmed))
		}
	}

	if len(normalized) == 0 {
		return []string{"stdout", resolvePath(baseDir, DefaultLogFileName)}
	}
	return normalized
}

func resolvePath(baseDir, value string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return ""
	}
	if filepath.IsAbs(trimmed) {
		return filepath.Clean(trimmed)
	}
	return filepath.Clean(filepath.Join(baseDir, trimmed))
}

func defaultPath(value, fallback, baseDir string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return filepath.Clean(fallback)
	}
	return resolvePath(baseDir, trimmed)
}

func canonicalPath(path string) (string, error) {
	abs, err := filepath.Abs(filepath.Clean(path))
	if err != nil {
		return "", err
	}

	current := abs
	suffix := make([]string, 0, 4)
	for {
		resolved, err := filepath.EvalSymlinks(current)
		if err == nil {
			for i := len(suffix) - 1; i >= 0; i-- {
				resolved = filepath.Join(resolved, suffix[i])
			}
			return filepath.Clean(resolved), nil
		}

		if !os.IsNotExist(err) {
			return "", err
		}

		parent := filepath.Dir(current)
		if parent == current {
			return abs, nil
		}

		suffix = append(suffix, filepath.Base(current))
		current = parent
	}
}

func defaultString(value, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return strings.TrimSpace(value)
}
