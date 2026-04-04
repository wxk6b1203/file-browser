package alibaba_oss

import (
	"context"
	"fmt"
	"io"
	"os"
	"path"

	"github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss"

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
// folder.Transferer — multipart upload via OSS Uploader
// ---------------------------------------------------------------------------

func (d *Driver) Upload(ctx context.Context, req *folder.TransferRequest, progressFn folder.ProgressFunc) error {
	client := d.ossClient()
	if client == nil {
		return fmt.Errorf("oss: upload %q: driver is closed", req.RemotePath)
	}

	f, err := os.Open(req.LocalPath)
	if err != nil {
		return fmt.Errorf("oss: upload: open local file %q: %w", req.LocalPath, err)
	}
	defer f.Close()

	info, err := f.Stat()
	if err != nil {
		return fmt.Errorf("oss: upload: stat local file %q: %w", req.LocalPath, err)
	}
	total := info.Size()
	modTime := folder.CloneTime(req.SourceModTime)
	if req.PreserveModTime && modTime == nil {
		value := info.ModTime()
		modTime = &value
	}

	key := d.fullKey(req.RemotePath)

	partSize := int64(defaultPartSize)
	if req.PartSize > 0 {
		partSize = req.PartSize
	}
	concurrency := defaultConcurrency
	if req.Concurrency > 0 {
		concurrency = req.Concurrency
	}

	uploader := client.NewUploader(func(uo *oss.UploaderOptions) {
		uo.PartSize = partSize
		uo.ParallelNum = concurrency
	})

	putReq := &oss.PutObjectRequest{
		Bucket: oss.Ptr(d.cfg.Bucket),
		Key:    oss.Ptr(key),
	}
	if req.ContentType != "" {
		putReq.ContentType = oss.Ptr(req.ContentType)
	}
	metadata := req.Metadata
	if req.PreserveModTime {
		metadata = folder.MergeMetadataWithModTime(metadata, modTime)
	}
	if len(metadata) > 0 {
		putReq.Metadata = metadata
	}

	// Wrap the file with a progress reader so multipart reads are tracked.
	body := folder.NewProgressReader(f, total, progressFn)

	_, err = uploader.UploadFrom(ctx, putReq, body)
	if err != nil {
		return fmt.Errorf("oss: upload %q: %w", req.RemotePath, err)
	}
	return nil
}

// ---------------------------------------------------------------------------
// folder.Transferer — download via OSS GetObject with progress
//
// The OSS SDK v2 Downloader.DownloadFile writes directly to a file path
// and does not expose a progress callback or WriterAt interface.
// We use GetObject (single-stream) with a progress writer for reliable
// progress tracking. For most desktop-app use cases this is sufficient.
// If parallel range download is needed, a manual implementation using
// ranged GetObject calls can be added later.
// ---------------------------------------------------------------------------

func (d *Driver) Download(ctx context.Context, req *folder.TransferRequest, progressFn folder.ProgressFunc) error {
	client := d.ossClient()
	if client == nil {
		return fmt.Errorf("oss: download %q: driver is closed", req.RemotePath)
	}

	key := d.fullKey(req.RemotePath)

	// Get total size for progress.
	var total int64
	headOut, err := client.HeadObject(ctx, &oss.HeadObjectRequest{
		Bucket: oss.Ptr(d.cfg.Bucket),
		Key:    oss.Ptr(key),
	})
	if err != nil {
		if isNotFound(err) {
			return fmt.Errorf("oss: download %q: %w", req.RemotePath, folder.ErrNotFound)
		}
		return fmt.Errorf("oss: download %q: head object: %w", req.RemotePath, err)
	}
	total = headOut.ContentLength
	modTime := folder.ResolveModTime(req.SourceModTime, headOut.Metadata, headOut.LastModified)

	// Ensure parent directory exists.
	if dir := path.Dir(req.LocalPath); dir != "" && dir != "." {
		if mkErr := os.MkdirAll(dir, 0o755); mkErr != nil {
			return fmt.Errorf("oss: download: mkdir %q: %w", dir, mkErr)
		}
	}

	out, err := client.GetObject(ctx, &oss.GetObjectRequest{
		Bucket: oss.Ptr(d.cfg.Bucket),
		Key:    oss.Ptr(key),
	})
	if err != nil {
		if isNotFound(err) {
			return fmt.Errorf("oss: download %q: %w", req.RemotePath, folder.ErrNotFound)
		}
		return fmt.Errorf("oss: download %q: %w", req.RemotePath, err)
	}
	defer out.Body.Close()

	f, err := os.Create(req.LocalPath)
	if err != nil {
		return fmt.Errorf("oss: download: create local file %q: %w", req.LocalPath, err)
	}

	pw := folder.NewProgressWriter(f, total, progressFn)
	buf := make([]byte, 256<<10) // 256 KiB

	if _, err := io.CopyBuffer(pw, out.Body, buf); err != nil {
		_ = f.Close()
		return fmt.Errorf("oss: download %q: %w", req.RemotePath, err)
	}
	if err := f.Close(); err != nil {
		return fmt.Errorf("oss: download: close local file %q: %w", req.LocalPath, err)
	}
	if req.PreserveModTime {
		if err := folder.ApplyLocalModTime(req.LocalPath, modTime); err != nil {
			return fmt.Errorf("oss: download: restore local mod time for %q: %w", req.LocalPath, err)
		}
	}
	return nil
}
