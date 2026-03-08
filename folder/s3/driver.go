package s3

import (
	"context"
	"fmt"
	"io"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"

	"github.com/wxk6b1203/file-util-manager/folder"
)

func init() {
	folder.RegisterDriver[Options]("s3", New)
}

// Driver implements folder.Manager, folder.Reader, folder.Writer,
// folder.HealthChecker and folder.Closer for Amazon S3 (and compatible) backends.
type Driver struct {
	folder.BaseDriver
	cfg *Options

	mu     sync.Mutex // guards client
	client *s3.Client
}

// New creates a new S3 driver. AK/SK/Region/Bucket are mandatory.
func New(_ context.Context, opt *folder.DriverOptions, cfg *Options) (folder.Manager, error) {
	if cfg.Region == "" {
		return nil, fmt.Errorf("s3: region is required")
	}
	if cfg.Bucket == "" {
		return nil, fmt.Errorf("s3: bucket is required")
	}
	if cfg.AccessKeyID == "" || cfg.AccessKeySecret == "" {
		return nil, fmt.Errorf("s3: accessKeyID and accessKeySecret are required")
	}

	// Normalize prefix: ensure it ends with "/" if non-empty.
	if cfg.Prefix != "" {
		cfg.Prefix = strings.TrimRight(cfg.Prefix, "/") + "/"
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

// buildClient creates an S3 client using explicit credentials (no env fallback).
func (d *Driver) buildClient() *s3.Client {
	creds := credentials.NewStaticCredentialsProvider(
		d.cfg.AccessKeyID,
		d.cfg.AccessKeySecret,
		d.cfg.SessionToken,
	)

	opts := func(o *s3.Options) {
		o.Region = d.cfg.Region
		o.Credentials = creds
		o.UsePathStyle = d.cfg.ForcePathStyle

		if d.cfg.Endpoint != "" {
			o.BaseEndpoint = aws.String(d.cfg.Endpoint)
		}
		if d.cfg.DisableSSL {
			o.EndpointOptions.DisableHTTPS = true
		}
	}

	return s3.New(s3.Options{}, opts)
}

// s3Client returns the shared client under the mutex.
func (d *Driver) s3Client() *s3.Client {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.client
}

// fullKey prepends the configured prefix to a relative path.
func (d *Driver) fullKey(relPath string) string {
	return d.cfg.Prefix + strings.TrimPrefix(relPath, "/")
}

// relPath strips the configured prefix from an S3 key.
func (d *Driver) relPath(key string) string {
	return strings.TrimPrefix(key, d.cfg.Prefix)
}

// -----------------------------------------------------------------------
// folder.Manager
// -----------------------------------------------------------------------

func (d *Driver) Capabilities() folder.Capabilities {
	caps := folder.BaseCapabilities()
	caps.CanRead = true
	caps.CanWrite = true
	caps.CanPresign = true
	// S3 has native copy; move is copy+delete (non-atomic).
	caps.AtomicMove = false
	caps.SupportsVersion = true
	return caps
}

func (d *Driver) Exist(ctx context.Context, filePath string) (bool, error) {
	return folder.ExistViaStat(d, ctx, filePath)
}

func (d *Driver) Rename(ctx context.Context, filePath string, newName string) error {
	dir := path.Dir(filePath)
	newPath := path.Join(dir, newName)
	if err := d.Copy(ctx, folder.PathOp{SrcPath: filePath, DstPath: newPath}); err != nil {
		return fmt.Errorf("s3: rename %q -> %q: %w", filePath, newName, err)
	}
	return d.Delete(ctx, filePath)
}

func (d *Driver) List(ctx context.Context, dir string, opt *folder.ListOptions) ([]*folder.FileInfo, error) {
	if opt == nil {
		opt = &folder.ListOptions{}
	}

	prefix := d.fullKey(dir)
	if prefix != "" && !strings.HasSuffix(prefix, "/") {
		prefix += "/"
	}

	// Apply extra prefix filter if provided.
	if opt.Prefix != "" {
		prefix += opt.Prefix
	}

	input := &s3.ListObjectsV2Input{
		Bucket: aws.String(d.cfg.Bucket),
		Prefix: aws.String(prefix),
	}
	if !opt.Recursive {
		input.Delimiter = aws.String("/")
	}
	if opt.Limit > 0 {
		input.MaxKeys = aws.Int32(int32(opt.Limit))
	}

	var result []*folder.FileInfo
	client := d.s3Client()

	paginator := s3.NewListObjectsV2Paginator(client, input)
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("s3: list %q: %w", dir, err)
		}

		// Directories (common prefixes).
		for _, cp := range page.CommonPrefixes {
			name := d.relPath(aws.ToString(cp.Prefix))
			name = strings.TrimSuffix(name, "/")
			name = path.Base(name)
			result = append(result, &folder.FileInfo{
				Name: name,
				Path: d.relPath(aws.ToString(cp.Prefix)),
				Type: folder.EntryTypeDirectory,
			})
		}

		// Files.
		for _, obj := range page.Contents {
			key := aws.ToString(obj.Key)
			// Skip the directory marker itself.
			if strings.HasSuffix(key, "/") {
				continue
			}
			rel := d.relPath(key)
			var lastMod *time.Time
			if obj.LastModified != nil {
				t := *obj.LastModified
				lastMod = &t
			}

			fi := &folder.FileInfo{
				Name:         path.Base(rel),
				Path:         rel,
				Type:         folder.EntryTypeFile,
				Size:         aws.ToInt64(obj.Size),
				LastModified: lastMod,
				ETag:         strings.Trim(aws.ToString(obj.ETag), "\""),
			}
			if obj.Owner != nil {
				fi.Owner = &folder.Owner{
					ID:   aws.ToString(obj.Owner.ID),
					Name: aws.ToString(obj.Owner.DisplayName),
				}
			}
			result = append(result, fi)
		}

		if opt.Limit > 0 && len(result) >= opt.Limit {
			result = result[:opt.Limit]
			break
		}
	}

	return result, nil
}

func (d *Driver) Stat(ctx context.Context, filePath string) (*folder.FileInfo, error) {
	key := d.fullKey(filePath)
	client := d.s3Client()

	out, err := client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(d.cfg.Bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		if isNotFound(err) {
			// Could be a directory — probe with a trailing slash list.
			return d.statDir(ctx, filePath)
		}
		return nil, fmt.Errorf("s3: stat %q: %w", filePath, err)
	}

	var lastMod *time.Time
	if out.LastModified != nil {
		t := *out.LastModified
		lastMod = &t
	}

	fi := &folder.FileInfo{
		Name:         path.Base(filePath),
		Path:         filePath,
		Type:         folder.EntryTypeFile,
		Size:         aws.ToInt64(out.ContentLength),
		LastModified: lastMod,
		ContentType:  aws.ToString(out.ContentType),
		ETag:         strings.Trim(aws.ToString(out.ETag), "\""),
	}
	if out.Metadata != nil {
		fi.Metadata = out.Metadata
	}
	return fi, nil
}

// statDir checks whether the path is a virtual directory (has children).
func (d *Driver) statDir(ctx context.Context, dir string) (*folder.FileInfo, error) {
	prefix := d.fullKey(dir)
	if !strings.HasSuffix(prefix, "/") {
		prefix += "/"
	}
	client := d.s3Client()
	out, err := client.ListObjectsV2(ctx, &s3.ListObjectsV2Input{
		Bucket:  aws.String(d.cfg.Bucket),
		Prefix:  aws.String(prefix),
		MaxKeys: aws.Int32(1),
	})
	if err != nil {
		return nil, fmt.Errorf("s3: stat dir %q: %w", dir, err)
	}
	if aws.ToInt32(out.KeyCount) == 0 {
		return nil, fmt.Errorf("s3: stat %q: %w", dir, folder.ErrNotFound)
	}
	return &folder.FileInfo{
		Name: path.Base(dir),
		Path: dir,
		Type: folder.EntryTypeDirectory,
	}, nil
}

func (d *Driver) Delete(ctx context.Context, filePath string) error {
	key := d.fullKey(filePath)

	// If it looks like a directory, recursively delete all children.
	if strings.HasSuffix(filePath, "/") {
		return d.deletePrefix(ctx, key)
	}

	client := d.s3Client()
	_, err := client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(d.cfg.Bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("s3: delete %q: %w", filePath, err)
	}
	return nil
}

// deletePrefix deletes all objects under a prefix (recursive directory delete).
func (d *Driver) deletePrefix(ctx context.Context, prefix string) error {
	client := d.s3Client()
	paginator := s3.NewListObjectsV2Paginator(client, &s3.ListObjectsV2Input{
		Bucket: aws.String(d.cfg.Bucket),
		Prefix: aws.String(prefix),
	})

	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return fmt.Errorf("s3: delete prefix %q list: %w", prefix, err)
		}
		if len(page.Contents) == 0 {
			continue
		}

		objs := make([]types.ObjectIdentifier, 0, len(page.Contents))
		for _, obj := range page.Contents {
			objs = append(objs, types.ObjectIdentifier{Key: obj.Key})
		}

		_, err = client.DeleteObjects(ctx, &s3.DeleteObjectsInput{
			Bucket: aws.String(d.cfg.Bucket),
			Delete: &types.Delete{Objects: objs, Quiet: aws.Bool(true)},
		})
		if err != nil {
			return fmt.Errorf("s3: delete prefix %q batch: %w", prefix, err)
		}
	}
	return nil
}

func (d *Driver) Copy(ctx context.Context, op folder.PathOp) error {
	src := d.fullKey(op.SrcPath)
	dst := d.fullKey(op.DstPath)
	client := d.s3Client()

	_, err := client.CopyObject(ctx, &s3.CopyObjectInput{
		Bucket:     aws.String(d.cfg.Bucket),
		CopySource: aws.String(d.cfg.Bucket + "/" + src),
		Key:        aws.String(dst),
	})
	if err != nil {
		return fmt.Errorf("s3: copy %q -> %q: %w", op.SrcPath, op.DstPath, err)
	}
	return nil
}

func (d *Driver) Move(ctx context.Context, op folder.PathOp) error {
	if err := d.Copy(ctx, op); err != nil {
		return err
	}
	return d.Delete(ctx, op.SrcPath)
}

func (d *Driver) Mkdir(ctx context.Context, dir string) error {
	key := d.fullKey(dir)
	if !strings.HasSuffix(key, "/") {
		key += "/"
	}
	client := d.s3Client()

	_, err := client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:        aws.String(d.cfg.Bucket),
		Key:           aws.String(key),
		ContentLength: aws.Int64(0),
	})
	if err != nil {
		return fmt.Errorf("s3: mkdir %q: %w", dir, err)
	}
	return nil
}

// -----------------------------------------------------------------------
// folder.Reader
// -----------------------------------------------------------------------

func (d *Driver) Read(ctx context.Context, filePath string) (io.ReadCloser, error) {
	key := d.fullKey(filePath)
	client := d.s3Client()

	out, err := client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(d.cfg.Bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		if isNotFound(err) {
			return nil, fmt.Errorf("s3: read %q: %w", filePath, folder.ErrNotFound)
		}
		return nil, fmt.Errorf("s3: read %q: %w", filePath, err)
	}
	return out.Body, nil
}

// -----------------------------------------------------------------------
// folder.Writer
// -----------------------------------------------------------------------

func (d *Driver) Write(ctx context.Context, filePath string, body io.Reader, opt *folder.WriteOptions) (*folder.FileInfo, error) {
	key := d.fullKey(filePath)
	client := d.s3Client()

	input := &s3.PutObjectInput{
		Bucket: aws.String(d.cfg.Bucket),
		Key:    aws.String(key),
		Body:   body,
	}
	if opt != nil {
		if opt.ContentType != "" {
			input.ContentType = aws.String(opt.ContentType)
		}
		if len(opt.Metadata) > 0 {
			input.Metadata = opt.Metadata
		}
	}

	out, err := client.PutObject(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("s3: write %q: %w", filePath, err)
	}

	fi := &folder.FileInfo{
		Name: path.Base(filePath),
		Path: filePath,
		Type: folder.EntryTypeFile,
		ETag: strings.Trim(aws.ToString(out.ETag), "\""),
	}
	return fi, nil
}

// -----------------------------------------------------------------------
// folder.HealthChecker
// -----------------------------------------------------------------------

func (d *Driver) Ping(ctx context.Context) error {
	client := d.s3Client()
	_, err := client.HeadBucket(ctx, &s3.HeadBucketInput{
		Bucket: aws.String(d.cfg.Bucket),
	})
	if err != nil {
		return fmt.Errorf("s3: ping bucket %q: %w", d.cfg.Bucket, err)
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
	key := d.fullKey(filePath)
	client := d.s3Client()
	presigner := s3.NewPresignClient(client)

	out, err := presigner.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(d.cfg.Bucket),
		Key:    aws.String(key),
	}, s3.WithPresignExpires(d.presignExpires(opt)))
	if err != nil {
		return "", fmt.Errorf("s3: presign read %q: %w", filePath, err)
	}
	return out.URL, nil
}

func (d *Driver) PresignWrite(ctx context.Context, filePath string, opt *folder.PresignOptions) (string, error) {
	key := d.fullKey(filePath)
	client := d.s3Client()
	presigner := s3.NewPresignClient(client)

	out, err := presigner.PresignPutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(d.cfg.Bucket),
		Key:    aws.String(key),
	}, s3.WithPresignExpires(d.presignExpires(opt)))
	if err != nil {
		return "", fmt.Errorf("s3: presign write %q: %w", filePath, err)
	}
	return out.URL, nil
}
