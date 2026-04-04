package local

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/wxk6b1203/file-util-manager/folder"
)

const (
	// copyBufSize is the buffer used by io.CopyBuffer (256 KiB).
	copyBufSize = 256 << 10

	// maxRecursionDepth guards recursive operations against pathological
	// nesting or symlink loops.
	maxRecursionDepth = 128
)

func init() {
	folder.RegisterDriver[Options]("Local",
		"Local file system — direct access to files and directories on the local machine.",
		New,
	)
}

// Driver implements folder.Manager, folder.Reader, folder.Writer,
// folder.HealthChecker and folder.Closer for the local file system.
type Driver struct {
	folder.BaseDriver
	cfg *Options
}

// ---------------------------------------------------------------------------
// Constructor
// ---------------------------------------------------------------------------

// New creates a new local file system driver.
func New(_ context.Context, opt *folder.DriverOptions, cfg *Options) (folder.Manager, error) {
	if cfg == nil {
		cfg = &Options{}
	}

	// Merge DriverOptions.Root into rootPath when specified.
	if opt != nil && opt.Root != "" {
		root := filepath.Clean(opt.Root)
		if cfg.RootPath != "" {
			cfg.RootPath = filepath.Join(root, cfg.RootPath)
		} else {
			cfg.RootPath = root
		}
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("local: invalid configuration: %w", err)
	}

	d := &Driver{
		BaseDriver: folder.NewBaseDriver(opt),
		cfg:        cfg,
	}
	return d, nil
}

// ---------------------------------------------------------------------------
// Path helpers
// ---------------------------------------------------------------------------

// fullPath maps a caller-supplied relative path to an absolute local path.
func (d *Driver) fullPath(relPath string) string {
	relPath = filepath.FromSlash(relPath)
	if d.cfg.RootPath == "" {
		return filepath.Clean(relPath)
	}
	return filepath.Join(d.cfg.RootPath, relPath)
}

// validFullPath returns the full path after a traversal check. If the resolved
// path escapes the configured root, folder.ErrInvalidPath is returned.
func (d *Driver) validFullPath(relPath string) (string, error) {
	fp := d.fullPath(relPath)

	// When no root is configured every path is allowed.
	if d.cfg.RootPath == "" {
		return fp, nil
	}

	// Resolve any symlinks to get the real path for comparison.
	resolved, err := filepath.EvalSymlinks(fp)
	if err != nil {
		// If the target doesn't exist yet, verify the parent.
		if os.IsNotExist(err) {
			parentResolved, pErr := filepath.EvalSymlinks(filepath.Dir(fp))
			if pErr != nil {
				// Parent doesn't exist either — still check the raw path.
				if !isSubPath(d.cfg.RootPath, fp) {
					return "", fmt.Errorf("local: %q: %w", relPath, folder.ErrInvalidPath)
				}
				return fp, nil
			}
			child := filepath.Join(parentResolved, filepath.Base(fp))
			if !isSubPath(d.cfg.RootPath, child) {
				return "", fmt.Errorf("local: %q: %w", relPath, folder.ErrInvalidPath)
			}
			return fp, nil
		}
		return "", fmt.Errorf("local: resolve %q: %w", relPath, err)
	}

	if !isSubPath(d.cfg.RootPath, resolved) {
		return "", fmt.Errorf("local: %q: %w", relPath, folder.ErrInvalidPath)
	}
	return fp, nil
}

// isSubPath checks whether child is equal to or a sub-path of root.
func isSubPath(root, child string) bool {
	root = filepath.Clean(root)
	child = filepath.Clean(child)

	rel, err := filepath.Rel(root, child)
	if err != nil {
		return false
	}

	rel = filepath.Clean(rel)
	if rel == "." {
		return true
	}
	if rel == ".." {
		return false
	}

	return !strings.HasPrefix(rel, ".."+string(filepath.Separator))
}

// toRelPath converts an absolute path back to a path relative to root.
func (d *Driver) toRelPath(absPath string) string {
	if d.cfg.RootPath == "" {
		return filepath.ToSlash(absPath)
	}
	rel, err := filepath.Rel(d.cfg.RootPath, absPath)
	if err != nil {
		return filepath.ToSlash(absPath)
	}
	return filepath.ToSlash(rel)
}

// ---------------------------------------------------------------------------
// folder.Manager
// ---------------------------------------------------------------------------

func (d *Driver) Capabilities() folder.Capabilities {
	caps := folder.BaseCapabilities()
	caps.CanRead = true
	caps.CanWrite = true
	caps.AtomicMove = true // os.Rename is atomic on the same filesystem
	return caps
}

func (d *Driver) Exist(ctx context.Context, filePath string) (bool, error) {
	return folder.ExistViaStat(d, ctx, filePath)
}

func (d *Driver) List(_ context.Context, dir string, opt *folder.ListOptions) ([]*folder.FileInfo, error) {
	if opt == nil {
		opt = &folder.ListOptions{}
	}

	fullDir, err := d.validFullPath(dir)
	if err != nil {
		return nil, err
	}

	if opt.Recursive {
		return d.listRecursive(fullDir, dir, opt, 0)
	}

	entries, err := os.ReadDir(fullDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("local: list %q: %w", dir, folder.ErrNotFound)
		}
		return nil, fmt.Errorf("local: list %q: %w", dir, err)
	}

	result := []*folder.FileInfo{}
	for _, entry := range entries {
		if opt.Prefix != "" && !strings.HasPrefix(entry.Name(), opt.Prefix) {
			continue
		}
		info, infoErr := entry.Info()
		if infoErr != nil {
			continue // skip entries we can't stat
		}
		fi := d.toFileInfo(info, dir)
		result = append(result, fi)
		if opt.Limit > 0 && len(result) >= opt.Limit {
			break
		}
	}
	return result, nil
}

func (d *Driver) Stat(_ context.Context, filePath string) (*folder.FileInfo, error) {
	full, err := d.validFullPath(filePath)
	if err != nil {
		return nil, err
	}

	linfo, err := os.Lstat(full)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("local: stat %q: %w", filePath, folder.ErrNotFound)
		}
		return nil, fmt.Errorf("local: stat %q: %w", filePath, err)
	}

	entryType := folder.EntryTypeFile
	info := linfo
	if linfo.Mode()&os.ModeSymlink != 0 {
		entryType = folder.EntryTypeSymlink
		if resolved, sErr := os.Stat(full); sErr == nil {
			info = resolved
		}
	} else if linfo.IsDir() {
		entryType = folder.EntryTypeDirectory
	}

	modTime := info.ModTime()
	return &folder.FileInfo{
		Name:         filepath.Base(full),
		Path:         filePath,
		Type:         entryType,
		Size:         info.Size(),
		LastModified: &modTime,
	}, nil
}

func (d *Driver) Rename(_ context.Context, filePath string, newName string) error {
	dir := filepath.ToSlash(filepath.Dir(filePath))
	newRelPath := dir + "/" + newName

	srcFull, err := d.validFullPath(filePath)
	if err != nil {
		return err
	}
	dstFull, err := d.validFullPath(newRelPath)
	if err != nil {
		return err
	}

	if err := os.Rename(srcFull, dstFull); err != nil {
		return fmt.Errorf("local: rename %q -> %q: %w", filePath, newName, err)
	}
	return nil
}

func (d *Driver) Delete(_ context.Context, filePath string) error {
	full, err := d.validFullPath(filePath)
	if err != nil {
		return err
	}

	info, err := os.Lstat(full)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // already gone
		}
		return fmt.Errorf("local: delete %q: %w", filePath, err)
	}

	if info.IsDir() {
		return d.removeAll(full, 0)
	}

	if err := os.Remove(full); err != nil {
		return fmt.Errorf("local: delete %q: %w", filePath, err)
	}
	return nil
}

func (d *Driver) Copy(_ context.Context, op folder.PathOp) error {
	srcFull, err := d.validFullPath(op.SrcPath)
	if err != nil {
		return err
	}
	dstFull, err := d.validFullPath(op.DstPath)
	if err != nil {
		return err
	}

	info, err := os.Stat(srcFull)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("local: copy %q -> %q: %w", op.SrcPath, op.DstPath, folder.ErrNotFound)
		}
		return fmt.Errorf("local: copy %q -> %q: %w", op.SrcPath, op.DstPath, err)
	}

	if info.IsDir() {
		return d.copyDir(srcFull, dstFull, 0)
	}
	return d.copyFile(srcFull, dstFull)
}

func (d *Driver) Move(_ context.Context, op folder.PathOp) error {
	srcFull, err := d.validFullPath(op.SrcPath)
	if err != nil {
		return err
	}
	dstFull, err := d.validFullPath(op.DstPath)
	if err != nil {
		return err
	}

	if err := os.Rename(srcFull, dstFull); err != nil {
		return fmt.Errorf("local: move %q -> %q: %w", op.SrcPath, op.DstPath, err)
	}
	return nil
}

func (d *Driver) Mkdir(_ context.Context, dir string) error {
	full, err := d.validFullPath(dir)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(full, 0o755); err != nil {
		return fmt.Errorf("local: mkdir %q: %w", dir, err)
	}
	return nil
}

func (d *Driver) SetDirectoryModTime(_ context.Context, dir string, modTime time.Time) error {
	full, err := d.validFullPath(dir)
	if err != nil {
		return err
	}
	if err := os.Chtimes(full, modTime, modTime); err != nil {
		return fmt.Errorf("local: set directory mod time %q: %w", dir, err)
	}
	return nil
}

// ---------------------------------------------------------------------------
// folder.Reader
// ---------------------------------------------------------------------------

func (d *Driver) Read(_ context.Context, filePath string) (io.ReadCloser, error) {
	full, err := d.validFullPath(filePath)
	if err != nil {
		return nil, err
	}

	f, err := os.Open(full)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("local: read %q: %w", filePath, folder.ErrNotFound)
		}
		return nil, fmt.Errorf("local: read %q: %w", filePath, err)
	}
	return f, nil
}

// ---------------------------------------------------------------------------
// folder.Writer
// ---------------------------------------------------------------------------

func (d *Driver) Write(_ context.Context, filePath string, body io.Reader, opt *folder.WriteOptions) (*folder.FileInfo, error) {
	full, err := d.validFullPath(filePath)
	if err != nil {
		return nil, err
	}

	// Ensure parent directory exists.
	parentDir := filepath.Dir(full)
	if err := os.MkdirAll(parentDir, 0o755); err != nil {
		return nil, fmt.Errorf("local: write %q: mkdir parent: %w", filePath, err)
	}

	f, err := os.Create(full)
	if err != nil {
		return nil, fmt.Errorf("local: write %q: %w", filePath, err)
	}

	buf := make([]byte, copyBufSize)
	n, err := io.CopyBuffer(f, body, buf)
	if err != nil {
		_ = f.Close()
		return nil, fmt.Errorf("local: write %q: %w", filePath, err)
	}

	if err := f.Close(); err != nil {
		return nil, fmt.Errorf("local: write %q: close: %w", filePath, err)
	}

	modTime := time.Now()
	if opt != nil && opt.ModTime != nil {
		modTime = *opt.ModTime
		if err := os.Chtimes(full, modTime, modTime); err != nil {
			return nil, fmt.Errorf("local: write %q: set mod time: %w", filePath, err)
		}
	}
	return &folder.FileInfo{
		Name:         filepath.Base(full),
		Path:         filePath,
		Type:         folder.EntryTypeFile,
		Size:         n,
		LastModified: &modTime,
	}, nil
}

// ---------------------------------------------------------------------------
// folder.HealthChecker
// ---------------------------------------------------------------------------

func (d *Driver) Ping(_ context.Context) error {
	root := d.cfg.RootPath
	if root == "" {
		root = "."
	}
	_, err := os.Stat(root)
	if err != nil {
		return fmt.Errorf("local: ping: %w", err)
	}
	return nil
}

// ---------------------------------------------------------------------------
// Internal helpers
// ---------------------------------------------------------------------------

// listRecursive walks the directory tree and collects all entries.
func (d *Driver) listRecursive(fullDir, relDir string, opt *folder.ListOptions, depth int) ([]*folder.FileInfo, error) {
	if depth >= maxRecursionDepth {
		return nil, fmt.Errorf("local: list %q: recursion depth limit (%d) exceeded", relDir, maxRecursionDepth)
	}

	entries, err := os.ReadDir(fullDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("local: list %q: %w", relDir, folder.ErrNotFound)
		}
		return nil, fmt.Errorf("local: list %q: %w", relDir, err)
	}

	result := []*folder.FileInfo{}
	for _, entry := range entries {
		if depth == 0 && opt.Prefix != "" && !strings.HasPrefix(entry.Name(), opt.Prefix) {
			continue
		}

		info, infoErr := entry.Info()
		if infoErr != nil {
			continue
		}
		fi := d.toFileInfo(info, relDir)
		result = append(result, fi)

		if opt.Limit > 0 && len(result) >= opt.Limit {
			return result[:opt.Limit], nil
		}

		if entry.IsDir() {
			childFull := filepath.Join(fullDir, entry.Name())
			childRel := strings.TrimSuffix(fi.Path, "/")
			sub, subErr := d.listRecursive(childFull, childRel, opt, depth+1)
			if subErr != nil {
				return nil, subErr
			}
			result = append(result, sub...)
			if opt.Limit > 0 && len(result) >= opt.Limit {
				return result[:opt.Limit], nil
			}
		}
	}

	return result, nil
}

// toFileInfo converts an os.FileInfo into a folder.FileInfo.
func (d *Driver) toFileInfo(info os.FileInfo, parentDir string) *folder.FileInfo {
	entryType := folder.EntryTypeFile
	relPath := parentDir + "/" + info.Name()
	if parentDir == "" || parentDir == "." {
		relPath = info.Name()
	}

	if info.IsDir() {
		entryType = folder.EntryTypeDirectory
		relPath += "/"
	} else if info.Mode()&os.ModeSymlink != 0 {
		entryType = folder.EntryTypeSymlink
	}

	modTime := info.ModTime()
	return &folder.FileInfo{
		Name:         info.Name(),
		Path:         relPath,
		Type:         entryType,
		Size:         info.Size(),
		LastModified: &modTime,
	}
}

// removeAll recursively removes a directory and all its contents.
func (d *Driver) removeAll(dirPath string, depth int) error {
	if depth >= maxRecursionDepth {
		return fmt.Errorf("local: removeAll %q: recursion depth limit (%d) exceeded", dirPath, maxRecursionDepth)
	}

	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return fmt.Errorf("local: readdir %q: %w", dirPath, err)
	}

	for _, entry := range entries {
		child := filepath.Join(dirPath, entry.Name())
		if entry.IsDir() {
			if err := d.removeAll(child, depth+1); err != nil {
				return err
			}
		} else {
			if err := os.Remove(child); err != nil {
				return fmt.Errorf("local: remove %q: %w", child, err)
			}
		}
	}

	if err := os.Remove(dirPath); err != nil {
		return fmt.Errorf("local: rmdir %q: %w", dirPath, err)
	}
	return nil
}

// copyFile copies a single file, preserving permissions.
func (d *Driver) copyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("local: open %q: %w", src, err)
	}
	defer srcFile.Close()

	srcInfo, err := srcFile.Stat()
	if err != nil {
		return fmt.Errorf("local: stat %q: %w", src, err)
	}

	// Ensure parent directory exists.
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return fmt.Errorf("local: mkdir parent %q: %w", dst, err)
	}

	dstFile, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, srcInfo.Mode())
	if err != nil {
		return fmt.Errorf("local: create %q: %w", dst, err)
	}

	buf := make([]byte, copyBufSize)
	if _, err := io.CopyBuffer(dstFile, srcFile, buf); err != nil {
		_ = dstFile.Close()
		return fmt.Errorf("local: copy data %q -> %q: %w", src, dst, err)
	}

	if err := dstFile.Close(); err != nil {
		return fmt.Errorf("local: close dst %q: %w", dst, err)
	}
	return nil
}

// copyDir recursively copies a directory.
func (d *Driver) copyDir(src, dst string, depth int) error {
	if depth >= maxRecursionDepth {
		return fmt.Errorf("local: copyDir %q: recursion depth limit (%d) exceeded", src, maxRecursionDepth)
	}

	srcInfo, err := os.Stat(src)
	if err != nil {
		return fmt.Errorf("local: stat %q: %w", src, err)
	}

	if err := os.MkdirAll(dst, srcInfo.Mode()); err != nil {
		return fmt.Errorf("local: mkdir %q: %w", dst, err)
	}

	entries, err := os.ReadDir(src)
	if err != nil {
		return fmt.Errorf("local: readdir %q: %w", src, err)
	}

	for _, entry := range entries {
		srcChild := filepath.Join(src, entry.Name())
		dstChild := filepath.Join(dst, entry.Name())
		if entry.IsDir() {
			if err := d.copyDir(srcChild, dstChild, depth+1); err != nil {
				return err
			}
		} else {
			if err := d.copyFile(srcChild, dstChild); err != nil {
				return err
			}
		}
	}
	return nil
}
