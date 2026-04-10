package webdav

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	pathpkg "path"
	"strings"
	"time"

	gowebdav "github.com/wxk6b1203/gowebdav"

	"github.com/wxk6b1203/file-util-manager/folder"
)

const maxRecursionDepth = 128

type webdavFile interface {
	os.FileInfo
	Path() string
	ContentType() string
	ETag() string
}

func init() {
	folder.RegisterDriver[Options]("WebDAV",
		"WebDAV — manage files on WebDAV-compatible servers over HTTP(S), including common NAS and cloud endpoints.",
		New,
	)
}

type Driver struct {
	folder.BaseDriver
	cfg    *Options
	client *gowebdav.Client
}

func New(_ context.Context, opt *folder.DriverOptions, cfg *Options) (folder.Manager, error) {
	if cfg == nil {
		return nil, fmt.Errorf("webdav: missing configuration")
	}

	if opt != nil && opt.Root != "" {
		root := normalizeRootPath(opt.Root)
		if cfg.RootPath != "" {
			cfg.RootPath = normalizeRootPath(pathpkg.Join(root, cfg.RootPath))
		} else {
			cfg.RootPath = root
		}
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("webdav: invalid configuration: %w", err)
	}

	client := newClient(cfg)
	driver := &Driver{
		BaseDriver: folder.NewBaseDriver(opt),
		cfg:        cfg,
		client:     client,
	}

	if err := driver.Ping(context.Background()); err != nil {
		return nil, fmt.Errorf("webdav: connect: %w", err)
	}

	return driver, nil
}

func newClient(cfg *Options) *gowebdav.Client {
	if strings.TrimSpace(cfg.BearerToken) != "" {
		return configureClient(gowebdav.NewBearerAuthClient(cfg.Endpoint, cfg.BearerToken), cfg)
	}
	return configureClient(gowebdav.NewClient(cfg.Endpoint, cfg.Username, cfg.Password), cfg)
}

func configureClient(client *gowebdav.Client, cfg *Options) *gowebdav.Client {
	if cfg == nil {
		return client
	}

	client.SetTimeout(timeDurationSec(cfg.TimeoutSec))

	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.Proxy = http.ProxyFromEnvironment
	if cfg.InsecureSkipVerify {
		transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true} //nolint:gosec
	}
	client.SetTransport(transport)
	return client
}

func timeDurationSec(timeoutSec int) (timeout time.Duration) {
	if timeoutSec <= 0 {
		timeoutSec = 30
	}
	return time.Duration(timeoutSec) * time.Second
}

func (d *Driver) Capabilities() folder.Capabilities {
	caps := folder.BaseCapabilities()
	caps.CanRead = true
	caps.CanWrite = true
	return caps
}

func (d *Driver) Ping(_ context.Context) error {
	return normalizeError("ping", "", d.client.Connect())
}

func (d *Driver) Exist(ctx context.Context, filePath string) (bool, error) {
	return folder.ExistViaStat(d, ctx, filePath)
}

func (d *Driver) List(_ context.Context, dir string, opt *folder.ListOptions) ([]*folder.FileInfo, error) {
	if opt == nil {
		opt = &folder.ListOptions{}
	}

	cleanDir, err := cleanRelativePath(dir)
	if err != nil {
		return nil, err
	}
	fullDir := d.fullPath(cleanDir)

	if opt.Recursive {
		return d.listRecursive(fullDir, cleanDir, opt, 0)
	}

	items, err := d.client.ReadDir(fullDir)
	if err != nil {
		return nil, normalizeError("list", cleanDir, err)
	}

	result := make([]*folder.FileInfo, 0, len(items))
	for _, item := range items {
		if opt.Prefix != "" && !strings.HasPrefix(item.Name(), opt.Prefix) {
			continue
		}
		result = append(result, d.toFileInfo(item, cleanDir))
		if opt.Limit > 0 && len(result) >= opt.Limit {
			break
		}
	}
	return result, nil
}

func (d *Driver) listRecursive(fullDir, relDir string, opt *folder.ListOptions, depth int) ([]*folder.FileInfo, error) {
	if depth > maxRecursionDepth {
		return nil, fmt.Errorf("webdav: list %q: max recursion depth exceeded", relDir)
	}

	items, err := d.client.ReadDir(fullDir)
	if err != nil {
		return nil, normalizeError("list", relDir, err)
	}

	result := []*folder.FileInfo{}
	for _, item := range items {
		if opt.Prefix != "" && !strings.HasPrefix(item.Name(), opt.Prefix) {
			continue
		}

		fi := d.toFileInfo(item, relDir)
		result = append(result, fi)
		if opt.Limit > 0 && len(result) >= opt.Limit {
			return result, nil
		}

		if !fi.IsDir() {
			continue
		}

		sub, err := d.listRecursive(d.fullPath(fi.Path), fi.Path, opt, depth+1)
		if err != nil {
			return nil, err
		}
		result = append(result, sub...)
		if opt.Limit > 0 && len(result) >= opt.Limit {
			return result[:opt.Limit], nil
		}
	}

	return result, nil
}

func (d *Driver) Stat(_ context.Context, filePath string) (*folder.FileInfo, error) {
	cleanPath, err := cleanRelativePath(filePath)
	if err != nil {
		return nil, err
	}

	info, err := d.client.Stat(d.fullPath(cleanPath))
	if err != nil {
		return nil, normalizeError("stat", cleanPath, err)
	}

	fi := d.toFileInfo(info, parentPath(cleanPath))
	fi.Path = cleanPath
	if info.IsDir() && fi.Path == "" {
		fi.Name = baseName(d.cfg.RootPath)
	}
	return fi, nil
}

func (d *Driver) Rename(_ context.Context, filePath string, newName string) error {
	cleanPath, err := cleanRelativePath(filePath)
	if err != nil {
		return err
	}
	trimmedName := strings.TrimSpace(newName)
	if trimmedName == "" {
		return fmt.Errorf("webdav: rename %q: %w", cleanPath, folder.ErrInvalidPath)
	}
	if strings.Contains(trimmedName, "/") || strings.Contains(trimmedName, "\\") {
		return fmt.Errorf("webdav: rename %q: %w", cleanPath, folder.ErrInvalidPath)
	}

	targetPath := joinRelativePath(parentPath(cleanPath), trimmedName)
	return normalizeError("rename", cleanPath, d.client.Rename(d.fullPath(cleanPath), d.fullPath(targetPath), false))
}

func (d *Driver) Delete(_ context.Context, filePath string) error {
	cleanPath, err := cleanRelativePath(filePath)
	if err != nil {
		return err
	}
	if cleanPath == "" {
		return fmt.Errorf("webdav: delete root: %w", folder.ErrInvalidPath)
	}
	return normalizeError("delete", cleanPath, d.client.RemoveAll(d.fullPath(cleanPath)))
}

func (d *Driver) Copy(_ context.Context, op folder.PathOp) error {
	srcPath, err := cleanRelativePath(op.SrcPath)
	if err != nil {
		return err
	}
	dstPath, err := cleanRelativePath(op.DstPath)
	if err != nil {
		return err
	}
	if srcPath == "" || dstPath == "" {
		return fmt.Errorf("webdav: copy %q -> %q: %w", srcPath, dstPath, folder.ErrInvalidPath)
	}
	return normalizeError("copy", srcPath, d.client.Copy(d.fullPath(srcPath), d.fullPath(dstPath), false))
}

func (d *Driver) Move(_ context.Context, op folder.PathOp) error {
	srcPath, err := cleanRelativePath(op.SrcPath)
	if err != nil {
		return err
	}
	dstPath, err := cleanRelativePath(op.DstPath)
	if err != nil {
		return err
	}
	if srcPath == "" || dstPath == "" {
		return fmt.Errorf("webdav: move %q -> %q: %w", srcPath, dstPath, folder.ErrInvalidPath)
	}
	return normalizeError("move", srcPath, d.client.Rename(d.fullPath(srcPath), d.fullPath(dstPath), false))
}

func (d *Driver) Mkdir(_ context.Context, filePath string) error {
	cleanPath, err := cleanRelativePath(filePath)
	if err != nil {
		return err
	}
	if cleanPath == "" {
		return nil
	}
	return normalizeError("mkdir", cleanPath, d.client.MkdirAll(d.fullPath(cleanPath), 0o755))
}

func (d *Driver) Read(_ context.Context, filePath string) (io.ReadCloser, error) {
	cleanPath, err := cleanRelativePath(filePath)
	if err != nil {
		return nil, err
	}
	reader, err := d.client.ReadStream(d.fullPath(cleanPath))
	if err != nil {
		return nil, normalizeError("read", cleanPath, err)
	}
	return reader, nil
}

func (d *Driver) Write(_ context.Context, filePath string, body io.Reader, _ *folder.WriteOptions) (*folder.FileInfo, error) {
	cleanPath, err := cleanRelativePath(filePath)
	if err != nil {
		return nil, err
	}
	if cleanPath == "" {
		return nil, fmt.Errorf("webdav: write root: %w", folder.ErrInvalidPath)
	}

	if err := d.client.WriteStream(d.fullPath(cleanPath), body, 0o644); err != nil {
		return nil, normalizeError("write", cleanPath, err)
	}
	return d.Stat(context.Background(), cleanPath)
}

func (d *Driver) fullPath(relPath string) string {
	relPath = normalizeRootPath(relPath)
	if d.cfg.RootPath == "" {
		return relPath
	}
	if relPath == "" {
		return d.cfg.RootPath
	}
	return pathpkg.Join(d.cfg.RootPath, relPath)
}

func (d *Driver) toRelativePath(fullPath string) string {
	cleaned := normalizeRootPath(fullPath)
	if d.cfg.RootPath == "" {
		return cleaned
	}
	root := d.cfg.RootPath
	if cleaned == root {
		return ""
	}
	if strings.HasPrefix(cleaned, root+"/") {
		return strings.TrimPrefix(cleaned, root+"/")
	}
	return cleaned
}

func (d *Driver) toFileInfo(info os.FileInfo, parent string) *folder.FileInfo {
	entryType := folder.EntryTypeFile
	if info.IsDir() {
		entryType = folder.EntryTypeDirectory
	}

	entryPath := joinRelativePath(parent, info.Name())
	contentType := ""
	etag := ""
	if wf, ok := info.(webdavFile); ok {
		entryPath = d.toRelativePath(wf.Path())
		contentType = wf.ContentType()
		etag = wf.ETag()
	}

	modTime := info.ModTime()
	return &folder.FileInfo{
		Name:         info.Name(),
		Path:         entryPath,
		Type:         entryType,
		Size:         info.Size(),
		LastModified: &modTime,
		ContentType:  contentType,
		ETag:         etag,
	}
}

func cleanRelativePath(value string) (string, error) {
	trimmed := strings.TrimSpace(strings.ReplaceAll(value, "\\", "/"))
	if trimmed == "" || trimmed == "." || trimmed == "/" {
		return "", nil
	}

	cleaned := pathpkg.Clean(trimmed)
	if cleaned == "." || cleaned == "/" {
		return "", nil
	}
	cleaned = strings.TrimPrefix(cleaned, "/")
	if cleaned == ".." || strings.HasPrefix(cleaned, "../") {
		return "", fmt.Errorf("webdav: %q: %w", value, folder.ErrInvalidPath)
	}
	return cleaned, nil
}

func joinRelativePath(parent, name string) string {
	parent = normalizeRootPath(parent)
	child := normalizeRootPath(name)
	if parent == "" {
		return child
	}
	if child == "" {
		return parent
	}
	return pathpkg.Join(parent, child)
}

func parentPath(value string) string {
	cleaned := normalizeRootPath(value)
	if cleaned == "" {
		return ""
	}
	parent := pathpkg.Dir(cleaned)
	if parent == "." || parent == "/" {
		return ""
	}
	return normalizeRootPath(parent)
}

func baseName(value string) string {
	cleaned := normalizeRootPath(value)
	if cleaned == "" {
		return ""
	}
	return pathpkg.Base(cleaned)
}

func normalizeError(op, targetPath string, err error) error {
	if err == nil {
		return nil
	}

	switch {
	case gowebdav.IsErrCode(err, http.StatusNotFound):
		return fmt.Errorf("webdav: %s %q: %w", op, targetPath, folder.ErrNotFound)
	case gowebdav.IsErrCode(err, http.StatusPreconditionFailed):
		return fmt.Errorf("webdav: %s %q: %w", op, targetPath, folder.ErrAlreadyExist)
	case gowebdav.IsErrCode(err, http.StatusConflict):
		return fmt.Errorf("webdav: %s %q: %w", op, targetPath, folder.ErrInvalidPath)
	}

	var pathErr *os.PathError
	if errors.As(err, &pathErr) {
		return fmt.Errorf("webdav: %s %q: %w", op, targetPath, pathErr)
	}
	return fmt.Errorf("webdav: %s %q: %w", op, targetPath, err)
}
