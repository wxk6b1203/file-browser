package fileops

import (
	"context"
	"path"
	"sort"
	"strings"

	"github.com/wxk6b1203/file-util-manager/connection"
	"github.com/wxk6b1203/file-util-manager/folder"
)

type Service struct {
	connections *connection.Service
}

func NewService(connections *connection.Service) *Service {
	return &Service{connections: connections}
}

func (s *Service) ListDirectory(ctx context.Context, connectionID, dir string) ([]*folder.FileInfo, error) {
	mgr, _, err := s.connections.Manager(ctx, connectionID)
	if err != nil {
		return nil, err
	}

	items, err := mgr.List(ctx, strings.TrimSpace(dir), &folder.ListOptions{})
	if err != nil {
		return nil, err
	}

	sort.Slice(items, func(i, j int) bool {
		left := items[i]
		right := items[j]
		if left.IsDir() != right.IsDir() {
			return left.IsDir()
		}
		return strings.ToLower(left.Name) < strings.ToLower(right.Name)
	})

	return items, nil
}

func (s *Service) CreateDirectory(ctx context.Context, connectionID, parentDir, name string) (*folder.FileInfo, error) {
	mgr, _, err := s.connections.Manager(ctx, connectionID)
	if err != nil {
		return nil, err
	}

	targetPath := joinChildPath(parentDir, name)
	if err := mgr.Mkdir(ctx, targetPath); err != nil {
		return nil, err
	}
	return mgr.Stat(ctx, targetPath)
}

func (s *Service) RenameEntry(ctx context.Context, connectionID, targetPath, newName string) (*folder.FileInfo, error) {
	mgr, _, err := s.connections.Manager(ctx, connectionID)
	if err != nil {
		return nil, err
	}

	cleanPath := cleanPath(targetPath)
	if err := mgr.Rename(ctx, cleanPath, strings.TrimSpace(newName)); err != nil {
		return nil, err
	}

	nextPath := joinChildPath(path.Dir(cleanPath), newName)
	return mgr.Stat(ctx, nextPath)
}

func (s *Service) DeleteEntry(ctx context.Context, connectionID, targetPath string) error {
	mgr, _, err := s.connections.Manager(ctx, connectionID)
	if err != nil {
		return err
	}

	return mgr.Delete(ctx, cleanPath(targetPath))
}

func joinChildPath(parentDir, name string) string {
	parent := cleanPath(parentDir)
	child := strings.Trim(strings.TrimSpace(name), "/")
	if parent == "" || parent == "." {
		return child
	}
	if child == "" {
		return parent
	}
	return path.Join(parent, child)
}

func cleanPath(value string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" || trimmed == "." || trimmed == "/" {
		return ""
	}
	cleaned := path.Clean(strings.ReplaceAll(trimmed, "\\", "/"))
	if cleaned == "." || cleaned == "/" {
		return ""
	}
	return strings.TrimPrefix(cleaned, "/")
}
