package sftp

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"

	"github.com/wxk6b1203/file-util-manager/folder"
)

const (
	// defaultDialTimeout prevents Dial from hanging indefinitely on unreachable hosts.
	defaultDialTimeout = 30 * time.Second

	// keepAliveInterval is how often the SSH keepalive probe is sent.
	keepAliveInterval = 30 * time.Second

	// copyBufSize is the buffer used by io.CopyBuffer when transferring file
	// data over the SFTP connection (256 KiB).
	copyBufSize = 256 << 10

	// maxRecursionDepth guards removeAll / copyDir / listRecursive against
	// pathological directory nesting or symlink loops.
	maxRecursionDepth = 128
)

func init() {
	folder.RegisterDriver[Options]("SFTP",
		"SSH File Transfer Protocol (SFTP) — a secure file transfer protocol that runs over an SSH connection, "+
			"providing encrypted access to remote file systems.",
		New,
	)
}

// Driver implements folder.Manager, folder.Reader, folder.Writer,
// folder.HealthChecker and folder.Closer for SFTP backends.
type Driver struct {
	folder.BaseDriver
	cfg *Options

	mu          sync.RWMutex // guards sshC, client
	reconnectMu sync.Mutex   // serializes reconnection attempts — separate from mu to avoid blocking readers during dial
	sshC        *ssh.Client
	client      *sftp.Client
	closed      atomic.Bool // true after explicit Close() — prevents reconnection
}

// ---------------------------------------------------------------------------
// Constructor
// ---------------------------------------------------------------------------

// New creates a new SFTP driver. Address/Username and at least one auth
// method (Password or PrivateKey) are mandatory.
func New(_ context.Context, opt *folder.DriverOptions, cfg *Options) (folder.Manager, error) {
	if cfg == nil {
		return nil, fmt.Errorf("sftp: missing configuration")
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("sftp: invalid configuration: %w", err)
	}

	// Merge DriverOptions.Root into rootPath when specified.
	if opt != nil && opt.Root != "" {
		root := strings.TrimRight(opt.Root, "/")
		if cfg.RootPath != "" {
			cfg.RootPath = root + "/" + strings.TrimLeft(cfg.RootPath, "/")
		} else {
			cfg.RootPath = root
		}
	}

	d := &Driver{
		BaseDriver: folder.NewBaseDriver(opt),
		cfg:        cfg,
	}

	if err := d.doConnect(); err != nil {
		return nil, fmt.Errorf("sftp: connect: %w", err)
	}

	return d, nil
}

// ---------------------------------------------------------------------------
// Connection lifecycle
// ---------------------------------------------------------------------------

// dial performs the SSH + SFTP handshake without holding any lock, so it
// will not block concurrent read operations.
func (d *Driver) dial() (*ssh.Client, *sftp.Client, error) {
	authMethods := make([]ssh.AuthMethod, 0, 2)

	if d.cfg.PrivateKey != "" || d.cfg.PrivateKeyPath != "" {
		key, err := loadPrivateKeyMaterial(d.cfg)
		if err != nil {
			return nil, nil, err
		}
		var signer ssh.Signer
		if d.cfg.Passphrase != "" {
			signer, err = ssh.ParsePrivateKeyWithPassphrase(key, []byte(d.cfg.Passphrase))
		} else {
			signer, err = ssh.ParsePrivateKey(key)
		}
		if err != nil {
			return nil, nil, fmt.Errorf("parse private key: %w", err)
		}
		authMethods = append(authMethods, ssh.PublicKeys(signer))
	}

	if d.cfg.Password != "" {
		authMethods = append(authMethods, ssh.Password(d.cfg.Password))
	}

	timeout := defaultDialTimeout
	if d.cfg.DialTimeoutSec > 0 {
		timeout = time.Duration(d.cfg.DialTimeoutSec) * time.Second
	}

	sshConfig := &ssh.ClientConfig{
		User:            d.cfg.Username,
		Auth:            authMethods,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), //nolint:gosec // TODO: support known_hosts
		Timeout:         timeout,
	}

	addr := fmt.Sprintf("%s:%d", d.cfg.Address, d.cfg.Port)
	sshConn, err := ssh.Dial("tcp", addr, sshConfig)
	if err != nil {
		return nil, nil, fmt.Errorf("ssh dial %s: %w", addr, err)
	}

	sftpClient, err := sftp.NewClient(sshConn)
	if err != nil {
		_ = sshConn.Close()
		return nil, nil, fmt.Errorf("sftp new client: %w", err)
	}

	return sshConn, sftpClient, nil
}

func loadPrivateKeyMaterial(cfg *Options) ([]byte, error) {
	if cfg == nil {
		return nil, fmt.Errorf("sftp: missing configuration")
	}

	privateKey := strings.TrimSpace(cfg.PrivateKey)
	if privateKey != "" {
		if looksLikePrivateKey(privateKey) {
			return []byte(privateKey), nil
		}
		if strings.TrimSpace(cfg.PrivateKeyPath) == "" {
			return readPrivateKeyFile(privateKey)
		}
	}

	privateKeyPath := strings.TrimSpace(cfg.PrivateKeyPath)
	if privateKeyPath == "" {
		return nil, fmt.Errorf("sftp: private key is empty")
	}
	return readPrivateKeyFile(privateKeyPath)
}

func looksLikePrivateKey(value string) bool {
	return strings.Contains(value, "-----BEGIN ") && strings.Contains(value, "PRIVATE KEY-----")
}

func readPrivateKeyFile(filePath string) ([]byte, error) {
	resolvedPath, err := expandUserPath(filePath)
	if err != nil {
		return nil, fmt.Errorf("resolve private key path %q: %w", filePath, err)
	}

	info, err := os.Stat(resolvedPath)
	if err != nil {
		return nil, fmt.Errorf("read private key path %q: %w", filePath, err)
	}
	if info.IsDir() {
		return nil, fmt.Errorf("read private key path %q: is a directory", filePath)
	}

	body, err := os.ReadFile(resolvedPath)
	if err != nil {
		return nil, fmt.Errorf("read private key path %q: %w", filePath, err)
	}
	if strings.TrimSpace(string(body)) == "" {
		return nil, fmt.Errorf("read private key path %q: file is empty", filePath)
	}
	return body, nil
}

func expandUserPath(filePath string) (string, error) {
	cleaned := strings.TrimSpace(filePath)
	if cleaned == "" {
		return "", fmt.Errorf("path is empty")
	}
	if cleaned == "~" {
		return os.UserHomeDir()
	}
	if strings.HasPrefix(cleaned, "~/") || strings.HasPrefix(cleaned, `~\`) {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		return filepath.Join(home, strings.TrimLeft(cleaned[1:], `/\`)), nil
	}
	return filepath.Clean(cleaned), nil
}

// doConnect dials a new connection and installs it, cleaning up any stale one.
// Safe to call without holding any lock (acquires them internally).
func (d *Driver) doConnect() error {
	if d.closed.Load() {
		return fmt.Errorf("sftp: driver is closed")
	}

	sshConn, sftpClient, err := d.dial()
	if err != nil {
		return err
	}

	d.mu.Lock()
	// Re-check after acquiring the lock — Close() may have raced.
	if d.closed.Load() {
		d.mu.Unlock()
		_ = sftpClient.Close()
		_ = sshConn.Close()
		return fmt.Errorf("sftp: driver is closed")
	}
	d.closeLocked() // clean up any stale connection
	d.sshC = sshConn
	d.client = sftpClient
	d.mu.Unlock()

	go d.keepAlive(sshConn)
	return nil
}

// closeLocked releases SSH/SFTP resources. Caller MUST hold d.mu write lock.
func (d *Driver) closeLocked() {
	if d.client != nil {
		_ = d.client.Close()
		d.client = nil
	}
	if d.sshC != nil {
		_ = d.sshC.Close()
		d.sshC = nil
	}
}

// getClient returns a live SFTP client, transparently reconnecting if the
// connection was lost. Returns an error only if the driver is explicitly
// closed or the reconnection dial fails.
func (d *Driver) getClient() (*sftp.Client, error) {
	// Fast path — read lock only.
	d.mu.RLock()
	c := d.client
	d.mu.RUnlock()

	if c != nil {
		return c, nil
	}
	if d.closed.Load() {
		return nil, fmt.Errorf("sftp: driver is closed")
	}

	// Slow path — reconnect. reconnectMu ensures only one goroutine dials;
	// others block here and re-use the result.
	d.reconnectMu.Lock()
	defer d.reconnectMu.Unlock()

	// Double-check: another goroutine may have reconnected while we waited.
	d.mu.RLock()
	c = d.client
	d.mu.RUnlock()
	if c != nil {
		return c, nil
	}
	if d.closed.Load() {
		return nil, fmt.Errorf("sftp: driver is closed")
	}

	if err := d.doConnect(); err != nil {
		return nil, fmt.Errorf("sftp: reconnect: %w", err)
	}

	d.mu.RLock()
	c = d.client
	d.mu.RUnlock()
	if c == nil {
		// Close() raced with doConnect and won — driver is shut down.
		return nil, fmt.Errorf("sftp: driver is closed")
	}
	return c, nil
}

// invalidateClient marks the current connection as dead so the next
// getClient call triggers a reconnect. No-op if the active client has
// already been replaced (i.e. another goroutine reconnected).
func (d *Driver) invalidateClient(old *sftp.Client) {
	if old == nil {
		return
	}
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.client != old {
		return // already reconnected
	}
	d.closeLocked()
}

// keepAlive sends periodic SSH keepalive probes. When the connection is
// dead it proactively invalidates the client so the next getClient call
// will reconnect immediately.
func (d *Driver) keepAlive(c *ssh.Client) {
	t := time.NewTicker(keepAliveInterval)
	defer t.Stop()

	for range t.C {
		// Bail out early if the driver was closed or the connection was
		// already replaced by a reconnect.
		d.mu.RLock()
		active := d.sshC == c
		d.mu.RUnlock()
		if d.closed.Load() || !active {
			return
		}

		_, _, err := c.SendRequest("keepalive@openssh.com", true, nil)
		if err != nil {
			d.mu.Lock()
			if d.sshC == c { // still our connection — invalidate
				d.closeLocked()
			}
			d.mu.Unlock()
			return
		}
	}
}

// ---------------------------------------------------------------------------
// Connection error detection & retry helpers
// ---------------------------------------------------------------------------

// isConnError classifies err as a transport / connection level failure that
// warrants a reconnection attempt, as opposed to an application-level error
// (e.g. "file not found").
func isConnError(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, io.EOF) ||
		errors.Is(err, io.ErrUnexpectedEOF) ||
		errors.Is(err, io.ErrClosedPipe) ||
		errors.Is(err, os.ErrClosed) ||
		errors.Is(err, net.ErrClosed) {
		return true
	}
	var netErr *net.OpError
	if errors.As(err, &netErr) {
		return true
	}
	// SFTP protocol-level connection errors.
	if errors.Is(err, sftp.ErrSshFxConnectionLost) ||
		errors.Is(err, sftp.ErrSshFxNoConnection) {
		return true
	}
	return false
}

// retryOnConn executes op with a live client. If op returns a connection
// error the client is invalidated, a reconnect is attempted, and op is
// retried exactly once. Suitable for idempotent / read-only operations.
func retryOnConn[T any](d *Driver, op func(*sftp.Client) (T, error)) (T, error) {
	client, err := d.getClient()
	if err != nil {
		var zero T
		return zero, err
	}

	result, err := op(client)
	if err == nil || !isConnError(err) {
		return result, err
	}

	// Connection error — invalidate, reconnect, retry once.
	d.invalidateClient(client)
	client, err = d.getClient()
	if err != nil {
		var zero T
		return zero, err
	}
	return op(client)
}

// retryVoid is the no-return-value variant of retryOnConn.
func retryVoid(d *Driver, op func(*sftp.Client) error) error {
	_, err := retryOnConn(d, func(c *sftp.Client) (struct{}, error) {
		return struct{}{}, op(c)
	})
	return err
}

// ---------------------------------------------------------------------------
// Path helpers
// ---------------------------------------------------------------------------

// fullPath maps a caller-supplied relative path to an absolute remote path
// by joining "/" + RootPath + relPath, with a traversal guard.
func (d *Driver) fullPath(relPath string) string {
	root := normalizeRemoteRoot(d.cfg.RootPath)
	p := path.Join(root, relPath)

	// When no root is configured every absolute path is valid.
	if root == "/" {
		return p
	}

	// Directory-boundary check: p must be the root itself or a child of it.
	// A naive HasPrefix("/dataXYZ", "/data") would pass — the extra "/"
	// ensures we compare on a directory boundary.
	if p != root && !strings.HasPrefix(p, root+"/") {
		return "" // sentinel — traversal detected
	}
	return p
}

func normalizeRemoteRoot(rootPath string) string {
	root := strings.TrimSpace(rootPath)
	if root == "" || root == "/" {
		return "/"
	}
	return path.Clean("/" + strings.TrimLeft(root, "/"))
}

// validFullPath is a convenience wrapper that returns folder.ErrInvalidPath
// when traversal is detected.
func (d *Driver) validFullPath(relPath string) (string, error) {
	fp := d.fullPath(relPath)
	if fp == "" {
		return "", fmt.Errorf("sftp: %q: %w", relPath, folder.ErrInvalidPath)
	}
	return fp, nil
}

// ---------------------------------------------------------------------------
// folder.Manager
// ---------------------------------------------------------------------------

func (d *Driver) Capabilities() folder.Capabilities {
	caps := folder.BaseCapabilities()
	caps.CanRead = true
	caps.CanWrite = true
	caps.AtomicMove = true
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

	return retryOnConn(d, func(client *sftp.Client) ([]*folder.FileInfo, error) {
		if opt.Recursive {
			return d.listRecursive(client, fullDir, dir, opt, 0)
		}

		entries, err := client.ReadDir(fullDir)
		if err != nil {
			if os.IsNotExist(err) {
				return nil, fmt.Errorf("sftp: list %q: %w", dir, folder.ErrNotFound)
			}
			return nil, fmt.Errorf("sftp: list %q: %w", dir, err)
		}

		result := []*folder.FileInfo{}
		for _, entry := range entries {
			if opt.Prefix != "" && !strings.HasPrefix(entry.Name(), opt.Prefix) {
				continue
			}
			fi := d.toFileInfo(entry, dir)
			result = append(result, fi)
			if opt.Limit > 0 && len(result) >= opt.Limit {
				break
			}
		}
		return result, nil
	})
}

func (d *Driver) Stat(_ context.Context, filePath string) (*folder.FileInfo, error) {
	full, err := d.validFullPath(filePath)
	if err != nil {
		return nil, err
	}

	return retryOnConn(d, func(client *sftp.Client) (*folder.FileInfo, error) {
		linfo, err := client.Lstat(full)
		if err != nil {
			if os.IsNotExist(err) {
				return nil, fmt.Errorf("sftp: stat %q: %w", filePath, folder.ErrNotFound)
			}
			return nil, fmt.Errorf("sftp: stat %q: %w", filePath, err)
		}

		entryType := folder.EntryTypeFile
		info := linfo
		if linfo.Mode()&os.ModeSymlink != 0 {
			entryType = folder.EntryTypeSymlink
			if resolved, err := client.Stat(full); err == nil {
				info = resolved
			}
		} else if linfo.IsDir() {
			entryType = folder.EntryTypeDirectory
		}

		modTime := info.ModTime()
		return &folder.FileInfo{
			Name:         path.Base(filePath),
			Path:         filePath,
			Type:         entryType,
			Size:         info.Size(),
			LastModified: &modTime,
		}, nil
	})
}

func (d *Driver) Rename(_ context.Context, filePath string, newName string) (retErr error) {
	client, err := d.getClient()
	if err != nil {
		return err
	}
	defer func() { d.invalidateOnConnError(client, retErr) }()

	dir := path.Dir(filePath)
	newPath := path.Join(dir, newName)

	srcFull, err := d.validFullPath(filePath)
	if err != nil {
		return err
	}
	dstFull, err := d.validFullPath(newPath)
	if err != nil {
		return err
	}

	if err := client.Rename(srcFull, dstFull); err != nil {
		return fmt.Errorf("sftp: rename %q -> %q: %w", filePath, newName, err)
	}
	return nil
}

func (d *Driver) Delete(_ context.Context, filePath string) (retErr error) {
	client, err := d.getClient()
	if err != nil {
		return err
	}
	defer func() { d.invalidateOnConnError(client, retErr) }()

	full, err := d.validFullPath(filePath)
	if err != nil {
		return err
	}

	// Optimistic: try Remove (file) first — saves one Stat round-trip.
	if rmErr := client.Remove(full); rmErr == nil {
		return nil
	} else if isConnError(rmErr) {
		return fmt.Errorf("sftp: delete %q: %w", filePath, rmErr)
	}

	info, err := client.Stat(full)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // already gone
		}
		return fmt.Errorf("sftp: delete %q: %w", filePath, err)
	}

	if info.IsDir() {
		return d.removeAll(client, full, 0)
	}

	if err := client.Remove(full); err != nil {
		return fmt.Errorf("sftp: delete %q: %w", filePath, err)
	}
	return nil
}

func (d *Driver) Copy(_ context.Context, op folder.PathOp) (retErr error) {
	client, err := d.getClient()
	if err != nil {
		return err
	}
	defer func() { d.invalidateOnConnError(client, retErr) }()

	srcFull, err := d.validFullPath(op.SrcPath)
	if err != nil {
		return err
	}
	dstFull, err := d.validFullPath(op.DstPath)
	if err != nil {
		return err
	}

	info, err := client.Stat(srcFull)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("sftp: copy %q -> %q: %w", op.SrcPath, op.DstPath, folder.ErrNotFound)
		}
		return fmt.Errorf("sftp: copy %q -> %q: %w", op.SrcPath, op.DstPath, err)
	}

	if info.IsDir() {
		return d.copyDir(client, srcFull, dstFull, 0)
	}
	return d.copyFile(client, srcFull, dstFull)
}

func (d *Driver) Move(_ context.Context, op folder.PathOp) (retErr error) {
	client, err := d.getClient()
	if err != nil {
		return err
	}
	defer func() { d.invalidateOnConnError(client, retErr) }()

	srcFull, err := d.validFullPath(op.SrcPath)
	if err != nil {
		return err
	}
	dstFull, err := d.validFullPath(op.DstPath)
	if err != nil {
		return err
	}

	if err := client.Rename(srcFull, dstFull); err != nil {
		return fmt.Errorf("sftp: move %q -> %q: %w", op.SrcPath, op.DstPath, err)
	}
	return nil
}

func (d *Driver) Mkdir(_ context.Context, dir string) error {
	full, err := d.validFullPath(dir)
	if err != nil {
		return err
	}
	// MkdirAll is idempotent — safe to retry.
	return retryVoid(d, func(client *sftp.Client) error {
		if err := client.MkdirAll(full); err != nil {
			return fmt.Errorf("sftp: mkdir %q: %w", dir, err)
		}
		return nil
	})
}

func (d *Driver) SetDirectoryModTime(_ context.Context, dir string, modTime time.Time) (retErr error) {
	client, err := d.getClient()
	if err != nil {
		return err
	}
	defer func() { d.invalidateOnConnError(client, retErr) }()

	full, err := d.validFullPath(dir)
	if err != nil {
		return err
	}
	if err := client.Chtimes(full, modTime, modTime); err != nil {
		return fmt.Errorf("sftp: set directory mod time %q: %w", dir, err)
	}
	return nil
}

// ---------------------------------------------------------------------------
// folder.Reader
// ---------------------------------------------------------------------------

func (d *Driver) Read(_ context.Context, filePath string) (_ io.ReadCloser, retErr error) {
	client, err := d.getClient()
	if err != nil {
		return nil, err
	}
	defer func() {
		if retErr != nil {
			d.invalidateOnConnError(client, retErr)
		}
	}()

	full, err := d.validFullPath(filePath)
	if err != nil {
		return nil, err
	}
	f, err := client.Open(full)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("sftp: read %q: %w", filePath, folder.ErrNotFound)
		}
		return nil, fmt.Errorf("sftp: read %q: %w", filePath, err)
	}
	return f, nil
}

// ---------------------------------------------------------------------------
// folder.Writer
// ---------------------------------------------------------------------------

func (d *Driver) Write(_ context.Context, filePath string, body io.Reader, opt *folder.WriteOptions) (_ *folder.FileInfo, retErr error) {
	client, err := d.getClient()
	if err != nil {
		return nil, err
	}
	defer func() { d.invalidateOnConnError(client, retErr) }()

	full, err := d.validFullPath(filePath)
	if err != nil {
		return nil, err
	}

	// Optimistic: try Create first; only MkdirAll on failure.
	f, err := client.Create(full)
	if err != nil {
		if isConnError(err) {
			return nil, fmt.Errorf("sftp: write %q: %w", filePath, err)
		}
		// Might be missing parent directory — create it and retry.
		dir := path.Dir(full)
		if mkErr := client.MkdirAll(dir); mkErr != nil {
			return nil, fmt.Errorf("sftp: write %q: mkdir parent: %w", filePath, mkErr)
		}
		f, err = client.Create(full)
		if err != nil {
			return nil, fmt.Errorf("sftp: write %q: %w", filePath, err)
		}
	}

	buf := make([]byte, copyBufSize)
	n, err := io.CopyBuffer(f, body, buf)
	if err != nil {
		_ = f.Close() // best-effort cleanup
		return nil, fmt.Errorf("sftp: write %q: %w", filePath, err)
	}

	// Close to flush buffered data — this error is significant.
	if err := f.Close(); err != nil {
		return nil, fmt.Errorf("sftp: write %q: close: %w", filePath, err)
	}

	modTime := folder.CloneTime(nil)
	if opt != nil && opt.ModTime != nil {
		modTime = folder.CloneTime(opt.ModTime)
		if err := client.Chtimes(full, *opt.ModTime, *opt.ModTime); err != nil {
			return nil, fmt.Errorf("sftp: write %q: set mod time: %w", filePath, err)
		}
	}

	return &folder.FileInfo{
		Name:         path.Base(filePath),
		Path:         filePath,
		Type:         folder.EntryTypeFile,
		Size:         n,
		LastModified: modTime,
	}, nil
}

// ---------------------------------------------------------------------------
// folder.HealthChecker
// ---------------------------------------------------------------------------

func (d *Driver) Ping(_ context.Context) error {
	return retryVoid(d, func(client *sftp.Client) error {
		if _, err := client.Getwd(); err != nil {
			return fmt.Errorf("sftp: ping: %w", err)
		}
		return nil
	})
}

// ---------------------------------------------------------------------------
// folder.Closer
// ---------------------------------------------------------------------------

func (d *Driver) Close() error {
	// Mark closed first — prevents any reconnection attempt.
	d.closed.Store(true)

	d.mu.Lock()
	defer d.mu.Unlock()

	var firstErr error
	if d.client != nil {
		if err := d.client.Close(); err != nil {
			firstErr = err
		}
		d.client = nil
	}
	if d.sshC != nil {
		if err := d.sshC.Close(); err != nil && firstErr == nil {
			firstErr = err
		}
		d.sshC = nil
	}

	if firstErr != nil {
		return fmt.Errorf("sftp: close: %w", firstErr)
	}
	return nil
}

// ---------------------------------------------------------------------------
// Internal helpers
// ---------------------------------------------------------------------------

// invalidateOnConnError calls invalidateClient when err is a connection error,
// ensuring the next getClient call triggers a reconnect.
func (d *Driver) invalidateOnConnError(client *sftp.Client, err error) {
	if err != nil && isConnError(err) {
		d.invalidateClient(client)
	}
}

// listRecursive walks the directory tree and collects all entries.
// Prefix filter is only applied at the top level (depth == 0) to stay
// consistent with S3/OSS drivers.
func (d *Driver) listRecursive(client *sftp.Client, fullDir, relDir string, opt *folder.ListOptions, depth int) ([]*folder.FileInfo, error) {
	if depth >= maxRecursionDepth {
		return nil, fmt.Errorf("sftp: list %q: recursion depth limit (%d) exceeded", relDir, maxRecursionDepth)
	}

	entries, err := client.ReadDir(fullDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("sftp: list %q: %w", relDir, folder.ErrNotFound)
		}
		return nil, fmt.Errorf("sftp: list %q: %w", relDir, err)
	}

	result := []*folder.FileInfo{}
	for _, entry := range entries {
		if depth == 0 && opt.Prefix != "" && !strings.HasPrefix(entry.Name(), opt.Prefix) {
			continue
		}

		fi := d.toFileInfo(entry, relDir)
		result = append(result, fi)

		if opt.Limit > 0 && len(result) >= opt.Limit {
			return result[:opt.Limit], nil
		}

		if entry.IsDir() {
			childFull := fullDir + "/" + entry.Name()
			childRel := strings.TrimSuffix(fi.Path, "/")
			sub, err := d.listRecursive(client, childFull, childRel, opt, depth+1)
			if err != nil {
				return nil, err
			}
			result = append(result, sub...)
			if opt.Limit > 0 && len(result) >= opt.Limit {
				return result[:opt.Limit], nil
			}
		}
	}

	return result, nil
}

// toFileInfo converts an os.FileInfo from ReadDir into a folder.FileInfo.
//
// NOTE: sftp.Client.ReadDir follows symlinks internally, so ModeSymlink is
// never set here. Symlink detection is handled in Stat() via Lstat.
func (d *Driver) toFileInfo(info os.FileInfo, parentDir string) *folder.FileInfo {
	entryType := folder.EntryTypeFile
	relPath := path.Join(parentDir, info.Name())

	if info.IsDir() {
		entryType = folder.EntryTypeDirectory
		relPath += "/"
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
func (d *Driver) removeAll(client *sftp.Client, dirPath string, depth int) error {
	if depth >= maxRecursionDepth {
		return fmt.Errorf("sftp: removeAll %q: recursion depth limit (%d) exceeded", dirPath, maxRecursionDepth)
	}

	entries, err := client.ReadDir(dirPath)
	if err != nil {
		return fmt.Errorf("sftp: readdir %q: %w", dirPath, err)
	}

	for _, entry := range entries {
		child := dirPath + "/" + entry.Name()
		if entry.IsDir() {
			if err := d.removeAll(client, child, depth+1); err != nil {
				return err
			}
		} else {
			if err := client.Remove(child); err != nil {
				return fmt.Errorf("sftp: remove %q: %w", child, err)
			}
		}
	}

	if err := client.RemoveDirectory(dirPath); err != nil {
		return fmt.Errorf("sftp: rmdir %q: %w", dirPath, err)
	}
	return nil
}

// copyFile copies a single file on the remote server.
// Data flows: remote → local memory buffer → remote.
func (d *Driver) copyFile(client *sftp.Client, src, dst string) error {
	srcFile, err := client.Open(src)
	if err != nil {
		return fmt.Errorf("sftp: open %q: %w", src, err)
	}
	defer srcFile.Close()

	dstFile, err := client.Create(dst)
	if err != nil {
		return fmt.Errorf("sftp: create %q: %w", dst, err)
	}

	buf := make([]byte, copyBufSize)
	if _, err := io.CopyBuffer(dstFile, srcFile, buf); err != nil {
		_ = dstFile.Close() // best-effort cleanup
		return fmt.Errorf("sftp: copy data %q -> %q: %w", src, dst, err)
	}

	// Close destination to flush buffered data — this error is significant.
	if err := dstFile.Close(); err != nil {
		return fmt.Errorf("sftp: close dst %q: %w", dst, err)
	}

	// Preserve permissions.
	if info, err := client.Stat(src); err == nil {
		_ = client.Chmod(dst, info.Mode())
	}
	return nil
}

// copyDir recursively copies a directory on the remote server.
func (d *Driver) copyDir(client *sftp.Client, src, dst string, depth int) error {
	if depth >= maxRecursionDepth {
		return fmt.Errorf("sftp: copyDir %q: recursion depth limit (%d) exceeded", src, maxRecursionDepth)
	}

	srcInfo, err := client.Stat(src)
	if err != nil {
		return fmt.Errorf("sftp: stat %q: %w", src, err)
	}

	if err := client.MkdirAll(dst); err != nil {
		return fmt.Errorf("sftp: mkdir %q: %w", dst, err)
	}
	_ = client.Chmod(dst, srcInfo.Mode())

	entries, err := client.ReadDir(src)
	if err != nil {
		return fmt.Errorf("sftp: readdir %q: %w", src, err)
	}

	for _, entry := range entries {
		srcChild := src + "/" + entry.Name()
		dstChild := dst + "/" + entry.Name()
		if entry.IsDir() {
			if err := d.copyDir(client, srcChild, dstChild, depth+1); err != nil {
				return err
			}
		} else {
			if err := d.copyFile(client, srcChild, dstChild); err != nil {
				return err
			}
		}
	}
	return nil
}
