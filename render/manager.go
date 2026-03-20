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

// Startup is called when the app starts
func (m *Manager) Startup(ctx context.Context) {
	m.ctx = ctx
	zap.S().Info("Render manager started")
}

// Shutdown is called when the app shuts down
func (m *Manager) Shutdown() {
	zap.S().Info("Render manager stopped")
}
