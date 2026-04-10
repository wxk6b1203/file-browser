package search

import (
	"context"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/wxk6b1203/file-util-manager/connection"
	_ "github.com/wxk6b1203/file-util-manager/folder/webdav"
	"github.com/wxk6b1203/file-util-manager/internal/testsupport/webdavtest"
	"golang.org/x/net/webdav"
)

func TestSearchWebDAVRespectsScopedRoot(t *testing.T) {
	ctx := context.Background()
	srv := webdavtest.NewServer(t, "", "", "")
	defer srv.Close()

	writeWebDAVFixture(t, srv.FS)

	repo := connection.NewFileRepository(t.TempDir() + "/connections.yaml")
	connSvc := connection.NewService(repo)
	searchSvc := NewService(connSvc, 2, 50)

	if _, err := connSvc.Save(ctx, connection.Definition{
		ID:      "webdav-search",
		Name:    "WebDAV Search",
		Driver:  "WebDAV",
		Enabled: true,
		Config: map[string]any{
			"endpoint": srv.URL,
			"rootPath": "/scoped",
		},
	}); err != nil {
		t.Fatalf("save connection: %v", err)
	}

	var (
		mu      sync.Mutex
		results []string
		done    = make(chan struct{})
	)

	_, err := searchSvc.Start(ctx, Request{
		Query:      "alpha",
		MaxResults: 10,
	}, func(event Event) {
		mu.Lock()
		defer mu.Unlock()

		if event.Type == EventTypeResult && event.Result != nil && event.Result.File != nil {
			results = append(results, event.Result.File.Path)
		}
		if event.Type == EventTypeCompleted {
			select {
			case <-done:
			default:
				close(done)
			}
		}
	})
	if err != nil {
		t.Fatalf("Start(): %v", err)
	}

	select {
	case <-done:
	case <-time.After(3 * time.Second):
		t.Fatal("search timed out")
	}

	mu.Lock()
	defer mu.Unlock()
	if len(results) != 1 {
		t.Fatalf("expected one scoped result, got %#v", results)
	}
	if results[0] != "docs/alpha.txt" {
		t.Fatalf("unexpected result path %q", results[0])
	}
}

func writeWebDAVFixture(t *testing.T, fs webdav.FileSystem) {
	t.Helper()

	ctx := context.Background()
	if err := fs.Mkdir(ctx, "/scoped/docs", 0o755); err != nil {
		t.Fatalf("mkdir docs: %v", err)
	}
	if err := fs.Mkdir(ctx, "/other", 0o755); err != nil {
		t.Fatalf("mkdir other: %v", err)
	}
	writeWebDAVFile(t, fs, "/scoped/docs/alpha.txt", "alpha")
	writeWebDAVFile(t, fs, "/other/alpha.txt", "alpha-outside-root")
}

func writeWebDAVFile(t *testing.T, fs webdav.FileSystem, filePath, body string) {
	t.Helper()

	ctx := context.Background()
	file, err := fs.OpenFile(ctx, filePath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0o644)
	if err != nil {
		t.Fatalf("OpenFile(%s): %v", filePath, err)
	}
	defer file.Close()

	if _, err := file.Write([]byte(body)); err != nil {
		t.Fatalf("Write(%s): %v", filePath, err)
	}
}
