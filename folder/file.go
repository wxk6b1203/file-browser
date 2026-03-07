package folder

import "time"

const (
	EntryTypeUnknown EntryType = iota
	EntryTypeFile
	EntryTypeDirectory
	EntryTypeSymlink
)

type EntryType int

type Owner struct {
	ID   string `json:"id" yaml:"id"`
	Name string `json:"name" yaml:"name"`
}

type FileInfo struct {
	Name         string            `json:"name" yaml:"name"`
	Path         string            `json:"path" yaml:"path"`
	Type         EntryType         `json:"type" yaml:"type"`
	Size         int64             `json:"size" yaml:"size"`
	LastModified *time.Time        `json:"lastModified,omitempty" yaml:"lastModified,omitempty"`
	Owner        *Owner            `json:"owner,omitempty" yaml:"owner,omitempty"`
	ContentType  string            `json:"contentType,omitempty" yaml:"contentType,omitempty"`
	ETag         string            `json:"etag,omitempty" yaml:"etag,omitempty"`
	Metadata     map[string]string `json:"metadata,omitempty" yaml:"metadata,omitempty"`
}

func (f *FileInfo) IsDir() bool {
	return f != nil && f.Type == EntryTypeDirectory
}

func (f *FileInfo) IsFile() bool {
	return f != nil && f.Type == EntryTypeFile
}
