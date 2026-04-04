package folder

import (
	"context"
	"time"
)

// ---------------------------------------------------------------------------
// Transfer direction & status enums
// ---------------------------------------------------------------------------

// TransferDirection indicates upload or download.
type TransferDirection int

const (
	TransferUpload TransferDirection = iota + 1
	TransferDownload
)

// TransferStatus represents the lifecycle state of a transfer task.
type TransferStatus int

const (
	TransferPending   TransferStatus = iota // queued, not yet started
	TransferRunning                         // actively transferring data
	TransferCompleted                       // finished successfully
	TransferFailed                          // finished with an error
	TransferCancelled                       // cancelled by the caller
)

// ---------------------------------------------------------------------------
// Transfer request & progress callback
// ---------------------------------------------------------------------------

// TransferRequest describes a single upload or download to submit.
type TransferRequest struct {
	RemotePath      string            `json:"remotePath" yaml:"remotePath"` // path on the remote storage
	LocalPath       string            `json:"localPath" yaml:"localPath"`   // local file path (upload source / download destination)
	ContentType     string            `json:"contentType,omitempty" yaml:"contentType,omitempty"`
	Metadata        map[string]string `json:"metadata,omitempty" yaml:"metadata,omitempty"`
	PreserveModTime bool              `json:"preserveModTime,omitempty" yaml:"preserveModTime,omitempty"`
	SourceModTime   *time.Time        `json:"sourceModTime,omitempty" yaml:"sourceModTime,omitempty"`
	PartSize        int64             `json:"partSize,omitempty" yaml:"partSize,omitempty"`       // multipart part size in bytes; 0 = backend default
	Concurrency     int               `json:"concurrency,omitempty" yaml:"concurrency,omitempty"` // parallel part count; 0 = backend default
}

// ProgressFunc is the callback signature for reporting transfer progress.
//   - transferred: cumulative bytes transferred so far
//   - total: total bytes (may be 0 if unknown)
type ProgressFunc func(transferred, total int64)

// ---------------------------------------------------------------------------
// Transfer task (progress snapshot)
// ---------------------------------------------------------------------------

// TransferTask represents the observable state of a single transfer.
// All fields are safe to serialize as JSON for the frontend.
type TransferTask struct {
	ID               string            `json:"id" yaml:"id"`
	DriverName       string            `json:"driverName" yaml:"driverName"`
	InstanceName     string            `json:"instanceName" yaml:"instanceName"`
	Direction        TransferDirection `json:"direction" yaml:"direction"`
	RemotePath       string            `json:"remotePath" yaml:"remotePath"`
	LocalPath        string            `json:"localPath" yaml:"localPath"`
	Status           TransferStatus    `json:"status" yaml:"status"`
	BytesTransferred int64             `json:"bytesTransferred" yaml:"bytesTransferred"`
	TotalBytes       int64             `json:"totalBytes" yaml:"totalBytes"`
	BytesPerSecond   int64             `json:"bytesPerSecond" yaml:"bytesPerSecond"`
	Error            string            `json:"error,omitempty" yaml:"error,omitempty"`
	CreatedAt        time.Time         `json:"createdAt" yaml:"createdAt"`
	StartedAt        *time.Time        `json:"startedAt,omitempty" yaml:"startedAt,omitempty"`
	CompletedAt      *time.Time        `json:"completedAt,omitempty" yaml:"completedAt,omitempty"`
}

// TransferEventType represents a task lifecycle change emitted by the
// transfer manager.
type TransferEventType string

const (
	TransferEventUpsert TransferEventType = "upsert"
	TransferEventRemove TransferEventType = "remove"
	TransferEventError  TransferEventType = "error"
)

// TransferEvent is the incremental event payload emitted when a transfer task
// changes or is removed.
type TransferEvent struct {
	Type    TransferEventType `json:"type" yaml:"type"`
	TaskID  string            `json:"taskId,omitempty" yaml:"taskId,omitempty"`
	Task    *TransferTask     `json:"task,omitempty" yaml:"task,omitempty"`
	Message string            `json:"message,omitempty" yaml:"message,omitempty"`
}

// TransferObserver consumes task lifecycle updates from TransferManager.
type TransferObserver func(event TransferEvent)

// ---------------------------------------------------------------------------
// Transferer – optional driver interface for optimized transfers
// ---------------------------------------------------------------------------

// Transferer is an optional capability for backends that support optimized
// (e.g. multipart / concurrent) upload and download with progress tracking.
//
// Drivers that do NOT implement this interface will fall back to the generic
// Reader/Writer path (single-stream, still with progress tracking).
type Transferer interface {
	// Upload transfers a local file to the remote storage path.
	// progressFn is called periodically with (bytesTransferred, totalBytes).
	Upload(ctx context.Context, req *TransferRequest, progressFn ProgressFunc) error

	// Download transfers a remote storage path to a local file.
	// progressFn is called periodically with (bytesTransferred, totalBytes).
	Download(ctx context.Context, req *TransferRequest, progressFn ProgressFunc) error
}
