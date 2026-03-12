package folder

import (
	"context"
	"fmt"
	"io"
	"os"
	"sort"
	"sync"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
)

// ---------------------------------------------------------------------------
// TransferManager – coordinates async transfers with progress tracking
// ---------------------------------------------------------------------------

const (
	// speedWindowSec is the sliding window (in seconds) used to compute
	// the instantaneous transfer speed.
	speedWindowSec = 3
)

// TransferManager manages the lifecycle of async upload/download tasks.
// It is safe for concurrent use.
type TransferManager struct {
	mu    sync.RWMutex
	tasks map[string]*taskEntry
}

// taskEntry is the internal bookkeeping for a single transfer.
type taskEntry struct {
	task   TransferTask
	cancel context.CancelFunc

	// progress tracking (atomic for lock-free updates from the transfer goroutine)
	bytesTransferred atomic.Int64
	totalBytes       atomic.Int64

	// speed computation
	speedMu      sync.Mutex
	speedSamples []speedSample
}

type speedSample struct {
	at    time.Time
	bytes int64
}

// NewTransferManager creates a ready-to-use TransferManager.
func NewTransferManager() *TransferManager {
	return &TransferManager{tasks: make(map[string]*taskEntry)}
}

// ---------------------------------------------------------------------------
// Submit – start a new async transfer
// ---------------------------------------------------------------------------

// Submit starts an async upload or download task. It returns the task ID
// immediately. The transfer runs in a background goroutine.
//
// mgr is the storage driver (must implement at least Reader or Writer for
// the fallback path; if it also implements Transferer the optimized path
// is used).
func (tm *TransferManager) Submit(
	mgr Manager,
	driverName, instanceName string,
	direction TransferDirection,
	req *TransferRequest,
) (string, error) {
	if req == nil {
		return "", fmt.Errorf("transfer: request is nil")
	}
	if req.RemotePath == "" {
		return "", fmt.Errorf("transfer: remotePath is required")
	}
	if req.LocalPath == "" {
		return "", fmt.Errorf("transfer: localPath is required")
	}

	id := uuid.New().String()
	now := time.Now()

	ctx, cancel := context.WithCancel(context.Background())

	entry := &taskEntry{
		task: TransferTask{
			ID:           id,
			DriverName:   driverName,
			InstanceName: instanceName,
			Direction:    direction,
			RemotePath:   req.RemotePath,
			LocalPath:    req.LocalPath,
			Status:       TransferPending,
			CreatedAt:    now,
		},
		cancel: cancel,
	}

	tm.mu.Lock()
	tm.tasks[id] = entry
	tm.mu.Unlock()

	go tm.run(ctx, entry, mgr, req)
	return id, nil
}

// ---------------------------------------------------------------------------
// Query / control
// ---------------------------------------------------------------------------

// Progress returns a snapshot of the given task, or nil if not found.
func (tm *TransferManager) Progress(taskID string) *TransferTask {
	tm.mu.RLock()
	entry, ok := tm.tasks[taskID]
	tm.mu.RUnlock()
	if !ok {
		return nil
	}
	return tm.snapshot(entry)
}

// List returns snapshots of all tasks (sorted by creation time, newest first).
func (tm *TransferManager) List() []*TransferTask {
	tm.mu.RLock()
	entries := make([]*taskEntry, 0, len(tm.tasks))
	for _, e := range tm.tasks {
		entries = append(entries, e)
	}
	tm.mu.RUnlock()

	result := make([]*TransferTask, len(entries))
	for i, e := range entries {
		result[i] = tm.snapshot(e)
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].CreatedAt.After(result[j].CreatedAt)
	})
	return result
}

// Cancel requests cancellation of a running task.
func (tm *TransferManager) Cancel(taskID string) error {
	tm.mu.RLock()
	entry, ok := tm.tasks[taskID]
	tm.mu.RUnlock()
	if !ok {
		return fmt.Errorf("transfer %q: %w", taskID, ErrNotFound)
	}
	entry.cancel()
	return nil
}

// Remove deletes a task from the manager. Only completed / failed / cancelled
// tasks can be removed. Returns ErrNotFound if the task doesn't exist.
func (tm *TransferManager) Remove(taskID string) error {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	entry, ok := tm.tasks[taskID]
	if !ok {
		return fmt.Errorf("transfer %q: %w", taskID, ErrNotFound)
	}
	switch entry.task.Status {
	case TransferPending, TransferRunning:
		return fmt.Errorf("transfer %q: cannot remove a %v task", taskID, entry.task.Status)
	}
	delete(tm.tasks, taskID)
	return nil
}

// RemoveAll removes all finished (completed/failed/cancelled) tasks.
func (tm *TransferManager) RemoveAll() {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	for id, entry := range tm.tasks {
		switch entry.task.Status {
		case TransferCompleted, TransferFailed, TransferCancelled:
			delete(tm.tasks, id)
		}
	}
}

// ---------------------------------------------------------------------------
// Internal: run the transfer in a goroutine
// ---------------------------------------------------------------------------

func (tm *TransferManager) run(ctx context.Context, entry *taskEntry, mgr Manager, req *TransferRequest) {
	// Mark running.
	now := time.Now()
	entry.task.StartedAt = &now
	entry.task.Status = TransferRunning

	progressFn := func(transferred, total int64) {
		entry.bytesTransferred.Store(transferred)
		if total > 0 {
			entry.totalBytes.Store(total)
		}
		// Record speed sample.
		entry.speedMu.Lock()
		entry.speedSamples = append(entry.speedSamples, speedSample{at: time.Now(), bytes: transferred})
		// Trim samples older than the window.
		cutoff := time.Now().Add(-time.Duration(speedWindowSec) * time.Second)
		start := 0
		for start < len(entry.speedSamples) && entry.speedSamples[start].at.Before(cutoff) {
			start++
		}
		if start > 0 && start < len(entry.speedSamples) {
			entry.speedSamples = entry.speedSamples[start-1:] // keep one sample before cutoff for delta
		}
		entry.speedMu.Unlock()
	}

	var err error
	if t, ok := mgr.(Transferer); ok {
		// Optimized path — driver handles multipart / concurrent transfer.
		switch entry.task.Direction {
		case TransferUpload:
			err = t.Upload(ctx, req, progressFn)
		case TransferDownload:
			err = t.Download(ctx, req, progressFn)
		default:
			err = fmt.Errorf("transfer: unknown direction %d", entry.task.Direction)
		}
	} else {
		// Fallback path — use Reader / Writer with progress wrapping.
		switch entry.task.Direction {
		case TransferUpload:
			err = tm.fallbackUpload(ctx, mgr, req, progressFn)
		case TransferDownload:
			err = tm.fallbackDownload(ctx, mgr, req, progressFn)
		default:
			err = fmt.Errorf("transfer: unknown direction %d", entry.task.Direction)
		}
	}

	// Finalize.
	finishedAt := time.Now()
	entry.task.CompletedAt = &finishedAt
	entry.task.BytesTransferred = entry.bytesTransferred.Load()
	entry.task.TotalBytes = entry.totalBytes.Load()

	if err != nil {
		if ctx.Err() != nil {
			entry.task.Status = TransferCancelled
			entry.task.Error = "cancelled"
		} else {
			entry.task.Status = TransferFailed
			entry.task.Error = err.Error()
		}
	} else {
		entry.task.Status = TransferCompleted
	}
}

// ---------------------------------------------------------------------------
// Fallback: single-stream upload via Writer
// ---------------------------------------------------------------------------

func (tm *TransferManager) fallbackUpload(ctx context.Context, mgr Manager, req *TransferRequest, progressFn ProgressFunc) error {
	w, ok := mgr.(Writer)
	if !ok {
		return fmt.Errorf("transfer: driver does not support upload (no Writer interface)")
	}

	f, err := os.Open(req.LocalPath)
	if err != nil {
		return fmt.Errorf("transfer: open local file %q: %w", req.LocalPath, err)
	}
	defer f.Close()

	info, err := f.Stat()
	if err != nil {
		return fmt.Errorf("transfer: stat local file %q: %w", req.LocalPath, err)
	}
	total := info.Size()

	body := NewProgressReader(f, total, progressFn)

	var opt *WriteOptions
	if req.ContentType != "" || len(req.Metadata) > 0 {
		opt = &WriteOptions{
			ContentType: req.ContentType,
			Metadata:    req.Metadata,
		}
	}

	_, err = w.Write(ctx, req.RemotePath, body, opt)
	return err
}

// ---------------------------------------------------------------------------
// Fallback: single-stream download via Reader
// ---------------------------------------------------------------------------

func (tm *TransferManager) fallbackDownload(ctx context.Context, mgr Manager, req *TransferRequest, progressFn ProgressFunc) error {
	r, ok := mgr.(Reader)
	if !ok {
		return fmt.Errorf("transfer: driver does not support download (no Reader interface)")
	}

	// Get total size for progress reporting.
	var total int64
	fi, err := mgr.Stat(ctx, req.RemotePath)
	if err == nil && fi != nil {
		total = fi.Size
	}

	rc, err := r.Read(ctx, req.RemotePath)
	if err != nil {
		return fmt.Errorf("transfer: read %q: %w", req.RemotePath, err)
	}
	defer rc.Close()

	// Ensure parent directory exists.
	if dir := parentDir(req.LocalPath); dir != "" {
		if mkErr := os.MkdirAll(dir, 0o755); mkErr != nil {
			return fmt.Errorf("transfer: mkdir %q: %w", dir, mkErr)
		}
	}

	f, err := os.Create(req.LocalPath)
	if err != nil {
		return fmt.Errorf("transfer: create local file %q: %w", req.LocalPath, err)
	}
	defer f.Close()

	pw := NewProgressWriter(f, total, progressFn)

	buf := make([]byte, 256<<10) // 256 KiB
	if _, err := io.CopyBuffer(pw, rc, buf); err != nil {
		return fmt.Errorf("transfer: download %q: %w", req.RemotePath, err)
	}
	return nil
}

// ---------------------------------------------------------------------------
// Snapshot helper
// ---------------------------------------------------------------------------

func (tm *TransferManager) snapshot(e *taskEntry) *TransferTask {
	t := e.task // shallow copy
	t.BytesTransferred = e.bytesTransferred.Load()
	t.TotalBytes = e.totalBytes.Load()
	t.BytesPerSecond = tm.computeSpeed(e)
	return &t
}

func (tm *TransferManager) computeSpeed(e *taskEntry) int64 {
	e.speedMu.Lock()
	defer e.speedMu.Unlock()

	n := len(e.speedSamples)
	if n < 2 {
		return 0
	}
	first := e.speedSamples[0]
	last := e.speedSamples[n-1]
	elapsed := last.at.Sub(first.at).Seconds()
	if elapsed <= 0 {
		return 0
	}
	return int64(float64(last.bytes-first.bytes) / elapsed)
}

// parentDir returns the parent directory of a file path.
func parentDir(p string) string {
	for i := len(p) - 1; i >= 0; i-- {
		if p[i] == '/' || p[i] == '\\' {
			return p[:i]
		}
	}
	return ""
}
