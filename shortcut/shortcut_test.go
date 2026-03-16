package shortcut

import (
	"sync"
	"sync/atomic"
	"testing"
)

func TestOnAndDispatch(t *testing.T) {
	d := NewDispatcher()
	var called atomic.Int32

	d.On("save", func() { called.Add(1) })
	d.On("save", func() { called.Add(1) })

	d.dispatch("save")
	if got := called.Load(); got != 2 {
		t.Errorf("expected 2 calls, got %d", got)
	}
}

func TestDispatchUnregistered(t *testing.T) {
	d := NewDispatcher()
	// Should not panic.
	d.dispatch("non-existent")
}

func TestDispatchPanicRecovery(t *testing.T) {
	d := NewDispatcher()
	var secondCalled atomic.Bool

	d.On("boom", func() { panic("oops") })
	d.On("boom", func() { secondCalled.Store(true) })

	// Should not panic; second handler should still run.
	d.dispatch("boom")
	if !secondCalled.Load() {
		t.Error("second handler should have been called despite first panicking")
	}
}

func TestConcurrentOn(t *testing.T) {
	d := NewDispatcher()
	var wg sync.WaitGroup

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			d.On("save", func() {})
		}()
	}
	wg.Wait()

	d.mu.RLock()
	n := len(d.handlers["save"])
	d.mu.RUnlock()
	if n != 100 {
		t.Errorf("expected 100 handlers, got %d", n)
	}
}

func TestStopIdempotent(t *testing.T) {
	d := NewDispatcher()
	// Stop before Listen — should be a no-op.
	d.Stop()
	d.Stop()
}
