// Package render handles frontend rendering state and drag-drop signals
package render

import (
	"context"

	"go.uber.org/zap"
)

// Manager receives frontend rendering signals and drag-drop events
// Methods on this struct are bound to the frontend via Wails
type Manager struct {
	ctx context.Context
}

// NewManager creates a new render manager
func NewManager() *Manager {
	return &Manager{}
}

// Start initialises the Manager with the Wails app context.
// Defined as a package-level function (not a method) so Wails does not bind
// it to the frontend.
func Start(m *Manager, ctx context.Context) {
	m.ctx = ctx
	zap.S().Info("Render manager started")
}

// Stop cleans up the Manager on application shutdown.
// Defined as a package-level function for the same reason as Start.
func Stop(m *Manager) {
	zap.S().Info("Render manager stopped")
}
