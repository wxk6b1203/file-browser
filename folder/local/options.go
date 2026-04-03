package local

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Options holds local-filesystem-specific configuration.
type Options struct {
	// RootPath is the base directory on the local machine.
	// All operations are scoped under this path.
	// An empty RootPath means the system root (use with caution).
	RootPath string `yaml:"rootPath,omitempty" json:"rootPath,omitempty"`
}

func (o *Options) Validate() error {
	o.RootPath = strings.TrimSpace(o.RootPath)

	if o.RootPath == "" {
		return nil // allow empty — means system root
	}

	// Normalize to absolute path.
	abs, err := filepath.Abs(o.RootPath)
	if err != nil {
		return fmt.Errorf("local: invalid rootPath %q: %w", o.RootPath, err)
	}
	o.RootPath = abs

	// Verify the root exists and is a directory.
	info, err := os.Stat(o.RootPath)
	if err != nil {
		return fmt.Errorf("local: rootPath %q: %w", o.RootPath, err)
	}
	if !info.IsDir() {
		return fmt.Errorf("local: rootPath %q is not a directory", o.RootPath)
	}

	// Canonicalize the root so traversal checks compare against the same
	// filesystem view as EvalSymlinks-based child resolution.
	resolved, err := filepath.EvalSymlinks(o.RootPath)
	if err != nil {
		return fmt.Errorf("local: resolve rootPath %q: %w", o.RootPath, err)
	}
	o.RootPath = resolved

	return nil
}
