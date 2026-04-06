package s3

import (
	"bytes"
	"context"
	"os"
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

// newDriverFromEnv constructs an S3 Driver from environment variables:
//
//	S3_ACCESS_KEY_ID      – AWS / compatible access key ID       (required)
//	S3_ACCESS_KEY_SECRET  – AWS / compatible access key secret   (required)
//	S3_REGION             – region, e.g. "us-east-1"             (required)
//	S3_BUCKET             – bucket name                          (required)
//	S3_ENDPOINT           – custom endpoint URL (MinIO, etc.)    (optional)
//	S3_PREFIX             – key prefix acting as a root dir      (optional)
//	S3_FORCE_PATH_STYLE   – "true" to enable path-style URLs     (optional)
//	S3_DISABLE_SSL        – "true" to use HTTP instead of HTTPS  (optional)
func newDriverFromEnv(t *testing.T) folder.Manager {
	t.Helper()

	cfg := &Options{
		AccessKeyID:     envOrSkip(t, "S3_ACCESS_KEY_ID"),
		AccessKeySecret: envOrSkip(t, "S3_ACCESS_KEY_SECRET"),
		Region:          envOrSkip(t, "S3_REGION"),
		Bucket:          envOrSkip(t, "S3_BUCKET"),
		Endpoint:        os.Getenv("S3_ENDPOINT"),
		Prefix:          os.Getenv("S3_PREFIX"),
		ForcePathStyle:  os.Getenv("S3_FORCE_PATH_STYLE") == "true",
		DisableSSL:      os.Getenv("S3_DISABLE_SSL") == "true",
	}

	drv, err := New(context.Background(), &folder.DriverOptions{Name: "test-s3"}, cfg)
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
// TestRunS3Driver – top-level suite that drives the whole lifecycle
// -----------------------------------------------------------------------

func TestRunS3Driver(t *testing.T) {
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
	})

	t.Run("Mkdir", func(t *testing.T) {
		if err := drv.Mkdir(ctx, dir); err != nil {
			t.Fatalf("Mkdir(%q) error: %v", dir, err)
		}
	})

	const fileName = "hello.txt"
	const fileContent = "hello, s3 driver"
	filePath := dir + "/" + fileName

	t.Run("Write", func(t *testing.T) {
		w, ok := drv.(folder.Writer)
		if !ok {
			t.Skip("driver does not implement Writer")
		}
		fi, err := w.Write(ctx, filePath, strings.NewReader(fileContent), &folder.WriteOptions{
			ContentType: "text/plain",
			Metadata:    map[string]string{"x-test-key": "x-test-value"},
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

	dirMoveSrc := dir + "/move-src"
	dirMoveDstParent := dir + "/move-dst"
	dirMoveDst := dirMoveDstParent + "/move-src"
	dirMoveSrcChild := dirMoveSrc + "/child.txt"
	dirMoveDstChild := dirMoveDst + "/child.txt"

	t.Run("MoveDirectory", func(t *testing.T) {
		w, ok := drv.(folder.Writer)
		if !ok {
			t.Skip("driver does not implement Writer")
		}
		if err := drv.Mkdir(ctx, dirMoveSrc); err != nil {
			t.Fatalf("Mkdir(%q) error: %v", dirMoveSrc, err)
		}
		if err := drv.Mkdir(ctx, dirMoveDstParent); err != nil {
			t.Fatalf("Mkdir(%q) error: %v", dirMoveDstParent, err)
		}
		if _, err := w.Write(ctx, dirMoveSrcChild, strings.NewReader("directory move"), nil); err != nil {
			t.Fatalf("Write(%q) error: %v", dirMoveSrcChild, err)
		}
		if err := drv.Move(ctx, folder.PathOp{SrcPath: dirMoveSrc, DstPath: dirMoveDst}); err != nil {
			t.Fatalf("Move directory error: %v", err)
		}
		srcExists, err := drv.Exist(ctx, dirMoveSrcChild)
		if err != nil {
			t.Fatalf("Exist(%q) after MoveDirectory error: %v", dirMoveSrcChild, err)
		}
		if srcExists {
			t.Errorf("Exist(%q) after MoveDirectory = true, want false", dirMoveSrcChild)
		}
		dstExists, err := drv.Exist(ctx, dirMoveDstChild)
		if err != nil {
			t.Fatalf("Exist(%q) after MoveDirectory error: %v", dirMoveDstChild, err)
		}
		if !dstExists {
			t.Errorf("Exist(%q) after MoveDirectory = false, want true", dirMoveDstChild)
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

	t.Run("PresignRead", func(t *testing.T) {
		ps, ok := drv.(folder.Presigner)
		if !ok {
			t.Skip("driver does not implement Presigner")
		}
		url, err := ps.PresignRead(ctx, filePath, &folder.PresignOptions{Expires: 5 * time.Minute})
		if err != nil {
			t.Fatalf("PresignRead() error: %v", err)
		}
		if url == "" {
			t.Error("PresignRead() returned empty URL")
		}
		t.Logf("PresignRead URL: %s", url)
	})

	t.Run("PresignWrite", func(t *testing.T) {
		ps, ok := drv.(folder.Presigner)
		if !ok {
			t.Skip("driver does not implement Presigner")
		}
		url, err := ps.PresignWrite(ctx, dir+"/upload-target.txt", &folder.PresignOptions{Expires: 5 * time.Minute})
		if err != nil {
			t.Fatalf("PresignWrite() error: %v", err)
		}
		if url == "" {
			t.Error("PresignWrite() returned empty URL")
		}
		t.Logf("PresignWrite URL: %s", url)
	})

	// --- cleanup: delete the whole test directory --------------------------
	t.Run("Delete", func(t *testing.T) {
		// Delete individual files first, then the directory marker.
		for _, p := range []string{filePath, renamedPath, dirMoveDst, dirMoveDstParent} {
			if err := drv.Delete(ctx, p); err != nil {
				t.Errorf("Delete(%q) error: %v", p, err)
			}
		}
		// Confirm they're gone.
		for _, p := range []string{filePath, renamedPath, dirMoveDstChild} {
			ok, err := drv.Exist(ctx, p)
			if err != nil {
				t.Errorf("Exist(%q) after Delete error: %v", p, err)
			}
			if ok {
				t.Errorf("Exist(%q) after Delete = true, want false", p)
			}
		}
	})
}
