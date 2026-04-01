package bootstrap

import (
	"fmt"

	"github.com/spf13/pflag"
	"github.com/wxk6b1203/file-util-manager/config"
	"github.com/wxk6b1203/file-util-manager/logging"
)

type StartupOptions struct {
	ConfigPath string
}

type Runtime struct {
	Startup         StartupOptions
	AppConfigPath   string
	AppConfigExists bool
	AppConfig       *config.AppConfig
}

func ParseStartupOptions(args []string) (*StartupOptions, error) {
	opts := &StartupOptions{}

	fs := pflag.NewFlagSet("file-browser", pflag.ContinueOnError)
	fs.ParseErrorsWhitelist.UnknownFlags = true
	fs.StringVarP(&opts.ConfigPath, "config", "c", "", "Path to the app config file (.yaml, .yml, .json)")

	if err := fs.Parse(args); err != nil {
		return nil, fmt.Errorf("parse startup options: %w", err)
	}

	return opts, nil
}

func Initialize(args []string) (*Runtime, error) {
	startup, err := ParseStartupOptions(args)
	if err != nil {
		return nil, err
	}

	loadedConfig, err := config.LoadAppConfig(startup.ConfigPath)
	if err != nil {
		return nil, err
	}

	logging.InitLogging(&logging.LogOptions{
		Level: loadedConfig.Config.Log.Level,
		Path:  append([]string(nil), loadedConfig.Config.Log.Outputs...),
	})

	return &Runtime{
		Startup:         *startup,
		AppConfigPath:   loadedConfig.Path,
		AppConfigExists: loadedConfig.Exists,
		AppConfig:       loadedConfig.Config,
	}, nil
}
