package folder

import "errors"

var (
	ErrNotFound     = errors.New("file entry not found")
	ErrAlreadyExist = errors.New("file entry already exists")
	ErrInvalidPath  = errors.New("invalid path")
	ErrUnsupported  = errors.New("operation not supported")
	ErrReadOnly     = errors.New("file system is read-only")
)

// IsNotFound reports whether err is or wraps ErrNotFound.
func IsNotFound(err error) bool {
	return errors.Is(err, ErrNotFound)
}
