// Package render handles frontend rendering state and drag-drop signals
package render

import (
	"go.uber.org/zap"
)

// OnDragSignal receives drag-drop signals from the frontend
// This is called when files are dragged into/out of the window or dropped
//
// Example usage from frontend:
//   await runtime.OnFileDrop(x, y, (paths) => {
//     await GoRender.OnDragSignal({ type: "drop", x, y, paths })
//   })
func (m *Manager) OnDragSignal(signal DragSignal) error {
	switch signal.Type {
	case "enter":
		zap.S().Infow("Drag enter signal received",
			"x", signal.X,
			"y", signal.Y,
		)
		// TODO: Show drag overlay mask

	case "leave":
		zap.S().Infow("Drag leave signal received",
			"x", signal.X,
			"y", signal.Y,
		)
		// TODO: Hide drag overlay mask

	case "drop":
		zap.S().Infow("Drop signal received",
			"x", signal.X,
			"y", signal.Y,
			"paths", signal.Paths,
		)
		// TODO: Process dropped files

	default:
		zap.S().Warnw("Unknown drag signal type",
			"type", signal.Type,
		)
	}

	return nil
}

// OnPanelFileDrop is called when files are dropped onto a specific SplitPane panel.
// GroupID and TabID identify which panel/tab received the drop (empty if unavailable).
// Paths carry the dropped file paths (same as a concurrent OnDragSignal "drop" event).
func (m *Manager) OnPanelFileDrop(signal PanelDropSignal) error {
	zap.S().Infow("Panel file drop received",
		"groupId", signal.GroupID,
		"tabId", signal.TabID,
		"x", signal.X,
		"y", signal.Y,
		"paths", signal.Paths,
	)
	// TODO: Route dropped files to the panel identified by GroupID/TabID
	return nil
}

// OnExternalDragEnter is called when external files enter the window
// This can be used to show a visual indicator that drop is allowed
func (m *Manager) OnExternalDragEnter(x, y float64) error {
	zap.S().Infow("External drag entered window",
		"x", x,
		"y", y,
	)
	// TODO: Notify frontend to show drag mask
	return nil
}

// OnExternalDragLeave is called when external files leave the window
// This can be used to hide the visual indicator
func (m *Manager) OnExternalDragLeave() error {
	zap.S().Info("External drag left window")
	// TODO: Notify frontend to hide drag mask
	return nil
}
