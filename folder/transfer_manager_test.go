package folder

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// ---------------------------------------------------------------------------
// Fake driver for testing the transfer manager fallback path
// ---------------------------------------------------------------------------

type fakeDriver struct {
	BaseDriver
	content string // the "remote" file content for downloads
}

func (f *fakeDriver) Capabilities() Capabilities {
	caps := BaseCapabilities()
	caps.CanRead = true
	caps.CanWrite = true
	return caps
}

func (f *fakeDriver) Stat(_ context.Context, _ string) (*FileInfo, error) {
	return &FileInfo{
		Name: "test.txt",
		Path: "test.txt",
		Type: EntryTypeFile,
		Size: int64(len(f.content)),
	}, nil
}

func (f *fakeDriver) Read(_ context.Context, _ string) (io.ReadCloser, error) {
	return io.NopCloser(strings.NewReader(f.content)), nil
}

func (f *fakeDriver) Write(_ context.Context, path string, body io.Reader, _ *WriteOptions) (*FileInfo, error) {
	data, err := io.ReadAll(body)
	if err != nil {
		return nil, err
	}
	f.content = string(data)
	return &FileInfo{
		Name: filepath.Base(path),
		Path: path,
		Type: EntryTypeFile,
		Size: int64(len(data)),
	}, nil
}

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

func TestTransferManager_FallbackUpload(t *testing.T) {
	// Create a temp file as the "local" file to upload.
	tmpDir := t.TempDir()
	localFile := filepath.Join(tmpDir, "upload.txt")
	content := "upload test content"
	if err := os.WriteFile(localFile, []byte(content), 0o644); err != nil {
		t.Fatalf("write temp file: %v", err)
	}

	drv := &fakeDriver{}
	tm := NewTransferManager()

	taskID, err := tm.Submit(drv, "fake", "inst1", TransferUpload, &TransferRequest{
		RemotePath: "remote/upload.txt",
		LocalPath:  localFile,
	})
	if err != nil {
		t.Fatalf("Submit error: %v", err)
	}
	if taskID == "" {
		t.Fatal("expected non-empty task ID")
	}

	// Wait for completion.
	if err := waitForTask(tm, taskID, 5*time.Second); err != nil {
		t.Fatalf("wait error: %v", err)
	}

	task := tm.Progress(taskID)
	if task == nil {
		t.Fatal("task not found")
	}
	if task.Status != TransferCompleted {
		t.Errorf("status = %v, want TransferCompleted; error = %q", task.Status, task.Error)
	}
	if task.BytesTransferred != int64(len(content)) {
		t.Errorf("bytesTransferred = %d, want %d", task.BytesTransferred, len(content))
	}

	// Verify the driver received the data.
	if drv.content != content {
		t.Errorf("driver content = %q, want %q", drv.content, content)
	}
}

func TestTransferManager_FallbackDownload(t *testing.T) {
	content := "download test content"
	drv := &fakeDriver{content: content}
	tm := NewTransferManager()

	tmpDir := t.TempDir()
	localFile := filepath.Join(tmpDir, "sub", "download.txt")

	taskID, err := tm.Submit(drv, "fake", "inst1", TransferDownload, &TransferRequest{
		RemotePath: "remote/download.txt",
		LocalPath:  localFile,
	})
	if err != nil {
		t.Fatalf("Submit error: %v", err)
	}

	if err := waitForTask(tm, taskID, 5*time.Second); err != nil {
		t.Fatalf("wait error: %v", err)
	}

	task := tm.Progress(taskID)
	if task.Status != TransferCompleted {
		t.Errorf("status = %v, want TransferCompleted; error = %q", task.Status, task.Error)
	}

	data, err := os.ReadFile(localFile)
	if err != nil {
		t.Fatalf("read local file: %v", err)
	}
	if string(data) != content {
		t.Errorf("local file content = %q, want %q", string(data), content)
	}
}

func TestTransferManager_Cancel(t *testing.T) {
	// Create a driver whose Read blocks until context is cancelled.
	drv := &slowDriver{}
	tm := NewTransferManager()

	tmpDir := t.TempDir()
	localFile := filepath.Join(tmpDir, "cancelled.txt")

	taskID, err := tm.Submit(drv, "slow", "inst1", TransferDownload, &TransferRequest{
		RemotePath: "remote/big.bin",
		LocalPath:  localFile,
	})
	if err != nil {
		t.Fatalf("Submit error: %v", err)
	}

	// Give it a moment to start.
	time.Sleep(100 * time.Millisecond)

	if err := tm.Cancel(taskID); err != nil {
		t.Fatalf("Cancel error: %v", err)
	}

	if err := waitForTask(tm, taskID, 5*time.Second); err != nil {
		t.Fatalf("wait error: %v", err)
	}

	task := tm.Progress(taskID)
	if task.Status != TransferCancelled {
		t.Errorf("status = %v, want TransferCancelled; error = %q", task.Status, task.Error)
	}
}

func TestTransferManager_List(t *testing.T) {
	content := "list test"
	drv := &fakeDriver{content: content}
	tm := NewTransferManager()

	tmpDir := t.TempDir()

	for i := range 3 {
		localFile := filepath.Join(tmpDir, fmt.Sprintf("file%d.txt", i))
		_ = os.WriteFile(localFile, []byte(content), 0o644)
		_, _ = tm.Submit(drv, "fake", "inst1", TransferUpload, &TransferRequest{
			RemotePath: fmt.Sprintf("remote/file%d.txt", i),
			LocalPath:  localFile,
		})
	}

	// Wait for all to complete.
	time.Sleep(500 * time.Millisecond)

	tasks := tm.List()
	if len(tasks) != 3 {
		t.Errorf("List() returned %d tasks, want 3", len(tasks))
	}
}

func TestTransferManager_RemoveAll(t *testing.T) {
	content := "removeall test"
	drv := &fakeDriver{content: content}
	tm := NewTransferManager()

	tmpDir := t.TempDir()
	localFile := filepath.Join(tmpDir, "rm.txt")
	_ = os.WriteFile(localFile, []byte(content), 0o644)

	taskID, _ := tm.Submit(drv, "fake", "inst1", TransferUpload, &TransferRequest{
		RemotePath: "remote/rm.txt",
		LocalPath:  localFile,
	})

	_ = waitForTask(tm, taskID, 5*time.Second)

	tm.RemoveAll()
	tasks := tm.List()
	if len(tasks) != 0 {
		t.Errorf("List() after RemoveAll returned %d tasks, want 0", len(tasks))
	}
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// slowDriver blocks on Read until context is cancelled.
type slowDriver struct {
	BaseDriver
}

func (s *slowDriver) Capabilities() Capabilities {
	caps := BaseCapabilities()
	caps.CanRead = true
	return caps
}

func (s *slowDriver) Stat(_ context.Context, _ string) (*FileInfo, error) {
	return &FileInfo{Name: "big.bin", Path: "big.bin", Type: EntryTypeFile, Size: 1 << 30}, nil
}

func (s *slowDriver) Read(ctx context.Context, _ string) (io.ReadCloser, error) {
	return &blockingReadCloser{ctx: ctx}, nil
}

type blockingReadCloser struct {
	ctx context.Context
}

func (b *blockingReadCloser) Read(_ []byte) (int, error) {
	<-b.ctx.Done()
	return 0, b.ctx.Err()
}

func (b *blockingReadCloser) Close() error { return nil }

func waitForTask(tm *TransferManager, taskID string, timeout time.Duration) error {
	deadline := time.After(timeout)
	for {
		select {
		case <-deadline:
			return fmt.Errorf("timeout waiting for task %q", taskID)
		default:
			task := tm.Progress(taskID)
			if task == nil {
				return fmt.Errorf("task %q not found", taskID)
			}
			switch task.Status {
			case TransferCompleted, TransferFailed, TransferCancelled:
				return nil
			}
			time.Sleep(50 * time.Millisecond)
		}
	}
}
