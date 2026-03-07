package s3

import (
	"context"

	"github.com/wxk6b1203/file-util-manager/folder"
)

func init() {
	folder.RegisterDriver[Options]("s3", New)
}

// Driver implements folder.Manager for Amazon S3 backends.
type Driver struct {
	folder.BaseDriver
	cfg *Options
}

// New creates a new S3 driver. Called automatically by the registry.
func New(_ context.Context, opt *folder.DriverOptions, cfg *Options) (folder.Manager, error) {
	return &Driver{
		BaseDriver: folder.NewBaseDriver(opt),
		cfg:        cfg,
	}, nil
}

func (d *Driver) Capabilities() folder.Capabilities {
	caps := folder.BaseCapabilities()
	caps.CanRead = true
	caps.CanWrite = true
	return caps
}
