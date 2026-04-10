package fileops

import (
	"context"
	"testing"

	"github.com/wxk6b1203/file-util-manager/connection"
	_ "github.com/wxk6b1203/file-util-manager/folder/webdav"
	"github.com/wxk6b1203/file-util-manager/internal/testsupport/webdavtest"
)

func TestWebDAVCreateListDeleteEntry(t *testing.T) {
	ctx := context.Background()
	srv := webdavtest.NewServer(t, "", "", "")
	defer srv.Close()

	repo := connection.NewFileRepository(t.TempDir() + "/connections.yaml")
	connSvc := connection.NewService(repo)
	fileSvc := NewService(connSvc)

	if _, err := connSvc.Save(ctx, connection.Definition{
		ID:      "webdav-fileops",
		Name:    "WebDAV FileOps",
		Driver:  "WebDAV",
		Enabled: true,
		Config: map[string]any{
			"endpoint": srv.URL,
			"rootPath": "/scoped",
		},
	}); err != nil {
		t.Fatalf("save connection: %v", err)
	}

	created, err := fileSvc.CreateDirectory(ctx, "webdav-fileops", "", "docs")
	if err != nil {
		t.Fatalf("CreateDirectory: %v", err)
	}
	if created == nil || !created.IsDir() || created.Path != "docs" {
		t.Fatalf("unexpected created dir: %#v", created)
	}

	items, err := fileSvc.ListDirectory(ctx, "webdav-fileops", "")
	if err != nil {
		t.Fatalf("ListDirectory: %v", err)
	}
	if len(items) != 1 || items[0].Path != "docs" {
		t.Fatalf("unexpected root items: %#v", items)
	}

	if err := fileSvc.DeleteEntry(ctx, "webdav-fileops", "docs"); err != nil {
		t.Fatalf("DeleteEntry: %v", err)
	}

	items, err = fileSvc.ListDirectory(ctx, "webdav-fileops", "")
	if err != nil {
		t.Fatalf("ListDirectory after delete: %v", err)
	}
	if len(items) != 0 {
		t.Fatalf("expected empty root after delete, got %#v", items)
	}
}
