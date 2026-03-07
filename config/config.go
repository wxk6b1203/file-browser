package config

type AppOptions struct {
	Log      *LogOptions      `json:"log,omitempty" yaml:"log,omitempty"`
	Location *LocationOptions `json:"location,omitempty" yaml:"location,omitempty"`
}

type LogOptions struct {
	Level string `json:"level" yaml:"level"`
	Path  string `json:"path" yaml:"path"`
}

type LocationOptions struct {
	Folders []*FolderOptions `json:"folders,omitempty" yaml:"folders,omitempty"`
}

type FolderOptions struct {
	Name    string         `json:"name" yaml:"name"`
	Driver  string         `json:"driver" yaml:"driver"`
	Root    string         `json:"root,omitempty" yaml:"root,omitempty"`
	Options map[string]any `json:"options,omitempty" yaml:"options,omitempty"`
}
