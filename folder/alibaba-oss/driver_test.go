package alibaba_oss

import (
	"bytes"
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/wxk6b1203/file-util-manager/folder"
)

// envOrSkip reads an environment variable and skips the test if it is not set.
func envOrSkip(t *testing.T, key string) string {
	t.Helper()
	v := os.Getenv(key)
	if v == "" {
		t.Skipf("environment variable %q is not set, skipping integration test", key)
	}
	return v
}

// newDriverFromEnv constructs an OSS Driver from environment variables:
//
//	OSS_ACCESS_KEY_ID      – Alibaba Cloud access key ID                  (required)
//	OSS_ACCESS_KEY_SECRET  – Alibaba Cloud access key secret              (required)
//	OSS_REGION             – region ID, e.g. "cn-hangzhou"                (required)
//	OSS_BUCKET             – bucket name                                  (required)
//	OSS_ENDPOINT           – custom endpoint URL                          (optional)
//	OSS_PREFIX             – key prefix acting as a virtual root dir      (optional)
//	OSS_SECURITY_TOKEN     – STS temporary security token                 (optional)
//	OSS_FORCE_PATH_STYLE   – "true" to enable path-style addressing       (optional)
//	OSS_USE_CNAME          – "true" if endpoint is a CName                (optional)
//	OSS_DISABLE_SSL        – "true" to use HTTP instead of HTTPS          (optional)
func newDriverFromEnv(t *testing.T) folder.Manager {
	t.Helper()

	cfg := &Options{
		AccessKeyID:     envOrSkip(t, "OSS_ACCESS_KEY_ID"),
		AccessKeySecret: envOrSkip(t, "OSS_ACCESS_KEY_SECRET"),
		Region:          envOrSkip(t, "OSS_REGION"),
		Bucket:          envOrSkip(t, "OSS_BUCKET"),
		Endpoint:        os.Getenv("OSS_ENDPOINT"),
		Prefix:          os.Getenv("OSS_PREFIX"),
		SecurityToken:   os.Getenv("OSS_SECURITY_TOKEN"),
		ForcePathStyle:  os.Getenv("OSS_FORCE_PATH_STYLE") == "true",
		UseCName:        os.Getenv("OSS_USE_CNAME") == "true",
		DisableSSL:      os.Getenv("OSS_DISABLE_SSL") == "true",
	}

	drv, err := New(context.Background(), &folder.DriverOptions{Name: "test-oss"}, cfg)
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
// TestRunOSSDriver – top-level suite that drives the whole lifecycle
// -----------------------------------------------------------------------

func TestRunOSSDriver(t *testing.T) {
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
		if !caps.CanPresign {
			t.Error("expected CanPresign = true")
		}
	})

	t.Run("Mkdir", func(t *testing.T) {
		if err := drv.Mkdir(ctx, dir); err != nil {
			t.Fatalf("Mkdir(%q) error: %v", dir, err)
		}
	})

	const fileName = "hello.txt"
	const fileContent = "hello, alibaba-oss driver"
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

	// --- cleanup: delete all test files and verify removal ----------------
	t.Run("Delete", func(t *testing.T) {
		for _, p := range []string{filePath, renamedPath} {
			if err := drv.Delete(ctx, p); err != nil {
				t.Errorf("Delete(%q) error: %v", p, err)
			}
		}
		for _, p := range []string{filePath, renamedPath} {
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
