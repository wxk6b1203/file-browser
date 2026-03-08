package folder

// Capabilities lets callers adapt behavior to backend-specific constraints.
type Capabilities struct {
	CanList         bool
	CanRead         bool
	CanWrite        bool
	CanDelete       bool
	CanCopy         bool
	CanMove         bool
	CanRename       bool
	CanMkdir        bool
	CanPresign      bool
	AtomicMove      bool
	SupportsVersion bool
}

func BaseCapabilities() Capabilities {
	return Capabilities{
		CanList:   true,
		CanDelete: true,
		CanCopy:   true,
		CanMove:   true,
		CanRename: true,
		CanMkdir:  true,
	}
}
