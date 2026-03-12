package folder

import (
	"bytes"
	"io"
	"sync/atomic"
	"testing"
)

func TestProgressReader(t *testing.T) {
	data := []byte("hello, progress reader test data")
	total := int64(len(data))

	var lastTransferred, lastTotal int64
	var callCount atomic.Int32

	fn := func(transferred, tot int64) {
		lastTransferred = transferred
		lastTotal = tot
		callCount.Add(1)
	}

	pr := NewProgressReader(bytes.NewReader(data), total, fn)
	buf, err := io.ReadAll(pr)
	if err != nil {
		t.Fatalf("ReadAll error: %v", err)
	}
	if !bytes.Equal(buf, data) {
		t.Errorf("data mismatch: got %q, want %q", buf, data)
	}
	if lastTransferred != total {
		t.Errorf("lastTransferred = %d, want %d", lastTransferred, total)
	}
	if lastTotal != total {
		t.Errorf("lastTotal = %d, want %d", lastTotal, total)
	}
	if callCount.Load() == 0 {
		t.Error("progress callback was never called")
	}
}

func TestProgressReader_NilFunc(t *testing.T) {
	data := []byte("no callback")
	pr := NewProgressReader(bytes.NewReader(data), int64(len(data)), nil)
	buf, err := io.ReadAll(pr)
	if err != nil {
		t.Fatalf("ReadAll error: %v", err)
	}
	if !bytes.Equal(buf, data) {
		t.Errorf("data mismatch")
	}
}

func TestProgressWriter(t *testing.T) {
	data := []byte("hello, progress writer test data")
	total := int64(len(data))

	var lastTransferred int64
	fn := func(transferred, tot int64) {
		lastTransferred = transferred
	}

	var buf bytes.Buffer
	pw := NewProgressWriter(&buf, total, fn)
	n, err := pw.Write(data)
	if err != nil {
		t.Fatalf("Write error: %v", err)
	}
	if n != len(data) {
		t.Errorf("Write returned %d, want %d", n, len(data))
	}
	if lastTransferred != total {
		t.Errorf("lastTransferred = %d, want %d", lastTransferred, total)
	}
	if !bytes.Equal(buf.Bytes(), data) {
		t.Errorf("data mismatch")
	}
}

func TestProgressWriterAt(t *testing.T) {
	total := int64(20)
	backing := make([]byte, total)

	var lastTransferred atomic.Int64
	fn := func(transferred, tot int64) {
		lastTransferred.Store(transferred)
	}

	// Use a bytesWriterAt helper to back the WriterAt.
	wa := &bytesWriterAt{buf: backing}
	pwa := NewProgressWriterAt(wa, total, fn)

	// Write at two offsets (simulating concurrent part downloads).
	if _, err := pwa.WriteAt([]byte("hello"), 0); err != nil {
		t.Fatalf("WriteAt(0) error: %v", err)
	}
	if _, err := pwa.WriteAt([]byte("world"), 10); err != nil {
		t.Fatalf("WriteAt(10) error: %v", err)
	}

	if lastTransferred.Load() != 10 { // 5 + 5
		t.Errorf("lastTransferred = %d, want 10", lastTransferred.Load())
	}
}

// bytesWriterAt is a simple in-memory io.WriterAt for testing.
type bytesWriterAt struct {
	buf []byte
}

func (b *bytesWriterAt) WriteAt(p []byte, off int64) (int, error) {
	copy(b.buf[off:], p)
	return len(p), nil
}
