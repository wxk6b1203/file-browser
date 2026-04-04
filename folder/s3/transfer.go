package s3

import (
	"context"
	"fmt"
	"os"
	"path"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"

	"github.com/wxk6b1203/file-util-manager/folder"
)

// Compile-time interface compliance check.
var _ folder.Transferer = (*Driver)(nil)

const (
	// defaultPartSize is the default multipart part size (10 MiB).
	defaultPartSize = 10 * 1024 * 1024

	// defaultConcurrency is the default number of parallel part uploads/downloads.
	defaultConcurrency = 5
)

// ---------------------------------------------------------------------------
// folder.Transferer — multipart upload via S3 Transfer Manager
// ---------------------------------------------------------------------------

func (d *Driver) Upload(ctx context.Context, req *folder.TransferRequest, progressFn folder.ProgressFunc) error {
	f, err := os.Open(req.LocalPath)
	if err != nil {
		return fmt.Errorf("s3: upload: open local file %q: %w", req.LocalPath, err)
	}
	defer f.Close()

	info, err := f.Stat()
	if err != nil {
		return fmt.Errorf("s3: upload: stat local file %q: %w", req.LocalPath, err)
	}
	total := info.Size()
	modTime := folder.CloneTime(req.SourceModTime)
	if req.PreserveModTime && modTime == nil {
		value := info.ModTime()
		modTime = &value
	}

	// Wrap the file with a progress reader.
	body := folder.NewProgressReader(f, total, progressFn)

	key := d.fullKey(req.RemotePath)
	client := d.s3Client()

	partSize := int64(defaultPartSize)
	if req.PartSize > 0 {
		partSize = req.PartSize
	}
	concurrency := defaultConcurrency
	if req.Concurrency > 0 {
		concurrency = req.Concurrency
	}

	uploader := manager.NewUploader(client, func(u *manager.Uploader) {
		u.PartSize = partSize
		u.Concurrency = concurrency
	})

	input := &s3.PutObjectInput{
		Bucket: aws.String(d.cfg.Bucket),
		Key:    aws.String(key),
		Body:   body,
	}
	if req.ContentType != "" {
		input.ContentType = aws.String(req.ContentType)
	}
	metadata := req.Metadata
	if req.PreserveModTime {
		metadata = folder.MergeMetadataWithModTime(metadata, modTime)
	}
	if len(metadata) > 0 {
		input.Metadata = metadata
	}

	_, err = uploader.Upload(ctx, input)
	if err != nil {
		return fmt.Errorf("s3: upload %q: %w", req.RemotePath, err)
	}
	return nil
}

// ---------------------------------------------------------------------------
// folder.Transferer — multipart download via S3 Transfer Manager
// ---------------------------------------------------------------------------

func (d *Driver) Download(ctx context.Context, req *folder.TransferRequest, progressFn folder.ProgressFunc) error {
	key := d.fullKey(req.RemotePath)
	client := d.s3Client()

	// Get total size for progress reporting.
	var total int64
	headOut, err := client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(d.cfg.Bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		if isNotFound(err) {
			return fmt.Errorf("s3: download %q: %w", req.RemotePath, folder.ErrNotFound)
		}
		return fmt.Errorf("s3: download %q: head object: %w", req.RemotePath, err)
	}
	total = aws.ToInt64(headOut.ContentLength)
	modTime := folder.ResolveModTime(req.SourceModTime, headOut.Metadata, headOut.LastModified)

	// Ensure parent directory exists.
	if dir := path.Dir(req.LocalPath); dir != "" && dir != "." {
		if mkErr := os.MkdirAll(dir, 0o755); mkErr != nil {
			return fmt.Errorf("s3: download: mkdir %q: %w", dir, mkErr)
		}
	}

	f, err := os.Create(req.LocalPath)
	if err != nil {
		return fmt.Errorf("s3: download: create local file %q: %w", req.LocalPath, err)
	}

	partSize := int64(defaultPartSize)
	if req.PartSize > 0 {
		partSize = req.PartSize
	}
	concurrency := defaultConcurrency
	if req.Concurrency > 0 {
		concurrency = req.Concurrency
	}

	downloader := manager.NewDownloader(client, func(dl *manager.Downloader) {
		dl.PartSize = partSize
		dl.Concurrency = concurrency
	})

	// Wrap the file with a progress WriterAt so concurrent part downloads
	// all feed into the progress callback.
	writerAt := folder.NewProgressWriterAt(f, total, progressFn)

	_, err = downloader.Download(ctx, writerAt, &s3.GetObjectInput{
		Bucket: aws.String(d.cfg.Bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		_ = f.Close()
		return fmt.Errorf("s3: download %q: %w", req.RemotePath, err)
	}
	if err := f.Close(); err != nil {
		return fmt.Errorf("s3: download: close local file %q: %w", req.LocalPath, err)
	}
	if req.PreserveModTime {
		if err := folder.ApplyLocalModTime(req.LocalPath, modTime); err != nil {
			return fmt.Errorf("s3: download: restore local mod time for %q: %w", req.LocalPath, err)
		}
	}
	return nil
}
