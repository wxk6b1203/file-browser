package alibaba_oss

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss"
	"github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss/credentials"

	"github.com/wxk6b1203/file-util-manager/folder"
)

func init() {
	folder.RegisterDriver[Options]("OSS",
		"Alibaba Cloud Object Storage Service (OSS) — a massive, secure, low-cost, and highly reliable "+
			"cloud storage service provided by Alibaba Cloud.",
		New,
	)
}

// Driver implements folder.Manager, folder.Reader, folder.Writer,
// folder.HealthChecker, folder.Presigner and folder.Closer for Alibaba Cloud OSS.
type Driver struct {
	folder.BaseDriver
	cfg *Options

	mu     sync.Mutex // guards client
	client *oss.Client
}

// New creates a new OSS driver. AK/SK/Region/Bucket are mandatory.
func New(_ context.Context, opt *folder.DriverOptions, cfg *Options) (folder.Manager, error) {
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("oss: invalid config: %w", err)
	}

	// Merge DriverOptions.Root into prefix when specified.
	if opt != nil && opt.Root != "" {
		root := strings.TrimRight(opt.Root, "/") + "/"
		cfg.Prefix = root + cfg.Prefix
	}

	d := &Driver{
		BaseDriver: folder.NewBaseDriver(opt),
		cfg:        cfg,
	}
	d.client = d.buildClient()
	return d, nil
}

// buildClient constructs an *oss.Client from explicit credentials.
func (d *Driver) buildClient() *oss.Client {
	cp := credentials.NewStaticCredentialsProvider(
		d.cfg.AccessKeyID,
		d.cfg.AccessKeySecret,
		d.cfg.SecurityToken,
	)

	ossConf := &oss.Config{
		Region:              oss.Ptr(d.cfg.Region),
		CredentialsProvider: cp,
	}

	if d.cfg.Endpoint != "" {
		ossConf.Endpoint = oss.Ptr(d.cfg.Endpoint)
	}
	if d.cfg.ForcePathStyle {
		ossConf.UsePathStyle = oss.Ptr(true)
	}
	if d.cfg.Region != "" {
		ossConf.Region = oss.Ptr(d.cfg.Region)
	}
	if d.cfg.UseCName {
		ossConf.UseCName = oss.Ptr(true)
	}
	if d.cfg.DisableSSL {
		ossConf.InsecureSkipVerify = oss.Ptr(true)
	}

	return oss.NewClient(ossConf)
}

// ossClient returns the shared client under the mutex.
func (d *Driver) ossClient() *oss.Client {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.client
}

// fullKey prepends the configured prefix to a relative path.
func (d *Driver) fullKey(relPath string) string {
	return d.cfg.Prefix + strings.TrimPrefix(relPath, "/")
}

// relPath strips the configured prefix from an OSS key.
func (d *Driver) relPath(key string) string {
	return strings.TrimPrefix(key, d.cfg.Prefix)
}

// isNotFound reports whether err is an OSS 404 / NoSuchKey response.
func isNotFound(err error) bool {
	var se *oss.ServiceError
	if errors.As(err, &se) {
		return se.HttpStatusCode() == http.StatusNotFound ||
			se.ErrorCode() == "NoSuchKey" ||
			se.ErrorCode() == "NoSuchBucket"
	}
	return false
}

// -----------------------------------------------------------------------
// folder.Manager
// -----------------------------------------------------------------------

func (d *Driver) Capabilities() folder.Capabilities {
	caps := folder.BaseCapabilities()
	caps.CanRead = true
	caps.CanWrite = true
	caps.CanPresign = true
	caps.CanTransfer = true // multipart upload via OSS Uploader
	caps.AtomicMove = false
	caps.SupportsVersion = true
	return caps
}

func (d *Driver) Exist(ctx context.Context, filePath string) (bool, error) {
	return folder.ExistViaStat(d, ctx, filePath)
}

func (d *Driver) List(ctx context.Context, dir string, opt *folder.ListOptions) ([]*folder.FileInfo, error) {
	if opt == nil {
		opt = &folder.ListOptions{}
	}

	prefix := d.fullKey(dir)
	if prefix != "" && !strings.HasSuffix(prefix, "/") {
		prefix += "/"
	}
	if opt.Prefix != "" {
		prefix += opt.Prefix
	}

	req := &oss.ListObjectsV2Request{
		Bucket:     oss.Ptr(d.cfg.Bucket),
		Prefix:     oss.Ptr(prefix),
		FetchOwner: true,
	}
	if !opt.Recursive {
		req.Delimiter = oss.Ptr("/")
	}
	if opt.Limit > 0 {
		req.MaxKeys = int32(opt.Limit)
	}

	client := d.ossClient()
	if client == nil {
		return nil, fmt.Errorf("oss: list %q: driver is closed", dir)
	}

	result := []*folder.FileInfo{}

	for {
		resp, err := client.ListObjectsV2(ctx, req)
		if err != nil {
			return nil, fmt.Errorf("oss: list %q: %w", dir, err)
		}

		// Virtual directories (common prefixes).
		for _, cp := range resp.CommonPrefixes {
			name := d.relPath(oss.ToString(cp.Prefix))
			name = strings.TrimSuffix(name, "/")
			name = path.Base(name)
			result = append(result, &folder.FileInfo{
				Name: name,
				Path: d.relPath(oss.ToString(cp.Prefix)),
				Type: folder.EntryTypeDirectory,
			})
		}

		// Files.
		for _, obj := range resp.Contents {
			key := oss.ToString(obj.Key)
			if strings.HasSuffix(key, "/") {
				if !opt.Recursive {
					continue
				}
				rel := d.relPath(key)
				if rel == "" {
					continue
				}
				result = append(result, &folder.FileInfo{
					Name:         path.Base(strings.TrimSuffix(rel, "/")),
					Path:         rel,
					Type:         folder.EntryTypeDirectory,
					LastModified: folder.ResolveModTime(nil, nil, obj.LastModified),
				})
				continue
			}
			rel := d.relPath(key)
			fi := &folder.FileInfo{
				Name:         path.Base(rel),
				Path:         rel,
				Type:         folder.EntryTypeFile,
				Size:         obj.Size,
				LastModified: obj.LastModified,
				ETag:         strings.Trim(oss.ToString(obj.ETag), "\""),
			}
			if obj.Owner != nil {
				fi.Owner = &folder.Owner{
					ID:   oss.ToString(obj.Owner.ID),
					Name: oss.ToString(obj.Owner.DisplayName),
				}
			}
			result = append(result, fi)
		}

		if opt.Limit > 0 && len(result) >= opt.Limit {
			result = result[:opt.Limit]
			break
		}

		if !resp.IsTruncated {
			break
		}
		req.ContinuationToken = resp.NextContinuationToken
	}

	return result, nil
}

func (d *Driver) Stat(ctx context.Context, filePath string) (*folder.FileInfo, error) {
	key := d.fullKey(filePath)
	client := d.ossClient()
	if client == nil {
		return nil, fmt.Errorf("oss: stat %q: driver is closed", filePath)
	}

	out, err := client.HeadObject(ctx, &oss.HeadObjectRequest{
		Bucket: oss.Ptr(d.cfg.Bucket),
		Key:    oss.Ptr(key),
	})
	if err != nil {
		if isNotFound(err) {
			if dirInfo, dirErr := d.statDirMarker(ctx, filePath); dirErr == nil {
				return dirInfo, nil
			}
			// Could be a virtual directory — probe with a trailing-slash listing.
			return d.statDir(ctx, filePath)
		}
		return nil, fmt.Errorf("oss: stat %q: %w", filePath, err)
	}

	lastModified := folder.ResolveModTime(nil, out.Metadata, out.LastModified)
	fi := &folder.FileInfo{
		Name:         path.Base(filePath),
		Path:         filePath,
		Type:         folder.EntryTypeFile,
		Size:         out.ContentLength,
		LastModified: lastModified,
		ContentType:  oss.ToString(out.ContentType),
		ETag:         strings.Trim(oss.ToString(out.ETag), "\""),
	}
	if len(out.Metadata) > 0 {
		fi.Metadata = out.Metadata
	}
	return fi, nil
}

func (d *Driver) statDirMarker(ctx context.Context, dir string) (*folder.FileInfo, error) {
	key := d.fullKey(dir)
	if !strings.HasSuffix(key, "/") {
		key += "/"
	}
	client := d.ossClient()
	if client == nil {
		return nil, fmt.Errorf("oss: stat %q: driver is closed", dir)
	}

	out, err := client.HeadObject(ctx, &oss.HeadObjectRequest{
		Bucket: oss.Ptr(d.cfg.Bucket),
		Key:    oss.Ptr(key),
	})
	if err != nil {
		return nil, err
	}

	fi := &folder.FileInfo{
		Name:         path.Base(strings.TrimSuffix(dir, "/")),
		Path:         strings.TrimSuffix(dir, "/"),
		Type:         folder.EntryTypeDirectory,
		LastModified: folder.ResolveModTime(nil, out.Metadata, out.LastModified),
	}
	if len(out.Metadata) > 0 {
		fi.Metadata = out.Metadata
	}
	return fi, nil
}

// statDir checks whether path is a virtual directory (has at least one child).
func (d *Driver) statDir(ctx context.Context, dir string) (*folder.FileInfo, error) {
	prefix := d.fullKey(dir)
	if !strings.HasSuffix(prefix, "/") {
		prefix += "/"
	}
	client := d.ossClient()
	resp, err := client.ListObjectsV2(ctx, &oss.ListObjectsV2Request{
		Bucket:  oss.Ptr(d.cfg.Bucket),
		Prefix:  oss.Ptr(prefix),
		MaxKeys: 1,
	})
	if err != nil {
		return nil, fmt.Errorf("oss: stat dir %q: %w", dir, err)
	}
	if len(resp.Contents) == 0 && len(resp.CommonPrefixes) == 0 {
		return nil, fmt.Errorf("oss: stat %q: %w", dir, folder.ErrNotFound)
	}
	return &folder.FileInfo{
		Name: path.Base(dir),
		Path: dir,
		Type: folder.EntryTypeDirectory,
	}, nil
}

func (d *Driver) Delete(ctx context.Context, filePath string) error {
	// If it looks like a directory, recursively delete all children.
	if strings.HasSuffix(filePath, "/") {
		return d.deletePrefix(ctx, d.fullKey(filePath))
	}

	client := d.ossClient()
	if client == nil {
		return fmt.Errorf("oss: delete %q: driver is closed", filePath)
	}

	_, err := client.DeleteObject(ctx, &oss.DeleteObjectRequest{
		Bucket: oss.Ptr(d.cfg.Bucket),
		Key:    oss.Ptr(d.fullKey(filePath)),
	})
	if err != nil {
		return fmt.Errorf("oss: delete %q: %w", filePath, err)
	}
	return nil
}

// deletePrefix deletes all objects under a prefix in batches of 1 000.
func (d *Driver) deletePrefix(ctx context.Context, prefix string) error {
	client := d.ossClient()
	if client == nil {
		return fmt.Errorf("oss: delete prefix %q: driver is closed", prefix)
	}

	var contToken *string
	for {
		resp, err := client.ListObjectsV2(ctx, &oss.ListObjectsV2Request{
			Bucket:            oss.Ptr(d.cfg.Bucket),
			Prefix:            oss.Ptr(prefix),
			MaxKeys:           1000,
			ContinuationToken: contToken,
		})
		if err != nil {
			return fmt.Errorf("oss: delete prefix %q list: %w", prefix, err)
		}
		if len(resp.Contents) == 0 {
			if !resp.IsTruncated {
				break
			}
			contToken = resp.NextContinuationToken
			continue
		}

		objs := make([]oss.DeleteObject, 0, len(resp.Contents))
		for _, obj := range resp.Contents {
			objs = append(objs, oss.DeleteObject{Key: obj.Key})
		}

		_, err = client.DeleteMultipleObjects(ctx, &oss.DeleteMultipleObjectsRequest{
			Bucket:  oss.Ptr(d.cfg.Bucket),
			Objects: objs,
			Quiet:   true,
		})
		if err != nil {
			return fmt.Errorf("oss: delete prefix %q batch: %w", prefix, err)
		}

		if !resp.IsTruncated {
			break
		}
		contToken = resp.NextContinuationToken
	}
	return nil
}

func (d *Driver) Copy(ctx context.Context, op folder.PathOp) error {
	client := d.ossClient()
	if client == nil {
		return fmt.Errorf("oss: copy %q -> %q: driver is closed", op.SrcPath, op.DstPath)
	}

	_, err := client.CopyObject(ctx, &oss.CopyObjectRequest{
		Bucket:    oss.Ptr(d.cfg.Bucket),
		Key:       oss.Ptr(d.fullKey(op.DstPath)),
		SourceKey: oss.Ptr(d.fullKey(op.SrcPath)),
	})
	if err != nil {
		return fmt.Errorf("oss: copy %q -> %q: %w", op.SrcPath, op.DstPath, err)
	}
	return nil
}

func (d *Driver) Move(ctx context.Context, op folder.PathOp) error {
	if err := d.Copy(ctx, op); err != nil {
		return err
	}
	return d.Delete(ctx, op.SrcPath)
}

func (d *Driver) Rename(ctx context.Context, filePath string, newName string) error {
	dir := path.Dir(filePath)
	newPath := path.Join(dir, newName)
	if err := d.Copy(ctx, folder.PathOp{SrcPath: filePath, DstPath: newPath}); err != nil {
		return fmt.Errorf("oss: rename %q -> %q: %w", filePath, newName, err)
	}
	return d.Delete(ctx, filePath)
}

func (d *Driver) Mkdir(ctx context.Context, dir string) error {
	key := d.fullKey(dir)
	if !strings.HasSuffix(key, "/") {
		key += "/"
	}
	client := d.ossClient()
	if client == nil {
		return fmt.Errorf("oss: mkdir %q: driver is closed", dir)
	}

	_, err := client.PutObject(ctx, &oss.PutObjectRequest{
		Bucket:        oss.Ptr(d.cfg.Bucket),
		Key:           oss.Ptr(key),
		ContentLength: oss.Ptr(int64(0)),
	})
	if err != nil {
		return fmt.Errorf("oss: mkdir %q: %w", dir, err)
	}
	return nil
}

func (d *Driver) SetDirectoryModTime(ctx context.Context, dir string, modTime time.Time) error {
	key := d.fullKey(dir)
	if !strings.HasSuffix(key, "/") {
		key += "/"
	}
	client := d.ossClient()
	if client == nil {
		return fmt.Errorf("oss: set directory mod time %q: driver is closed", dir)
	}
	metadata := folder.MergeMetadataWithModTime(nil, &modTime)
	_, err := client.PutObject(ctx, &oss.PutObjectRequest{
		Bucket:        oss.Ptr(d.cfg.Bucket),
		Key:           oss.Ptr(key),
		ContentLength: oss.Ptr(int64(0)),
		Metadata:      metadata,
	})
	if err != nil {
		return fmt.Errorf("oss: set directory mod time %q: %w", dir, err)
	}
	return nil
}

// -----------------------------------------------------------------------
// folder.Reader
// -----------------------------------------------------------------------

func (d *Driver) Read(ctx context.Context, filePath string) (io.ReadCloser, error) {
	client := d.ossClient()
	if client == nil {
		return nil, fmt.Errorf("oss: read %q: driver is closed", filePath)
	}

	out, err := client.GetObject(ctx, &oss.GetObjectRequest{
		Bucket: oss.Ptr(d.cfg.Bucket),
		Key:    oss.Ptr(d.fullKey(filePath)),
	})
	if err != nil {
		if isNotFound(err) {
			return nil, fmt.Errorf("oss: read %q: %w", filePath, folder.ErrNotFound)
		}
		return nil, fmt.Errorf("oss: read %q: %w", filePath, err)
	}
	return out.Body, nil
}

// -----------------------------------------------------------------------
// folder.Writer
// -----------------------------------------------------------------------

func (d *Driver) Write(ctx context.Context, filePath string, body io.Reader, opt *folder.WriteOptions) (*folder.FileInfo, error) {
	client := d.ossClient()
	if client == nil {
		return nil, fmt.Errorf("oss: write %q: driver is closed", filePath)
	}
	metadata := map[string]string(nil)

	req := &oss.PutObjectRequest{
		Bucket: oss.Ptr(d.cfg.Bucket),
		Key:    oss.Ptr(d.fullKey(filePath)),
		Body:   body,
	}
	if opt != nil {
		if opt.ContentType != "" {
			req.ContentType = oss.Ptr(opt.ContentType)
		}
		metadata = folder.MergeMetadataWithModTime(opt.Metadata, opt.ModTime)
		if len(metadata) > 0 {
			req.Metadata = metadata
		}
	}

	out, err := client.PutObject(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("oss: write %q: %w", filePath, err)
	}

	fi := &folder.FileInfo{
		Name: path.Base(filePath),
		Path: filePath,
		Type: folder.EntryTypeFile,
		ETag: strings.Trim(oss.ToString(out.ETag), "\""),
	}
	if len(metadata) > 0 {
		fi.Metadata = metadata
	}
	if opt != nil {
		fi.LastModified = folder.CloneTime(opt.ModTime)
	}
	return fi, nil
}

// -----------------------------------------------------------------------
// folder.HealthChecker
// -----------------------------------------------------------------------

func (d *Driver) Ping(ctx context.Context) error {
	client := d.ossClient()
	if client == nil {
		return fmt.Errorf("oss: ping bucket %q: driver is closed", d.cfg.Bucket)
	}

	_, err := client.HeadObject(ctx, &oss.HeadObjectRequest{
		Bucket: oss.Ptr(d.cfg.Bucket),
		Key:    oss.Ptr(".ping"),
	})
	// A 404 means the bucket is reachable — the placeholder key simply doesn't exist.
	if err != nil && !isNotFound(err) {
		return fmt.Errorf("oss: ping bucket %q: %w", d.cfg.Bucket, err)
	}
	return nil
}

// -----------------------------------------------------------------------
// folder.Closer
// -----------------------------------------------------------------------

func (d *Driver) Close() error {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.client = nil
	return nil
}

// -----------------------------------------------------------------------
// folder.Presigner
// -----------------------------------------------------------------------

const defaultPresignExpires = 15 * time.Minute

func (d *Driver) presignExpires(opt *folder.PresignOptions) time.Duration {
	if opt != nil && opt.Expires > 0 {
		return opt.Expires
	}
	return defaultPresignExpires
}

func (d *Driver) PresignRead(ctx context.Context, filePath string, opt *folder.PresignOptions) (string, error) {
	client := d.ossClient()
	if client == nil {
		return "", fmt.Errorf("oss: presign read %q: driver is closed", filePath)
	}

	result, err := client.Presign(ctx, &oss.GetObjectRequest{
		Bucket: oss.Ptr(d.cfg.Bucket),
		Key:    oss.Ptr(d.fullKey(filePath)),
	}, oss.PresignExpires(d.presignExpires(opt)))
	if err != nil {
		return "", fmt.Errorf("oss: presign read %q: %w", filePath, err)
	}
	return result.URL, nil
}

func (d *Driver) PresignWrite(ctx context.Context, filePath string, opt *folder.PresignOptions) (string, error) {
	client := d.ossClient()
	if client == nil {
		return "", fmt.Errorf("oss: presign write %q: driver is closed", filePath)
	}

	result, err := client.Presign(ctx, &oss.PutObjectRequest{
		Bucket: oss.Ptr(d.cfg.Bucket),
		Key:    oss.Ptr(d.fullKey(filePath)),
	}, oss.PresignExpires(d.presignExpires(opt)))
	if err != nil {
		return "", fmt.Errorf("oss: presign write %q: %w", filePath, err)
	}
	return result.URL, nil
}
