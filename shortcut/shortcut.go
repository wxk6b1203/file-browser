package shortcut

import (
	"context"
	"sync"

	wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"
	"go.uber.org/zap"
)

// ---------------------------------------------------------------------------
// Event name constants — keep frontend & Go in sync
// ---------------------------------------------------------------------------

const (
	// EventShortcutFired is emitted by the frontend composable when a
	// shortcut's keydown is matched. Go listens for this via Dispatcher.
	// Go can also emit this to programmatically trigger a shortcut on
	// the frontend side.
	EventShortcutFired = "shortcut:fired"
)

// ---------------------------------------------------------------------------
// HandlerFunc — callback signature for Go-side shortcut reactions
// ---------------------------------------------------------------------------

// HandlerFunc is a callback invoked when the frontend fires a shortcut.
type HandlerFunc func()

// ---------------------------------------------------------------------------
// Dispatcher — receives "shortcut:fired" events and dispatches to Go handlers
// ---------------------------------------------------------------------------

// Dispatcher listens for shortcut events emitted by the frontend and
// dispatches them to Go-side handlers. It is safe for concurrent use.
//
// Usage:
//
//	d := shortcut.NewDispatcher()
//	d.On("save", func() { /* handle save */ })
//	d.On("delete", func() { /* handle delete */ })
//	// in OnStartup:
//	d.Listen(ctx)
type Dispatcher struct {
	mu       sync.RWMutex
	handlers map[string][]HandlerFunc
	cancel   func() // returned by EventsOn, used to unsubscribe
}

// NewDispatcher creates a Dispatcher ready for registering handlers.
func NewDispatcher() *Dispatcher {
	return &Dispatcher{handlers: make(map[string][]HandlerFunc)}
}

// On registers a Go-side callback for the given shortcut ID.
// Multiple handlers per ID are allowed; they execute in registration order.
func (d *Dispatcher) On(shortcutID string, fn HandlerFunc) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.handlers[shortcutID] = append(d.handlers[shortcutID], fn)
}

// Listen starts listening for EventShortcutFired events from the frontend.
// Call this once in OnStartup after the Wails context is available.
func (d *Dispatcher) Listen(ctx context.Context) {
	d.cancel = wailsRuntime.EventsOn(ctx, EventShortcutFired, func(data ...interface{}) {
		if len(data) == 0 {
			return
		}
		id, ok := data[0].(string)
		if !ok {
			return
		}
		d.dispatch(id)
	})
}

// Stop unsubscribes from the Wails event. Safe to call multiple times.
func (d *Dispatcher) Stop() {
	if d.cancel != nil {
		d.cancel()
		d.cancel = nil
	}
}

// Emit lets Go programmatically trigger a shortcut on the frontend side.
// The frontend composable will invoke any registered JS callbacks for this ID.
func (d *Dispatcher) Emit(ctx context.Context, shortcutID string) {
	wailsRuntime.EventsEmit(ctx, EventShortcutFired, shortcutID)
}

// dispatch invokes all registered handlers for the given shortcut ID.
func (d *Dispatcher) dispatch(id string) {
	d.mu.RLock()
	fns := make([]HandlerFunc, len(d.handlers[id]))
	copy(fns, d.handlers[id])
	d.mu.RUnlock()

	for _, fn := range fns {
		func() {
			defer func() {
				if r := recover(); r != nil {
					zap.S().Errorw("shortcut handler panicked", "id", id, "panic", r)
				}
			}()
			fn()
		}()
	}
}
