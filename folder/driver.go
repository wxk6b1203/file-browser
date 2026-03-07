package folder

import (
	"context"
	"io"
)

type PathOp struct {
	SrcPath string
	DstPath string
}

type ListOptions struct {
	Recursive bool
	Limit     int
	Prefix    string
}

type WriteOptions struct {
	ContentType string
	Metadata    map[string]string
}

// Manager defines common file-system operations across local/remote backends.
type Manager interface {
	Name() string
	Capabilities() Capabilities

	List(ctx context.Context, path string, opt *ListOptions) ([]*FileInfo, error)
	Stat(ctx context.Context, path string) (*FileInfo, error)
	Delete(ctx context.Context, path string) error
	Copy(ctx context.Context, op PathOp) error
	Move(ctx context.Context, op PathOp) error
	Mkdir(ctx context.Context, path string) error
}

// Reader is an optional capability for backends that support file download/streaming.
type Reader interface {
	Read(ctx context.Context, path string) (io.ReadCloser, error)
}

// Writer is an optional capability for backends that support file upload/streaming.
type Writer interface {
	Write(ctx context.Context, path string, body io.Reader, opt *WriteOptions) (*FileInfo, error)
}

// HealthChecker is an optional capability for connection-level readiness checks.
type HealthChecker interface {
	Ping(ctx context.Context) error
}

// Closer is an optional capability for releasing driver resources.
type Closer interface {
	Close() error
}

// ---------------------------------------------------------------------------
// BaseDriver – embed in concrete drivers to inherit default (ErrUnsupported)
// implementations for every Manager method. Override only what the backend
// actually supports.
// ---------------------------------------------------------------------------

// BaseDriver provides default stub implementations for all Manager methods.
// Concrete drivers embed *BaseDriver and override only the methods they support.
type BaseDriver struct {
	Opt *DriverOptions
}

func NewBaseDriver(opt *DriverOptions) BaseDriver {
	return BaseDriver{Opt: opt}
}

func (b *BaseDriver) Name() string {
	if b.Opt != nil && b.Opt.Name != "" {
		return b.Opt.Name
	}
	if b.Opt != nil {
		return b.Opt.Driver
	}
	return ""
}

func (b *BaseDriver) Options() *DriverOptions { return b.Opt }

func (b *BaseDriver) Capabilities() Capabilities { return BaseCapabilities() }

func (b *BaseDriver) List(_ context.Context, _ string, _ *ListOptions) ([]*FileInfo, error) {
	return nil, ErrUnsupported
}

func (b *BaseDriver) Stat(_ context.Context, _ string) (*FileInfo, error) {
	return nil, ErrUnsupported
}

func (b *BaseDriver) Delete(_ context.Context, _ string) error {
	return ErrUnsupported
}

func (b *BaseDriver) Copy(_ context.Context, _ PathOp) error {
	return ErrUnsupported
}

func (b *BaseDriver) Move(_ context.Context, _ PathOp) error {
	return ErrUnsupported
}

func (b *BaseDriver) Mkdir(_ context.Context, _ string) error {
	return ErrUnsupported
}
