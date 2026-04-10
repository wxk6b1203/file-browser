package webdav

import (
	"context"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/wxk6b1203/file-util-manager/folder"
	"github.com/wxk6b1203/file-util-manager/internal/testsupport/webdavtest"
)

func TestDriverCRUDAndScopedRoot(t *testing.T) {
	ctx := context.Background()
	srv := webdavtest.NewServer(t, "", "", "")
	defer srv.Close()

	cfg := &Options{
		Endpoint: srv.URL,
		RootPath: "/scoped",
	}
	mgr, err := New(ctx, &folder.DriverOptions{Name: "webdav-test", Driver: "WebDAV"}, cfg)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	drv := mgr.(*Driver)
	if err := drv.Mkdir(ctx, "nested"); err != nil {
		t.Fatalf("Mkdir() error = %v", err)
	}

	if _, err := drv.Write(ctx, "nested/hello.txt", strings.NewReader("hello"), nil); err != nil {
		t.Fatalf("Write() error = %v", err)
	}

	items, err := drv.List(ctx, "", &folder.ListOptions{})
	if err != nil {
		t.Fatalf("List(root) error = %v", err)
	}
	if len(items) != 1 || items[0].Path != "nested" || !items[0].IsDir() {
		t.Fatalf("unexpected root items: %#v", items)
	}

	recursive, err := drv.List(ctx, "", &folder.ListOptions{Recursive: true})
	if err != nil {
		t.Fatalf("List(recursive) error = %v", err)
	}
	if len(recursive) != 2 {
		t.Fatalf("recursive items len = %d", len(recursive))
	}

	reader, err := drv.Read(ctx, "nested/hello.txt")
	if err != nil {
		t.Fatalf("Read() error = %v", err)
	}
	defer reader.Close()
	body, _ := io.ReadAll(reader)
	if string(body) != "hello" {
		t.Fatalf("unexpected file body %q", string(body))
	}

	if err := drv.Rename(ctx, "nested/hello.txt", "renamed.txt"); err != nil {
		t.Fatalf("Rename() error = %v", err)
	}
	if err := drv.Copy(ctx, folder.PathOp{SrcPath: "nested/renamed.txt", DstPath: "copied.txt"}); err != nil {
		t.Fatalf("Copy() error = %v", err)
	}
	if err := drv.Move(ctx, folder.PathOp{SrcPath: "copied.txt", DstPath: "nested/copied.txt"}); err != nil {
		t.Fatalf("Move() error = %v", err)
	}
	if err := drv.Delete(ctx, "nested/copied.txt"); err != nil {
		t.Fatalf("Delete(file) error = %v", err)
	}
	if err := drv.Delete(ctx, "nested"); err != nil {
		t.Fatalf("Delete(dir) error = %v", err)
	}

	items, err = drv.List(ctx, "", &folder.ListOptions{})
	if err != nil {
		t.Fatalf("List(after delete) error = %v", err)
	}
	if len(items) != 0 {
		t.Fatalf("expected empty root after delete, got %#v", items)
	}
}

func TestDriverBearerAuth(t *testing.T) {
	ctx := context.Background()
	srv := webdavtest.NewServer(t, "", "", "secret-token")
	defer srv.Close()

	mgr, err := New(ctx, &folder.DriverOptions{Name: "webdav-bearer", Driver: "WebDAV"}, &Options{
		Endpoint:    srv.URL,
		BearerToken: "secret-token",
	})
	if err != nil {
		t.Fatalf("New() bearer error = %v", err)
	}

	if err := mgr.(folder.HealthChecker).Ping(ctx); err != nil {
		t.Fatalf("Ping() error = %v", err)
	}
}

func TestDriverBasicAuth(t *testing.T) {
	ctx := context.Background()
	srv := webdavtest.NewServer(t, "user", "password", "")
	defer srv.Close()

	mgr, err := New(ctx, &folder.DriverOptions{Name: "webdav-basic", Driver: "WebDAV"}, &Options{
		Endpoint: srv.URL,
		Username: "user",
		Password: "password",
	})
	if err != nil {
		t.Fatalf("New() basic error = %v", err)
	}

	if err := mgr.(folder.HealthChecker).Ping(ctx); err != nil {
		t.Fatalf("Ping() error = %v", err)
	}
}

func TestDriverAuthFailure(t *testing.T) {
	ctx := context.Background()
	srv := webdavtest.NewServer(t, "user", "password", "")
	defer srv.Close()

	_, err := New(ctx, &folder.DriverOptions{Name: "webdav-basic", Driver: "WebDAV"}, &Options{
		Endpoint: srv.URL,
		Username: "user",
		Password: "wrong-password",
	})
	if err == nil {
		t.Fatal("expected auth failure")
	}
}

func TestDriverRejectsTraversalPath(t *testing.T) {
	ctx := context.Background()
	srv := webdavtest.NewServer(t, "", "", "")
	defer srv.Close()

	mgr, err := New(ctx, &folder.DriverOptions{Name: "webdav-test", Driver: "WebDAV"}, &Options{
		Endpoint: srv.URL,
		RootPath: "/scoped",
	})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	drv := mgr.(*Driver)
	_, err = drv.Stat(ctx, "../outside.txt")
	if err == nil {
		t.Fatal("expected invalid path error")
	}
	if !errors.Is(err, folder.ErrInvalidPath) {
		t.Fatalf("expected ErrInvalidPath, got %v", err)
	}
}

func TestDriverMergesDriverRootIntoWebDAVRootPath(t *testing.T) {
	ctx := context.Background()
	srv := webdavtest.NewServer(t, "", "", "")
	defer srv.Close()

	mgr, err := New(ctx, &folder.DriverOptions{
		Name:   "webdav-test",
		Driver: "WebDAV",
		Root:   "/tenant-a",
	}, &Options{
		Endpoint: srv.URL,
		RootPath: "/scoped",
	})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	drv := mgr.(*Driver)
	if drv.cfg.RootPath != "tenant-a/scoped" {
		t.Fatalf("RootPath = %q", drv.cfg.RootPath)
	}
}
