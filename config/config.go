package config

type AppOptions struct {
	Log *LogOptions `json:"log,omitempty" yaml:"log,omitempty"`
}

type LogOptions struct {
	Level string `json:"level" yaml:"level"`
	Path  string `json:"path" yaml:"path"`
}
type LocationOptions struct {
	Folders []*FolderOptions[any] `json:"drivers,omitempty" yaml:"drivers,omitempty"`
}

type FolderOptions[T any] struct {
	Name   string `json:"name" yaml:"name"`
	Driver string `json:"folder,omitempty" yaml:"folder,omitempty"`
}
