package local

import (
	"bytes"
	"context"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/wxk6b1203/file-util-manager/folder"
)

// newTestDriver creates a local driver rooted in a temporary directory.
func newTestDriver(t *testing.T) (folder.Manager, string) {
	t.Helper()
	root := t.TempDir()
	mgr, err := New(context.Background(), &folder.DriverOptions{
		Name:   "test-local",
		Driver: "Local",
	}, &Options{RootPath: root})
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}
	return mgr, root
}

func TestMkdirAndList(t *testing.T) {
	mgr, _ := newTestDriver(t)
	ctx := context.Background()

	if err := mgr.Mkdir(ctx, "subdir"); err != nil {
		t.Fatalf("Mkdir error: %v", err)
	}

	items, err := mgr.List(ctx, ".", nil)
	if err != nil {
		t.Fatalf("List error: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(items))
	}
	if items[0].Name != "subdir" {
		t.Errorf("expected name 'subdir', got %q", items[0].Name)
	}
	if items[0].Type != folder.EntryTypeDirectory {
		t.Errorf("expected directory type, got %v", items[0].Type)
	}
}

func TestWriteReadStat(t *testing.T) {
	mgr, _ := newTestDriver(t)
	ctx := context.Background()

	w, ok := mgr.(folder.Writer)
	if !ok {
		t.Fatal("driver does not implement Writer")
	}
	r, ok := mgr.(folder.Reader)
	if !ok {
		t.Fatal("driver does not implement Reader")
	}

	content := []byte("hello local driver")
	fi, err := w.Write(ctx, "test.txt", bytes.NewReader(content), nil)
	if err != nil {
		t.Fatalf("Write error: %v", err)
	}
	if fi.Size != int64(len(content)) {
		t.Errorf("Write returned size %d, want %d", fi.Size, len(content))
	}

	// Stat
	info, err := mgr.Stat(ctx, "test.txt")
	if err != nil {
		t.Fatalf("Stat error: %v", err)
	}
	if info.Size != int64(len(content)) {
		t.Errorf("Stat size %d, want %d", info.Size, len(content))
	}
	if info.Type != folder.EntryTypeFile {
		t.Errorf("Stat type %v, want file", info.Type)
	}

	// Read
	rc, err := r.Read(ctx, "test.txt")
	if err != nil {
		t.Fatalf("Read error: %v", err)
	}
	defer rc.Close()
	got, err := io.ReadAll(rc)
	if err != nil {
		t.Fatalf("ReadAll error: %v", err)
	}
	if !bytes.Equal(got, content) {
		t.Errorf("Read content = %q, want %q", got, content)
	}
}

func TestExist(t *testing.T) {
	mgr, _ := newTestDriver(t)
	ctx := context.Background()

	exists, err := mgr.Exist(ctx, "nope.txt")
	if err != nil {
		t.Fatalf("Exist error: %v", err)
	}
	if exists {
		t.Error("expected Exist=false for non-existent file")
	}

	if err := mgr.Mkdir(ctx, "exists-dir"); err != nil {
		t.Fatalf("Mkdir error: %v", err)
	}
	exists, err = mgr.Exist(ctx, "exists-dir")
	if err != nil {
		t.Fatalf("Exist error: %v", err)
	}
	if !exists {
		t.Error("expected Exist=true for created dir")
	}
}

func TestRename(t *testing.T) {
	mgr, root := newTestDriver(t)
	ctx := context.Background()

	w := mgr.(folder.Writer)
	if _, err := w.Write(ctx, "old.txt", bytes.NewReader([]byte("data")), nil); err != nil {
		t.Fatalf("Write error: %v", err)
	}

	if err := mgr.Rename(ctx, "old.txt", "new.txt"); err != nil {
		t.Fatalf("Rename error: %v", err)
	}

	// old should not exist
	if _, err := os.Stat(filepath.Join(root, "old.txt")); !os.IsNotExist(err) {
		t.Error("old.txt should not exist after rename")
	}
	// new should exist
	if _, err := os.Stat(filepath.Join(root, "new.txt")); err != nil {
		t.Errorf("new.txt should exist after rename: %v", err)
	}
}

func TestDelete(t *testing.T) {
	mgr, _ := newTestDriver(t)
	ctx := context.Background()

	w := mgr.(folder.Writer)
	if _, err := w.Write(ctx, "del.txt", bytes.NewReader([]byte("bye")), nil); err != nil {
		t.Fatalf("Write error: %v", err)
	}

	if err := mgr.Delete(ctx, "del.txt"); err != nil {
		t.Fatalf("Delete error: %v", err)
	}

	exists, _ := mgr.Exist(ctx, "del.txt")
	if exists {
		t.Error("file should not exist after delete")
	}
}

func TestDeleteDir(t *testing.T) {
	mgr, _ := newTestDriver(t)
	ctx := context.Background()

	w := mgr.(folder.Writer)
	if _, err := w.Write(ctx, "dir/sub/file.txt", bytes.NewReader([]byte("nested")), nil); err != nil {
		t.Fatalf("Write error: %v", err)
	}

	if err := mgr.Delete(ctx, "dir"); err != nil {
		t.Fatalf("Delete dir error: %v", err)
	}

	exists, _ := mgr.Exist(ctx, "dir")
	if exists {
		t.Error("dir should not exist after recursive delete")
	}
}

func TestCopyFile(t *testing.T) {
	mgr, _ := newTestDriver(t)
	ctx := context.Background()

	w := mgr.(folder.Writer)
	r := mgr.(folder.Reader)
	content := []byte("copy me")
	if _, err := w.Write(ctx, "src.txt", bytes.NewReader(content), nil); err != nil {
		t.Fatalf("Write error: %v", err)
	}

	if err := mgr.Copy(ctx, folder.PathOp{SrcPath: "src.txt", DstPath: "dst.txt"}); err != nil {
		t.Fatalf("Copy error: %v", err)
	}

	// Verify dst content
	rc, err := r.Read(ctx, "dst.txt")
	if err != nil {
		t.Fatalf("Read dst error: %v", err)
	}
	defer rc.Close()
	got, _ := io.ReadAll(rc)
	if !bytes.Equal(got, content) {
		t.Errorf("Copy content = %q, want %q", got, content)
	}

	// src should still exist
	exists, _ := mgr.Exist(ctx, "src.txt")
	if !exists {
		t.Error("src should still exist after copy")
	}
}

func TestMoveFile(t *testing.T) {
	mgr, _ := newTestDriver(t)
	ctx := context.Background()

	w := mgr.(folder.Writer)
	if _, err := w.Write(ctx, "moveme.txt", bytes.NewReader([]byte("moved")), nil); err != nil {
		t.Fatalf("Write error: %v", err)
	}

	if err := mgr.Move(ctx, folder.PathOp{SrcPath: "moveme.txt", DstPath: "moved.txt"}); err != nil {
		t.Fatalf("Move error: %v", err)
	}

	exists, _ := mgr.Exist(ctx, "moveme.txt")
	if exists {
		t.Error("source should not exist after move")
	}
	exists, _ = mgr.Exist(ctx, "moved.txt")
	if !exists {
		t.Error("destination should exist after move")
	}
}

func TestListRecursive(t *testing.T) {
	mgr, _ := newTestDriver(t)
	ctx := context.Background()

	w := mgr.(folder.Writer)
	if _, err := w.Write(ctx, "a.txt", bytes.NewReader([]byte("a")), nil); err != nil {
		t.Fatal(err)
	}
	if _, err := w.Write(ctx, "sub/b.txt", bytes.NewReader([]byte("b")), nil); err != nil {
		t.Fatal(err)
	}

	items, err := mgr.List(ctx, ".", &folder.ListOptions{Recursive: true})
	if err != nil {
		t.Fatalf("List recursive error: %v", err)
	}

	// Should have: a.txt, sub/, sub/b.txt = 3 entries
	if len(items) != 3 {
		t.Errorf("expected 3 entries, got %d", len(items))
		for _, item := range items {
			t.Logf("  %s (type=%d)", item.Path, item.Type)
		}
	}
}

func TestPing(t *testing.T) {
	mgr, _ := newTestDriver(t)
	ctx := context.Background()

	hc, ok := mgr.(folder.HealthChecker)
	if !ok {
		t.Fatal("driver does not implement HealthChecker")
	}
	if err := hc.Ping(ctx); err != nil {
		t.Fatalf("Ping error: %v", err)
	}
}

func TestCapabilities(t *testing.T) {
	mgr, _ := newTestDriver(t)
	caps := mgr.Capabilities()

	if !caps.CanList {
		t.Error("expected CanList")
	}
	if !caps.CanRead {
		t.Error("expected CanRead")
	}
	if !caps.CanWrite {
		t.Error("expected CanWrite")
	}
	if !caps.CanDelete {
		t.Error("expected CanDelete")
	}
	if !caps.CanCopy {
		t.Error("expected CanCopy")
	}
	if !caps.CanMove {
		t.Error("expected CanMove")
	}
	if !caps.CanRename {
		t.Error("expected CanRename")
	}
	if !caps.CanMkdir {
		t.Error("expected CanMkdir")
	}
	if !caps.AtomicMove {
		t.Error("expected AtomicMove")
	}
}

func TestNotFoundErrors(t *testing.T) {
	mgr, _ := newTestDriver(t)
	ctx := context.Background()

	_, err := mgr.Stat(ctx, "nonexistent.txt")
	if !folder.IsNotFound(err) {
		t.Errorf("Stat on missing file: expected ErrNotFound, got %v", err)
	}

	r := mgr.(folder.Reader)
	_, err = r.Read(ctx, "nonexistent.txt")
	if !folder.IsNotFound(err) {
		t.Errorf("Read on missing file: expected ErrNotFound, got %v", err)
	}

	_, err = mgr.List(ctx, "nonexistent-dir", nil)
	if !folder.IsNotFound(err) {
		t.Errorf("List on missing dir: expected ErrNotFound, got %v", err)
	}
}
