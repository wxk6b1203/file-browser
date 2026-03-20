// Package render handles frontend rendering state and drag-drop signals
package render

// NodeRect represents the bounding rectangle of a rendered tree node
// This mirrors the frontend TypeScript type from Tabs/types.ts
type NodeRect struct {
	// X position relative to the viewport
	X float64 `json:"x"`
	// Y position relative to the viewport
	Y float64 `json:"y"`
	// Width in pixels
	Width float64 `json:"width"`
	// Height in pixels
	Height float64 `json:"height"`
}

// DragSignal represents a file drag-drop event from the OS/frontend
type DragSignal struct {
	// Type of drag event: "enter", "leave", "drop"
	Type string `json:"type"`
	// X position relative to viewport
	X float64 `json:"x"`
	// Y position relative to viewport
	Y float64 `json:"y"`
	// Absolute file paths being dragged (only valid for "drop" events)
	Paths []string `json:"paths,omitempty"`
}

