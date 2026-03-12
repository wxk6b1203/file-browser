package folder

import (
	"io"
	"sync/atomic"
)

// ---------------------------------------------------------------------------
// progressReader – wraps io.Reader with byte-counting callback
// ---------------------------------------------------------------------------

// progressReader wraps an io.Reader and invokes fn after every Read with
// the cumulative number of bytes read and the total (if known).
type progressReader struct {
	r     io.Reader
	total int64 // total expected bytes; 0 if unknown
	read  atomic.Int64
	fn    ProgressFunc
}

// NewProgressReader returns an io.Reader that reports progress via fn.
// total may be 0 when the size is unknown.
func NewProgressReader(r io.Reader, total int64, fn ProgressFunc) io.Reader {
	if fn == nil {
		return r
	}
	return &progressReader{r: r, total: total, fn: fn}
}

func (pr *progressReader) Read(p []byte) (int, error) {
	n, err := pr.r.Read(p)
	if n > 0 {
		current := pr.read.Add(int64(n))
		pr.fn(current, pr.total)
	}
	return n, err
}

// If the underlying reader supports Seek we propagate it so that the
// S3/OSS upload managers can compute content length and retry parts.
func (pr *progressReader) Seek(offset int64, whence int) (int64, error) {
	seeker, ok := pr.r.(io.Seeker)
	if !ok {
		return 0, io.ErrNoProgress
	}
	pos, err := seeker.Seek(offset, whence)
	if err == nil && offset == 0 && whence == io.SeekStart {
		// Reset counter on full rewind (retry scenario).
		pr.read.Store(0)
	}
	return pos, err
}

// ReadAt delegates to the underlying ReaderAt if available.
func (pr *progressReader) ReadAt(p []byte, off int64) (int, error) {
	ra, ok := pr.r.(io.ReaderAt)
	if !ok {
		return 0, io.ErrNoProgress
	}
	n, err := ra.ReadAt(p, off)
	if n > 0 {
		current := pr.read.Add(int64(n))
		pr.fn(current, pr.total)
	}
	return n, err
}

// ---------------------------------------------------------------------------
// progressWriter – wraps io.Writer with byte-counting callback
// ---------------------------------------------------------------------------

// progressWriter wraps an io.Writer and invokes fn after every Write.
type progressWriter struct {
	w       io.Writer
	total   int64
	written atomic.Int64
	fn      ProgressFunc
}

// NewProgressWriter returns an io.Writer that reports progress via fn.
func NewProgressWriter(w io.Writer, total int64, fn ProgressFunc) io.Writer {
	if fn == nil {
		return w
	}
	return &progressWriter{w: w, total: total, fn: fn}
}

func (pw *progressWriter) Write(p []byte) (int, error) {
	n, err := pw.w.Write(p)
	if n > 0 {
		current := pw.written.Add(int64(n))
		pw.fn(current, pw.total)
	}
	return n, err
}

// ---------------------------------------------------------------------------
// progressWriterAt – wraps io.WriterAt for concurrent downloads (S3 manager)
// ---------------------------------------------------------------------------

// progressWriterAt wraps an io.WriterAt and atomically tracks total bytes
// written across concurrent WriteAt calls. Used by the S3 download manager.
type progressWriterAt struct {
	w       io.WriterAt
	total   int64
	written atomic.Int64
	fn      ProgressFunc
}

// NewProgressWriterAt returns an io.WriterAt that reports progress via fn.
func NewProgressWriterAt(w io.WriterAt, total int64, fn ProgressFunc) io.WriterAt {
	if fn == nil {
		return w
	}
	return &progressWriterAt{w: w, total: total, fn: fn}
}

func (pw *progressWriterAt) WriteAt(p []byte, off int64) (int, error) {
	n, err := pw.w.WriteAt(p, off)
	if n > 0 {
		current := pw.written.Add(int64(n))
		pw.fn(current, pw.total)
	}
	return n, err
}
