package fileops

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/wxk6b1203/file-util-manager/connection"
	_ "github.com/wxk6b1203/file-util-manager/folder/local"
)

func TestListDirectorySortsDirectoriesFirst(t *testing.T) {
	ctx := context.Background()
	root := t.TempDir()

	if err := os.Mkdir(filepath.Join(root, "beta-dir"), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "alpha.txt"), []byte("a"), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	repo := connection.NewFileRepository(filepath.Join(t.TempDir(), "connections.yaml"))
	connSvc := connection.NewService(repo)
	fileSvc := NewService(connSvc)

	if _, err := connSvc.Save(ctx, connection.Definition{
		ID:      "local-list",
		Name:    "Local List",
		Driver:  "Local",
		Enabled: true,
		Config: map[string]any{
			"rootPath": root,
		},
	}); err != nil {
		t.Fatalf("save connection: %v", err)
	}

	items, err := fileSvc.ListDirectory(ctx, "local-list", "")
	if err != nil {
		t.Fatalf("ListDirectory: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("expected two entries, got %d", len(items))
	}
	if !items[0].IsDir() || items[0].Name != "beta-dir" {
		t.Fatalf("expected directory first, got %#v", items[0])
	}
	if items[1].IsDir() || items[1].Name != "alpha.txt" {
		t.Fatalf("expected file second, got %#v", items[1])
	}
}

func TestCreateRenameDeleteDirectory(t *testing.T) {
	ctx := context.Background()
	root := t.TempDir()

	repo := connection.NewFileRepository(filepath.Join(t.TempDir(), "connections.yaml"))
	connSvc := connection.NewService(repo)
	fileSvc := NewService(connSvc)

	if _, err := connSvc.Save(ctx, connection.Definition{
		ID:      "local-mutate",
		Name:    "Local Mutate",
		Driver:  "Local",
		Enabled: true,
		Config: map[string]any{
			"rootPath": root,
		},
	}); err != nil {
		t.Fatalf("save connection: %v", err)
	}

	created, err := fileSvc.CreateDirectory(ctx, "local-mutate", "", "alpha")
	if err != nil {
		t.Fatalf("CreateDirectory: %v", err)
	}
	if created == nil || created.Name != "alpha" || !created.IsDir() {
		t.Fatalf("unexpected created directory: %#v", created)
	}

	renamed, err := fileSvc.RenameEntry(ctx, "local-mutate", "alpha", "beta")
	if err != nil {
		t.Fatalf("RenameEntry: %v", err)
	}
	if renamed == nil || renamed.Name != "beta" {
		t.Fatalf("unexpected renamed directory: %#v", renamed)
	}

	if err := fileSvc.DeleteEntry(ctx, "local-mutate", "beta"); err != nil {
		t.Fatalf("DeleteEntry: %v", err)
	}

	if _, err := os.Stat(filepath.Join(root, "beta")); !os.IsNotExist(err) {
		t.Fatalf("expected renamed directory to be deleted, stat err=%v", err)
	}
}
