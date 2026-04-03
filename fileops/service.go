package fileops

import (
	"context"
	"fmt"
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

	if items == nil {
		return []*folder.FileInfo{}, nil
	}

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

func (s *Service) MoveEntry(ctx context.Context, connectionID, sourcePath, targetDir string) (*folder.FileInfo, error) {
	mgr, _, err := s.connections.Manager(ctx, connectionID)
	if err != nil {
		return nil, err
	}

	cleanSourcePath := cleanPath(sourcePath)
	if cleanSourcePath == "" {
		return nil, fmt.Errorf("source path is required: %w", folder.ErrInvalidPath)
	}

	info, err := mgr.Stat(ctx, cleanSourcePath)
	if err != nil {
		return nil, err
	}
	if info == nil {
		return nil, fmt.Errorf("source path %q: %w", cleanSourcePath, folder.ErrNotFound)
	}

	cleanTargetDir := cleanPath(targetDir)
	targetPath := joinChildPath(cleanTargetDir, path.Base(cleanSourcePath))
	if targetPath == cleanSourcePath {
		return info, nil
	}

	if info.IsDir() && isPathWithin(targetPath, cleanSourcePath) {
		return nil, fmt.Errorf("cannot move directory %q into %q: %w", cleanSourcePath, targetPath, folder.ErrInvalidPath)
	}

	if cleanTargetDir != "" {
		targetInfo, err := mgr.Stat(ctx, cleanTargetDir)
		if err != nil {
			return nil, err
		}
		if targetInfo == nil || !targetInfo.IsDir() {
			return nil, fmt.Errorf("target directory %q: %w", cleanTargetDir, folder.ErrInvalidPath)
		}
	}

	if err := mgr.Move(ctx, folder.PathOp{
		SrcPath: cleanSourcePath,
		DstPath: targetPath,
	}); err != nil {
		return nil, err
	}

	return mgr.Stat(ctx, targetPath)
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

func isPathWithin(candidate, parent string) bool {
	cleanCandidate := cleanPath(candidate)
	cleanParent := cleanPath(parent)
	if cleanCandidate == "" || cleanParent == "" {
		return false
	}
	return cleanCandidate == cleanParent || strings.HasPrefix(cleanCandidate, cleanParent+"/")
}
