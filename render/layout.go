// Package render handles frontend rendering state and drag-drop signals
package render

import (
	"fmt"

	"go.uber.org/zap"
)

// OnLayoutChange receives the complete layout state from the frontend
// This is called whenever the split panel layout changes
//
// The rects map contains the bounding rectangles of all rendered tree nodes
// keyed by their node IDs. This corresponds to getAllNodeRects() in Tabs.vue
//
// Example usage from frontend:
//
//	const rects = tabsRef.getAllNodeRects()
//	await GoRender.OnLayoutChange(Object.fromEntries(rects))
func (m *Manager) OnLayoutChange(rects NodeRectMap) error {
	zap.S().Infow("Layout change received",
		"node_count", len(rects),
	)

	// Print all node rects
	fmt.Println("=== Layout Change ===")
	fmt.Printf("Total nodes: %d\n\n", len(rects))

	for nodeID, rect := range rects {
		if rect == nil {
			fmt.Printf("Node: %s has nil rect\n", nodeID)
			continue
		}
		fmt.Printf("Node: %s\n", nodeID)
		fmt.Printf("  Position: (%.1f, %.1f)\n", rect.X, rect.Y)
		fmt.Printf("  Size: %.1f x %.1f\n", rect.Width, rect.Height)
		fmt.Println()
	}
	fmt.Println("=====================")

	return nil
}

// OnPanelResize is called when a specific panel is resized
// This provides fine-grained updates for individual panels
func (m *Manager) OnPanelResize(nodeID string, rect NodeRect) error {
	zap.S().Infow("Panel resize received",
		"node_id", nodeID,
		"x", rect.X,
		"y", rect.Y,
		"width", rect.Width,
		"height", rect.Height,
	)

	fmt.Printf("Panel resized: %s at (%.1f, %.1f) size %.1f x %.1f\n",
		nodeID, rect.X, rect.Y, rect.Width, rect.Height,
	)

	return nil
}

// OnPanelActivate is called when a tab or panel becomes active
func (m *Manager) OnPanelActivate(nodeID string, tabID string) error {
	zap.S().Infow("Panel activated",
		"node_id", nodeID,
		"tab_id", tabID,
	)
	return nil
}
