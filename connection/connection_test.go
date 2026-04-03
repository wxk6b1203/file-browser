package connection

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	_ "github.com/wxk6b1203/file-util-manager/folder/local"
)

func TestFileRepositorySaveListDelete(t *testing.T) {
	ctx := context.Background()
	repo := NewFileRepository(filepath.Join(t.TempDir(), "connections.yaml"))

	saved, err := repo.Save(ctx, Definition{
		Name:    "Local Root",
		Driver:  "Local",
		Enabled: true,
		Config: map[string]any{
			"rootPath": ".",
		},
	})
	if err != nil {
		t.Fatalf("Save: %v", err)
	}
	if saved.ID == "" {
		t.Fatalf("expected generated id")
	}

	defs, err := repo.List(ctx)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(defs) != 1 {
		t.Fatalf("expected one connection, got %d", len(defs))
	}

	if err := repo.Delete(ctx, saved.ID); err != nil {
		t.Fatalf("Delete: %v", err)
	}

	defs, err = repo.List(ctx)
	if err != nil {
		t.Fatalf("List after delete: %v", err)
	}
	if len(defs) != 0 {
		t.Fatalf("expected empty repository, got %d entries", len(defs))
	}
}

func TestServiceOpenCloseLocalConnection(t *testing.T) {
	ctx := context.Background()
	root := t.TempDir()
	repo := NewFileRepository(filepath.Join(t.TempDir(), "connections.yaml"))
	svc := NewService(repo)

	saved, err := svc.Save(ctx, Definition{
		ID:      "local-test",
		Name:    "Local Test",
		Driver:  "Local",
		Enabled: true,
		Config: map[string]any{
			"rootPath": root,
		},
	})
	if err != nil {
		t.Fatalf("Save: %v", err)
	}

	state, err := svc.Open(ctx, saved.ID)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	if !state.Connected {
		t.Fatalf("expected connection to be opened")
	}
	if state.Capabilities == nil || !state.Capabilities.CanList {
		t.Fatalf("expected capabilities to be populated: %#v", state.Capabilities)
	}

	if err := svc.Close(ctx, saved.ID); err != nil {
		t.Fatalf("Close: %v", err)
	}

	states := svc.ListStates()
	if len(states) != 1 {
		t.Fatalf("expected one state entry, got %d", len(states))
	}
	if states[0].Connected {
		t.Fatalf("expected connection state to be closed")
	}
}

func TestServiceTestLocalConnectionWithoutSaving(t *testing.T) {
	ctx := context.Background()
	root := t.TempDir()
	repo := NewFileRepository(filepath.Join(t.TempDir(), "connections.yaml"))
	svc := NewService(repo)

	state, err := svc.Test(ctx, Definition{
		Driver:  "Local",
		Enabled: true,
		Config: map[string]any{
			"rootPath": root,
		},
	})
	if err != nil {
		t.Fatalf("Test: %v", err)
	}
	if state == nil || !state.Connected {
		t.Fatalf("expected test connection to succeed: %#v", state)
	}
	if state.Capabilities == nil || !state.Capabilities.CanList {
		t.Fatalf("expected capabilities to be populated: %#v", state)
	}
}

func TestServiceSaveReconnectsConnectedInstanceWithUpdatedConfig(t *testing.T) {
	ctx := context.Background()
	rootA := t.TempDir()
	rootB := t.TempDir()

	if err := os.WriteFile(filepath.Join(rootA, "alpha.txt"), []byte("alpha"), 0o644); err != nil {
		t.Fatalf("write alpha: %v", err)
	}
	if err := os.WriteFile(filepath.Join(rootB, "beta.txt"), []byte("beta"), 0o644); err != nil {
		t.Fatalf("write beta: %v", err)
	}

	repo := NewFileRepository(filepath.Join(t.TempDir(), "connections.yaml"))
	svc := NewService(repo)

	saved, err := svc.Save(ctx, Definition{
		ID:      "local-edit-live",
		Name:    "Local Edit Live",
		Driver:  "Local",
		Enabled: true,
		Config: map[string]any{
			"rootPath": rootA,
		},
	})
	if err != nil {
		t.Fatalf("initial save: %v", err)
	}

	if _, err := svc.Open(ctx, saved.ID); err != nil {
		t.Fatalf("open: %v", err)
	}

	if _, err := svc.Save(ctx, Definition{
		ID:      saved.ID,
		Name:    saved.Name,
		Driver:  saved.Driver,
		Enabled: true,
		Config: map[string]any{
			"rootPath": rootB,
		},
		}); err != nil {
		t.Fatalf("save updated config: %v", err)
	}

	mgr, _, err := svc.Manager(ctx, saved.ID)
	if err != nil {
		t.Fatalf("manager after save: %v", err)
	}

	items, err := mgr.List(ctx, "", nil)
	if err != nil {
		t.Fatalf("list after save: %v", err)
	}
	if len(items) != 1 || items[0].Name != "beta.txt" {
		t.Fatalf("expected reconnected instance to use updated root, got %#v", items)
	}

	states := svc.ListStates()
	if len(states) != 1 || !states[0].Connected {
		t.Fatalf("expected state to remain connected after save, got %#v", states)
	}
}
