package sftp

import (
	"context"

	"github.com/wxk6b1203/file-util-manager/folder"
)

func init() {
	folder.RegisterDriver[Options]("sftp", New)
}

// Driver implements folder.Manager for SFTP backends.
type Driver struct {
	folder.BaseDriver
	cfg *Options
}

// New creates a new SFTP driver. Called automatically by the registry.
func New(_ context.Context, opt *folder.DriverOptions, cfg *Options) (folder.Manager, error) {
	if cfg.Port == 0 {
		cfg.Port = 22
	}
	return &Driver{
		BaseDriver: folder.NewBaseDriver(opt),
		cfg:        cfg,
	}, nil
}

func (d *Driver) Capabilities() folder.Capabilities {
	caps := folder.BaseCapabilities()
	caps.CanRead = true
	caps.CanWrite = true
	caps.AtomicMove = true
	return caps
}
