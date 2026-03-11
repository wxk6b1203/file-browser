package sftp

import (
	"bytes"
	"context"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/wxk6b1203/file-util-manager/folder"
)

// envOrSkip reads an environment variable; skips the test if it is not set.
func envOrSkip(t *testing.T, key string) string {
	t.Helper()
	v := os.Getenv(key)
	if v == "" {
		t.Skipf("environment variable %q is not set, skipping integration test", key)
	}
	return v
}

// newDriverFromEnv constructs an SFTP Driver from environment variables:
//
//	SFTP_ADDRESS          – hostname or IP of the SFTP server               (required)
//	SFTP_USERNAME         – SSH username                                    (required)
//	SFTP_PASSWORD         – password authentication                         (required if no private key)
//	SFTP_PRIVATE_KEY      – PEM-encoded private key                         (required if no password)
//	SFTP_PASSPHRASE       – passphrase for encrypted private key            (optional)
//	SFTP_PORT             – SSH port, default 22                            (optional)
//	SFTP_ROOT_PATH        – remote root path for scoping operations         (optional)
//	SFTP_DIAL_TIMEOUT_SEC – connection timeout in seconds                   (optional)
func newDriverFromEnv(t *testing.T) folder.Manager {
	t.Helper()

	address := envOrSkip(t, "SFTP_ADDRESS")
	username := envOrSkip(t, "SFTP_USERNAME")

	password := os.Getenv("SFTP_PASSWORD")
	privateKey := os.Getenv("SFTP_PRIVATE_KEY")
	if password == "" && privateKey == "" {
		t.Skip("neither SFTP_PASSWORD nor SFTP_PRIVATE_KEY is set, skipping integration test")
	}

	port := 22
	if v := os.Getenv("SFTP_PORT"); v != "" {
		p, err := strconv.Atoi(v)
		if err != nil {
			t.Fatalf("SFTP_PORT %q is not a valid integer: %v", v, err)
		}
		port = p
	}

	dialTimeout := 0
	if v := os.Getenv("SFTP_DIAL_TIMEOUT_SEC"); v != "" {
		d, err := strconv.Atoi(v)
		if err != nil {
			t.Fatalf("SFTP_DIAL_TIMEOUT_SEC %q is not a valid integer: %v", v, err)
		}
		dialTimeout = d
	}

	cfg := &Options{
		Address:        address,
		Port:           port,
		Username:       username,
		Password:       password,
		PrivateKey:     privateKey,
		Passphrase:     os.Getenv("SFTP_PASSPHRASE"),
		RootPath:       os.Getenv("SFTP_ROOT_PATH"),
		DialTimeoutSec: dialTimeout,
	}

	drv, err := New(context.Background(), &folder.DriverOptions{Name: "test-sftp"}, cfg)
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}
	t.Cleanup(func() {
		if c, ok := drv.(folder.Closer); ok {
			_ = c.Close()
		}
	})
	return drv
}

// testDir returns a unique test directory path scoped to this test run so
// parallel / repeated runs don't collide with each other.
func testDir(t *testing.T) string {
	return "test-runs/" + t.Name() + "-" + time.Now().Format("20060102-150405")
}

// -----------------------------------------------------------------------
// TestRunSFTPDriver – top-level suite that drives the whole lifecycle
// -----------------------------------------------------------------------

func TestRunSFTPDriver(t *testing.T) {
	drv := newDriverFromEnv(t)
	ctx := context.Background()
	dir := testDir(t)

	t.Run("Ping", func(t *testing.T) {
		hc, ok := drv.(folder.HealthChecker)
		if !ok {
			t.Skip("driver does not implement HealthChecker")
		}
		if err := hc.Ping(ctx); err != nil {
			t.Fatalf("Ping() error: %v", err)
		}
	})

	t.Run("Capabilities", func(t *testing.T) {
		caps := drv.Capabilities()
		if !caps.CanRead {
			t.Error("expected CanRead = true")
		}
		if !caps.CanWrite {
			t.Error("expected CanWrite = true")
		}
		if !caps.AtomicMove {
			t.Error("expected AtomicMove = true")
		}
	})

	t.Run("Mkdir", func(t *testing.T) {
		if err := drv.Mkdir(ctx, dir); err != nil {
			t.Fatalf("Mkdir(%q) error: %v", dir, err)
		}
	})

	const fileName = "hello.txt"
	const fileContent = "hello, sftp driver"
	filePath := dir + "/" + fileName

	t.Run("Write", func(t *testing.T) {
		w, ok := drv.(folder.Writer)
		if !ok {
			t.Skip("driver does not implement Writer")
		}
		fi, err := w.Write(ctx, filePath, strings.NewReader(fileContent), &folder.WriteOptions{
			ContentType: "text/plain",
		})
		if err != nil {
			t.Fatalf("Write() error: %v", err)
		}
		if fi.Name != fileName {
			t.Errorf("Write() fi.Name = %q, want %q", fi.Name, fileName)
		}
	})

	t.Run("Stat", func(t *testing.T) {
		fi, err := drv.Stat(ctx, filePath)
		if err != nil {
			t.Fatalf("Stat(%q) error: %v", filePath, err)
		}
		if fi.Name != fileName {
			t.Errorf("Stat() Name = %q, want %q", fi.Name, fileName)
		}
		if fi.Size != int64(len(fileContent)) {
			t.Errorf("Stat() Size = %d, want %d", fi.Size, len(fileContent))
		}
		if fi.Type != folder.EntryTypeFile {
			t.Errorf("Stat() Type = %v, want EntryTypeFile", fi.Type)
		}
	})

	t.Run("Exist", func(t *testing.T) {
		ok, err := drv.Exist(ctx, filePath)
		if err != nil {
			t.Fatalf("Exist() error: %v", err)
		}
		if !ok {
			t.Errorf("Exist(%q) = false, want true", filePath)
		}
	})

	t.Run("ExistNonExistent", func(t *testing.T) {
		ok, err := drv.Exist(ctx, dir+"/does-not-exist.txt")
		if err != nil {
			t.Fatalf("Exist() error: %v", err)
		}
		if ok {
			t.Error("Exist() = true for non-existent path, want false")
		}
	})

	t.Run("List", func(t *testing.T) {
		files, err := drv.List(ctx, dir, nil)
		if err != nil {
			t.Fatalf("List(%q) error: %v", dir, err)
		}
		found := false
		for _, f := range files {
			if f.Name == fileName {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("List(%q): file %q not found in listing", dir, fileName)
		}
	})

	t.Run("ListWithPrefix", func(t *testing.T) {
		files, err := drv.List(ctx, dir, &folder.ListOptions{Prefix: "hello"})
		if err != nil {
			t.Fatalf("List(%q, prefix=hello) error: %v", dir, err)
		}
		if len(files) != 1 {
			t.Errorf("List with prefix: expected 1 entry, got %d", len(files))
		}
		if len(files) > 0 && files[0].Name != fileName {
			t.Errorf("List with prefix: got %q, want %q", files[0].Name, fileName)
		}
	})

	t.Run("ListWithLimit", func(t *testing.T) {
		// Write a second file so we can test the limit.
		w, ok := drv.(folder.Writer)
		if !ok {
			t.Skip("driver does not implement Writer")
		}
		secondPath := dir + "/second.txt"
		if _, err := w.Write(ctx, secondPath, strings.NewReader("second"), nil); err != nil {
			t.Fatalf("Write second file error: %v", err)
		}

		files, err := drv.List(ctx, dir, &folder.ListOptions{Limit: 1})
		if err != nil {
			t.Fatalf("List(%q, limit=1) error: %v", dir, err)
		}
		if len(files) != 1 {
			t.Errorf("List with limit=1: expected 1 entry, got %d", len(files))
		}

		// Clean up the second file.
		_ = drv.Delete(ctx, secondPath)
	})

	t.Run("Read", func(t *testing.T) {
		r, ok := drv.(folder.Reader)
		if !ok {
			t.Skip("driver does not implement Reader")
		}
		rc, err := r.Read(ctx, filePath)
		if err != nil {
			t.Fatalf("Read(%q) error: %v", filePath, err)
		}
		defer func() { _ = rc.Close() }()

		buf := new(bytes.Buffer)
		if _, err = buf.ReadFrom(rc); err != nil {
			t.Fatalf("Read() body read error: %v", err)
		}
		if got := buf.String(); got != fileContent {
			t.Errorf("Read() content = %q, want %q", got, fileContent)
		}
	})

	copyPath := dir + "/hello-copy.txt"

	t.Run("Copy", func(t *testing.T) {
		if err := drv.Copy(ctx, folder.PathOp{SrcPath: filePath, DstPath: copyPath}); err != nil {
			t.Fatalf("Copy() error: %v", err)
		}
		ok, err := drv.Exist(ctx, copyPath)
		if err != nil {
			t.Fatalf("Exist() after Copy error: %v", err)
		}
		if !ok {
			t.Errorf("Exist(%q) after Copy = false, want true", copyPath)
		}

		// Verify copied content.
		r, ok := drv.(folder.Reader)
		if !ok {
			t.Skip("driver does not implement Reader")
		}
		rc, err := r.Read(ctx, copyPath)
		if err != nil {
			t.Fatalf("Read(%q) error: %v", copyPath, err)
		}
		defer func() { _ = rc.Close() }()
		buf := new(bytes.Buffer)
		if _, err = buf.ReadFrom(rc); err != nil {
			t.Fatalf("Read copy body error: %v", err)
		}
		if got := buf.String(); got != fileContent {
			t.Errorf("Copy content = %q, want %q", got, fileContent)
		}
	})

	movedPath := dir + "/hello-moved.txt"

	t.Run("Move", func(t *testing.T) {
		if err := drv.Move(ctx, folder.PathOp{SrcPath: copyPath, DstPath: movedPath}); err != nil {
			t.Fatalf("Move() error: %v", err)
		}
		srcExists, err := drv.Exist(ctx, copyPath)
		if err != nil {
			t.Fatalf("Exist() after Move (src) error: %v", err)
		}
		if srcExists {
			t.Errorf("Exist(%q) after Move = true, want false (src should be gone)", copyPath)
		}
		dstExists, err := drv.Exist(ctx, movedPath)
		if err != nil {
			t.Fatalf("Exist() after Move (dst) error: %v", err)
		}
		if !dstExists {
			t.Errorf("Exist(%q) after Move = false, want true (dst should exist)", movedPath)
		}
	})

	renamedPath := dir + "/hello-renamed.txt"

	t.Run("Rename", func(t *testing.T) {
		if err := drv.Rename(ctx, movedPath, "hello-renamed.txt"); err != nil {
			t.Fatalf("Rename() error: %v", err)
		}
		ok, err := drv.Exist(ctx, renamedPath)
		if err != nil {
			t.Fatalf("Exist() after Rename error: %v", err)
		}
		if !ok {
			t.Errorf("Exist(%q) after Rename = false, want true", renamedPath)
		}
	})

	t.Run("MkdirNested", func(t *testing.T) {
		nested := dir + "/a/b/c"
		if err := drv.Mkdir(ctx, nested); err != nil {
			t.Fatalf("Mkdir(%q) error: %v", nested, err)
		}
		ok, err := drv.Exist(ctx, nested)
		if err != nil {
			t.Fatalf("Exist(%q) error: %v", nested, err)
		}
		if !ok {
			t.Errorf("Exist(%q) = false after Mkdir, want true", nested)
		}
	})

	t.Run("ListRecursive", func(t *testing.T) {
		files, err := drv.List(ctx, dir, &folder.ListOptions{Recursive: true})
		if err != nil {
			t.Fatalf("List(%q, recursive) error: %v", dir, err)
		}
		if len(files) == 0 {
			t.Error("List recursive returned 0 entries, expected at least 1")
		}
		t.Logf("List recursive returned %d entries:", len(files))
		for _, f := range files {
			t.Logf("  %s (type=%d, size=%d)", f.Path, f.Type, f.Size)
		}
	})

	t.Run("CopyDir", func(t *testing.T) {
		// Write a file inside a subdirectory to test recursive copy.
		w, ok := drv.(folder.Writer)
		if !ok {
			t.Skip("driver does not implement Writer")
		}
		srcDir := dir + "/copy-src"
		if _, err := w.Write(ctx, srcDir+"/nested/file.txt", strings.NewReader("nested content"), nil); err != nil {
			t.Fatalf("Write nested file error: %v", err)
		}

		dstDir := dir + "/copy-dst"
		if err := drv.Copy(ctx, folder.PathOp{SrcPath: srcDir, DstPath: dstDir}); err != nil {
			t.Fatalf("Copy dir error: %v", err)
		}

		ok2, err := drv.Exist(ctx, dstDir+"/nested/file.txt")
		if err != nil {
			t.Fatalf("Exist() after CopyDir error: %v", err)
		}
		if !ok2 {
			t.Error("expected nested file to exist after directory copy")
		}
	})

	t.Run("WriteAutoCreateParent", func(t *testing.T) {
		w, ok := drv.(folder.Writer)
		if !ok {
			t.Skip("driver does not implement Writer")
		}
		deepPath := dir + "/auto/parent/deep/file.txt"
		fi, err := w.Write(ctx, deepPath, strings.NewReader("auto parent"), nil)
		if err != nil {
			t.Fatalf("Write with auto parent error: %v", err)
		}
		if fi.Name != "file.txt" {
			t.Errorf("Write fi.Name = %q, want %q", fi.Name, "file.txt")
		}
		ok2, err := drv.Exist(ctx, deepPath)
		if err != nil {
			t.Fatalf("Exist() error: %v", err)
		}
		if !ok2 {
			t.Errorf("Exist(%q) = false after Write, want true", deepPath)
		}
	})

	t.Run("StatNotFound", func(t *testing.T) {
		_, err := drv.Stat(ctx, dir+"/definitely-not-here.txt")
		if !folder.IsNotFound(err) {
			t.Errorf("Stat on non-existent file: expected ErrNotFound, got %v", err)
		}
	})

	t.Run("ReadNotFound", func(t *testing.T) {
		r, ok := drv.(folder.Reader)
		if !ok {
			t.Skip("driver does not implement Reader")
		}
		_, err := r.Read(ctx, dir+"/definitely-not-here.txt")
		if !folder.IsNotFound(err) {
			t.Errorf("Read on non-existent file: expected ErrNotFound, got %v", err)
		}
	})

	t.Run("ListNotFound", func(t *testing.T) {
		_, err := drv.List(ctx, dir+"/nonexistent-dir", nil)
		if !folder.IsNotFound(err) {
			t.Errorf("List on non-existent dir: expected ErrNotFound, got %v", err)
		}
	})

	// --- cleanup: delete the whole test directory --------------------------
	t.Run("Delete", func(t *testing.T) {
		// Delete the top-level test directory recursively.
		if err := drv.Delete(ctx, dir); err != nil {
			t.Fatalf("Delete(%q) error: %v", dir, err)
		}

		// Confirm it's gone.
		ok, err := drv.Exist(ctx, dir)
		if err != nil {
			t.Fatalf("Exist(%q) after Delete error: %v", dir, err)
		}
		if ok {
			t.Errorf("Exist(%q) after Delete = true, want false", dir)
		}
	})
}
